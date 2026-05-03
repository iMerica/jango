package migrations

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/iMerica/jango/orm"
)

func TestCreateModelStateFromMeta(t *testing.T) {
	meta := &orm.ModelMeta{
		AppLabel:  "polls",
		ModelName: "Question",
		TableName: "polls_question",
		PKField:   "ID",
		Fields: []orm.FieldDef{
			orm.BigAutoField("ID"),
			orm.CharField("QuestionText", 200),
			orm.DateTimeField("PubDate"),
		},
		Indexes: []orm.IndexDef{
			{Name: "polls_question_pub_date_idx", Fields: []string{"pub_date"}},
		},
	}

	ms := ModelStateFromMeta(meta)

	if ms.AppLabel != "polls" {
		t.Errorf("expected AppLabel polls, got %s", ms.AppLabel)
	}
	if ms.ModelName != "Question" {
		t.Errorf("expected ModelName Question, got %s", ms.ModelName)
	}
	if ms.TableName != "polls_question" {
		t.Errorf("expected TableName polls_question, got %s", ms.TableName)
	}
	if ms.PKField != "ID" {
		t.Errorf("expected PKField ID, got %s", ms.PKField)
	}
	if len(ms.Fields) != 3 {
		t.Errorf("expected 3 fields, got %d", len(ms.Fields))
	}
	if len(ms.Indexes) != 1 {
		t.Errorf("expected 1 index, got %d", len(ms.Indexes))
	}
}

func TestModelStateClone(t *testing.T) {
	ms := &ModelState{
		AppLabel:  "polls",
		ModelName: "Question",
		TableName: "polls_question",
		PKField:   "ID",
		Fields: []orm.FieldDef{
			orm.BigAutoField("ID"),
			orm.CharField("QuestionText", 200),
		},
	}

	clone := ms.Clone()

	if clone.AppLabel != ms.AppLabel {
		t.Errorf("clone AppLabel mismatch")
	}
	if clone.ModelName != ms.ModelName {
		t.Errorf("clone ModelName mismatch")
	}
	if len(clone.Fields) != len(ms.Fields) {
		t.Errorf("clone Fields length mismatch")
	}

	clone.Fields[0] = orm.BigAutoField("DifferentID")
	if ms.Fields[0].Name == "DifferentID" {
		t.Error("modifying clone affected original")
	}
}

func TestModelStateFieldOperations(t *testing.T) {
	ms := &ModelState{
		AppLabel:  "polls",
		ModelName: "Question",
		TableName: "polls_question",
		PKField:   "ID",
		Fields: []orm.FieldDef{
			orm.BigAutoField("ID"),
			orm.CharField("QuestionText", 200),
		},
	}

	field := orm.DateTimeField("PubDate")
	ms.AddField(field)

	if len(ms.Fields) != 3 {
		t.Errorf("expected 3 fields after AddField, got %d", len(ms.Fields))
	}

	_, found := ms.FieldByName("PubDate")
	if !found {
		t.Error("expected to find PubDate field")
	}

	ms.RemoveField("PubDate")
	if len(ms.Fields) != 2 {
		t.Errorf("expected 2 fields after RemoveField, got %d", len(ms.Fields))
	}

	_, found = ms.FieldByName("PubDate")
	if found {
		t.Error("expected PubDate field to be removed")
	}
}

func TestProjectStateAddRemoveModel(t *testing.T) {
	ps := NewProjectState()

	ms := &ModelState{
		AppLabel:  "polls",
		ModelName: "Question",
		TableName: "polls_question",
		PKField:   "ID",
		Fields: []orm.FieldDef{
			orm.BigAutoField("ID"),
		},
	}

	err := ps.AddModel(ms)
	if err != nil {
		t.Errorf("AddModel failed: %v", err)
	}

	retrieved, ok := ps.GetModel("polls", "Question")
	if !ok {
		t.Error("expected to find model polls.Question")
	}
	if retrieved.ModelName != "Question" {
		t.Errorf("expected Question, got %s", retrieved.ModelName)
	}

	err = ps.AddModel(ms)
	if err == nil {
		t.Error("expected error for duplicate model")
	}

	err = ps.RemoveModel("polls", "Question")
	if err != nil {
		t.Errorf("RemoveModel failed: %v", err)
	}

	_, ok = ps.GetModel("polls", "Question")
	if ok {
		t.Error("expected model to be removed")
	}
}

