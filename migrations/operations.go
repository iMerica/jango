package migrations

import (
	"fmt"

	"github.com/iMerica/jango/orm"
)

type CreateModel struct {
	Name        string
	TableName   string
	PKField     string
	Fields      []orm.FieldDef
	Indexes     []orm.IndexDef
	Constraints []orm.ConstraintDef
	Options     orm.ModelOptions
	Managers    map[string]*orm.ManagerDef
}

func (op *CreateModel) StateForwards(appLabel string, state *ProjectState) error {
	ms := &ModelState{
		AppLabel:    appLabel,
		ModelName:   op.Name,
		TableName:   tableNameFor(appLabel, op.Name, op.TableName),
		PKField:     op.PKField,
		Fields:      op.Fields,
		Indexes:     op.Indexes,
		Constraints: op.Constraints,
		Options:     op.Options,
		Managers:    op.Managers,
	}
	if ms.Managers == nil {
		ms.Managers = make(map[string]*orm.ManagerDef)
		ms.Managers["objects"] = &orm.ManagerDef{Name: "objects"}
	}
	return state.AddModel(ms)
}

func (op *CreateModel) StateBackwards(appLabel string, state *ProjectState) error {
	return state.RemoveModel(appLabel, op.Name)
}

func (op *CreateModel) DatabaseForwards(appLabel string, state *ProjectState, editor SchemaEditor) error {
	ms, ok := state.GetModel(appLabel, op.Name)
	if !ok {
		return fmt.Errorf("migrations: create model %s: model not found in state", op.Name)
	}
	return editor.CreateModel(ms)
}

func (op *CreateModel) DatabaseBackwards(appLabel string, state *ProjectState, editor SchemaEditor) error {
	ms := &ModelState{
		AppLabel:  appLabel,
		ModelName: op.Name,
		TableName: tableNameFor(appLabel, op.Name, op.TableName),
		Fields:    op.Fields,
	}
	return editor.DeleteModel(ms)
}

func (op *CreateModel) Describe() string {
	return fmt.Sprintf("Create model %s", op.Name)
}

type DeleteModel struct {
	Name      string
	TableName string
}

func (op *DeleteModel) StateForwards(appLabel string, state *ProjectState) error {
	return state.RemoveModel(appLabel, op.Name)
}

func (op *DeleteModel) StateBackwards(appLabel string, state *ProjectState) error {
	return fmt.Errorf("migrations: DeleteModel.StateBackwards requires reversible migration")
}

func (op *DeleteModel) DatabaseForwards(appLabel string, state *ProjectState, editor SchemaEditor) error {
	ms := &ModelState{
		AppLabel:  appLabel,
		ModelName: op.Name,
		TableName: tableNameFor(appLabel, op.Name, op.TableName),
	}
	return editor.DeleteModel(ms)
}

func (op *DeleteModel) DatabaseBackwards(appLabel string, state *ProjectState, editor SchemaEditor) error {
	return fmt.Errorf("migrations: DeleteModel.DatabaseBackwards requires reversible migration")
}

func (op *DeleteModel) Describe() string {
	return fmt.Sprintf("Delete model %s", op.Name)
}

type AddField struct {
	ModelName string
	FieldName string
	Field     orm.FieldDef
	Preserve  bool
}

func (op *AddField) StateForwards(appLabel string, state *ProjectState) error {
	ms, ok := state.GetModel(appLabel, op.ModelName)
	if !ok {
		return fmt.Errorf("migrations: add field %s to %s: model not found in state", op.FieldName, op.ModelName)
	}
	ms.AddField(op.Field)
	return nil
}

func (op *AddField) StateBackwards(appLabel string, state *ProjectState) error {
	ms, ok := state.GetModel(appLabel, op.ModelName)
	if !ok {
		return fmt.Errorf("migrations: remove field %s from %s: model not found in state", op.FieldName, op.ModelName)
	}
	ms.RemoveField(op.FieldName)
	return nil
}

