package orm

import (
	"fmt"
	"time"
)

type FieldType string

const (
	AutoFieldType             FieldType = "auto"
	BigAutoFieldType          FieldType = "bigauto"
	SmallAutoFieldType        FieldType = "smallauto"
	CharFieldType             FieldType = "char"
	TextFieldType             FieldType = "text"
	SlugFieldType             FieldType = "slug"
	EmailFieldType            FieldType = "email"
	URLFieldType              FieldType = "url"
	UUIDFieldType             FieldType = "uuid"
	IPAddressFieldType        FieldType = "ipaddress"
	GenericIPFieldType        FieldType = "genericip"
	IntFieldType              FieldType = "int"
	BigIntFieldType           FieldType = "bigint"
	SmallIntFieldType         FieldType = "smallint"
	PositiveIntFieldType      FieldType = "positiveint"
	PositiveSmallIntFieldType FieldType = "positivesmallint"
	PositiveBigIntFieldType   FieldType = "positivebigint"
	FloatFieldType            FieldType = "float"
	DoubleFieldType           FieldType = "double"
	DecimalFieldType          FieldType = "decimal"
	BooleanFieldType          FieldType = "boolean"
	NullBooleanFieldType      FieldType = "nullboolean"
	DateFieldType             FieldType = "date"
	TimeFieldType             FieldType = "time"
	DateTimeFieldType         FieldType = "datetime"
	DurationFieldType         FieldType = "duration"
	FileFieldType             FieldType = "file"
	ImageFieldType            FieldType = "image"
	FilePathFieldType         FieldType = "filepath"
	JSONFieldType             FieldType = "json"
	ArrayFieldType            FieldType = "array"
	ForeignKeyType            FieldType = "foreignkey"
	OneToOneType              FieldType = "onetoonetype"
	ManyToManyType            FieldType = "manytomany"
)

type OnDelete string

const (
	Cascade    OnDelete = "cascade"
	Protect    OnDelete = "protect"
	SetNull    OnDelete = "set_null"
	SetDefault OnDelete = "set_default"
	Set        OnDelete = "set"
	DoNothing  OnDelete = "do_nothing"
	Restrict   OnDelete = "restrict"
)

type FieldDef struct {
	Name           string
	DBColumn       string
	FieldType      FieldType
	PrimaryKey     bool
	Auto           bool
	Nullable       bool
	Unique         bool
	UniqueForDate  string
	UniqueForMonth string
	UniqueForYear  string
	DBIndex        bool
	DbTablespace   string
	Default        interface{}
	Choices        []Choice
	HelpText       string
	VerboseName    string
	Blank          bool
	Editable       bool
	Serial         bool
	Validators     []ValidatorFunc
	ErrorMessages  map[string]string

	MaxLength int
	Decimals  int
	MinValue  interface{}
	MaxValue  interface{}

	AutoNow    bool
	AutoNowAdd bool

	RelatedModel     string
	RelatedName      string
	RelatedQueryName string
	OnDelete         OnDelete
	DBConstraint     bool
	SelfRelation     bool

	Through       string
	ThroughFields []string
	Symmetric     bool

	DBComment string
}

type Choice struct {
	Value interface{}
	Label string
}

type ValidatorFunc func(value interface{}) error

type IndexDef struct {
	Name         string
	Fields       []string
	Unique       bool
	Condition    string
	Include      []string
	Opclasses    []string
	Tables       string
	DBTablespace string
	Concurrently bool
}

type ConstraintDef struct {
	Name       string
	Check      string
	Unique     []string
	Condition  string
	Deferrable string
	Deferred   bool
	Violated   string
}

func AutoField(name string, opts ...FieldOption) FieldDef {
	f := FieldDef{
		Name:       name,
		FieldType:  AutoFieldType,
		PrimaryKey: true,
		Auto:       true,
		Editable:   false,
	}
	for _, opt := range opts {
		opt(&f)
	}
	return f
}

func BigAutoField(name string, opts ...FieldOption) FieldDef {
	f := FieldDef{
		Name:       name,
		FieldType:  BigAutoFieldType,
		PrimaryKey: true,
		Auto:       true,
		Editable:   false,
	}
	for _, opt := range opts {
		opt(&f)
	}
	return f
}

func SmallAutoField(name string, opts ...FieldOption) FieldDef {
	f := FieldDef{
		Name:       name,
		FieldType:  SmallAutoFieldType,
		PrimaryKey: true,
		Auto:       true,
		Editable:   false,
	}
	for _, opt := range opts {
		opt(&f)
	}
	return f
}

