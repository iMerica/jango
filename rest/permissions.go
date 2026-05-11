package rest

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/iMerica/jango/auth"
)

type Permission interface {
	HasPermission(req *APIRequest, view interface{}) bool
	HasObjectPermission(req *APIRequest, view interface{}, obj interface{}) bool
	Message() string
}

type AllowAny struct{}

func (p AllowAny) HasPermission(req *APIRequest, view interface{}) bool { return true }
func (p AllowAny) HasObjectPermission(req *APIRequest, view interface{}, obj interface{}) bool {
	return true
}
func (p AllowAny) Message() string { return "" }

type IsAuthenticated struct{}

func (p IsAuthenticated) HasPermission(req *APIRequest, view interface{}) bool {
	return userIsAuthenticated(req.User)
}
func (p IsAuthenticated) HasObjectPermission(req *APIRequest, view interface{}, obj interface{}) bool {
	return p.HasPermission(req, view)
}
func (p IsAuthenticated) Message() string { return "authentication credentials were not provided" }

type ModelPermissions struct{}

func (p ModelPermissions) HasPermission(req *APIRequest, view interface{}) bool {
	if req.Method == http.MethodOptions {
		return true
	}
	meta := modelMetaFromView(view)
	if meta == nil {
		return false
	}
	action := permissionAction(req.Method)
	if action == "" {
		return true
	}
	return userHasPerm(req.User, fmt.Sprintf("%s.%s_%s", meta.AppLabel, action, strings.ToLower(meta.ModelName)))
}

func (p ModelPermissions) HasObjectPermission(req *APIRequest, view interface{}, obj interface{}) bool {
	return p.HasPermission(req, view)
}
func (p ModelPermissions) Message() string { return "permission denied" }

type ObjectPermissionFunc struct {
	Check  func(req *APIRequest, view interface{}, obj interface{}) bool
	Detail string
}

func (p ObjectPermissionFunc) HasPermission(req *APIRequest, view interface{}) bool { return true }
func (p ObjectPermissionFunc) HasObjectPermission(req *APIRequest, view interface{}, obj interface{}) bool {
	if p.Check == nil {
		return true
	}
	return p.Check(req, view, obj)
}
func (p ObjectPermissionFunc) Message() string {
	if p.Detail != "" {
		return p.Detail
	}
	return "permission denied"
}

func checkPermissions(req *APIRequest, view interface{}, permissions []Permission) *APIResponse {
	for _, permission := range permissions {
		if !permission.HasPermission(req, view) {
			return ErrorResponse(permission.Message(), http.StatusForbidden)
		}
	}
	return nil
}

func checkObjectPermissions(req *APIRequest, view interface{}, permissions []Permission, obj interface{}) *APIResponse {
	for _, permission := range permissions {
		if !permission.HasObjectPermission(req, view, obj) {
			return ErrorResponse(permission.Message(), http.StatusForbidden)
		}
	}
	return nil
}

func userIsAuthenticated(user interface{}) bool {
	switch u := user.(type) {
	case interface{ IsAuthenticated() bool }:
		return u.IsAuthenticated()
	case *auth.User:
		return u.IsAuthenticated()
	default:
		return user != nil
	}
}

func userHasPerm(user interface{}, perm string) bool {
	if checker, ok := user.(interface{ HasPerm(string) bool }); ok {
		return checker.HasPerm(perm)
	}
	return false
}

func permissionAction(method string) string {
	switch method {
	case http.MethodGet, http.MethodHead:
		return "view"
	case http.MethodPost:
		return "add"
	case http.MethodPut, http.MethodPatch:
		return "change"
	case http.MethodDelete:
		return "delete"
	default:
		return ""
	}
}