func (op *AddField) DatabaseForwards(appLabel string, state *ProjectState, editor SchemaEditor) error {
	ms, ok := state.GetModel(appLabel, op.ModelName)
	if !ok {
		return fmt.Errorf("migrations: add field %s: model not found", op.FieldName)
	}
	return editor.AddColumn(ms, op.Field)
}

func (op *AddField) DatabaseBackwards(appLabel string, state *ProjectState, editor SchemaEditor) error {
	ms, ok := state.GetModel(appLabel, op.ModelName)
	if !ok {
		return fmt.Errorf("migrations: remove field %s: model not found", op.FieldName)
	}
	return editor.RemoveColumn(ms, op.Field)
}

func (op *AddField) Describe() string {
	return fmt.Sprintf("Add field %s to %s", op.FieldName, op.ModelName)
}

type RemoveField struct {
	ModelName string
	FieldName string
	Field     orm.FieldDef
}

func (op *RemoveField) StateForwards(appLabel string, state *ProjectState) error {
	ms, ok := state.GetModel(appLabel, op.ModelName)
	if !ok {
		return fmt.Errorf("migrations: remove field %s from %s: model not found", op.FieldName, op.ModelName)
	}
	ms.RemoveField(op.FieldName)
	return nil
}

func (op *RemoveField) StateBackwards(appLabel string, state *ProjectState) error {
	ms, ok := state.GetModel(appLabel, op.ModelName)
	if !ok {
		return fmt.Errorf("migrations: add field %s to %s: model not found", op.FieldName, op.ModelName)
	}
	ms.AddField(op.Field)
	return nil
}

func (op *RemoveField) DatabaseForwards(appLabel string, state *ProjectState, editor SchemaEditor) error {
	ms, ok := state.GetModel(appLabel, op.ModelName)
	if !ok {
		return fmt.Errorf("migrations: remove field %s: model not found", op.FieldName)
	}
	return editor.RemoveColumn(ms, op.Field)
}

func (op *RemoveField) DatabaseBackwards(appLabel string, state *ProjectState, editor SchemaEditor) error {
	ms, ok := state.GetModel(appLabel, op.ModelName)
	if !ok {
		return fmt.Errorf("migrations: add field %s: model not found", op.FieldName)
	}
	return editor.AddColumn(ms, op.Field)
}

func (op *RemoveField) Describe() string {
	return fmt.Sprintf("Remove field %s from %s", op.FieldName, op.ModelName)
}

type AlterField struct {
	ModelName string
	FieldName string
	OldField  orm.FieldDef
	NewField  orm.FieldDef
}

func (op *AlterField) StateForwards(appLabel string, state *ProjectState) error {
	ms, ok := state.GetModel(appLabel, op.ModelName)
	if !ok {
		return fmt.Errorf("migrations: alter field %s on %s: model not found", op.FieldName, op.ModelName)
	}
	ms.ReplaceField(op.FieldName, op.NewField)
	return nil
}

func (op *AlterField) StateBackwards(appLabel string, state *ProjectState) error {
	ms, ok := state.GetModel(appLabel, op.ModelName)
	if !ok {
		return fmt.Errorf("migrations: alter field %s on %s: model not found", op.FieldName, op.ModelName)
	}
	ms.ReplaceField(op.FieldName, op.OldField)
	return nil
}

func (op *AlterField) DatabaseForwards(appLabel string, state *ProjectState, editor SchemaEditor) error {
	ms, ok := state.GetModel(appLabel, op.ModelName)
	if !ok {
		return fmt.Errorf("migrations: alter field %s: model not found", op.FieldName)
	}
	return editor.AlterColumn(ms, op.OldField, op.NewField)
}

func (op *AlterField) DatabaseBackwards(appLabel string, state *ProjectState, editor SchemaEditor) error {
	ms, ok := state.GetModel(appLabel, op.ModelName)
	if !ok {
		return fmt.Errorf("migrations: alter field %s: model not found", op.FieldName)
	}
	return editor.AlterColumn(ms, op.NewField, op.OldField)
}

func (op *AlterField) Describe() string {
	return fmt.Sprintf("Alter field %s on %s", op.FieldName, op.ModelName)
}

