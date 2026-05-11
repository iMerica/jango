package rest

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/iMerica/jango/orm"
)

type Serializer[T any] interface {
	Serialize(instance interface{}) (map[string]interface{}, error)
	SerializeList(instances []*T) ([]map[string]interface{}, error)
	Fields() []string
	ModelMeta() *orm.ModelMeta
	SetContext(ctx SerializerContext)
	Bind(data map[string]interface{}) error
	BindPartial(data map[string]interface{}) error
	IsValid() bool
	Errors() ValidationError
	ValidatedData() map[string]interface{}
	Create(ctx context.Context, qs *orm.QuerySet[T]) (*T, error)
	Update(ctx context.Context, qs *orm.QuerySet[T], instance *T) (*T, error)
}

type SerializerOption func(*serializerOptions)

type serializerOptions struct {
	fields           []string
	exclude          []string
	fieldOptions     map[string]FieldOptions
	fieldValidators  map[string][]func(interface{}) error
	objectValidators []func(map[string]interface{}) error
	createFunc       func(context.Context, map[string]interface{}) (interface{}, error)
	updateFunc       func(context.Context, interface{}, map[string]interface{}) (interface{}, error)
}

type SerializerContext struct {
	Request *APIRequest
	View    interface{}
	Format  string
}

type ValidationError map[string][]string

func (e ValidationError) Error() string {
	if len(e) == 0 {
		return "validation failed"
	}
	parts := make([]string, 0, len(e))
	for field, messages := range e {
		parts = append(parts, field+": "+strings.Join(messages, ", "))
	}
	return strings.Join(parts, "; ")
}

func (e ValidationError) Add(field, message string) {
	if field == "" {
		field = "non_field_errors"
	}
	e[field] = append(e[field], message)
}

type FieldOptions struct {
	ReadOnly  bool
	WriteOnly bool
	Required  *bool
	Default   interface{}
}

type SerializerFieldInfo struct {
	Name      string
	Type      string
	Required  bool
	ReadOnly  bool
	WriteOnly bool
	Default   interface{}
	MaxLength int
	HelpText  string
}

type SerializerSchemaProvider interface {
	SchemaName() string
	SchemaFields() []SerializerFieldInfo
}

type FieldOption func(*FieldOptions)

func ReadOnly() FieldOption {
	return func(opts *FieldOptions) { opts.ReadOnly = true }
}

func WriteOnly() FieldOption {
	return func(opts *FieldOptions) { opts.WriteOnly = true }
}

func Required(required bool) FieldOption {
	return func(opts *FieldOptions) { opts.Required = &required }
}

func Default(value interface{}) FieldOption {
	return func(opts *FieldOptions) { opts.Default = value }
}

func Fields(fields ...string) SerializerOption {
	return func(opts *serializerOptions) {
		opts.fields = append(opts.fields, fields...)
	}
}

func Exclude(fields ...string) SerializerOption {
	return func(opts *serializerOptions) {
		opts.exclude = append(opts.exclude, fields...)
	}
}

func Field(name string, fieldOpts ...FieldOption) SerializerOption {
	return func(opts *serializerOptions) {
		if opts.fieldOptions == nil {
			opts.fieldOptions = make(map[string]FieldOptions)
		}
		cfg := opts.fieldOptions[name]
		for _, opt := range fieldOpts {
			opt(&cfg)
		}
		opts.fieldOptions[name] = cfg
	}
}

func ValidateField(name string, validator func(interface{}) error) SerializerOption {
	return func(opts *serializerOptions) {
		if opts.fieldValidators == nil {
			opts.fieldValidators = make(map[string][]func(interface{}) error)
		}
		opts.fieldValidators[name] = append(opts.fieldValidators[name], validator)
	}
}

func ValidateObject(validator func(map[string]interface{}) error) SerializerOption {
	return func(opts *serializerOptions) {
		opts.objectValidators = append(opts.objectValidators, validator)
	}
}

func WithCreate[T any](fn func(context.Context, map[string]interface{}) (*T, error)) SerializerOption {
	return func(opts *serializerOptions) {
		opts.createFunc = func(ctx context.Context, data map[string]interface{}) (interface{}, error) {
			return fn(ctx, data)
		}
	}
}