func CharField(name string, maxLength int, opts ...FieldOption) FieldDef {
	f := FieldDef{
		Name:      name,
		FieldType: CharFieldType,
		MaxLength: maxLength,
		Editable:  true,
	}
	for _, opt := range opts {
		opt(&f)
	}
	return f
}

func TextField(name string, opts ...FieldOption) FieldDef {
	f := FieldDef{
		Name:      name,
		FieldType: TextFieldType,
		Editable:  true,
	}
	for _, opt := range opts {
		opt(&f)
	}
	return f
}

func SlugField(name string, maxLength int, opts ...FieldOption) FieldDef {
	f := FieldDef{
		Name:      name,
		FieldType: SlugFieldType,
		MaxLength: maxLength,
		Editable:  true,
	}
	for _, opt := range opts {
		opt(&f)
	}
	return f
}

func EmailField(name string, maxLength int, opts ...FieldOption) FieldDef {
	if maxLength <= 0 {
		maxLength = 254
	}
	f := FieldDef{
		Name:      name,
		FieldType: EmailFieldType,
		MaxLength: maxLength,
		Editable:  true,
	}
	for _, opt := range opts {
		opt(&f)
	}
	return f
}

func URLField(name string, maxLength int, opts ...FieldOption) FieldDef {
	if maxLength <= 0 {
		maxLength = 200
	}
	f := FieldDef{
		Name:      name,
		FieldType: URLFieldType,
		MaxLength: maxLength,
		Editable:  true,
	}
	for _, opt := range opts {
		opt(&f)
	}
	return f
}

func UUIDField(name string, opts ...FieldOption) FieldDef {
	f := FieldDef{
		Name:      name,
		FieldType: UUIDFieldType,
		Editable:  true,
	}
	for _, opt := range opts {
		opt(&f)
	}
	return f
}

func IPAddressField(name string, opts ...FieldOption) FieldDef {
	f := FieldDef{
		Name:      name,
		FieldType: IPAddressFieldType,
		Editable:  true,
	}
	for _, opt := range opts {
		opt(&f)
	}
	return f
}

func GenericIPField(name string, opts ...FieldOption) FieldDef {
	f := FieldDef{
		Name:      name,
		FieldType: GenericIPFieldType,
		Editable:  true,
	}
	for _, opt := range opts {
		opt(&f)
	}
	return f
}

func IntegerField(name string, opts ...FieldOption) FieldDef {
	f := FieldDef{
		Name:      name,
		FieldType: IntFieldType,
		Editable:  true,
	}
	for _, opt := range opts {
		opt(&f)
	}
	return f
}

func BigIntegerField(name string, opts ...FieldOption) FieldDef {
	f := FieldDef{
		Name:      name,
		FieldType: BigIntFieldType,
		Editable:  true,
	}
	for _, opt := range opts {
		opt(&f)
	}
	return f
}

func SmallIntegerField(name string, opts ...FieldOption) FieldDef {
	f := FieldDef{
		Name:      name,
		FieldType: SmallIntFieldType,
		Editable:  true,
	}
	for _, opt := range opts {
		opt(&f)
	}
	return f
}

func PositiveIntegerField(name string, opts ...FieldOption) FieldDef {
	f := FieldDef{
		Name:      name,
		FieldType: PositiveIntFieldType,
		MinValue:  0,
		Editable:  true,
	}
	for _, opt := range opts {
		opt(&f)
	}
	return f
}

func PositiveSmallIntegerField(name string, opts ...FieldOption) FieldDef {
	f := FieldDef{
		Name:      name,
		FieldType: PositiveSmallIntFieldType,
		MinValue:  0,
		Editable:  true,
	}
	for _, opt := range opts {
		opt(&f)
	}
	return f
}

func PositiveBigIntegerField(name string, opts ...FieldOption) FieldDef {
	f := FieldDef{
		Name:      name,
		FieldType: PositiveBigIntFieldType,
		MinValue:  0,
		Editable:  true,
	}
	for _, opt := range opts {
		opt(&f)
	}
	return f
}

func FloatField(name string, opts ...FieldOption) FieldDef {
	f := FieldDef{
		Name:      name,
		FieldType: FloatFieldType,
		Editable:  true,
	}
	for _, opt := range opts {
		opt(&f)
	}
	return f
}

func DoubleField(name string, opts ...FieldOption) FieldDef {
	f := FieldDef{
		Name:      name,
		FieldType: DoubleFieldType,
		Editable:  true,
	}
	for _, opt := range opts {
		opt(&f)
	}
	return f
}