type RenameField struct {
	ModelName string
	OldName   string
	NewName   string
}

func (op *RenameField) StateForwards(appLabel string, state *ProjectState) error {
	ms, ok := state.GetModel(appLabel, op.ModelName)
	if !ok {
		return fmt.Errorf("migrations: rename field %s on %s: model not found", op.OldName, op.ModelName)
	}
	field, ok := ms.FieldByName(op.OldName)
	if !ok {
		return fmt.Errorf("migrations: field %s not found on %s", op.OldName, op.ModelName)
	}
	ms.RemoveField(op.OldName)
	field.Name = op.NewName
	ms.AddField(field)
	return nil
}

func (op *RenameField) StateBackwards(appLabel string, state *ProjectState) error {
	ms, ok := state.GetModel(appLabel, op.ModelName)
	if !ok {
		return fmt.Errorf("migrations: rename field %s on %s: model not found", op.NewName, op.ModelName)
	}
	field, ok := ms.FieldByName(op.NewName)
	if !ok {
		return fmt.Errorf("migrations: field %s not found on %s", op.NewName, op.ModelName)
	}
	ms.RemoveField(op.NewName)
	field.Name = op.OldName
	ms.AddField(field)
	return nil
}

func (op *RenameField) DatabaseForwards(appLabel string, state *ProjectState, editor SchemaEditor) error {
	ms, ok := state.GetModel(appLabel, op.ModelName)
	if !ok {
		return fmt.Errorf("migrations: rename field: model not found")
	}
	oldCol := op.OldName
	newCol := op.NewName
	return editor.RenameColumn(ms, oldCol, newCol)
}

func (op *RenameField) DatabaseBackwards(appLabel string, state *ProjectState, editor SchemaEditor) error {
	ms, ok := state.GetModel(appLabel, op.ModelName)
	if !ok {
		return fmt.Errorf("migrations: rename field: model not found")
	}
	return editor.RenameColumn(ms, op.NewName, op.OldName)
}

func (op *RenameField) Describe() string {
	return fmt.Sprintf("Rename field %s to %s on %s", op.OldName, op.NewName, op.ModelName)
}

type AddIndex struct {
	ModelName string
	Index     orm.IndexDef
}

func (op *AddIndex) StateForwards(appLabel string, state *ProjectState) error {
	ms, ok := state.GetModel(appLabel, op.ModelName)
	if !ok {
		return fmt.Errorf("migrations: add index: model %s not found", op.ModelName)
	}
	ms.AddIndex(op.Index)
	return nil
}

func (op *AddIndex) StateBackwards(appLabel string, state *ProjectState) error {
	ms, ok := state.GetModel(appLabel, op.ModelName)
	if !ok {
		return fmt.Errorf("migrations: remove index: model %s not found", op.ModelName)
	}
	ms.RemoveIndex(op.Index.Name)
	return nil
}

func (op *AddIndex) DatabaseForwards(appLabel string, state *ProjectState, editor SchemaEditor) error {
	ms, ok := state.GetModel(appLabel, op.ModelName)
	if !ok {
		return fmt.Errorf("migrations: add index: model not found")
	}
	return editor.AddIndex(ms, op.Index)
}

func (op *AddIndex) DatabaseBackwards(appLabel string, state *ProjectState, editor SchemaEditor) error {
	ms, ok := state.GetModel(appLabel, op.ModelName)
	if !ok {
		return fmt.Errorf("migrations: remove index: model not found")
	}
	return editor.RemoveIndex(ms, op.Index)
}

func (op *AddIndex) Describe() string {
	return fmt.Sprintf("Add index %s on %s", op.Index.Name, op.ModelName)
}

type RemoveIndex struct {
	ModelName string
	Index     orm.IndexDef
}

func (op *RemoveIndex) StateForwards(appLabel string, state *ProjectState) error {
	ms, ok := state.GetModel(appLabel, op.ModelName)
	if !ok {
		return fmt.Errorf("migrations: remove index: model %s not found", op.ModelName)
	}
	ms.RemoveIndex(op.Index.Name)
	return nil
}

