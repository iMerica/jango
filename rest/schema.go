package rest

import (
	"net/http"
	"reflect"
	"strings"

	"github.com/iMerica/jango/orm"
	"github.com/iMerica/jango/urls"
)

type OpenAPIParameter struct {
	Name        string
	In          string
	Required    bool
	Description string
	Schema      map[string]interface{}
}

type SchemaGenerator struct {
	Title       string
	Version     string
	Description string
	Patterns    []urls.Pattern
	Models      []*orm.ModelMeta
	Serializers []SerializerSchemaProvider
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
		metadata, ok := patternRouteMetadata(pattern)
		if !ok {
			metadata = fallbackRouteMetadata(pattern)
		}
		for _, method := range metadata.Methods {
			if method == http.MethodHead || method == http.MethodOptions {
				continue
			}
			methods[strings.ToLower(method)] = g.operation(pattern, metadata, method)
		}
		paths[path] = methods
	}
	components := map[string]interface{}{"schemas": g.schemas()}
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

func (g SchemaGenerator) operation(pattern urls.Pattern, metadata RouteMetadata, method string) map[string]interface{} {
	action := actionNameForMethod(metadata, method)
	tag := strings.Trim(pattern.Path, "/")
	if idx := strings.Index(tag, "/"); idx >= 0 {
		tag = tag[:idx]
	}
	if tag == "" && metadata.ModelMeta != nil {
		tag = metadata.ModelMeta.ModelName
	}
	operationID := operationID(pattern.Name, method, action)
	parameters := openAPIParameters(pattern.Path, metadata)
	responses := map[string]interface{}{
		"200": map[string]interface{}{"description": "OK"},
	}
	if method == http.MethodPost {
		responses["201"] = map[string]interface{}{"description": "Created"}
	}
	if method == http.MethodDelete {
		responses["204"] = map[string]interface{}{"description": "No Content"}
	}
	if metadata.Throttled {
		responses["429"] = map[string]interface{}{"description": "Too Many Requests"}
	}
	operation := map[string]interface{}{
		"operationId": operationID,
		"tags":        []string{tag},
		"parameters":  parameters,
		"responses":   responses,
	}
	if metadata.AuthRequired {
		operation["security"] = []map[string][]string{{"JanGOAuth": []string{}}}
	}
	if method == http.MethodPost || method == http.MethodPut || method == http.MethodPatch {
		if schemaRef := serializerSchemaRef(metadata.Serializer); schemaRef != nil {
			operation["requestBody"] = map[string]interface{}{
				"required": true,
				"content": map[string]interface{}{
					"application/json": map[string]interface{}{"schema": schemaRef},
				},
			}
		}
	}
	if schemaRef := serializerSchemaRef(metadata.Serializer); schemaRef != nil && method != http.MethodDelete {
		responses["200"].(map[string]interface{})["content"] = map[string]interface{}{
			"application/json": map[string]interface{}{"schema": responseSchemaForAction(schemaRef, action)},
		}
	}
	return operation
}

func fallbackRouteMetadata(pattern urls.Pattern) RouteMetadata {
	return RouteMetadata{Methods: []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodOptions}}
}

func actionNameForMethod(metadata RouteMetadata, method string) string {
	for _, action := range metadata.Actions {
		if strings.EqualFold(action.Method, method) {
			return action.Name
		}
	}
	switch method {
	case http.MethodGet:
		return "retrieve"
	case http.MethodPost:
		return "create"
	case http.MethodPut:
		return "update"
	case http.MethodPatch:
		return "partial_update"
	case http.MethodDelete:
		return "destroy"
	default:
		return strings.ToLower(method)
	}
}

func operationID(patternName, method, action string) string {
	base := strings.ReplaceAll(patternName, "-", "_")
	base = strings.ReplaceAll(base, ":", "_")
	if base == "" {
		base = strings.ToLower(method)
	}
	if action != "" && !strings.Contains(base, action) {
		return base + "_" + action
	}
	return base
}

func openAPIParameters(path string, metadata RouteMetadata) []map[string]interface{} {
	var params []map[string]interface{}
	for _, name := range pathParamNames(path) {
		params = append(params, map[string]interface{}{
			"name":     name,
			"in":       "path",
			"required": true,
			"schema":   map[string]interface{}{"type": pathParamType(path, name)},
		})
	}
	for _, field := range metadata.FilterFields {
		params = append(params, queryParam(field, "string", "Filter by "+field))
	}
	if len(metadata.SearchFields) > 0 {
		params = append(params, queryParam("search", "string", "Search query"))
	}
	if len(metadata.OrderingFields) > 0 {
		params = append(params, queryParam("ordering", "string", "Comma-separated ordering fields"))
	}
	params = append(params, paginationParameters(metadata.Paginator)...)
	if metadata.Versioning != nil {
		for _, param := range metadata.Versioning.SchemaParameters() {
			params = append(params, openAPIParameterMap(param))
		}
	}
	return params
}