func DecimalField(name string, maxDigits, decimalPlaces int, opts ...FieldOption) FieldDef {
	f := FieldDef{
		Name:      name,
		FieldType: DecimalFieldType,
		MaxLength: maxDigits,
		Decimals:  decimalPlaces,
		Editable:  true,
	}
	for _, opt := range opts {
		opt(&f)
	}
	return f
}

func BooleanField(name string, opts ...FieldOption) FieldDef {
	f := FieldDef{
		Name:      name,
		FieldType: BooleanFieldType,
		Editable:  true,
	}
	for _, opt := range opts {
		opt(&f)
	}
	return f
}

func NullBooleanField(name string, opts ...FieldOption) FieldDef {
	f := FieldDef{
		Name:      name,
		FieldType: NullBooleanFieldType,
		Nullable:  true,
		Editable:  true,
	}
	for _, opt := range opts {
		opt(&f)
	}
	return f
}

func DateField(name string, opts ...FieldOption) FieldDef {
	f := FieldDef{
		Name:      name,
		FieldType: DateFieldType,
		Editable:  true,
	}
	for _, opt := range opts {
		opt(&f)
	}
	return f
}

func TimeField(name string, opts ...FieldOption) FieldDef {
	f := FieldDef{
		Name:      name,
		FieldType: TimeFieldType,
		Editable:  true,
	}
	for _, opt := range opts {
		opt(&f)
	}
	return f
}

func DateTimeField(name string, opts ...FieldOption) FieldDef {
	f := FieldDef{
		Name:      name,
		FieldType: DateTimeFieldType,
		Editable:  true,
	}
	for _, opt := range opts {
		opt(&f)
	}
	return f
}

func DurationField(name string, opts ...FieldOption) FieldDef {
	f := FieldDef{
		Name:      name,
		FieldType: DurationFieldType,
		Editable:  true,
	}
	for _, opt := range opts {
		opt(&f)
	}
	return f
}

func FileField(name string, maxLength int, opts ...FieldOption) FieldDef {
	f := FieldDef{
		Name:      name,
		FieldType: FileFieldType,
		MaxLength: maxLength,
		Editable:  true,
	}
	for _, opt := range opts {
		opt(&f)
	}
	return f
}

func ImageField(name string, maxLength int, opts ...FieldOption) FieldDef {
	f := FieldDef{
		Name:      name,
		FieldType: ImageFieldType,
		MaxLength: maxLength,
		Editable:  true,
	}
	for _, opt := range opts {
		opt(&f)
	}
	return f
}

func FilePathField(name string, opts ...FieldOption) FieldDef {
	f := FieldDef{
		Name:      name,
		FieldType: FilePathFieldType,
		Editable:  true,
	}
	for _, opt := range opts {
		opt(&f)
	}
	return f
}

func JSONField(name string, opts ...FieldOption) FieldDef {
	f := FieldDef{
		Name:      name,
		FieldType: JSONFieldType,
		Nullable:  true,
		Editable:  true,
	}
	for _, opt := range opts {
		opt(&f)
	}
	return f
}

func ForeignKey(name string, relatedModel string, opts ...FieldOption) FieldDef {
	f := FieldDef{
		Name:         name,
		FieldType:    ForeignKeyType,
		RelatedModel: relatedModel,
		OnDelete:     Cascade,
		RelatedName:  "+",
		DBConstraint: true,
		Editable:     true,
		DBIndex:      true,
	}
	for _, opt := range opts {
		opt(&f)
	}
	if f.DBColumn == "" {
		f.DBColumn = GoFieldToDBColumn(name) + "_id"
	}
	return f
}

func OneToOneField(name string, relatedModel string, opts ...FieldOption) FieldDef {
	f := FieldDef{
		Name:         name,
		FieldType:    OneToOneType,
		RelatedModel: relatedModel,
		OnDelete:     Cascade,
		RelatedName:  "+",
		DBConstraint: true,
		Unique:       true,
		Editable:     true,
		DBIndex:      true,
	}
	for _, opt := range opts {
		opt(&f)
	}
	if f.DBColumn == "" {
		f.DBColumn = GoFieldToDBColumn(name) + "_id"
	}
	return f
}

func ManyToManyField(name string, relatedModel string, opts ...FieldOption) FieldDef {
	f := FieldDef{
		Name:         name,
		FieldType:    ManyToManyType,
		RelatedModel: relatedModel,
		RelatedName:  "+",
		Symmetric:    true,
		Editable:     true,
	}
	for _, opt := range opts {
		opt(&f)
	}
	return f
}

