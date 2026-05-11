package forms

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/iMerica/jango/orm"
)

type Field interface {
	Name() string
	Validate(value interface{}) error
	Clean(value interface{}) (interface{}, error)
	IsRequired() bool
}

type Validator func(interface{}) error

type BaseField struct {
	FieldName  string
	Required   bool
	Validators []Validator
}

func (f BaseField) Name() string     { return f.FieldName }
func (f BaseField) IsRequired() bool { return f.Required }
func (f BaseField) Validate(v interface{}) error {
	if f.Required && emptyValue(v) {
		return fmt.Errorf("this field is required")
	}
	for _, validator := range f.Validators {
		if err := validator(v); err != nil {
			return err
		}
	}
	return nil
}

type CharField struct {
	BaseField
	MaxLength int
}

func (f CharField) Clean(value interface{}) (interface{}, error) {
	if err := f.Validate(value); err != nil {
		return "", err
	}
	if !f.Required && emptyValue(value) {
		return "", nil
	}
	s := strings.TrimSpace(fmt.Sprint(value))
	if f.MaxLength > 0 && len(s) > f.MaxLength {
		return "", fmt.Errorf("ensure this value has at most %d characters", f.MaxLength)
	}
	return s, nil
}

type IntegerField struct{ BaseField }

func (f IntegerField) Clean(value interface{}) (interface{}, error) {
	if err := f.Validate(value); err != nil {
		return int64(0), err
	}
	if !f.Required && emptyValue(value) {
		return nil, nil
	}
	switch v := value.(type) {
	case int64:
		return v, nil
	case int:
		return int64(v), nil
	case string:
		parsed, err := strconv.ParseInt(strings.TrimSpace(v), 10, 64)
		if err != nil {
			return nil, fmt.Errorf("enter a whole number")
		}
		return parsed, nil
	default:
		return nil, fmt.Errorf("enter a whole number")
	}
}

type BooleanField struct{ BaseField }

func (f BooleanField) Clean(value interface{}) (interface{}, error) {
	if err := f.Validate(value); err != nil {
		return false, err
	}
	if !f.Required && emptyValue(value) {
		return false, nil
	}
	switch v := value.(type) {
	case bool:
		return v, nil
	case string:
		parsed, err := strconv.ParseBool(strings.TrimSpace(v))
		if err != nil {
			return nil, fmt.Errorf("enter either true or false")
		}
		return parsed, nil
	default:
		return nil, fmt.Errorf("enter either true or false")
	}
}

type Form struct {
	Fields      map[string]Field
	Errors      map[string][]string
	CleanedData map[string]interface{}
	bound       bool
	clean       func(map[string]interface{}) error
}

func NewForm(fields map[string]Field) *Form {
	return &Form{
		Fields:      fields,
		Errors:      make(map[string][]string),
		CleanedData: make(map[string]interface{}),
	}
}

func (f *Form) IsBound() bool {
	return f.bound
}

func (f *Form) IsValid() bool {
	if !f.bound {
		return false
	}
	return len(f.Errors) == 0
}

func (f *Form) AddError(field, message string) {
	f.Errors[field] = append(f.Errors[field], message)
}

func (f *Form) Bind(data map[string]interface{}) bool {
	f.bound = true
	f.Errors = make(map[string][]string)
	f.CleanedData = make(map[string]interface{})
	for name, field := range f.Fields {
		raw, ok := data[name]
		if !ok {
			raw = ""
		}
		cleaned, err := field.Clean(raw)
		if err != nil {
			f.AddError(name, err.Error())
			continue
		}
		f.CleanedData[name] = cleaned
	}
	if len(f.Errors) == 0 && f.clean != nil {
		if err := f.clean(f.CleanedData); err != nil {
			f.AddError("__all__", err.Error())
		}
	}
	return f.IsValid()
}

func (f *Form) BindValues(values url.Values) bool {
	data := make(map[string]interface{}, len(values))
	for key, vals := range values {
		if len(vals) > 0 {
			data[key] = vals[0]
		}
	}
	return f.Bind(data)
}

func (f *Form) SetClean(fn func(map[string]interface{}) error) {
	f.clean = fn
}

func NewModelForm(meta *orm.ModelMeta, fields ...string) *Form {
	allowed := make(map[string]bool, len(fields))
	for _, field := range fields {
		allowed[field] = true
	}
	formFields := make(map[string]Field)
	for _, field := range meta.ConcreteFields() {
		if len(allowed) > 0 && !allowed[field.Name] && !allowed[meta.DBColumnForField(field.Name)] {
			continue
		}
		if field.PrimaryKey && field.Auto {
			continue
		}
		name := meta.DBColumnForField(field.Name)
		required := !field.Nullable && !field.Blank && field.Default == nil
		base := BaseField{FieldName: name, Required: required}
		switch field.FieldType {
		case orm.IntFieldType, orm.BigIntFieldType, orm.SmallIntFieldType,
			orm.PositiveIntFieldType, orm.PositiveBigIntFieldType, orm.PositiveSmallIntFieldType,
			orm.ForeignKeyType, orm.OneToOneType:
			formFields[name] = IntegerField{BaseField: base}
		case orm.BooleanFieldType, orm.NullBooleanFieldType:
			formFields[name] = BooleanField{BaseField: base}
		default:
			formFields[name] = CharField{BaseField: base, MaxLength: field.MaxLength}
		}
	}
	return NewForm(formFields)
}

func emptyValue(value interface{}) bool {
	if value == nil {
		return true
	}
	if s, ok := value.(string); ok {
		return strings.TrimSpace(s) == ""
	}
	return false
}