func WithUpdate[T any](fn func(context.Context, *T, map[string]interface{}) (*T, error)) SerializerOption {
	return func(opts *serializerOptions) {
		opts.updateFunc = func(ctx context.Context, instance interface{}, data map[string]interface{}) (interface{}, error) {
			typed, ok := instance.(*T)
			if !ok {
				return nil, fmt.Errorf("rest: update hook expected *T")
			}
			return fn(ctx, typed, data)
		}
	}
}

type ModelSerializer[T any] struct {
	meta      *orm.ModelMeta
	bindings  []fieldBinding
	options   serializerOptions
	context   SerializerContext
	input     map[string]interface{}
	validated map[string]interface{}
	errors    ValidationError
	partial   bool
	err       error
}

type fieldBinding struct {
	field       orm.FieldDef
	outputName  string
	structNames []string
	options     FieldOptions
}

func NewModelSerializer[T any](meta *orm.ModelMeta, opts ...SerializerOption) *ModelSerializer[T] {
	cfg := serializerOptions{}
	for _, opt := range opts {
		opt(&cfg)
	}
	s := &ModelSerializer[T]{meta: meta, options: cfg}
	s.bindings, s.err = buildFieldBindings(meta, cfg)
	return s
}

func (s *ModelSerializer[T]) ModelMeta() *orm.ModelMeta {
	return s.meta
}

func (s *ModelSerializer[T]) Fields() []string {
	fields := make([]string, 0, len(s.bindings))
	for _, b := range s.bindings {
		if b.options.WriteOnly {
			continue
		}
		fields = append(fields, b.outputName)
	}
	return fields
}

func (s *ModelSerializer[T]) SchemaName() string {
	if s.meta == nil {
		return "Serializer"
	}
	return s.meta.AppLabel + "." + s.meta.ModelName
}

func (s *ModelSerializer[T]) SchemaFields() []SerializerFieldInfo {
	fields := make([]SerializerFieldInfo, 0, len(s.bindings))
	for _, binding := range s.bindings {
		fields = append(fields, SerializerFieldInfo{
			Name:      binding.outputName,
			Type:      openAPIType(binding.field),
			Required:  serializerFieldRequired(binding),
			ReadOnly:  binding.options.ReadOnly || (binding.field.PrimaryKey && binding.field.Auto),
			WriteOnly: binding.options.WriteOnly,
			Default:   binding.options.Default,
			MaxLength: binding.field.MaxLength,
			HelpText:  binding.field.HelpText,
		})
	}
	return fields
}

func (s *ModelSerializer[T]) SetContext(ctx SerializerContext) {
	s.context = ctx
}

func (s *ModelSerializer[T]) Serialize(instance interface{}) (map[string]interface{}, error) {
	if s.err != nil {
		return nil, s.err
	}
	if s.meta == nil {
		return nil, fmt.Errorf("rest: model serializer requires model metadata")
	}
	values, err := instanceValue(instance)
	if err != nil {
		return nil, err
	}
	result := make(map[string]interface{}, len(s.bindings))
	for _, binding := range s.bindings {
		if binding.options.WriteOnly {
			continue
		}
		result[binding.outputName] = readBindingValue(values, binding)
	}
	return result, nil
}

func (s *ModelSerializer[T]) SerializeList(instances []*T) ([]map[string]interface{}, error) {
	result := make([]map[string]interface{}, 0, len(instances))
	for _, instance := range instances {
		serialized, err := s.Serialize(instance)
		if err != nil {
			return nil, err
		}
		result = append(result, serialized)
	}
	return result, nil
}

func (s *ModelSerializer[T]) Bind(data map[string]interface{}) error {
	return s.bind(data, false)
}

func (s *ModelSerializer[T]) BindPartial(data map[string]interface{}) error {
	return s.bind(data, true)
}

