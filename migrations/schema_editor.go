package migrations

import (
	"context"
	"fmt"
	"strings"

	"github.com/iMerica/jango/orm"
)

type SchemaEditor interface {
	CreateModel(state *ModelState) error
	DeleteModel(state *ModelState) error
	AddColumn(model *ModelState, field orm.FieldDef) error
	RemoveColumn(model *ModelState, field orm.FieldDef) error
	AlterColumn(model *ModelState, oldField, newField orm.FieldDef) error
	RenameColumn(model *ModelState, oldName, newName string) error
	AddIndex(model *ModelState, index orm.IndexDef) error
	RemoveIndex(model *ModelState, index orm.IndexDef) error
	AddConstraint(model *ModelState, constraint orm.ConstraintDef) error
	RemoveConstraint(model *ModelState, constraint orm.ConstraintDef) error
	AlterModelTable(oldTable, newTable string) error
	ExecuteSQL(sql string) error
}

type PostgresSchemaEditor struct {
	db *orm.DB
	tx orm.Tx
}

func NewPostgresSchemaEditor(db *orm.DB) *PostgresSchemaEditor {
	return &PostgresSchemaEditor{db: db}
}

func (e *PostgresSchemaEditor) beginTx(ctx context.Context) error {
	tx, err := e.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("migrations: begin transaction: %w", err)
	}
	e.tx = tx
	return nil
}

func (e *PostgresSchemaEditor) commitTx() error {
	if e.tx == nil {
		return nil
	}
	err := e.tx.Commit()
	e.tx = nil
	return err
}

func (e *PostgresSchemaEditor) rollbackTx() error {
	if e.tx == nil {
		return nil
	}
	err := e.tx.Rollback()
	e.tx = nil
	return err
}

func (e *PostgresSchemaEditor) exec(ctx context.Context, sql string, args ...interface{}) error {
	if e.tx != nil {
		_, err := e.tx.Exec(ctx, sql, args...)
		return err
	}
	_, err := e.db.Exec(ctx, sql, args...)
	return err
}

func (e *PostgresSchemaEditor) CreateModel(state *ModelState) error {
	ctx := context.Background()
	var b strings.Builder
	b.WriteString("CREATE TABLE ")
	b.WriteString(quoteIdent(state.TableName))
	b.WriteString(" (")

	first := true
	for _, f := range state.Fields {
		if f.FieldType == orm.ManyToManyType {
			continue
		}
		if !first {
			b.WriteString(", ")
		}
		first = false
		b.WriteString(quoteIdent(columnName(f)))
		b.WriteString(" ")
		b.WriteString(orm.InferDBType(f))
		colConstraints := buildColumnConstraints(f)
		if colConstraints != "" {
			b.WriteString(" ")
			b.WriteString(colConstraints)
		}
	}

	for _, c := range state.Constraints {
		if !first {
			b.WriteString(", ")
		}
		first = false
		b.WriteString(buildConstraintDDL(c, state.TableName))
	}

	b.WriteString(")")

	if err := e.exec(ctx, b.String()); err != nil {
		return fmt.Errorf("migrations: create table %s: %w", state.TableName, err)
	}

	for _, idx := range state.Indexes {
		if err := e.AddIndex(state, idx); err != nil {
			return err
		}
	}

	return nil
}

func (e *PostgresSchemaEditor) DeleteModel(state *ModelState) error {
	ctx := context.Background()
	sql := fmt.Sprintf("DROP TABLE IF EXISTS %s CASCADE", quoteIdent(state.TableName))
	return e.exec(ctx, sql)
}

func (e *PostgresSchemaEditor) AddColumn(model *ModelState, field orm.FieldDef) error {
	ctx := context.Background()
	var b strings.Builder
	b.WriteString("ALTER TABLE ")
	b.WriteString(quoteIdent(model.TableName))
	b.WriteString(" ADD COLUMN ")
	b.WriteString(quoteIdent(columnName(field)))
	b.WriteString(" ")
	b.WriteString(orm.InferDBType(field))

	constraints := buildColumnConstraints(field)
	if constraints != "" {
		b.WriteString(" ")
		b.WriteString(constraints)
	}

	return e.exec(ctx, b.String())
}

func (e *PostgresSchemaEditor) RemoveColumn(model *ModelState, field orm.FieldDef) error {
	ctx := context.Background()
	sql := fmt.Sprintf("ALTER TABLE %s DROP COLUMN %s",
		quoteIdent(model.TableName),
		quoteIdent(columnName(field)))
	return e.exec(ctx, sql)
}

