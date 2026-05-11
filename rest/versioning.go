package rest

import (
	"fmt"
	"strings"
)

type VersioningStrategy interface {
	DetermineVersion(req *APIRequest, view interface{}) (string, error)
	VersioningMetadata() map[string]interface{}
	SchemaParameters() []OpenAPIParameter
}

type QueryParameterVersioning struct {
	ParamName       string
	DefaultVersion  string
	AllowedVersions []string
}

func (v QueryParameterVersioning) DetermineVersion(req *APIRequest, view interface{}) (string, error) {
	param := v.ParamName
	if param == "" {
		param = "version"
	}
	version := strings.TrimSpace(req.QueryParams().Get(param))
	if version == "" {
		version = v.DefaultVersion
	}
	return validateVersion(version, v.AllowedVersions)
}

func (v QueryParameterVersioning) VersioningMetadata() map[string]interface{} {
	param := v.ParamName
	if param == "" {
		param = "version"
	}
	return map[string]interface{}{"type": "query", "param": param, "default": v.DefaultVersion, "allowed": v.AllowedVersions}
}

func (v QueryParameterVersioning) SchemaParameters() []OpenAPIParameter {
	param := v.ParamName
	if param == "" {
		param = "version"
	}
	return []OpenAPIParameter{{
		Name:        param,
		In:          "query",
		Required:    false,
		Description: "API version",
		Schema:      versionSchema(v.AllowedVersions),
	}}
}

type HeaderVersioning struct {
	HeaderName      string
	DefaultVersion  string
	AllowedVersions []string
}

func (v HeaderVersioning) DetermineVersion(req *APIRequest, view interface{}) (string, error) {
	header := v.HeaderName
	if header == "" {
		header = "X-API-Version"
	}
	version := strings.TrimSpace(req.Header.Get(header))
	if version == "" {
		version = v.DefaultVersion
	}
	return validateVersion(version, v.AllowedVersions)
}

func (v HeaderVersioning) VersioningMetadata() map[string]interface{} {
	header := v.HeaderName
	if header == "" {
		header = "X-API-Version"
	}
	return map[string]interface{}{"type": "header", "header": header, "default": v.DefaultVersion, "allowed": v.AllowedVersions}
}

func (v HeaderVersioning) SchemaParameters() []OpenAPIParameter {
	header := v.HeaderName
	if header == "" {
		header = "X-API-Version"
	}
	return []OpenAPIParameter{{
		Name:        header,
		In:          "header",
		Required:    false,
		Description: "API version",
		Schema:      versionSchema(v.AllowedVersions),
	}}
}

type URLPathVersioning struct {
	ParamName       string
	DefaultVersion  string
	AllowedVersions []string
}

func (v URLPathVersioning) DetermineVersion(req *APIRequest, view interface{}) (string, error) {
	param := v.ParamName
	if param == "" {
		param = "version"
	}
	version := strings.TrimSpace(req.URLParam(param))
	if version == "" {
		version = v.DefaultVersion
	}
	return validateVersion(version, v.AllowedVersions)
}

func (v URLPathVersioning) VersioningMetadata() map[string]interface{} {
	param := v.ParamName
	if param == "" {
		param = "version"
	}
	return map[string]interface{}{"type": "path", "param": param, "default": v.DefaultVersion, "allowed": v.AllowedVersions}
}

func (v URLPathVersioning) SchemaParameters() []OpenAPIParameter {
	param := v.ParamName
	if param == "" {
		param = "version"
	}
	return []OpenAPIParameter{{
		Name:        param,
		In:          "path",
		Required:    true,
		Description: "API version",
		Schema:      versionSchema(v.AllowedVersions),
	}}
}

type NamespaceVersioning struct {
	ParamName       string
	DefaultVersion  string
	AllowedVersions []string
}

func (v NamespaceVersioning) DetermineVersion(req *APIRequest, view interface{}) (string, error) {
	param := v.ParamName
	if param == "" {
		param = "namespace"
	}
	version := strings.TrimSpace(req.URLParam(param))
	if version == "" {
		version = v.DefaultVersion
	}
	return validateVersion(version, v.AllowedVersions)
}

func (v NamespaceVersioning) VersioningMetadata() map[string]interface{} {
	param := v.ParamName
	if param == "" {
		param = "namespace"
	}
	return map[string]interface{}{"type": "namespace", "param": param, "default": v.DefaultVersion, "allowed": v.AllowedVersions}
}

func (v NamespaceVersioning) SchemaParameters() []OpenAPIParameter {
	param := v.ParamName
	if param == "" {
		param = "namespace"
	}
	return []OpenAPIParameter{{
		Name:        param,
		In:          "path",
		Required:    true,
		Description: "API namespace version",
		Schema:      versionSchema(v.AllowedVersions),
	}}
}

func validateVersion(version string, allowed []string) (string, error) {
	if version == "" || len(allowed) == 0 {
		return version, nil
	}
	for _, candidate := range allowed {
		if version == candidate {
			return version, nil
		}
	}
	return "", fmt.Errorf("unsupported API version %q", version)
}

func versionSchema(allowed []string) map[string]interface{} {
	schema := map[string]interface{}{"type": "string"}
	if len(allowed) > 0 {
		enum := make([]interface{}, 0, len(allowed))
		for _, version := range allowed {
			enum = append(enum, version)
		}
		schema["enum"] = enum
	}
	return schema
}