func (op *RemoveIndex) StateBackwards(appLabel string, state *ProjectState) error {
	ms, ok := state.GetModel(appLabel, op.ModelName)
	if !ok {
		return fmt.Errorf("migrations: add index: model %s not found", op.ModelName)
	}
	ms.AddIndex(op.Index)
	return nil
}

func (op *RemoveIndex) DatabaseForwards(appLabel string, state *ProjectState, editor SchemaEditor) error {
	ms, ok := state.GetModel(appLabel, op.ModelName)
	if !ok {
		return fmt.Errorf("migrations: remove index: model not found")
	}
	return editor.RemoveIndex(ms, op.Index)
}

func (op *RemoveIndex) DatabaseBackwards(appLabel string, state *ProjectState, editor SchemaEditor) error {
	ms, ok := state.GetModel(appLabel, op.ModelName)
	if !ok {
		return fmt.Errorf("migrations: add index: model not found")
	}
	return editor.AddIndex(ms, op.Index)
}

func (op *RemoveIndex) Describe() string {
	return fmt.Sprintf("Remove index %s on %s", op.Index.Name, op.ModelName)
}

type AddConstraint struct {
	ModelName  string
	Constraint orm.ConstraintDef
}

func (op *AddConstraint) StateForwards(appLabel string, state *ProjectState) error {
	ms, ok := state.GetModel(appLabel, op.ModelName)
	if !ok {
		return fmt.Errorf("migrations: add constraint: model %s not found", op.ModelName)
	}
	ms.AddConstraint(op.Constraint)
	return nil
}

func (op *AddConstraint) StateBackwards(appLabel string, state *ProjectState) error {
	ms, ok := state.GetModel(appLabel, op.ModelName)
	if !ok {
		return fmt.Errorf("migrations: remove constraint: model %s not found", op.ModelName)
	}
	ms.RemoveConstraint(op.Constraint.Name)
	return nil
}

func (op *AddConstraint) DatabaseForwards(appLabel string, state *ProjectState, editor SchemaEditor) error {
	ms, ok := state.GetModel(appLabel, op.ModelName)
	if !ok {
		return fmt.Errorf("migrations: add constraint: model not found")
	}
	return editor.AddConstraint(ms, op.Constraint)
}

func (op *AddConstraint) DatabaseBackwards(appLabel string, state *ProjectState, editor SchemaEditor) error {
	ms, ok := state.GetModel(appLabel, op.ModelName)
	if !ok {
		return fmt.Errorf("migrations: remove constraint: model not found")
	}
	return editor.RemoveConstraint(ms, op.Constraint)
}

func (op *AddConstraint) Describe() string {
	return fmt.Sprintf("Add constraint %s on %s", op.Constraint.Name, op.ModelName)
}

type RemoveConstraint struct {
	ModelName  string
	Constraint orm.ConstraintDef
}

func (op *RemoveConstraint) StateForwards(appLabel string, state *ProjectState) error {
	ms, ok := state.GetModel(appLabel, op.ModelName)
	if !ok {
		return fmt.Errorf("migrations: remove constraint: model %s not found", op.ModelName)
	}
	ms.RemoveConstraint(op.Constraint.Name)
	return nil
}

func (op *RemoveConstraint) StateBackwards(appLabel string, state *ProjectState) error {
	ms, ok := state.GetModel(appLabel, op.ModelName)
	if !ok {
		return fmt.Errorf("migrations: add constraint: model %s not found", op.ModelName)
	}
	ms.AddConstraint(op.Constraint)
	return nil
}

func (op *RemoveConstraint) DatabaseForwards(appLabel string, state *ProjectState, editor SchemaEditor) error {
	ms, ok := state.GetModel(appLabel, op.ModelName)
	if !ok {
		return fmt.Errorf("migrations: remove constraint: model not found")
	}
	return editor.RemoveConstraint(ms, op.Constraint)
}

