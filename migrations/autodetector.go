package migrations

import (
	"fmt"
	"reflect"

	"github.com/iMerica/jango/orm"
)

type AutoDetector struct {
	beforeState *ProjectState
	afterState  *ProjectState
	appLabel    string
}

func NewAutoDetector(appLabel string, beforeState, afterState *ProjectState) *AutoDetector {
	return &AutoDetector{
		appLabel:    appLabel,
		beforeState: beforeState,
		afterState:  afterState,
	}
}

func NewAutoDetectorFromRegistry(appLabel string) *AutoDetector {
	beforeState := BuildProjectStateFromMigrations(appLabel)
	afterState := ProjectStateFromRegistry()
	return NewAutoDetector(appLabel, beforeState, afterState)
}

func (d *AutoDetector) Detect() ([]Operation, error) {
	var operations []Operation

	addedModels, removedModels, changedModels := d.diffModels()

	for _, name := range addedModels {
		ms, ok := d.afterState.GetModel(d.appLabel, name)
		if !ok {
			continue
		}
		operations = append(operations, &CreateModel{
			Name:        ms.ModelName,
			TableName:   ms.TableName,
			PKField:     ms.PKField,
			Fields:      ms.Fields,
			Indexes:     ms.Indexes,
			Constraints: ms.Constraints,
			Options:     ms.Options,
			Managers:    ms.Managers,
		})
	}

	for _, name := range removedModels {
		ms, ok := d.beforeState.GetModel(d.appLabel, name)
		if !ok {
			continue
		}
		operations = append(operations, &DeleteModel{
			Name:      ms.ModelName,
			TableName: ms.TableName,
		})
	}

	for _, name := range changedModels {
		before, _ := d.beforeState.GetModel(d.appLabel, name)
		after, _ := d.afterState.GetModel(d.appLabel, name)

		ops := d.diffModelFields(before, after)
		ops = append(ops, d.diffModelIndexes(before, after)...)
		ops = append(ops, d.diffModelConstraints(before, after)...)
		ops = append(ops, d.diffModelOptions(before, after)...)

		operations = append(operations, ops...)
	}

	return operations, nil
}

func (d *AutoDetector) diffModels() (added, removed, changed []string) {
	beforeModels := d.beforeState.Models()
	afterModels := d.afterState.Models()

	beforeForApp := make(map[string]*ModelState)
	for _, ms := range beforeModels {
		if ms.AppLabel == d.appLabel {
			beforeForApp[ms.ModelName] = ms
		}
	}

	afterForApp := make(map[string]*ModelState)
	for _, ms := range afterModels {
		if ms.AppLabel == d.appLabel {
			afterForApp[ms.ModelName] = ms
		}
	}

	for name := range afterForApp {
		if _, exists := beforeForApp[name]; !exists {
			added = append(added, name)
		}
	}

	for name := range beforeForApp {
		if _, exists := afterForApp[name]; !exists {
			removed = append(removed, name)
		}
	}

	for name := range afterForApp {
		if before, exists := beforeForApp[name]; exists {
			after := afterForApp[name]
			if modelsDiffer(before, after) {
				changed = append(changed, name)
			}
		}
	}

	return added, removed, changed
}

func modelsDiffer(before, after *ModelState) bool {
	if before.TableName != after.TableName {
		return true
	}
	if len(before.Fields) != len(after.Fields) {
		return true
	}
	if len(before.Indexes) != len(after.Indexes) {
		return true
	}
	if len(before.Constraints) != len(after.Constraints) {
		return true
	}

	beforeFields := make(map[string]orm.FieldDef)
	for _, f := range before.Fields {
		beforeFields[f.Name] = f
	}
	afterFields := make(map[string]orm.FieldDef)
	for _, f := range after.Fields {
		afterFields[f.Name] = f
	}

	for name, bf := range beforeFields {
		af, exists := afterFields[name]
		if !exists {
			return true
		}
		if fieldDiffer(bf, af) {
			return true
		}
	}
	for name := range afterFields {
		if _, exists := beforeFields[name]; !exists {
			return true
		}
	}

	return false
}

func fieldDiffer(a, b orm.FieldDef) bool {
	if a.FieldType != b.FieldType {
		return true
	}
	if a.Nullable != b.Nullable {
		return true
	}
	if a.Unique != b.Unique {
		return true
	}
	if a.MaxLength != b.MaxLength {
		return true
	}
	if a.PrimaryKey != b.PrimaryKey {
		return true
	}
	if a.DBColumn != b.DBColumn {
		return true
	}
	if !defaultsEqual(a.Default, b.Default) {
		return true
	}
	if a.RelatedModel != b.RelatedModel {
		return true
	}
	if a.OnDelete != b.OnDelete {
		return true
	}
	if a.DBIndex != b.DBIndex {
		return true
	}
	if a.DBConstraint != b.DBConstraint {
		return true
	}
	if a.Auto != b.Auto {
		return true
	}
	return false
}

