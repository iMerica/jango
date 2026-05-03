package migrations

import (
	"fmt"

	"github.com/iMerica/jango/orm"
)

type ModelState struct {
	AppLabel    string
	ModelName   string
	TableName   string
	PKField     string
	Fields      []orm.FieldDef
	Indexes     []orm.IndexDef
	Constraints []orm.ConstraintDef
	Options     orm.ModelOptions
	Managers    map[string]*orm.ManagerDef
}

func ModelStateFromMeta(meta *orm.ModelMeta) *ModelState {
	fields := make([]orm.FieldDef, len(meta.Fields))
	copy(fields, meta.Fields)

	indexes := make([]orm.IndexDef, len(meta.Indexes))
	copy(indexes, meta.Indexes)

	constraints := make([]orm.ConstraintDef, len(meta.Constraints))
	copy(constraints, meta.Constraints)

	ms := &ModelState{
		AppLabel:    meta.AppLabel,
		ModelName:   meta.ModelName,
		TableName:   meta.TableName,
		PKField:     meta.PKField,
		Fields:      fields,
		Indexes:     indexes,
		Constraints: constraints,
		Options:     meta.Options,
		Managers:    make(map[string]*orm.ManagerDef),
	}

	for k, v := range meta.Managers {
		ms.Managers[k] = &orm.ManagerDef{Name: v.Name}
	}

	return ms
}

func (ms *ModelState) Key() string {
	return ms.AppLabel + "." + ms.ModelName
}

func (ms *ModelState) FieldByName(name string) (orm.FieldDef, bool) {
	for _, f := range ms.Fields {
		if f.Name == name {
			return f, true
		}
	}
	return orm.FieldDef{}, false
}

func (ms *ModelState) Clone() *ModelState {
	fields := make([]orm.FieldDef, len(ms.Fields))
	copy(fields, ms.Fields)

	indexes := make([]orm.IndexDef, len(ms.Indexes))
	copy(indexes, ms.Indexes)

	constraints := make([]orm.ConstraintDef, len(ms.Constraints))
	copy(constraints, ms.Constraints)

	managers := make(map[string]*orm.ManagerDef)
	for k, v := range ms.Managers {
		managers[k] = &orm.ManagerDef{Name: v.Name}
	}

	return &ModelState{
		AppLabel:    ms.AppLabel,
		ModelName:   ms.ModelName,
		TableName:   ms.TableName,
		PKField:     ms.PKField,
		Fields:      fields,
		Indexes:     indexes,
		Constraints: constraints,
		Options:     ms.Options,
		Managers:    managers,
	}
}

func (ms *ModelState) RemoveField(name string) {
	for i, f := range ms.Fields {
		if f.Name == name {
			ms.Fields = append(ms.Fields[:i], ms.Fields[i+1:]...)
			break
		}
	}
}

func (ms *ModelState) ReplaceField(name string, newField orm.FieldDef) {
	for i, f := range ms.Fields {
		if f.Name == name {
			ms.Fields[i] = newField
			return
		}
	}
}

func (ms *ModelState) AddField(newField orm.FieldDef) {
	ms.Fields = append(ms.Fields, newField)
}

func (ms *ModelState) RemoveIndex(name string) {
	for i, idx := range ms.Indexes {
		if idx.Name == name {
			ms.Indexes = append(ms.Indexes[:i], ms.Indexes[i+1:]...)
			return
		}
	}
}

func (ms *ModelState) AddIndex(idx orm.IndexDef) {
	ms.Indexes = append(ms.Indexes, idx)
}

func (ms *ModelState) RemoveConstraint(name string) {
	for i, c := range ms.Constraints {
		if c.Name == name {
			ms.Constraints = append(ms.Constraints[:i], ms.Constraints[i+1:]...)
			return
		}
	}
}

func (ms *ModelState) AddConstraint(c orm.ConstraintDef) {
	ms.Constraints = append(ms.Constraints, c)
}

type ProjectState struct {
	models map[string]*ModelState
}

func NewProjectState() *ProjectState {
	return &ProjectState{
		models: make(map[string]*ModelState),
	}
}

func ProjectStateFromRegistry() *ProjectState {
	state := NewProjectState()
	for _, meta := range orm.GlobalRegistry().AllModels() {
		ms := ModelStateFromMeta(meta)
		state.models[ms.Key()] = ms
	}
	return state
}

func (ps *ProjectState) GetModel(appLabel, modelName string) (*ModelState, bool) {
	key := appLabel + "." + modelName
	ms, ok := ps.models[key]
	return ms, ok
}

func (ps *ProjectState) AddModel(ms *ModelState) error {
	key := ms.Key()
	if _, exists := ps.models[key]; exists {
		return fmt.Errorf("migrations: model %s already exists in project state", key)
	}
	ps.models[key] = ms
	return nil
}

func (ps *ProjectState) RemoveModel(appLabel, modelName string) error {
	key := appLabel + "." + modelName
	if _, exists := ps.models[key]; !exists {
		return fmt.Errorf("migrations: model %s does not exist in project state", key)
	}
	delete(ps.models, key)
	return nil
}

func (ps *ProjectState) Models() map[string]*ModelState {
	return ps.models
}

func (ps *ProjectState) Clone() *ProjectState {
	newState := NewProjectState()
	for k, ms := range ps.models {
		newState.models[k] = ms.Clone()
	}
	return newState
}

func (ps *ProjectState) ApplyOperation(appLabel string, op Operation) error {
	return op.StateForwards(appLabel, ps)
}

func (ps *ProjectState) ReverseOperation(appLabel string, op Operation) error {
	return op.StateBackwards(appLabel, ps)
}