func TestProjectStateApplyOperation(t *testing.T) {
	ps := NewProjectState()

	op := &CreateModel{
		Name:      "Question",
		TableName: "polls_question",
		PKField:   "ID",
		Fields: []orm.FieldDef{
			orm.BigAutoField("ID"),
			orm.CharField("QuestionText", 200),
		},
	}

	err := ps.ApplyOperation("polls", op)
	if err != nil {
		t.Errorf("ApplyOperation failed: %v", err)
	}

	_, ok := ps.GetModel("polls", "Question")
	if !ok {
		t.Error("expected model polls.Question to exist after operation")
	}

	// Test removing model via DeleteModel
	delOp := &DeleteModel{
		Name:      "Question",
		TableName: "polls_question",
	}

	err = ps.ApplyOperation("polls", delOp)
	if err != nil {
		t.Errorf("ApplyOperation DeleteModel failed: %v", err)
	}

	_, ok = ps.GetModel("polls", "Question")
	if ok {
		t.Error("expected model to be removed after DeleteModel")
	}
}

func TestCreateModelOperation(t *testing.T) {
	op := &CreateModel{
		Name:      "Question",
		TableName: "polls_question",
		PKField:   "ID",
		Fields: []orm.FieldDef{
			orm.BigAutoField("ID"),
			orm.CharField("QuestionText", 200),
			orm.DateTimeField("PubDate"),
		},
	}

	desc := op.Describe()
	if desc != "Create model Question" {
		t.Errorf("unexpected description: %s", desc)
	}

	ps := NewProjectState()
	err := op.StateForwards("polls", ps)
	if err != nil {
		t.Errorf("StateForwards failed: %v", err)
	}

	ms, ok := ps.GetModel("polls", "Question")
	if !ok {
		t.Fatal("expected model to exist in project state")
	}
	if ms.ModelName != "Question" {
		t.Errorf("expected Question, got %s", ms.ModelName)
	}
	if len(ms.Fields) != 3 {
		t.Errorf("expected 3 fields, got %d", len(ms.Fields))
	}

	err = op.StateBackwards("polls", ps)
	if err != nil {
		t.Errorf("StateBackwards failed: %v", err)
	}

	_, ok = ps.GetModel("polls", "Question")
	if ok {
		t.Error("expected model to be removed after StateBackwards")
	}
}

func TestAddRemoveFieldOperation(t *testing.T) {
	ps := NewProjectState()
	createOp := &CreateModel{
		Name:      "Question",
		TableName: "polls_question",
		PKField:   "ID",
		Fields: []orm.FieldDef{
			orm.BigAutoField("ID"),
		},
	}
	createOp.StateForwards("polls", ps)

	addOp := &AddField{
		ModelName: "Question",
		FieldName: "QuestionText",
		Field:     orm.CharField("QuestionText", 200),
	}

	err := addOp.StateForwards("polls", ps)
	if err != nil {
		t.Errorf("AddField StateForwards failed: %v", err)
	}

	ms, _ := ps.GetModel("polls", "Question")
	if len(ms.Fields) != 2 {
		t.Errorf("expected 2 fields after AddField, got %d", len(ms.Fields))
	}

	_, found := ms.FieldByName("QuestionText")
	if !found {
		t.Error("expected to find QuestionText field")
	}

	removeOp := &RemoveField{
		ModelName: "Question",
		FieldName: "QuestionText",
		Field:     orm.CharField("QuestionText", 200),
	}

	err = removeOp.StateForwards("polls", ps)
	if err != nil {
		t.Errorf("RemoveField StateForwards failed: %v", err)
	}

	ms, _ = ps.GetModel("polls", "Question")
	if len(ms.Fields) != 1 {
		t.Errorf("expected 1 field after RemoveField, got %d", len(ms.Fields))
	}

	err = removeOp.StateBackwards("polls", ps)
	if err != nil {
		t.Errorf("RemoveField StateBackwards failed: %v", err)
	}

	ms, _ = ps.GetModel("polls", "Question")
	if len(ms.Fields) != 2 {
		t.Errorf("expected 2 fields after StateBackwards, got %d", len(ms.Fields))
	}
}