func (op *RemoveConstraint) DatabaseBackwards(appLabel string, state *ProjectState, editor SchemaEditor) error {
	ms, ok := state.GetModel(appLabel, op.ModelName)
	if !ok {
		return fmt.Errorf("migrations: add constraint: model not found")
	}
	return editor.AddConstraint(ms, op.Constraint)
}

func (op *RemoveConstraint) Describe() string {
	return fmt.Sprintf("Remove constraint %s on %s", op.Constraint.Name, op.ModelName)
}

type AlterModelTable struct {
	ModelName string
	OldTable  string
	NewTable  string
}

func (op *AlterModelTable) StateForwards(appLabel string, state *ProjectState) error {
	ms, ok := state.GetModel(appLabel, op.ModelName)
	if !ok {
		return fmt.Errorf("migrations: alter table: model %s not found", op.ModelName)
	}
	ms.TableName = op.NewTable
	return nil
}

func (op *AlterModelTable) StateBackwards(appLabel string, state *ProjectState) error {
	ms, ok := state.GetModel(appLabel, op.ModelName)
	if !ok {
		return fmt.Errorf("migrations: alter table: model %s not found", op.ModelName)
	}
	ms.TableName = op.OldTable
	return nil
}

func (op *AlterModelTable) DatabaseForwards(appLabel string, state *ProjectState, editor SchemaEditor) error {
	return editor.AlterModelTable(op.OldTable, op.NewTable)
}

func (op *AlterModelTable) DatabaseBackwards(appLabel string, state *ProjectState, editor SchemaEditor) error {
	return editor.AlterModelTable(op.NewTable, op.OldTable)
}

func (op *AlterModelTable) Describe() string {
	return fmt.Sprintf("Rename table %s to %s", op.OldTable, op.NewTable)
}

type AlterModelOptions struct {
	ModelName  string
	OldOptions orm.ModelOptions
	NewOptions orm.ModelOptions
}

func (op *AlterModelOptions) StateForwards(appLabel string, state *ProjectState) error {
	ms, ok := state.GetModel(appLabel, op.ModelName)
	if !ok {
		return fmt.Errorf("migrations: alter options: model %s not found", op.ModelName)
	}
	ms.Options = op.NewOptions
	return nil
}

func (op *AlterModelOptions) StateBackwards(appLabel string, state *ProjectState) error {
	ms, ok := state.GetModel(appLabel, op.ModelName)
	if !ok {
		return fmt.Errorf("migrations: alter options: model %s not found", op.ModelName)
	}
	ms.Options = op.OldOptions
	return nil
}

func (op *AlterModelOptions) DatabaseForwards(appLabel string, state *ProjectState, editor SchemaEditor) error {
	return nil
}

func (op *AlterModelOptions) DatabaseBackwards(appLabel string, state *ProjectState, editor SchemaEditor) error {
	return nil
}

func (op *AlterModelOptions) Describe() string {
	return fmt.Sprintf("Alter model options for %s", op.ModelName)
}

type RenameModel struct {
	OldName  string
	NewName  string
	OldTable string
	NewTable string
}

func (op *RenameModel) StateForwards(appLabel string, state *ProjectState) error {
	oldKey := state.models[appLabel+"."+op.OldName]
	if oldKey == nil {
		return fmt.Errorf("migrations: rename model %s: not found", op.OldName)
	}
	delete(state.models, appLabel+"."+op.OldName)
	oldKey.ModelName = op.NewName
	oldKey.TableName = op.NewTable
	state.models[appLabel+"."+op.NewName] = oldKey
	return nil
}

func (op *RenameModel) StateBackwards(appLabel string, state *ProjectState) error {
	newKey := state.models[appLabel+"."+op.NewName]
	if newKey == nil {
		return fmt.Errorf("migrations: rename model %s: not found", op.NewName)
	}
	delete(state.models, appLabel+"."+op.NewName)
	newKey.ModelName = op.OldName
	newKey.TableName = op.OldTable
	state.models[appLabel+"."+op.OldName] = newKey
	return nil
}

