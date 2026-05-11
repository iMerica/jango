package rest

import (
	"strings"

	"github.com/iMerica/jango/orm"
)

type Metadata interface {
	DetermineMetadata(req *APIRequest, view interface{}, meta *orm.ModelMeta) map[string]interface{}
}

type SimpleMetadata struct{}

func (m SimpleMetadata) DetermineMetadata(req *APIRequest, view interface{}, meta *orm.ModelMeta) map[string]interface{} {
	data := map[string]interface{}{
		"name":        "",
		"description": "",
		"renders":     []string{"application/json", "text/html"},
		"parses":      []string{"application/json", "application/x-www-form-urlencoded", "multipart/form-data"},
	}
	if meta != nil {
		data["name"] = meta.Options.VerboseName
		if data["name"] == "" {
			data["name"] = strings.ToLower(meta.ModelName)
		}
		fields := make(map[string]interface{})
		for _, field := range meta.ConcreteFields() {
			fields[meta.DBColumnForField(field.Name)] = map[string]interface{}{
				"type":       string(field.FieldType),
				"required":   !field.Nullable && !field.Blank && field.Default == nil && !field.Auto,
				"read_only":  field.PrimaryKey && field.Auto,
				"max_length": field.MaxLength,
				"help_text":  field.HelpText,
			}
		}
		data["fields"] = fields
	}
	return data
}