func TestAlterFieldOperation(t *testing.T) {
	ps := NewProjectState()
	createOp := &CreateModel{
		Name:      "Question",
		TableName: "polls_question",
		PKField:   "ID",
		Fields: []orm.FieldDef{
			orm.BigAutoField("ID"),
			orm.CharField("QuestionText", 200),
		},
	}
	createOp.StateForwards("polls", ps)

	alterOp := &AlterField{
		ModelName: "Question",
		FieldName: "QuestionText",
		OldField:  orm.CharField("QuestionText", 200),
		NewField:  orm.CharField("QuestionText", 500),
	}

	err := alterOp.StateForwards("polls", ps)
	if err != nil {
		t.Errorf("AlterField StateForwards failed: %v", err)
	}

	ms, _ := ps.GetModel("polls", "Question")
	f, found := ms.FieldByName("QuestionText")
	if !found {
		t.Fatal("expected to find QuestionText field")
	}
	if f.MaxLength != 500 {
		t.Errorf("expected MaxLength 500, got %d", f.MaxLength)
	}

	err = alterOp.StateBackwards("polls", ps)
	if err != nil {
		t.Errorf("AlterField StateBackwards failed: %v", err)
	}

	ms, _ = ps.GetModel("polls", "Question")
	f, found = ms.FieldByName("QuestionText")
	if !found {
		t.Fatal("expected to find QuestionText field after backward")
	}
	if f.MaxLength != 200 {
		t.Errorf("expected MaxLength 200 after backward, got %d", f.MaxLength)
	}
}

func TestRenameFieldOperation(t *testing.T) {
	ps := NewProjectState()
	createOp := &CreateModel{
		Name:      "Question",
		TableName: "polls_question",
		PKField:   "ID",
		Fields: []orm.FieldDef{
			orm.BigAutoField("ID"),
			orm.CharField("QuestionText", 200),
		},
	}
	createOp.StateForwards("polls", ps)

	renameOp := &RenameField{
		ModelName: "Question",
		OldName:   "QuestionText",
		NewName:   "Text",
	}

	err := renameOp.StateForwards("polls", ps)
	if err != nil {
		t.Errorf("RenameField StateForwards failed: %v", err)
	}

	ms, _ := ps.GetModel("polls", "Question")
	_, found := ms.FieldByName("Text")
	if !found {
		t.Error("expected to find 'Text' field after rename")
	}
	_, found = ms.FieldByName("QuestionText")
	if found {
		t.Error("expected 'QuestionText' field to be gone after rename")
	}

	err = renameOp.StateBackwards("polls", ps)
	if err != nil {
		t.Errorf("RenameField StateBackwards failed: %v", err)
	}

	ms, _ = ps.GetModel("polls", "Question")
	_, found = ms.FieldByName("QuestionText")
	if !found {
		t.Error("expected 'QuestionText' field to be restored after backward")
	}
}

func TestAddRemoveIndexOperation(t *testing.T) {
	ps := NewProjectState()
	createOp := &CreateModel{
		Name:      "Question",
		TableName: "polls_question",
		PKField:   "ID",
		Fields: []orm.FieldDef{
			orm.BigAutoField("ID"),
			orm.CharField("QuestionText", 200),
		},
	}
	createOp.StateForwards("polls", ps)

	idx := orm.IndexDef{
		Name:   "polls_q_text_idx",
		Fields: []string{"question_text"},
	}

	addOp := &AddIndex{ModelName: "Question", Index: idx}
	err := addOp.StateForwards("polls", ps)
	if err != nil {
		t.Errorf("AddIndex StateForwards failed: %v", err)
	}

	ms, _ := ps.GetModel("polls", "Question")
	if len(ms.Indexes) != 1 {
		t.Errorf("expected 1 index, got %d", len(ms.Indexes))
	}

	removeOp := &RemoveIndex{ModelName: "Question", Index: idx}
	err = removeOp.StateForwards("polls", ps)
	if err != nil {
		t.Errorf("RemoveIndex StateForwards failed: %v", err)
	}

	ms, _ = ps.GetModel("polls", "Question")
	if len(ms.Indexes) != 0 {
		t.Errorf("expected 0 indexes, got %d", len(ms.Indexes))
	}

	err = removeOp.StateBackwards("polls", ps)
	if err != nil {
		t.Errorf("RemoveIndex StateBackwards failed: %v", err)
	}

	ms, _ = ps.GetModel("polls", "Question")
	if len(ms.Indexes) != 1 {
		t.Errorf("expected 1 index after backward, got %d", len(ms.Indexes))
	}
}