type FieldOption func(*FieldDef)

func WithNullable(f *FieldDef) { f.Nullable = true }
func WithUnique(f *FieldDef)   { f.Unique = true }
func WithDefault(v interface{}) FieldOption {
	return func(f *FieldDef) { f.Default = v }
}
func WithOnDelete(od OnDelete) FieldOption {
	return func(f *FieldDef) { f.OnDelete = od }
}
func WithRelatedName(name string) FieldOption {
	return func(f *FieldDef) { f.RelatedName = name }
}
func WithRelatedQueryName(name string) FieldOption {
	return func(f *FieldDef) { f.RelatedQueryName = name }
}
func WithDBColumn(col string) FieldOption {
	return func(f *FieldDef) { f.DBColumn = col }
}
func WithDBIndex(f *FieldDef) { f.DBIndex = true }
func WithThrough(through string) FieldOption {
	return func(f *FieldDef) { f.Through = through }
}
func WithThroughFields(fields ...string) FieldOption {
	return func(f *FieldDef) { f.ThroughFields = fields }
}
func WithSymmetric(b bool) FieldOption {
	return func(f *FieldDef) { f.Symmetric = b }
}
func WithDBConstraint(b bool) FieldOption {
	return func(f *FieldDef) { f.DBConstraint = b }
}
func WithAutoNow(f *FieldDef)    { f.AutoNow = true }
func WithAutoNowAdd(f *FieldDef) { f.AutoNowAdd = true }
func WithChoices(choices []Choice) FieldOption {
	return func(f *FieldDef) { f.Choices = choices }
}
func WithHelpText(text string) FieldOption {
	return func(f *FieldDef) { f.HelpText = text }
}
func WithVerboseName(name string) FieldOption {
	return func(f *FieldDef) { f.VerboseName = name }
}
func WithDBComment(comment string) FieldOption {
	return func(f *FieldDef) { f.DBComment = comment }
}
func WithValidators(validators ...ValidatorFunc) FieldOption {
	return func(f *FieldDef) { f.Validators = append(f.Validators, validators...) }
}

type ModelOptions struct {
	TableName          string
	Ordering           []string
	DefaultOrdering    []string
	VerboseName        string
	VerboseNamePlural  string
	DefaultManagerName string
	UniqueTogether     [][]string
	Indexes            []IndexDef
	Constraints        []ConstraintDef
	DBTableComment     string
	Managed            bool
	AppLabel           string
	Swappable          string
	DefaultAutoField   FieldType
	DefaultPermissions []string
}

type ModelMeta struct {
	AppLabel        string
	ModelName       string
	TableName       string
	PKField         string
	Fields          []FieldDef
	Indexes         []IndexDef
	Constraints     []ConstraintDef
	DefaultOrdering []string
	Options         ModelOptions
	Managers        map[string]*ManagerDef
}

func (m *ModelMeta) FieldByName(name string) (FieldDef, bool) {
	for _, f := range m.Fields {
		if f.Name == name {
			return f, true
		}
	}
	return FieldDef{}, false
}

func (m *ModelMeta) FieldForNameOrColumn(name string) (FieldDef, bool) {
	for _, f := range m.Fields {
		if f.Name == name || f.DBColumn == name || GoFieldToDBColumn(f.Name) == name {
			return f, true
		}
	}
	return FieldDef{}, false
}

func (m *ModelMeta) DBColumnForField(name string) string {
	f, ok := m.FieldForNameOrColumn(name)
	if !ok {
		return GoFieldToDBColumn(name)
	}
	if f.DBColumn != "" {
		return f.DBColumn
	}
	return GoFieldToDBColumn(f.Name)
}

func (m *ModelMeta) PKColumn() string {
	return m.DBColumnForField(m.PKField)
}

func (m *ModelMeta) AutoPKField() (FieldDef, bool) {
	if m.PKField == "" {
		return FieldDef{}, false
	}
	f, ok := m.FieldForNameOrColumn(m.PKField)
	if !ok {
		return FieldDef{}, false
	}
	return f, f.Auto
}

func (m *ModelMeta) IsRelation(name string) bool {
	f, ok := m.FieldForNameOrColumn(name)
	if !ok {
		return false
	}
	return f.FieldType == ForeignKeyType || f.FieldType == OneToOneType || f.FieldType == ManyToManyType
}