func (s *ModelSerializer[T]) bind(data map[string]interface{}, partial bool) error {
	s.input = data
	s.partial = partial
	s.validated = make(map[string]interface{})
	s.errors = make(ValidationError)
	if s.err != nil {
		s.errors.Add("non_field_errors", s.err.Error())
		return s.errors
	}
	if s.meta == nil {
		s.errors.Add("non_field_errors", "model metadata is required")
		return s.errors
	}

	for _, binding := range s.bindings {
		if binding.options.ReadOnly {
			continue
		}
		value, ok := lookupInput(data, binding.outputName, binding.field.Name)
		if !ok {
			if binding.options.Default != nil {
				s.validated[validatedKey(binding)] = binding.options.Default
				continue
			}
			if !partial && serializerFieldRequired(binding) {
				s.errors.Add(binding.outputName, "this field is required")
			}
			continue
		}
		cleaned, err := coerceSerializerValue(binding.field, value)
		if err != nil {
			s.errors.Add(binding.outputName, err.Error())
			continue
		}
		for _, validator := range s.options.fieldValidators[binding.outputName] {
			if err := validator(cleaned); err != nil {
				s.errors.Add(binding.outputName, err.Error())
			}
		}
		for _, validator := range s.options.fieldValidators[binding.field.Name] {
			if err := validator(cleaned); err != nil {
				s.errors.Add(binding.outputName, err.Error())
			}
		}
		s.validated[validatedKey(binding)] = cleaned
	}
	if len(s.errors) == 0 {
		for _, validator := range s.options.objectValidators {
			if err := validator(s.validated); err != nil {
				s.errors.Add("non_field_errors", err.Error())
			}
		}
	}
	if len(s.errors) > 0 {
		return s.errors
	}
	return nil
}

func (s *ModelSerializer[T]) IsValid() bool {
	return s.errors != nil && len(s.errors) == 0
}

func (s *ModelSerializer[T]) Errors() ValidationError {
	if s.errors == nil {
		return ValidationError{}
	}
	return s.errors
}

func (s *ModelSerializer[T]) ValidatedData() map[string]interface{} {
	data := make(map[string]interface{}, len(s.validated))
	for k, v := range s.validated {
		data[k] = v
	}
	return data
}

func (s *ModelSerializer[T]) Create(ctx context.Context, qs *orm.QuerySet[T]) (*T, error) {
	if s.options.createFunc != nil {
		created, err := s.options.createFunc(ctx, s.ValidatedData())
		if err != nil {
			return nil, err
		}
		typed, ok := created.(*T)
		if !ok {
			return nil, fmt.Errorf("rest: create hook returned unexpected type")
		}
		return typed, nil
	}
	if qs == nil {
		return nil, fmt.Errorf("rest: queryset is required")
	}
	instance := new(T)
	created, err := qs.BaseQuerySet.Create(ctx, s.validated)
	if err != nil {
		return nil, err
	}
	if err := assignMapToStruct(instance, s.validated); err != nil {
		return nil, err
	}
	_ = assignMapToStruct(instance, created)
	return instance, nil
}

func (s *ModelSerializer[T]) Update(ctx context.Context, qs *orm.QuerySet[T], instance *T) (*T, error) {
	if instance == nil {
		return nil, fmt.Errorf("rest: instance is required")
	}
	if s.options.updateFunc != nil {
		updated, err := s.options.updateFunc(ctx, instance, s.ValidatedData())
		if err != nil {
			return nil, err
		}
		typed, ok := updated.(*T)
		if !ok {
			return nil, fmt.Errorf("rest: update hook returned unexpected type")
		}
		return typed, nil
	}
	if err := assignMapToStruct(instance, s.validated); err != nil {
		return nil, err
	}
	if qs != nil {
		pkValue, ok := readPKValue(instance, s.meta)
		if ok && len(s.validated) > 0 {
			if _, err := qs.Filter(orm.L(s.meta.PKColumn(), pkValue)).BaseQuerySet.Update(ctx, s.validated); err != nil {
				return nil, err
			}
		}
	}
	return instance, nil
}

