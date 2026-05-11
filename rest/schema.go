package rest

import (
	"strings"

	"github.com/iMerica/jango/orm"
	"github.com/iMerica/jango/urls"
)

type SchemaGenerator struct {
	Title       string
	Version     string
	Description string
	Patterns    []urls.Pattern
	Models      []*orm.ModelMeta
}

func (g SchemaGenerator) OpenAPI() map[string]interface{} {
	title := g.Title
	if title == "" {
		title = "JanGO API"
	}
	version := g.Version
	if version == "" {
		version = "0.1.0"
	}
	paths := make(map[string]interface{})
	for _, pattern := range g.Patterns {
		if pattern.IsInclude {
			continue
		}
		path := openAPIPath(pattern.Path)
		methods := map[string]interface{}{}
		for _, method := range []string{"get", "post", "put", "patch", "delete", "options"} {
			methods[method] = map[string]interface{}{
				"responses": map[string]interface{}{
					"200": map[string]interface{}{"description": "OK"},
				},
			}
		}
		paths[path] = methods
	}
	components := map[string]interface{}{"schemas": g.modelSchemas()}
	return map[string]interface{}{
		"openapi": "3.0.3",
		"info": map[string]interface{}{
			"title":       title,
			"version":     version,
			"description": g.Description,
		},
		"paths":      paths,
		"components": components,
	}
}

func (g SchemaGenerator) modelSchemas() map[string]interface{} {
	models := g.Models
	if len(models) == 0 {
		models = orm.GlobalRegistry().AllModels()
	}
	schemas := make(map[string]interface{}, len(models))
	for _, meta := range models {
		properties := make(map[string]interface{})
		required := []string{}
		for _, field := range meta.ConcreteFields() {
			name := meta.DBColumnForField(field.Name)
			properties[name] = map[string]interface{}{"type": openAPIType(field)}
			if !field.Nullable && !field.Blank && field.Default == nil && !field.Auto {
				required = append(required, name)
			}
		}
		schemas[meta.AppLabel+"."+meta.ModelName] = map[string]interface{}{
			"type":       "object",
			"properties": properties,
			"required":   required,
		}
	}
	return schemas
}

func openAPIPath(path string) string {
	path = strings.ReplaceAll(path, "<int:", "{")
	path = strings.ReplaceAll(path, "<slug:", "{")
	path = strings.ReplaceAll(path, "<str:", "{")
	path = strings.ReplaceAll(path, "<uuid:", "{")
	path = strings.ReplaceAll(path, "<", "{")
	path = strings.ReplaceAll(path, ">", "}")
	return path
}

func openAPIType(field orm.FieldDef) string {
	switch field.FieldType {
	case orm.AutoFieldType, orm.BigAutoFieldType, orm.SmallAutoFieldType,
		orm.IntFieldType, orm.BigIntFieldType, orm.SmallIntFieldType,
		orm.PositiveIntFieldType, orm.PositiveBigIntFieldType, orm.PositiveSmallIntFieldType,
		orm.ForeignKeyType, orm.OneToOneType:
		return "integer"
	case orm.BooleanFieldType, orm.NullBooleanFieldType:
		return "boolean"
	case orm.FloatFieldType, orm.DoubleFieldType, orm.DecimalFieldType:
		return "number"
	case orm.JSONFieldType, orm.ArrayFieldType:
		return "object"
	default:
		return "string"
	}
}
