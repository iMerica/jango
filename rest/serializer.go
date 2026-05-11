package rest

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/iMerica/jango/orm"
)

type Serializer[T any] interface {
	Serialize(instance interface{}) (map[string]interface{}, error)
	SerializeList(instances []*T) ([]map[string]interface{}, error)
	Fields() []string
	ModelMeta() *orm.ModelMeta
}

type SerializerOption func(*serializerOptions)

type serializerOptions struct {
	fields  []string
	exclude []string
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

type ModelSerializer[T any] struct {
	meta     *orm.ModelMeta
	bindings []fieldBinding
	err      error
}

type fieldBinding struct {
	field       orm.FieldDef
	outputName  string
	structNames []string
}

func NewModelSerializer[T any](meta *orm.ModelMeta, opts ...SerializerOption) *ModelSerializer[T] {
	cfg := serializerOptions{}
	for _, opt := range opts {
		opt(&cfg)
	}
	s := &ModelSerializer[T]{meta: meta}
	s.bindings, s.err = buildFieldBindings(meta, cfg)
	return s
}

func (s *ModelSerializer[T]) ModelMeta() *orm.ModelMeta {
	return s.meta
}

func (s *ModelSerializer[T]) Fields() []string {
	fields := make([]string, len(s.bindings))
	for i, b := range s.bindings {
		fields[i] = b.outputName
	}
	return fields
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
		bindings = append(bindings, fieldBinding{
			field:      field,
			outputName: column,
			structNames: []string{
				field.Name,
				dbColumnToGoField(column),
			},
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

var _ Serializer[struct{}] = (*ModelSerializer[struct{}])(nil)
