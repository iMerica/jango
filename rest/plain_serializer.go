package rest

import (
	"fmt"
	"reflect"
	"time"
)

type PlainField struct {
	Name       string
	Type       string
	Required   bool
	ReadOnly   bool
	WriteOnly  bool
	Default    interface{}
	MaxLength  int
	HelpText   string
	Validators []func(interface{}) error
}

type PlainSerializerOption func(*PlainSerializer)

type PlainSerializer struct {
	Name             string
	fields           []PlainField
	objectValidators []func(map[string]interface{}) error
	context          SerializerContext
	input            map[string]interface{}
	validated        map[string]interface{}
	errors           ValidationError
	partial          bool
}

func NewPlainSerializer(name string, fields []PlainField, opts ...PlainSerializerOption) *PlainSerializer {
	s := &PlainSerializer{Name: name, fields: append([]PlainField(nil), fields...)}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

func PlainObjectValidator(validator func(map[string]interface{}) error) PlainSerializerOption {
	return func(s *PlainSerializer) {
		s.objectValidators = append(s.objectValidators, validator)
	}
}

func (s *PlainSerializer) SetContext(ctx SerializerContext) {
	s.context = ctx
}

func (s *PlainSerializer) ModelMeta() interface{} {
	return nil
}

func (s *PlainSerializer) Fields() []string {
	fields := make([]string, 0, len(s.fields))
	for _, field := range s.fields {
		if field.WriteOnly {
			continue
		}
		fields = append(fields, field.Name)
	}
	return fields
}

func (s *PlainSerializer) SchemaName() string {
	if s.Name == "" {
		return "PlainSerializer"
	}
	return s.Name
}

func (s *PlainSerializer) SchemaFields() []SerializerFieldInfo {
	fields := make([]SerializerFieldInfo, 0, len(s.fields))
	for _, field := range s.fields {
		fields = append(fields, SerializerFieldInfo{
			Name:      field.Name,
			Type:      normalizedPlainFieldType(field.Type),
			Required:  field.Required,
			ReadOnly:  field.ReadOnly,
			WriteOnly: field.WriteOnly,
			Default:   field.Default,
			MaxLength: field.MaxLength,
			HelpText:  field.HelpText,
		})
	}
	return fields
}

func (s *PlainSerializer) Serialize(instance interface{}) (map[string]interface{}, error) {
	values, err := plainValueMap(instance)
	if err != nil {
		return nil, err
	}
	result := make(map[string]interface{}, len(s.fields))
	for _, field := range s.fields {
		if field.WriteOnly {
			continue
		}
		if value, ok := values[field.Name]; ok {
			result[field.Name] = value
		}
	}
	return result, nil
}

func (s *PlainSerializer) Bind(data map[string]interface{}) error {
	return s.bind(data, false)
}

func (s *PlainSerializer) BindPartial(data map[string]interface{}) error {
	return s.bind(data, true)
}

func (s *PlainSerializer) bind(data map[string]interface{}, partial bool) error {
	s.input = data
	s.partial = partial
	s.validated = make(map[string]interface{})
	s.errors = make(ValidationError)
	for _, field := range s.fields {
		if field.ReadOnly {
			continue
		}
		value, ok := data[field.Name]
		if !ok {
			if field.Default != nil {
				s.validated[field.Name] = field.Default
				continue
			}
			if !partial && field.Required {
				s.errors.Add(field.Name, "this field is required")
			}
			continue
		}
		cleaned, err := coercePlainValue(field, value)
		if err != nil {
			s.errors.Add(field.Name, err.Error())
			continue
		}
		for _, validator := range field.Validators {
			if err := validator(cleaned); err != nil {
				s.errors.Add(field.Name, err.Error())
			}
		}
		s.validated[field.Name] = cleaned
	}
	if len(s.errors) == 0 {
		for _, validator := range s.objectValidators {
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

func (s *PlainSerializer) IsValid() bool {
	return s.errors != nil && len(s.errors) == 0
}

func (s *PlainSerializer) Errors() ValidationError {
	if s.errors == nil {
		return ValidationError{}
	}
	return s.errors
}

func (s *PlainSerializer) ValidatedData() map[string]interface{} {
	data := make(map[string]interface{}, len(s.validated))
	for k, v := range s.validated {
		data[k] = v
	}
	return data
}

func plainValueMap(instance interface{}) (map[string]interface{}, error) {
	if instance == nil {
		return nil, fmt.Errorf("rest: cannot serialize nil instance")
	}
	if data, ok := instance.(map[string]interface{}); ok {
		return data, nil
	}
	value := reflect.ValueOf(instance)
	for value.Kind() == reflect.Ptr {
		if value.IsNil() {
			return nil, fmt.Errorf("rest: cannot serialize nil instance")
		}
		value = value.Elem()
	}
	if value.Kind() != reflect.Struct {
		return nil, fmt.Errorf("rest: plain serializer expected struct or map, got %s", value.Kind())
	}
	data := make(map[string]interface{}, value.NumField())
	typ := value.Type()
	for i := 0; i < value.NumField(); i++ {
		field := value.Field(i)
		if !field.CanInterface() {
			continue
		}
		name := typ.Field(i).Name
		data[name] = field.Interface()
	}
	return data, nil
}

func coercePlainValue(field PlainField, value interface{}) (interface{}, error) {
	if value == nil {
		return nil, nil
	}
	switch normalizedPlainFieldType(field.Type) {
	case "integer":
		return toInt64(value, field.Name)
	case "number":
		return toFloat64(value, field.Name)
	case "boolean":
		return toBoolValue(value, field.Name)
	case "string":
		text, ok := value.(string)
		if !ok {
			return nil, fmt.Errorf("%s must be a string", field.Name)
		}
		if field.MaxLength > 0 && len(text) > field.MaxLength {
			return nil, fmt.Errorf("%s must be at most %d characters", field.Name, field.MaxLength)
		}
		return text, nil
	case "datetime":
		return toTime(value, time.RFC3339, field.Name+" must be an RFC3339 timestamp")
	case "date":
		return toTime(value, "2006-01-02", field.Name+" must be a date in YYYY-MM-DD format")
	case "array":
		if reflect.TypeOf(value).Kind() != reflect.Slice {
			return nil, fmt.Errorf("%s must be an array", field.Name)
		}
		return value, nil
	case "object":
		if _, ok := value.(map[string]interface{}); !ok {
			return nil, fmt.Errorf("%s must be an object", field.Name)
		}
		return value, nil
	default:
		return value, nil
	}
}

func normalizedPlainFieldType(fieldType string) string {
	switch fieldType {
	case "", "char", "text":
		return "string"
	case "int":
		return "integer"
	case "float", "decimal":
		return "number"
	default:
		return fieldType
	}
}

var _ SerializerSchemaProvider = (*PlainSerializer)(nil)