func queryParam(name, typ, description string) map[string]interface{} {
	return map[string]interface{}{
		"name":        name,
		"in":          "query",
		"required":    false,
		"description": description,
		"schema":      map[string]interface{}{"type": typ},
	}
}

func paginationParameters(paginator interface{}) []map[string]interface{} {
	if paginator == nil {
		return []map[string]interface{}{queryParam("limit", "integer", "Maximum number of results"), queryParam("offset", "integer", "Result offset")}
	}
	name := reflect.TypeOf(paginator).String()
	switch {
	case strings.Contains(name, "PageNumberPagination"):
		return []map[string]interface{}{queryParam("page", "integer", "Page number"), queryParam("page_size", "integer", "Page size")}
	case strings.Contains(name, "CursorPagination"):
		return []map[string]interface{}{queryParam("cursor", "string", "Pagination cursor")}
	default:
		return []map[string]interface{}{queryParam("limit", "integer", "Maximum number of results"), queryParam("offset", "integer", "Result offset")}
	}
}

func openAPIParameterMap(param OpenAPIParameter) map[string]interface{} {
	return map[string]interface{}{
		"name":        param.Name,
		"in":          param.In,
		"required":    param.Required,
		"description": param.Description,
		"schema":      param.Schema,
	}
}

func pathParamNames(path string) []string {
	var names []string
	remaining := path
	for {
		start := strings.Index(remaining, "<")
		if start < 0 {
			return names
		}
		end := strings.Index(remaining[start:], ">")
		if end < 0 {
			return names
		}
		spec := remaining[start+1 : start+end]
		if colon := strings.Index(spec, ":"); colon >= 0 {
			spec = spec[colon+1:]
		}
		names = append(names, spec)
		remaining = remaining[start+end+1:]
	}
}

func pathParamType(path, name string) string {
	marker := "<int:" + name + ">"
	if strings.Contains(path, marker) {
		return "integer"
	}
	return "string"
}

func serializerSchemaRef(serializer interface{}) map[string]interface{} {
	provider, ok := serializer.(SerializerSchemaProvider)
	if !ok || provider == nil {
		return nil
	}
	return map[string]interface{}{"$ref": "#/components/schemas/" + provider.SchemaName()}
}

func responseSchemaForAction(schemaRef map[string]interface{}, action string) map[string]interface{} {
	if action == "list" {
		return map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"count":   map[string]interface{}{"type": "integer"},
				"limit":   map[string]interface{}{"type": "integer"},
				"offset":  map[string]interface{}{"type": "integer"},
				"results": map[string]interface{}{"type": "array", "items": schemaRef},
			},
		}
	}
	return schemaRef
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

func (g SchemaGenerator) schemas() map[string]interface{} {
	schemas := g.modelSchemas()
	for _, provider := range g.Serializers {
		if provider == nil {
			continue
		}
		schemas[provider.SchemaName()] = schemaFromSerializer(provider)
	}
	for _, pattern := range g.Patterns {
		metadata, ok := patternRouteMetadata(pattern)
		if !ok {
			continue
		}
		provider, ok := metadata.Serializer.(SerializerSchemaProvider)
		if !ok || provider == nil {
			continue
		}
		schemas[provider.SchemaName()] = schemaFromSerializer(provider)
	}
	return schemas
}

func patternRouteMetadata(pattern urls.Pattern) (RouteMetadata, bool) {
	if metadata, ok := pattern.Metadata.(RouteMetadata); ok {
		return metadata, true
	}
	return lookupRouteMetadata(pattern.Handler)
}

func schemaFromSerializer(provider SerializerSchemaProvider) map[string]interface{} {
	properties := make(map[string]interface{})
	required := []string{}
	for _, field := range provider.SchemaFields() {
		schema := map[string]interface{}{"type": field.Type}
		if field.MaxLength > 0 {
			schema["maxLength"] = field.MaxLength
		}
		if field.ReadOnly {
			schema["readOnly"] = true
		}
		if field.WriteOnly {
			schema["writeOnly"] = true
		}
		properties[field.Name] = schema
		if field.Required && !field.ReadOnly {
			required = append(required, field.Name)
		}
	}
	schema := map[string]interface{}{
		"type":       "object",
		"properties": properties,
	}
	if len(required) > 0 {
		schema["required"] = required
	}
	return schema
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