func TestAddRemoveConstraintOperation(t *testing.T) {
	ps := NewProjectState()
	createOp := &CreateModel{
		Name:      "Question",
		TableName: "polls_question",
		PKField:   "ID",
		Fields: []orm.FieldDef{
			orm.BigAutoField("ID"),
			orm.CharField("QuestionText", 200),
		},
	}
	createOp.StateForwards("polls", ps)

	constraint := orm.ConstraintDef{
		Name:   "polls_q_text_uniq",
		Unique: []string{"question_text"},
	}

	addOp := &AddConstraint{ModelName: "Question", Constraint: constraint}
	err := addOp.StateForwards("polls", ps)
	if err != nil {
		t.Errorf("AddConstraint StateForwards failed: %v", err)
	}

	ms, _ := ps.GetModel("polls", "Question")
	if len(ms.Constraints) != 1 {
		t.Errorf("expected 1 constraint, got %d", len(ms.Constraints))
	}

	removeOp := &RemoveConstraint{ModelName: "Question", Constraint: constraint}
	err = removeOp.StateForwards("polls", ps)
	if err != nil {
		t.Errorf("RemoveConstraint StateForwards failed: %v", err)
	}

	ms, _ = ps.GetModel("polls", "Question")
	if len(ms.Constraints) != 0 {
		t.Errorf("expected 0 constraints, got %d", len(ms.Constraints))
	}
}

func TestAlterModelTableOperation(t *testing.T) {
	ps := NewProjectState()
	createOp := &CreateModel{
		Name:      "Question",
		TableName: "polls_question",
		PKField:   "ID",
		Fields:    []orm.FieldDef{orm.BigAutoField("ID")},
	}
	createOp.StateForwards("polls", ps)

	alterOp := &AlterModelTable{
		ModelName: "Question",
		OldTable:  "polls_question",
		NewTable:  "surveys_question",
	}

	err := alterOp.StateForwards("polls", ps)
	if err != nil {
		t.Errorf("AlterModelTable StateForwards failed: %v", err)
	}

	ms, _ := ps.GetModel("polls", "Question")
	if ms.TableName != "surveys_question" {
		t.Errorf("expected TableName surveys_question, got %s", ms.TableName)
	}

	err = alterOp.StateBackwards("polls", ps)
	if err != nil {
		t.Errorf("AlterModelTable StateBackwards failed: %v", err)
	}

	ms, _ = ps.GetModel("polls", "Question")
	if ms.TableName != "polls_question" {
		t.Errorf("expected TableName polls_question, got %s", ms.TableName)
	}
}

func TestRenameModelOperation(t *testing.T) {
	ps := NewProjectState()
	createOp := &CreateModel{
		Name:      "Question",
		TableName: "polls_question",
		PKField:   "ID",
		Fields:    []orm.FieldDef{orm.BigAutoField("ID")},
	}
	createOp.StateForwards("polls", ps)

	renameOp := &RenameModel{
		OldName:  "Question",
		NewName:  "Poll",
		OldTable: "polls_question",
		NewTable: "polls_poll",
	}

	err := renameOp.StateForwards("polls", ps)
	if err != nil {
		t.Errorf("RenameModel StateForwards failed: %v", err)
	}

	ms, ok := ps.GetModel("polls", "Poll")
	if !ok {
		t.Fatal("expected to find model polls.Poll")
	}
	if ms.ModelName != "Poll" {
		t.Errorf("expected ModelName Poll, got %s", ms.ModelName)
	}

	_, ok = ps.GetModel("polls", "Question")
	if ok {
		t.Error("expected polls.Question to be gone after rename")
	}

	err = renameOp.StateBackwards("polls", ps)
	if err != nil {
		t.Errorf("RenameModel StateBackwards failed: %v", err)
	}

	_, ok = ps.GetModel("polls", "Question")
	if !ok {
		t.Error("expected polls.Question to be restored after backward")
	}
}