func (e *PostgresSchemaEditor) AlterColumn(model *ModelState, oldField, newField orm.FieldDef) error {
	ctx := context.Background()
	tableName := quoteIdent(model.TableName)
	colName := quoteIdent(columnName(newField))

	typeChange := orm.InferDBType(oldField) != orm.InferDBType(newField)
	nullChange := oldField.Nullable != newField.Nullable
	uniqueChange := oldField.Unique != newField.Unique
	defaultChange := !defaultsEqual(oldField.Default, newField.Default)
	maxLenChange := oldField.MaxLength != newField.MaxLength

	if typeChange || nullChange || maxLenChange {
		var b strings.Builder
		b.WriteString("ALTER TABLE ")
		b.WriteString(tableName)
		b.WriteString(" ALTER COLUMN ")
		b.WriteString(colName)
		b.WriteString(" TYPE ")
		b.WriteString(orm.InferDBType(newField))
		b.WriteString(" USING ")
		b.WriteString(colName)
		b.WriteString("::")
		b.WriteString(orm.InferDBType(newField))

		if err := e.exec(ctx, b.String()); err != nil {
			return err
		}
	}

	if nullChange {
		var sql string
		if newField.Nullable {
			sql = fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s DROP NOT NULL", tableName, colName)
		} else {
			sql = fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s SET NOT NULL", tableName, colName)
		}
		if err := e.exec(ctx, sql); err != nil {
			return err
		}
	}

	if defaultChange {
		var sql string
		if newField.Default == nil {
			sql = fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s DROP DEFAULT", tableName, colName)
		} else {
			sql = fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s SET DEFAULT %s",
				tableName, colName, formatValue(newField.Default))
		}
		if err := e.exec(ctx, sql); err != nil {
			return err
		}
	}

	if uniqueChange {
		if newField.Unique {
			idxName := model.TableName + "_" + columnName(newField) + "_uniq"
			sql := fmt.Sprintf("ALTER TABLE %s ADD CONSTRAINT %s UNIQUE (%s)",
				tableName, quoteIdent(idxName), colName)
			if err := e.exec(ctx, sql); err != nil {
				return err
			}
		} else {
			idxName := model.TableName + "_" + columnName(newField) + "_uniq"
			sql := fmt.Sprintf("ALTER TABLE %s DROP CONSTRAINT %s",
				tableName, quoteIdent(idxName))
			_ = e.exec(ctx, sql)
		}
	}

	return nil
}

func (e *PostgresSchemaEditor) RenameColumn(model *ModelState, oldName, newName string) error {
	ctx := context.Background()
	sql := fmt.Sprintf("ALTER TABLE %s RENAME COLUMN %s TO %s",
		quoteIdent(model.TableName),
		quoteIdent(oldName),
		quoteIdent(newName))
	return e.exec(ctx, sql)
}

func (e *PostgresSchemaEditor) AddIndex(model *ModelState, index orm.IndexDef) error {
	ctx := context.Background()
	cols := make([]string, len(index.Fields))
	for i, f := range index.Fields {
		cols[i] = quoteIdent(f)
	}

	var b strings.Builder
	b.WriteString("CREATE ")
	if index.Unique {
		b.WriteString("UNIQUE ")
	}
	b.WriteString("INDEX ")
	if index.Concurrently {
		b.WriteString("CONCURRENTLY ")
	}
	if index.Name != "" {
		b.WriteString(quoteIdent(index.Name))
		b.WriteString(" ")
	}
	b.WriteString("ON ")
	b.WriteString(quoteIdent(model.TableName))
	if len(index.Opclasses) > 0 {
		b.WriteString(" (")
		for i, col := range cols {
			if i > 0 {
				b.WriteString(", ")
			}
			b.WriteString(col)
			if i < len(index.Opclasses) {
				b.WriteString(" ")
				b.WriteString(index.Opclasses[i])
			}
		}
		b.WriteString(")")
	} else {
		b.WriteString(" (")
		b.WriteString(strings.Join(cols, ", "))
		b.WriteString(")")
	}

	if index.Condition != "" {
		b.WriteString(" WHERE ")
		b.WriteString(index.Condition)
	}

	return e.exec(ctx, b.String())
}

