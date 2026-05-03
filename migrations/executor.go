package migrations

import (
	"context"
	"fmt"
	"log/slog"
	"sort"
	"strings"
	"time"

	"github.com/iMerica/jango/orm"
	"github.com/iMerica/jango/signals"
)

const migrationsTable = "django_migrations"

type Executor struct {
	db       *orm.DB
	editor   *PostgresSchemaEditor
	appLabel string
}

func NewExecutor(db *orm.DB, appLabel string) *Executor {
	return &Executor{
		db:       db,
		editor:   NewPostgresSchemaEditor(db),
		appLabel: appLabel,
	}
}

func (e *Executor) EnsureMigrationTable(ctx context.Context) error {
	sql := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (
		id SERIAL PRIMARY KEY,
		app VARCHAR(255) NOT NULL,
		name VARCHAR(255) NOT NULL,
		applied TIMESTAMP WITH TIME ZONE NOT NULL
	)`, quoteIdent(migrationsTable))
	_, err := e.db.Exec(ctx, sql)
	if err != nil {
		return fmt.Errorf("migrations: cannot create %s table: %w", migrationsTable, err)
	}

	sql = fmt.Sprintf(`CREATE UNIQUE INDEX IF NOT EXISTS %s ON %s (app, name)`,
		quoteIdent(migrationsTable+"_app_name_idx"), quoteIdent(migrationsTable))
	_, err = e.db.Exec(ctx, sql)
	return err
}

func (e *Executor) GetAppliedMigrations(ctx context.Context) (map[string]bool, error) {
	applied := make(map[string]bool)

	rows, err := e.db.Query(ctx, fmt.Sprintf("SELECT app, name FROM %s", quoteIdent(migrationsTable)))
	if err != nil {
		return applied, nil
	}
	defer rows.Close()

	for rows.Next() {
		var app, name string
		if err := rows.Scan(&app, &name); err != nil {
			continue
		}
		applied[app+"."+name] = true
	}

	return applied, nil
}

func (e *Executor) ApplyAll(ctx context.Context) ([]string, error) {
	if err := e.EnsureMigrationTable(ctx); err != nil {
		return nil, err
	}

	applied, err := e.GetAppliedMigrations(ctx)
	if err != nil {
		return nil, err
	}

	migrations := GetMigrationsForApp(e.appLabel)

	type namedMigration struct {
		name string
		m    *Migration
	}
	var sorted []namedMigration
	for _, m := range migrations {
		sorted = append(sorted, namedMigration{name: m.Name, m: m})
	}
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].name < sorted[j].name
	})

	var appliedNames []string
	for _, nm := range sorted {
		key := e.appLabel + "." + nm.name
		if applied[key] {
			slog.Info("migrations: already applied", "migration", key)
			continue
		}

		slog.Info("migrations: applying", "migration", key)
		if err := e.applyOne(ctx, nm.m); err != nil {
			return appliedNames, fmt.Errorf("migrations: error applying %s: %w", key, err)
		}

		if err := e.recordApplied(ctx, nm.m); err != nil {
			return appliedNames, fmt.Errorf("migrations: error recording %s: %w", key, err)
		}

		appliedNames = append(appliedNames, key)
		slog.Info("migrations: applied", "migration", key)
	}

	signals.PostMigrate.Send(e.appLabel, map[string]interface{}{
		"app_label": e.appLabel,
	})

	return appliedNames, nil
}

func (e *Executor) ApplyMigration(ctx context.Context, migration *Migration) error {
	if err := e.EnsureMigrationTable(ctx); err != nil {
		return err
	}

	return e.applyOne(ctx, migration)
}

func (e *Executor) applyOne(ctx context.Context, m *Migration) error {
	state := BuildProjectStateForAppBefore(e.appLabel, m.Name)

	for _, op := range m.Operations {
		if err := op.StateForwards(m.AppLabel, state); err != nil {
			return fmt.Errorf("migrations: state forward error for %s: %w", op.Describe(), err)
		}

		if err := op.DatabaseForwards(m.AppLabel, state, e.editor); err != nil {
			return fmt.Errorf("migrations: database forward error for %s: %w", op.Describe(), err)
		}
	}

	return nil
}

func (e *Executor) Reverse(ctx context.Context, migration *Migration) error {
	if err := e.EnsureMigrationTable(ctx); err != nil {
		return err
	}

	state := BuildProjectStateForAppBefore(e.appLabel, migration.Name)

	for i := len(migration.Operations) - 1; i >= 0; i-- {
		op := migration.Operations[i]

		if err := op.StateBackwards(migration.AppLabel, state); err != nil {
			return fmt.Errorf("migrations: state backward error for %s: %w", op.Describe(), err)
		}

		if err := op.DatabaseBackwards(migration.AppLabel, state, e.editor); err != nil {
			return fmt.Errorf("migrations: database backward error for %s: %w", op.Describe(), err)
		}
	}

	return e.unrecordApplied(ctx, migration)
}

func (e *Executor) recordApplied(ctx context.Context, m *Migration) error {
	sql := fmt.Sprintf("INSERT INTO %s (app, name, applied) VALUES ($1, $2, $3)",
		quoteIdent(migrationsTable))
	_, err := e.db.Exec(ctx, sql, m.AppLabel, m.Name, time.Now().UTC())
	return err
}

func (e *Executor) unrecordApplied(ctx context.Context, m *Migration) error {
	sql := fmt.Sprintf("DELETE FROM %s WHERE app = $1 AND name = $2",
		quoteIdent(migrationsTable))
	_, err := e.db.Exec(ctx, sql, m.AppLabel, m.Name)
	return err
}

func (e *Executor) GetMigrationSQL(migration *Migration) ([]string, error) {
	state := BuildProjectStateForAppBefore(e.appLabel, migration.Name)
	var statements []string

	for _, op := range migration.Operations {
		sql, err := generateOperationSQL(op, migration.AppLabel, state)
		if err != nil {
			return nil, err
		}
		if sql != "" {
			statements = append(statements, sql)
		}

		if err := op.StateForwards(migration.AppLabel, state); err != nil {
			return nil, err
		}
	}

	return statements, nil
}

func generateOperationSQL(op Operation, appLabel string, state *ProjectState) (string, error) {
	sqlEditor := &SQLLoggingEditor{}

	if err := op.StateForwards(appLabel, state); err != nil {
		return "", err
	}

	_, ok := state.GetModel(appLabel, extractModelName(op))
	if !ok && needsModel(op) {
		_ = &ModelState{
			AppLabel:  appLabel,
			ModelName: extractModelName(op),
			TableName: appLabel + "_" + extractModelName(op),
		}
	}

	if err := op.DatabaseForwards(appLabel, state, sqlEditor); err != nil {
		return "", err
	}

	if len(sqlEditor.statements) == 0 {
		return "", nil
	}

	return strings.Join(sqlEditor.statements, ";\n") + ";", nil
}

func needsModel(op Operation) bool {
	switch op.(type) {
	case *RunSQL, *RunPython, *AlterModelOptions, *SeparateDatabaseAndState:
		return false
	default:
		return true
	}
}

func extractModelName(op Operation) string {
	switch o := op.(type) {
	case *CreateModel:
		return o.Name
	case *DeleteModel:
		return o.Name
	case *AddField:
		return o.ModelName
	case *RemoveField:
		return o.ModelName
	case *AlterField:
		return o.ModelName
	case *RenameField:
		return o.ModelName
	case *AddIndex:
		return o.ModelName
	case *RemoveIndex:
		return o.ModelName
	case *AddConstraint:
		return o.ModelName
	case *RemoveConstraint:
		return o.ModelName
	case *AlterModelTable:
		return o.ModelName
	case *AlterModelOptions:
		return o.ModelName
	case *RenameModel:
		return o.OldName
	default:
		return ""
	}
}

type SQLLoggingEditor struct {
	statements []string
}

func (e *SQLLoggingEditor) CreateModel(state *ModelState) error {
	return nil
}

func (e *SQLLoggingEditor) DeleteModel(state *ModelState) error {
	return nil
}

func (e *SQLLoggingEditor) AddColumn(model *ModelState, field orm.FieldDef) error {
	return nil
}

func (e *SQLLoggingEditor) RemoveColumn(model *ModelState, field orm.FieldDef) error {
	return nil
}

func (e *SQLLoggingEditor) AlterColumn(model *ModelState, oldField, newField orm.FieldDef) error {
	return nil
}

func (e *SQLLoggingEditor) RenameColumn(model *ModelState, oldName, newName string) error {
	return nil
}

func (e *SQLLoggingEditor) AddIndex(model *ModelState, index orm.IndexDef) error {
	return nil
}

func (e *SQLLoggingEditor) RemoveIndex(model *ModelState, index orm.IndexDef) error {
	return nil
}

func (e *SQLLoggingEditor) AddConstraint(model *ModelState, constraint orm.ConstraintDef) error {
	return nil
}

func (e *SQLLoggingEditor) RemoveConstraint(model *ModelState, constraint orm.ConstraintDef) error {
	return nil
}

func (e *SQLLoggingEditor) AlterModelTable(oldTable, newTable string) error {
	return nil
}

func (e *SQLLoggingEditor) ExecuteSQL(sql string) error {
	e.statements = append(e.statements, sql)
	return nil
}

func BuildProjectStateForAppBefore(appLabel, beforeMigration string) *ProjectState {
	state := NewProjectState()
	migrations := GetMigrationsForApp(appLabel)

	sorted := make([]*Migration, len(migrations))
	copy(sorted, migrations)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Name < sorted[j].Name
	})

	for _, m := range sorted {
		if m.Name == beforeMigration {
			break
		}
		for _, op := range m.Operations {
			if err := op.StateForwards(appLabel, state); err != nil {
				slog.Warn("migrations: applying state forward", "migration", m.Name, "error", err)
			}
		}
	}

	return state
}