func TestAlterModelOptionsOperation(t *testing.T) {
	ps := NewProjectState()
	createOp := &CreateModel{
		Name:      "Question",
		TableName: "polls_question",
		PKField:   "ID",
		Fields:    []orm.FieldDef{orm.BigAutoField("ID")},
		Options: orm.ModelOptions{
			VerboseName: "question",
			Managed:     true,
		},
	}
	createOp.StateForwards("polls", ps)

	alterOp := &AlterModelOptions{
		ModelName:  "Question",
		OldOptions: orm.ModelOptions{VerboseName: "question", Managed: true},
		NewOptions: orm.ModelOptions{VerboseName: "poll question", Managed: true},
	}

	err := alterOp.StateForwards("polls", ps)
	if err != nil {
		t.Errorf("AlterModelOptions StateForwards failed: %v", err)
	}

	ms, _ := ps.GetModel("polls", "Question")
	if ms.Options.VerboseName != "poll question" {
		t.Errorf("expected VerboseName 'poll question', got %s", ms.Options.VerboseName)
	}
}

func TestRunPythonOperation(t *testing.T) {
	executed := false
	op := &RunPython{
		Code: func() error {
			executed = true
			return nil
		},
		ReverseCode: func() error {
			executed = false
			return nil
		},
		Hint: "test data migration",
	}

	ps := NewProjectState()
	err := op.StateForwards("polls", ps)
	if err != nil {
		t.Errorf("RunPython StateForwards failed: %v", err)
	}

	desc := op.Describe()
	if desc != "test data migration" {
		t.Errorf("expected 'test data migration', got %s", desc)
	}

	ps2 := NewProjectState()
	editor := &SQLLoggingEditor{}

	err = op.DatabaseForwards("polls", ps2, editor)
	if err != nil {
		t.Errorf("RunPython DatabaseForwards failed: %v", err)
	}
	if !executed {
		t.Error("expected Code function to be called")
	}

	err = op.DatabaseBackwards("polls", ps2, editor)
	if err != nil {
		t.Errorf("RunPython DatabaseBackwards failed: %v", err)
	}
	if executed {
		t.Error("expected ReverseCode to reset executed flag")
	}
}

func TestRunSQLOperation(t *testing.T) {
	op := &RunSQL{
		SQL:        "INSERT INTO polls_question (question_text) VALUES ('test')",
		ReverseSQL: "DELETE FROM polls_question WHERE question_text = 'test'",
	}

	ps := NewProjectState()
	err := op.StateForwards("polls", ps)
	if err != nil {
		t.Errorf("RunSQL StateForwards failed: %v", err)
	}

	editor := &SQLLoggingEditor{}
	err = op.DatabaseForwards("polls", ps, editor)
	if err != nil {
		t.Errorf("RunSQL DatabaseForwards failed: %v", err)
	}
	if len(editor.statements) != 1 {
		t.Errorf("expected 1 SQL statement, got %d", len(editor.statements))
	}

	err = op.DatabaseBackwards("polls", ps, editor)
	if err != nil {
		t.Errorf("RunSQL DatabaseBackwards failed: %v", err)
	}
}

func TestRunSQLNoReverse(t *testing.T) {
	op := &RunSQL{
		SQL: "INSERT INTO polls_question (question_text) VALUES ('test')",
	}

	editor := &SQLLoggingEditor{}
	err := op.DatabaseBackwards("polls", nil, editor)
	if err == nil {
		t.Error("expected error for missing reverse SQL")
	}
}