func (d *AutoDetector) diffModelFields(before, after *ModelState) []Operation {
	var operations []Operation

	beforeFields := make(map[string]orm.FieldDef)
	for _, f := range before.Fields {
		beforeFields[f.Name] = f
	}
	afterFields := make(map[string]orm.FieldDef)
	for _, f := range after.Fields {
		afterFields[f.Name] = f
	}

	for name, af := range afterFields {
		bf, exists := beforeFields[name]
		if !exists {
			operations = append(operations, &AddField{
				ModelName: after.ModelName,
				FieldName: af.Name,
				Field:     af,
			})
		} else if fieldDiffer(bf, af) {
			operations = append(operations, &AlterField{
				ModelName: after.ModelName,
				FieldName: af.Name,
				OldField:  bf,
				NewField:  af,
			})
		}
	}

	for name, bf := range beforeFields {
		if _, exists := afterFields[name]; !exists {
			operations = append(operations, &RemoveField{
				ModelName: after.ModelName,
				FieldName: bf.Name,
				Field:     bf,
			})
		}
	}

	return operations
}

func (d *AutoDetector) diffModelIndexes(before, after *ModelState) []Operation {
	var operations []Operation

	beforeIndexes := make(map[string]orm.IndexDef)
	for _, idx := range before.Indexes {
		if idx.Name != "" {
			beforeIndexes[idx.Name] = idx
		}
	}
	afterIndexes := make(map[string]orm.IndexDef)
	for _, idx := range after.Indexes {
		if idx.Name != "" {
			afterIndexes[idx.Name] = idx
		}
	}

	for name, ai := range afterIndexes {
		bi, exists := beforeIndexes[name]
		if !exists {
			operations = append(operations, &AddIndex{
				ModelName: after.ModelName,
				Index:     ai,
			})
		} else if !indexesEqual(bi, ai) {
			operations = append(operations, &RemoveIndex{
				ModelName: after.ModelName,
				Index:     bi,
			})
			operations = append(operations, &AddIndex{
				ModelName: after.ModelName,
				Index:     ai,
			})
		}
	}

	for name, bi := range beforeIndexes {
		if _, exists := afterIndexes[name]; !exists {
			operations = append(operations, &RemoveIndex{
				ModelName: after.ModelName,
				Index:     bi,
			})
		}
	}

	return operations
}

func indexesEqual(a, b orm.IndexDef) bool {
	if a.Name != b.Name {
		return false
	}
	if a.Unique != b.Unique {
		return false
	}
	if !reflect.DeepEqual(a.Fields, b.Fields) {
		return false
	}
	if a.Condition != b.Condition {
		return false
	}
	return true
}

func (d *AutoDetector) diffModelConstraints(before, after *ModelState) []Operation {
	var operations []Operation

	beforeConstraints := make(map[string]orm.ConstraintDef)
	for _, c := range before.Constraints {
		if c.Name != "" {
			beforeConstraints[c.Name] = c
		}
	}
	afterConstraints := make(map[string]orm.ConstraintDef)
	for _, c := range after.Constraints {
		if c.Name != "" {
			afterConstraints[c.Name] = c
		}
	}

	for name, ac := range afterConstraints {
		bc, exists := beforeConstraints[name]
		if !exists {
			operations = append(operations, &AddConstraint{
				ModelName:  after.ModelName,
				Constraint: ac,
			})
		} else if !constraintsEqual(bc, ac) {
			operations = append(operations, &RemoveConstraint{
				ModelName:  after.ModelName,
				Constraint: bc,
			})
			operations = append(operations, &AddConstraint{
				ModelName:  after.ModelName,
				Constraint: ac,
			})
		}
	}

	for name, bc := range beforeConstraints {
		if _, exists := afterConstraints[name]; !exists {
			operations = append(operations, &RemoveConstraint{
				ModelName:  after.ModelName,
				Constraint: bc,
			})
		}
	}

	return operations
}

func constraintsEqual(a, b orm.ConstraintDef) bool {
	if a.Name != b.Name {
		return false
	}
	if a.Check != b.Check {
		return false
	}
	if !reflect.DeepEqual(a.Unique, b.Unique) {
		return false
	}
	if a.Condition != b.Condition {
		return false
	}
	return true
}

func (d *AutoDetector) diffModelOptions(before, after *ModelState) []Operation {
	var operations []Operation

	if before.TableName != after.TableName {
		operations = append(operations, &AlterModelTable{
			ModelName: after.ModelName,
			OldTable:  before.TableName,
			NewTable:  after.TableName,
		})
	}

	beforeManaged := before.Options.Managed
	afterManaged := after.Options.Managed
	if beforeManaged != afterManaged {
		operations = append(operations, &AlterModelOptions{
			ModelName:  after.ModelName,
			OldOptions: before.Options,
			NewOptions: after.Options,
		})
	}

	return operations
}

func BuildProjectStateFromMigrations(appLabel string) *ProjectState {
	state := NewProjectState()
	migrations := GetMigrationsForApp(appLabel)
	for _, m := range migrations {
		for _, op := range m.Operations {
			if err := op.StateForwards(appLabel, state); err != nil {
				fmt.Printf("migrations: warning: applying state for %s: %v\n", m.Name, err)
			}
		}
	}
	return state
}

func BuildFullProjectState() *ProjectState {
	state := NewProjectState()
	allMigrations := GetAllMigrations()
	for appLabel, migrations := range allMigrations {
		for _, m := range migrations {
			for _, op := range m.Operations {
				if err := op.StateForwards(appLabel, state); err != nil {
					fmt.Printf("migrations: warning: applying state for %s: %v\n", m.Name, err)
				}
			}
		}
	}
	return state
}