func (e *PostgresSchemaEditor) RemoveIndex(model *ModelState, index orm.IndexDef) error {
	ctx := context.Background()
	if index.Name != "" {
		sql := fmt.Sprintf("DROP INDEX IF EXISTS %s", quoteIdent(index.Name))
		return e.exec(ctx, sql)
	}

	cols := make([]string, len(index.Fields))
	for i, f := range index.Fields {
		cols[i] = f
	}
	constraintName := model.TableName + "_" + strings.Join(cols, "_") + "_idx"
	sql := fmt.Sprintf("DROP INDEX IF EXISTS %s", quoteIdent(constraintName))
	return e.exec(ctx, sql)
}

func (e *PostgresSchemaEditor) AddConstraint(model *ModelState, constraint orm.ConstraintDef) error {
	ctx := context.Background()
	sql := fmt.Sprintf("ALTER TABLE %s ADD %s",
		quoteIdent(model.TableName),
		buildConstraintDDL(constraint, model.TableName))
	return e.exec(ctx, sql)
}

func (e *PostgresSchemaEditor) RemoveConstraint(model *ModelState, constraint orm.ConstraintDef) error {
	ctx := context.Background()
	if constraint.Name != "" {
		sql := fmt.Sprintf("ALTER TABLE %s DROP CONSTRAINT IF EXISTS %s",
			quoteIdent(model.TableName), quoteIdent(constraint.Name))
		return e.exec(ctx, sql)
	}
	return nil
}

func (e *PostgresSchemaEditor) AlterModelTable(oldTable, newTable string) error {
	ctx := context.Background()
	sql := fmt.Sprintf("ALTER TABLE %s RENAME TO %s",
		quoteIdent(oldTable), quoteIdent(newTable))
	return e.exec(ctx, sql)
}

func (e *PostgresSchemaEditor) ExecuteSQL(sql string) error {
	ctx := context.Background()
	return e.exec(ctx, sql)
}

func columnName(f orm.FieldDef) string {
	if f.DBColumn != "" {
		return f.DBColumn
	}
	if f.FieldType == orm.ForeignKeyType || f.FieldType == orm.OneToOneType {
		return f.Name + "_id"
	}
	return f.Name
}

func buildColumnConstraints(f orm.FieldDef) string {
	var parts []string

	if f.PrimaryKey {
		parts = append(parts, "PRIMARY KEY")
	}
	if !f.Nullable && !f.PrimaryKey {
		parts = append(parts, "NOT NULL")
	}
	if f.Unique && !f.PrimaryKey {
		parts = append(parts, "UNIQUE")
	}
	if f.Default != nil && !f.Auto {
		parts = append(parts, fmt.Sprintf("DEFAULT %s", formatValue(f.Default)))
	}
	if f.DBIndex && !f.PrimaryKey && !f.Unique {
		parts = append(parts, fmt.Sprintf("/* DBINDEX: %s */", f.Name))
	}

	return strings.Join(parts, " ")
}

func buildConstraintDDL(c orm.ConstraintDef, tableName string) string {
	var b strings.Builder

	if len(c.Unique) > 0 {
		b.WriteString("CONSTRAINT ")
		if c.Name != "" {
			b.WriteString(quoteIdent(c.Name))
		} else {
			b.WriteString(quoteIdent(tableName + "_" + strings.Join(c.Unique, "_") + "_uniq"))
		}
		b.WriteString(" UNIQUE (")
		cols := make([]string, len(c.Unique))
		for i, col := range c.Unique {
			cols[i] = quoteIdent(col)
		}
		b.WriteString(strings.Join(cols, ", "))
		b.WriteString(")")
	} else if c.Check != "" {
		b.WriteString("CONSTRAINT ")
		if c.Name != "" {
			b.WriteString(quoteIdent(c.Name))
		}
		b.WriteString(" CHECK (")
		b.WriteString(c.Check)
		b.WriteString(")")
	}

	return b.String()
}

func formatValue(v interface{}) string {
	switch val := v.(type) {
	case string:
		return fmt.Sprintf("'%s'", strings.ReplaceAll(val, "'", "''"))
	case int:
		return fmt.Sprintf("%d", val)
	case int64:
		return fmt.Sprintf("%d", val)
	case float64:
		return fmt.Sprintf("%f", val)
	case bool:
		if val {
			return "TRUE"
		}
		return "FALSE"
	default:
		return fmt.Sprintf("'%v'", v)
	}
}

func defaultsEqual(a, b interface{}) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return fmt.Sprintf("%v", a) == fmt.Sprintf("%v", b)
}

func quoteIdent(name string) string {
	return `"` + strings.ReplaceAll(name, `"`, `""`) + `"`
}

func InferDBType(f orm.FieldDef) string {
	return orm.InferDBType(f)
}