func (op *RenameModel) DatabaseForwards(appLabel string, state *ProjectState, editor SchemaEditor) error {
	return editor.AlterModelTable(op.OldTable, op.NewTable)
}

func (op *RenameModel) DatabaseBackwards(appLabel string, state *ProjectState, editor SchemaEditor) error {
	return editor.AlterModelTable(op.NewTable, op.OldTable)
}

func (op *RenameModel) Describe() string {
	return fmt.Sprintf("Rename model %s to %s", op.OldName, op.NewName)
}

type RunPython struct {
	Code        func() error
	ReverseCode func() error
	Hint        string
}

func (op *RunPython) StateForwards(appLabel string, state *ProjectState) error {
	return nil
}

func (op *RunPython) StateBackwards(appLabel string, state *ProjectState) error {
	return nil
}

func (op *RunPython) DatabaseForwards(appLabel string, state *ProjectState, editor SchemaEditor) error {
	if op.Code == nil {
		return fmt.Errorf("migrations: RunPython: no forward function provided")
	}
	return op.Code()
}

func (op *RunPython) DatabaseBackwards(appLabel string, state *ProjectState, editor SchemaEditor) error {
	if op.ReverseCode == nil {
		return fmt.Errorf("migrations: RunPython: no reverse function provided")
	}
	return op.ReverseCode()
}

func (op *RunPython) Describe() string {
	if op.Hint != "" {
		return op.Hint
	}
	return "RunPython"
}

type RunSQL struct {
	SQL        string
	ReverseSQL string
	Hint       string
}

func (op *RunSQL) StateForwards(appLabel string, state *ProjectState) error {
	return nil
}

func (op *RunSQL) StateBackwards(appLabel string, state *ProjectState) error {
	return nil
}

func (op *RunSQL) DatabaseForwards(appLabel string, state *ProjectState, editor SchemaEditor) error {
	if op.SQL == "" {
		return nil
	}
	return editor.ExecuteSQL(op.SQL)
}

func (op *RunSQL) DatabaseBackwards(appLabel string, state *ProjectState, editor SchemaEditor) error {
	if op.ReverseSQL == "" {
		return fmt.Errorf("migrations: RunSQL: no reverse SQL provided")
	}
	return editor.ExecuteSQL(op.ReverseSQL)
}

func (op *RunSQL) Describe() string {
	if op.Hint != "" {
		return op.Hint
	}
	return "RunSQL"
}

type SeparateDatabaseAndState struct {
	StateOperations    []Operation
	DatabaseOperations []Operation
}

func (op *SeparateDatabaseAndState) StateForwards(appLabel string, state *ProjectState) error {
	for _, sop := range op.StateOperations {
		if err := sop.StateForwards(appLabel, state); err != nil {
			return err
		}
	}
	return nil
}

func (op *SeparateDatabaseAndState) StateBackwards(appLabel string, state *ProjectState) error {
	for i := len(op.StateOperations) - 1; i >= 0; i-- {
		if err := op.StateOperations[i].StateBackwards(appLabel, state); err != nil {
			return err
		}
	}
	return nil
}

func (op *SeparateDatabaseAndState) DatabaseForwards(appLabel string, state *ProjectState, editor SchemaEditor) error {
	for _, dop := range op.DatabaseOperations {
		if err := dop.DatabaseForwards(appLabel, state, editor); err != nil {
			return err
		}
	}
	return nil
}

func (op *SeparateDatabaseAndState) DatabaseBackwards(appLabel string, state *ProjectState, editor SchemaEditor) error {
	for i := len(op.DatabaseOperations) - 1; i >= 0; i-- {
		if err := op.DatabaseOperations[i].DatabaseBackwards(appLabel, state, editor); err != nil {
			return err
		}
	}
	return nil
}

func (op *SeparateDatabaseAndState) Describe() string {
	return "SeparateDatabaseAndState"
}

type AddFieldNoDefault struct {
	AddField
}

func tableNameFor(appLabel, modelName, explicitTable string) string {
	if explicitTable != "" {
		return explicitTable
	}
	return appLabel + "_" + modelName
}
