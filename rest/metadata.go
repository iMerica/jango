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
	if provider, ok := view.(interface {
		serializerSchemaProvider() SerializerSchemaProvider
	}); ok {
		if serializer := provider.serializerSchemaProvider(); serializer != nil {
			data["fields"] = metadataFieldsFromSerializer(serializer)
		}
	} else if provider, ok := view.(SerializerSchemaProvider); ok {
		data["fields"] = metadataFieldsFromSerializer(provider)
	}
	if versioned, ok := view.(interface{ versioningStrategy() VersioningStrategy }); ok {
		if strategy := versioned.versioningStrategy(); strategy != nil {
			data["versioning"] = strategy.VersioningMetadata()
		}
	}
	return data
}

func metadataFieldsFromSerializer(serializer SerializerSchemaProvider) map[string]interface{} {
	fields := make(map[string]interface{})
	for _, field := range serializer.SchemaFields() {
		fields[field.Name] = map[string]interface{}{
			"type":       field.Type,
			"required":   field.Required,
			"read_only":  field.ReadOnly,
			"write_only": field.WriteOnly,
			"max_length": field.MaxLength,
			"help_text":  field.HelpText,
		}
	}
	return fields
}

func (v ModelViewSet[T]) serializerSchemaProvider() SerializerSchemaProvider {
	provider, _ := v.Serializer.(SerializerSchemaProvider)
	return provider
}

func (v ModelViewSet[T]) versioningStrategy() VersioningStrategy {
	return v.Versioning
}
