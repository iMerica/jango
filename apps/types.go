package apps

type App interface {
	Config() *AppConfig
}

type ReadyHook func() error

type ModelInfo struct {
	AppLabel   string
	ModelName  string
	TableName  string
	PKField    string
	Fields     []FieldInfo
	Relations  []RelationInfo
	Indexes    []IndexInfo
	Constraints []ConstraintInfo
}

type FieldInfo struct {
	Name         string
	Type         string
	Nullable     bool
	Unique       bool
	DBColumn     string
	MaxLength    int
	Default      interface{}
	IsRelation   bool
	RelatedModel string
	RelatedName  string
	OnDelete     string
}

type RelationInfo struct {
	Name         string
	Type         string
	RelatedModel string
	RelatedName  string
	OnDelete     string
	Through      string
}

type IndexInfo struct {
	Name         string
	Fields       []string
	Unique       bool
	Condition    string
	Opclasses    []string
}

type ConstraintInfo struct {
	Name      string
	Unique    []string
	Check     string
	Condition string
}

type CommandInfo struct {
	Name        string
	AppLabel    string
	Description string
}