func buildFieldBindings(meta *orm.ModelMeta, cfg serializerOptions) ([]fieldBinding, error) {
	if meta == nil {
		return nil, fmt.Errorf("rest: model serializer requires model metadata")
	}
	if len(cfg.fields) > 0 && len(cfg.exclude) > 0 {
		return nil, fmt.Errorf("rest: Fields and Exclude cannot be used together")
	}

	var fields []orm.FieldDef
	if len(cfg.fields) > 0 {
		for _, name := range cfg.fields {
			field, ok := meta.FieldForNameOrColumn(name)
			if !ok {
				return nil, fmt.Errorf("rest: unknown serializer field %q on %s.%s", name, meta.AppLabel, meta.ModelName)
			}
			if field.FieldType == orm.ManyToManyType {
				return nil, fmt.Errorf("rest: many-to-many serializer field %q is not supported yet", name)
			}
			fields = append(fields, field)
		}
	} else {
		excluded := make(map[string]bool, len(cfg.exclude))
		for _, name := range cfg.exclude {
			field, ok := meta.FieldForNameOrColumn(name)
			if !ok {
				return nil, fmt.Errorf("rest: unknown serializer exclude field %q on %s.%s", name, meta.AppLabel, meta.ModelName)
			}
			excluded[field.DBColumn] = true
		}
		for _, field := range meta.ConcreteFields() {
			if excluded[field.DBColumn] {
				continue
			}
			fields = append(fields, field)
		}
	}

	bindings := make([]fieldBinding, 0, len(fields))
	for _, field := range fields {
		column := meta.DBColumnForField(field.Name)
		fieldOptions := cfg.fieldOptions[column]
		if namedOptions, ok := cfg.fieldOptions[field.Name]; ok {
			fieldOptions = mergeFieldOptions(fieldOptions, namedOptions)
		}
		bindings = append(bindings, fieldBinding{
			field:      field,
			outputName: column,
			structNames: []string{
				field.Name,
				dbColumnToGoField(column),
			},
			options: fieldOptions,
		})
	}
	return bindings, nil
}

func instanceValue(instance interface{}) (reflect.Value, error) {
	if instance == nil {
		return reflect.Value{}, fmt.Errorf("rest: cannot serialize nil instance")
	}
	val := reflect.ValueOf(instance)
	for val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return reflect.Value{}, fmt.Errorf("rest: cannot serialize nil instance")
		}
		val = val.Elem()
	}
	if val.Kind() != reflect.Struct && val.Kind() != reflect.Map {
		return reflect.Value{}, fmt.Errorf("rest: serializer expected struct or map, got %s", val.Kind())
	}
	return val, nil
}

func readBindingValue(val reflect.Value, binding fieldBinding) interface{} {
	if val.Kind() == reflect.Map {
		for _, key := range []string{binding.outputName, binding.field.Name} {
			mapVal := val.MapIndex(reflect.ValueOf(key))
			if mapVal.IsValid() {
				return mapVal.Interface()
			}
		}
		return nil
	}
	for _, name := range binding.structNames {
		field := val.FieldByName(name)
		if field.IsValid() && field.CanInterface() {
			return field.Interface()
		}
	}
	return nil
}

func dbColumnToGoField(column string) string {
	parts := strings.Split(column, "_")
	for i, part := range parts {
		if strings.EqualFold(part, "id") {
			parts[i] = "ID"
			continue
		}
		if part == "" {
			continue
		}
		parts[i] = strings.ToUpper(part[:1]) + part[1:]
	}
	return strings.Join(parts, "")
}

func mergeFieldOptions(left, right FieldOptions) FieldOptions {
	if right.ReadOnly {
		left.ReadOnly = true
	}
	if right.WriteOnly {
		left.WriteOnly = true
	}
	if right.Required != nil {
		left.Required = right.Required
	}
	if right.Default != nil {
		left.Default = right.Default
	}
	return left
}

func lookupInput(data map[string]interface{}, names ...string) (interface{}, bool) {
	for _, name := range names {
		if value, ok := data[name]; ok {
			return value, true
		}
	}
	return nil, false
}

func serializerFieldRequired(binding fieldBinding) bool {
	if binding.options.Required != nil {
		return *binding.options.Required
	}
	field := binding.field
	if field.PrimaryKey && field.Auto {
		return false
	}
	if field.AutoNow || field.AutoNowAdd || field.Nullable || field.Blank || field.Default != nil {
		return false
	}
	return true
}

func validatedKey(binding fieldBinding) string {
	if len(binding.structNames) > 1 && binding.structNames[1] != "" {
		return binding.structNames[1]
	}
	return binding.field.Name
}