func (m *ModelMeta) ConcreteFields() []FieldDef {
	var result []FieldDef
	for _, f := range m.Fields {
		if f.FieldType != ManyToManyType {
			result = append(result, f)
		}
	}
	return result
}

func (m *ModelMeta) LocalFields() []FieldDef {
	var result []FieldDef
	for _, f := range m.Fields {
		if f.FieldType != ManyToManyType && f.FieldType != ForeignKeyType && f.FieldType != OneToOneType {
			result = append(result, f)
		}
	}
	return result
}

func (m *ModelMeta) Relations() []FieldDef {
	var result []FieldDef
	for _, f := range m.Fields {
		if f.FieldType == ForeignKeyType || f.FieldType == OneToOneType || f.FieldType == ManyToManyType {
			result = append(result, f)
		}
	}
	return result
}

func (m *ModelMeta) FullTableName() string {
	if m.TableName != "" {
		return m.TableName
	}
	if m.AppLabel != "" {
		return m.AppLabel + "_" + m.TableName
	}
	return m.ModelName
}

type ManagerDef struct {
	Name string
}

type ModelInstance interface {
	TableName() string
	PKValue() interface{}
	SetPKValue(interface{})
}

func DefaultModelOptions() ModelOptions {
	return ModelOptions{
		Managed:            true,
		DefaultAutoField:   AutoFieldType,
		DefaultPermissions: []string{"add", "change", "delete", "view"},
		DefaultManagerName: "objects",
	}
}

func InferDBType(f FieldDef) string {
	switch f.FieldType {
	case AutoFieldType, SmallAutoFieldType:
		return "SERIAL"
	case BigAutoFieldType:
		return "BIGSERIAL"
	case CharFieldType, SlugFieldType, EmailFieldType, URLFieldType:
		maxLen := f.MaxLength
		if maxLen <= 0 {
			maxLen = 255
		}
		return fmt.Sprintf("VARCHAR(%d)", maxLen)
	case TextFieldType:
		return "TEXT"
	case IPAddressFieldType:
		return "INET"
	case GenericIPFieldType:
		return "INET"
	case UUIDFieldType:
		return "UUID"
	case IntFieldType, PositiveIntFieldType:
		return "INTEGER"
	case BigIntFieldType, PositiveBigIntFieldType:
		return "BIGINT"
	case SmallIntFieldType, PositiveSmallIntFieldType:
		return "SMALLINT"
	case FloatFieldType:
		return "REAL"
	case DoubleFieldType:
		return "DOUBLE PRECISION"
	case DecimalFieldType:
		return fmt.Sprintf("NUMERIC(%d,%d)", f.MaxLength, f.Decimals)
	case BooleanFieldType:
		return "BOOLEAN"
	case NullBooleanFieldType:
		return "BOOLEAN"
	case DateFieldType:
		return "DATE"
	case TimeFieldType:
		return "TIME"
	case DateTimeFieldType:
		return "TIMESTAMPTZ"
	case DurationFieldType:
		return "INTERVAL"
	case FileFieldType, ImageFieldType, FilePathFieldType:
		return "VARCHAR(255)"
	case JSONFieldType:
		return "JSONB"
	case ArrayFieldType:
		return "TEXT[]"
	case ForeignKeyType:
		return "BIGINT"
	case OneToOneType:
		return "BIGINT"
	case ManyToManyType:
		return ""
	default:
		return "TEXT"
	}
}

func zeroValueForField(f FieldDef) interface{} {
	switch f.FieldType {
	case AutoFieldType, BigAutoFieldType, SmallAutoFieldType, IntFieldType, BigIntFieldType,
		SmallIntFieldType, PositiveIntFieldType, PositiveSmallIntFieldType, PositiveBigIntFieldType:
		return int64(0)
	case FloatFieldType, DoubleFieldType, DecimalFieldType:
		return float64(0)
	case BooleanFieldType, NullBooleanFieldType:
		return false
	case CharFieldType, TextFieldType, SlugFieldType, EmailFieldType, URLFieldType,
		FileFieldType, ImageFieldType, FilePathFieldType, IPAddressFieldType, GenericIPFieldType:
		return ""
	case UUIDFieldType:
		return ""
	case DateFieldType:
		return time.Time{}
	case TimeFieldType:
		return time.Time{}
	case DateTimeFieldType:
		return time.Time{}
	case DurationFieldType:
		return time.Duration(0)
	case JSONFieldType:
		return nil
	case ForeignKeyType, OneToOneType:
		return int64(0)
	case ManyToManyType:
		return nil
	default:
		return nil
	}
}
