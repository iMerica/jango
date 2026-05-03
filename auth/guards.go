package auth

import (
	"fmt"
	"strings"

	jangohttp "github.com/iMerica/jango/http"
)

func LoginRequired(viewFn jangohttp.ViewFunc) jangohttp.ViewFunc {
	return func(req *jangohttp.Request) jangohttp.Response {
		if req.User == nil {
			return jangohttp.NewRedirectResponse("/login/")
		}
		user, ok := req.User.(*User)
		if !ok || user.IsAnonymous() {
			return jangohttp.NewRedirectResponse("/login/")
		}
		return viewFn(req)
	}
}

func LoginRequiredAPI(viewFn jangohttp.ViewFunc) jangohttp.ViewFunc {
	return func(req *jangohttp.Request) jangohttp.Response {
		if req.User == nil {
			return jangohttp.NewUnauthorizedResponse("Authentication required")
		}
		user, ok := req.User.(*User)
		if !ok || user.IsAnonymous() {
			return jangohttp.NewUnauthorizedResponse("Authentication required")
		}
		return viewFn(req)
	}
}

func PermissionRequired(perm string, loginURL string) func(jangohttp.ViewFunc) jangohttp.ViewFunc {
	if loginURL == "" {
		loginURL = "/login/"
	}
	return func(viewFn jangohttp.ViewFunc) jangohttp.ViewFunc {
		return func(req *jangohttp.Request) jangohttp.Response {
			user, ok := req.User.(*User)
			if !ok || user.IsAnonymous() {
				return jangohttp.NewRedirectResponse(loginURL)
			}
			if !user.HasPerm(perm) {
				return jangohttp.NewForbiddenResponse("Permission denied")
			}
			return viewFn(req)
		}
	}
}

func PermissionRequiredAPI(perm string) func(jangohttp.ViewFunc) jangohttp.ViewFunc {
	return func(viewFn jangohttp.ViewFunc) jangohttp.ViewFunc {
		return func(req *jangohttp.Request) jangohttp.Response {
			user, ok := req.User.(*User)
			if !ok || user.IsAnonymous() {
				return jangohttp.NewUnauthorizedResponse("Authentication required")
			}
			if !user.HasPerm(perm) {
				return jangohttp.NewForbiddenResponse("Permission denied")
			}
			return viewFn(req)
		}
	}
}

func CreatePerModelPermissions(appLabel, modelName string) []string {
	actions := []string{"add", "change", "delete", "view"}
	var perms []string
	for _, action := range actions {
		perms = append(perms, fmt.Sprintf("%s.%s_%s", appLabel, action, strings.ToLower(modelName)))
	}
	return perms
}