func coerceSerializerValue(field orm.FieldDef, value interface{}) (interface{}, error) {
	if value == nil {
		if field.Nullable {
			return nil, nil
		}
		return nil, fmt.Errorf("null is not allowed")
	}
	switch field.FieldType {
	case orm.AutoFieldType, orm.BigAutoFieldType, orm.SmallAutoFieldType,
		orm.IntFieldType, orm.BigIntFieldType, orm.SmallIntFieldType,
		orm.PositiveIntFieldType, orm.PositiveBigIntFieldType, orm.PositiveSmallIntFieldType,
		orm.ForeignKeyType, orm.OneToOneType:
		return toInt64(value, field.DBColumn)
	case orm.BooleanFieldType, orm.NullBooleanFieldType:
		return toBoolValue(value, field.DBColumn)
	case orm.FloatFieldType, orm.DoubleFieldType, orm.DecimalFieldType:
		return toFloat64(value, field.DBColumn)
	case orm.DateTimeFieldType:
		return toTime(value, time.RFC3339, field.DBColumn+" must be an RFC3339 timestamp")
	case orm.DateFieldType:
		return toTime(value, "2006-01-02", field.DBColumn+" must be a date in YYYY-MM-DD format")
	case orm.TimeFieldType:
		return toTime(value, "15:04:05", field.DBColumn+" must be a time in HH:MM:SS format")
	default:
		return value, nil
	}
}

func toInt64(value interface{}, name string) (interface{}, error) {
	switch v := value.(type) {
	case int:
		return int64(v), nil
	case int8:
		return int64(v), nil
	case int16:
		return int64(v), nil
	case int32:
		return int64(v), nil
	case int64:
		return v, nil
	case uint:
		return int64(v), nil
	case uint8:
		return int64(v), nil
	case uint16:
		return int64(v), nil
	case uint32:
		return int64(v), nil
	case uint64:
		return int64(v), nil
	case float64:
		return int64(v), nil
	case string:
		parsed, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("%s must be an integer", name)
		}
		return parsed, nil
	default:
		return nil, fmt.Errorf("%s must be an integer", name)
	}
}

func toBoolValue(value interface{}, name string) (interface{}, error) {
	switch v := value.(type) {
	case bool:
		return v, nil
	case string:
		parsed, err := strconv.ParseBool(v)
		if err != nil {
			return nil, fmt.Errorf("%s must be a boolean", name)
		}
		return parsed, nil
	default:
		return nil, fmt.Errorf("%s must be a boolean", name)
	}
}

func toFloat64(value interface{}, name string) (interface{}, error) {
	switch v := value.(type) {
	case float32:
		return float64(v), nil
	case float64:
		return v, nil
	case int:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case string:
		parsed, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return nil, fmt.Errorf("%s must be a number", name)
		}
		return parsed, nil
	default:
		return nil, fmt.Errorf("%s must be a number", name)
	}
}

func toTime(value interface{}, layout, message string) (interface{}, error) {
	if t, ok := value.(time.Time); ok {
		return t, nil
	}
	raw, ok := value.(string)
	if !ok {
		return nil, errors.New(message)
	}
	parsed, err := time.Parse(layout, raw)
	if err != nil {
		return nil, errors.New(message)
	}
	return parsed, nil
}

func assignMapToStruct(instance interface{}, data map[string]interface{}) error {
	val := reflect.ValueOf(instance)
	if val.Kind() != reflect.Ptr || val.IsNil() {
		return fmt.Errorf("rest: instance must be a pointer")
	}
	val = val.Elem()
	if val.Kind() != reflect.Struct {
		return fmt.Errorf("rest: instance must point to a struct")
	}
	for name, value := range data {
		field := val.FieldByName(name)
		if !field.IsValid() || !field.CanSet() || value == nil {
			continue
		}
		rv := reflect.ValueOf(value)
		if rv.Type().AssignableTo(field.Type()) {
			field.Set(rv)
			continue
		}
		if rv.Type().ConvertibleTo(field.Type()) {
			field.Set(rv.Convert(field.Type()))
		}
	}
	return nil
}

func readPKValue(instance interface{}, meta *orm.ModelMeta) (interface{}, bool) {
	if meta == nil || meta.PKField == "" || instance == nil {
		return nil, false
	}
	val := reflect.ValueOf(instance)
	for val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return nil, false
		}
		val = val.Elem()
	}
	if val.Kind() != reflect.Struct {
		return nil, false
	}
	field := val.FieldByName(meta.PKField)
	if !field.IsValid() || !field.CanInterface() {
		return nil, false
	}
	return field.Interface(), true
}

var _ Serializer[struct{}] = (*ModelSerializer[struct{}])(nil)
var _ SerializerSchemaProvider = (*ModelSerializer[struct{}])(nil)