func TestMigrationRegistry(t *testing.T) {
	ResetMigrations()

	m := Migration{
		AppLabel:     "polls",
		Name:         "0001_initial",
		Dependencies: []Dependency{},
		Operations:   []Operation{},
	}

	RegisterMigration(m)

	retrieved := GetMigrationsForApp("polls")
	if len(retrieved) != 1 {
		t.Errorf("expected 1 migration for polls, got %d", len(retrieved))
	}
	if retrieved[0].Name != "0001_initial" {
		t.Errorf("expected 0001_initial, got %s", retrieved[0].Name)
	}

	_, ok := GetMigration("polls", "0001_initial")
	if !ok {
		t.Error("expected to find migration polls.0001_initial")
	}

	ResetMigrations()
	retrieved = GetMigrationsForApp("polls")
	if len(retrieved) != 0 {
		t.Errorf("expected 0 migrations after reset, got %d", len(retrieved))
	}
}

func TestMigrationKey(t *testing.T) {
	m := Migration{
		AppLabel: "polls",
		Name:     "0001_initial",
	}
	if m.Key() != "polls.0001_initial" {
		t.Errorf("expected polls.0001_initial, got %s", m.Key())
	}
}

func TestDependency(t *testing.T) {
	d := Dependency{AppLabel: "polls", Name: "0001_initial"}
	if d.String() != "polls.0001_initial" {
		t.Errorf("expected polls.0001_initial, got %s", d.String())
	}

	d2 := Dependency{Name: "0001_initial"}
	if d2.String() != "0001_initial" {
		t.Errorf("expected 0001_initial, got %s", d2.String())
	}
}

func TestGraphTopologicalSort(t *testing.T) {
	graph := NewGraph()

	m1 := &Migration{AppLabel: "polls", Name: "0001_initial"}
	m2 := &Migration{AppLabel: "polls", Name: "0002_add_field", Dependencies: []Dependency{{AppLabel: "polls", Name: "0001_initial"}}}
	m3 := &Migration{AppLabel: "polls", Name: "0003_alter", Dependencies: []Dependency{{AppLabel: "polls", Name: "0002_add_field"}}}

	graph.Add(m1)
	graph.Add(m2)
	graph.Add(m3)

	sorted, err := graph.TopologicalSort()
	if err != nil {
		t.Errorf("TopologicalSort failed: %v", err)
	}

	if len(sorted) != 3 {
		t.Errorf("expected 3 migrations, got %d", len(sorted))
	}

	order := make(map[string]int)
	for i, m := range sorted {
		order[m.Key()] = i
	}
	if order["polls.0001_initial"] > order["polls.0002_add_field"] {
		t.Error("expected 0001 before 0002")
	}
	if order["polls.0002_add_field"] > order["polls.0003_alter"] {
		t.Error("expected 0002 before 0003")
	}
}

func TestAutoDetector(t *testing.T) {
	before := NewProjectState()
	after := NewProjectState()

	after.AddModel(&ModelState{
		AppLabel:  "polls",
		ModelName: "Question",
		TableName: "polls_question",
		PKField:   "ID",
		Fields: []orm.FieldDef{
			orm.BigAutoField("ID"),
			orm.CharField("QuestionText", 200),
		},
	})

	detector := NewAutoDetector("polls", before, after)
	ops, err := detector.Detect()
	if err != nil {
		t.Fatalf("Detect failed: %v", err)
	}

	if len(ops) != 1 {
		t.Fatalf("expected 1 operation, got %d", len(ops))
	}

	createModel, ok := ops[0].(*CreateModel)
	if !ok {
		t.Fatalf("expected CreateModel, got %T", ops[0])
	}
	if createModel.Name != "Question" {
		t.Errorf("expected Question, got %s", createModel.Name)
	}
}

