package forms

type Field interface {
	Name() string
	Validate(value interface{}) error
	Clean(value interface{}) (interface{}, error)
	IsRequired() bool
}

type Form struct {
	Fields    map[string]Field
	Errors    map[string][]string
	CleanedData map[string]interface{}
	bound     bool
}

func NewForm(fields map[string]Field) *Form {
	return &Form{
		Fields: fields,
		Errors: make(map[string][]string),
		CleanedData: make(map[string]interface{}),
	}
}

func (f *Form) IsBound() bool {
	return f.bound
}

func (f *Form) IsValid() bool {
	return len(f.Errors) == 0
}

func (f *Form) AddError(field, message string) {
	f.Errors[field] = append(f.Errors[field], message)
}