func TestAutoDetectorAddField(t *testing.T) {
	before := NewProjectState()
	after := NewProjectState()

	before.AddModel(&ModelState{
		AppLabel:  "polls",
		ModelName: "Question",
		TableName: "polls_question",
		PKField:   "ID",
		Fields: []orm.FieldDef{
			orm.BigAutoField("ID"),
		},
	})

	after.AddModel(&ModelState{
		AppLabel:  "polls",
		ModelName: "Question",
		TableName: "polls_question",
		PKField:   "ID",
		Fields: []orm.FieldDef{
			orm.BigAutoField("ID"),
			orm.CharField("QuestionText", 200),
		},
	})

	detector := NewAutoDetector("polls", before, after)
	ops, err := detector.Detect()
	if err != nil {
		t.Fatalf("Detect failed: %v", err)
	}

	if len(ops) != 1 {
		t.Fatalf("expected 1 operation, got %d", len(ops))
	}

	addField, ok := ops[0].(*AddField)
	if !ok {
		t.Fatalf("expected AddField, got %T", ops[0])
	}
	if addField.FieldName != "QuestionText" {
		t.Errorf("expected QuestionText, got %s", addField.FieldName)
	}
}

func TestAutoDetectorRemoveField(t *testing.T) {
	before := NewProjectState()
	after := NewProjectState()

	before.AddModel(&ModelState{
		AppLabel:  "polls",
		ModelName: "Question",
		TableName: "polls_question",
		PKField:   "ID",
		Fields: []orm.FieldDef{
			orm.BigAutoField("ID"),
			orm.CharField("QuestionText", 200),
		},
	})

	after.AddModel(&ModelState{
		AppLabel:  "polls",
		ModelName: "Question",
		TableName: "polls_question",
		PKField:   "ID",
		Fields: []orm.FieldDef{
			orm.BigAutoField("ID"),
		},
	})

	detector := NewAutoDetector("polls", before, after)
	ops, err := detector.Detect()
	if err != nil {
		t.Fatalf("Detect failed: %v", err)
	}

	found := false
	for _, op := range ops {
		if _, ok := op.(*RemoveField); ok {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected to find a RemoveField operation")
	}
}

func TestAutoDetectorAlterField(t *testing.T) {
	before := NewProjectState()
	after := NewProjectState()

	before.AddModel(&ModelState{
		AppLabel:  "polls",
		ModelName: "Question",
		TableName: "polls_question",
		PKField:   "ID",
		Fields: []orm.FieldDef{
			orm.BigAutoField("ID"),
			orm.CharField("QuestionText", 200),
		},
	})

	after.AddModel(&ModelState{
		AppLabel:  "polls",
		ModelName: "Question",
		TableName: "polls_question",
		PKField:   "ID",
		Fields: []orm.FieldDef{
			orm.BigAutoField("ID"),
			orm.CharField("QuestionText", 500),
		},
	})

	detector := NewAutoDetector("polls", before, after)
	ops, err := detector.Detect()
	if err != nil {
		t.Fatalf("Detect failed: %v", err)
	}

	found := false
	for _, op := range ops {
		if _, ok := op.(*AlterField); ok {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected to find an AlterField operation")
	}
}

func TestAutoDetectorNoChanges(t *testing.T) {
	state := NewProjectState()
	state.AddModel(&ModelState{
		AppLabel:  "polls",
		ModelName: "Question",
		TableName: "polls_question",
		PKField:   "ID",
		Fields: []orm.FieldDef{
			orm.BigAutoField("ID"),
			orm.CharField("QuestionText", 200),
		},
	})

	detector := NewAutoDetector("polls", state, state)
	ops, err := detector.Detect()
	if err != nil {
		t.Fatalf("Detect failed: %v", err)
	}

	if len(ops) != 0 {
		t.Errorf("expected 0 operations for identical states, got %d", len(ops))
	}
}

func TestDiscoverMigrationsEmpty(t *testing.T) {
	tmpDir := t.TempDir()
	appDir := filepath.Join(tmpDir, "myapp")
	os.MkdirAll(filepath.Join(appDir, "migrations"), 0755)

	infos, err := DiscoverMigrations(appDir)
	if err != nil {
		t.Fatalf("DiscoverMigrations failed: %v", err)
	}
	if len(infos) != 0 {
		t.Errorf("expected 0 migrations in empty dir, got %d", len(infos))
	}
}

func TestDiscoverMigrationsNonexistent(t *testing.T) {
	tmpDir := t.TempDir()
	infos, err := DiscoverMigrations(filepath.Join(tmpDir, "nonexistent"))
	if err != nil {
		t.Fatalf("DiscoverMigrations on nonexistent dir failed: %v", err)
	}
	if len(infos) != 0 {
		t.Errorf("expected 0 migrations for nonexistent dir, got %d", len(infos))
	}
}

func TestNextMigrationNumber(t *testing.T) {
	tmpDir := t.TempDir()
	appDir := filepath.Join(tmpDir, "myapp")
	migDir := filepath.Join(appDir, "migrations")
	os.MkdirAll(migDir, 0755)

	num, err := NextMigrationNumber(appDir)
	if err != nil {
		t.Fatalf("NextMigrationNumber failed: %v", err)
	}
	if num != 1 {
		t.Errorf("expected 1 for empty migrations dir, got %d", num)
	}

	os.WriteFile(filepath.Join(migDir, "0001_initial.go"), []byte("package migrations"), 0644)
	num, err = NextMigrationNumber(appDir)
	if err != nil {
		t.Fatalf("NextMigrationNumber failed: %v", err)
	}
	if num != 2 {
		t.Errorf("expected 2 after one migration, got %d", num)
	}
}

func TestGenerateMigrationCode(t *testing.T) {
	m := Migration{
		AppLabel: "polls",
		Name:     "0001_initial",
		Operations: []Operation{
			&CreateModel{
				Name:      "Question",
				TableName: "polls_question",
				PKField:   "ID",
				Fields: []orm.FieldDef{
					orm.BigAutoField("ID"),
					orm.CharField("QuestionText", 200),
				},
			},
		},
		IsInitial: true,
	}

	code, err := GenerateMigrationCode(m)
	if err != nil {
		t.Fatalf("GenerateMigrationCode failed: %v", err)
	}

	if len(code) == 0 {
		t.Error("expected non-empty generated code")
	}

	if !containsString(code, "package migrations") {
		t.Error("expected 'package migrations' in generated code")
	}
	if !containsString(code, "RegisterMigration") {
		t.Error("expected 'RegisterMigration' in generated code")
	}
	if !containsString(code, "0001_initial") {
		t.Error("expected '0001_initial' in generated code")
	}
	if !containsString(code, "CreateModel") {
		t.Error("expected 'CreateModel' in generated code")
	}
	if !containsString(code, "IsInitial") {
		t.Error("expected 'IsInitial' in generated code")
	}
}

func TestSeparateDatabaseAndState(t *testing.T) {
	ps := NewProjectState()
	createState := &CreateModel{
		Name:      "Question",
		TableName: "polls_question",
		PKField:   "ID",
		Fields:    []orm.FieldDef{orm.BigAutoField("ID")},
	}

	op := &SeparateDatabaseAndState{
		StateOperations:    []Operation{createState},
		DatabaseOperations: []Operation{},
	}

	err := op.StateForwards("polls", ps)
	if err != nil {
		t.Errorf("SeparateDatabaseAndState StateForwards failed: %v", err)
	}

	_, ok := ps.GetModel("polls", "Question")
	if !ok {
		t.Error("expected model to exist in project state after state operation")
	}
}

func TestProjectStateFromRegistry(t *testing.T) {
	state := ProjectStateFromRegistry()
	if state == nil {
		t.Error("expected non-nil ProjectState")
	}
}

func TestSQLLoggingEditor(t *testing.T) {
	editor := &SQLLoggingEditor{}

	err := editor.ExecuteSQL("SELECT 1")
	if err != nil {
		t.Errorf("ExecuteSQL failed: %v", err)
	}
	if len(editor.statements) != 1 {
		t.Errorf("expected 1 statement, got %d", len(editor.statements))
	}
	if editor.statements[0] != "SELECT 1" {
		t.Errorf("expected 'SELECT 1', got %s", editor.statements[0])
	}
}

func TestTableNameFor(t *testing.T) {
	if tableNameFor("polls", "Question", "") != "polls_Question" {
		t.Error("expected default table name")
	}
	if tableNameFor("polls", "Question", "custom_table") != "custom_table" {
		t.Error("expected explicit table name")
	}
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstr(s, substr))
}

func containsSubstr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
