package admin

import (
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"github.com/iMerica/jango/auth"
	jangohttp "github.com/iMerica/jango/http"
	"github.com/iMerica/jango/orm"
)

type ModelAdmin struct {
	Model          *orm.ModelMeta
	ListDisplay    []string
	SearchFields   []string
	ReadOnlyFields []string
	Actions        map[string]Action
}

type Action func(req *jangohttp.Request, meta *orm.ModelMeta, ids []string) error

type Site struct {
	Name     string
	registry map[string]*ModelAdmin
}

func NewSite(name string) *Site {
	if name == "" {
		name = "admin"
	}
	return &Site{Name: name, registry: make(map[string]*ModelAdmin)}
}

var DefaultSite = NewSite("admin")

func (s *Site) Register(meta *orm.ModelMeta, admin *ModelAdmin) error {
	if meta == nil {
		return fmt.Errorf("admin: model metadata is required")
	}
	if admin == nil {
		admin = &ModelAdmin{}
	}
	admin.Model = meta
	s.registry[key(meta)] = admin
	return nil
}

func (s *Site) Unregister(meta *orm.ModelMeta) {
	delete(s.registry, key(meta))
}

func (s *Site) Get(meta *orm.ModelMeta) (*ModelAdmin, bool) {
	admin, ok := s.registry[key(meta)]
	return admin, ok
}

func (s *Site) Models() []*ModelAdmin {
	models := make([]*ModelAdmin, 0, len(s.registry))
	for _, admin := range s.registry {
		models = append(models, admin)
	}
	return models
}

func (s *Site) Index(req *jangohttp.Request) jangohttp.Response {
	if !hasAdminAccess(req.User) {
		return jangohttp.NewTextResponseWithStatus("permission denied", http.StatusForbidden)
	}
	var b strings.Builder
	b.WriteString("<h1>JanGO administration</h1><ul>")
	for _, admin := range s.Models() {
		label := template.HTMLEscapeString(admin.Model.AppLabel + "." + admin.Model.ModelName)
		b.WriteString("<li>" + label + "</li>")
	}
	b.WriteString("</ul>")
	return jangohttp.NewHTMLResponse(b.String())
}

func (s *Site) ModelList(req *jangohttp.Request, meta *orm.ModelMeta) jangohttp.Response {
	if !hasAdminAccess(req.User) || !hasModelPerm(req.User, meta, "view") {
		return jangohttp.NewTextResponseWithStatus("permission denied", http.StatusForbidden)
	}
	return jangohttp.NewHTMLResponse("<h1>" + template.HTMLEscapeString(meta.ModelName) + "</h1>")
}

func hasAdminAccess(user interface{}) bool {
	switch u := user.(type) {
	case *auth.User:
		return u.IsAuthenticated() && (u.IsStaff || u.IsAdmin || u.IsSuperuser)
	case interface{ IsAuthenticated() bool }:
		return u.IsAuthenticated()
	default:
		return false
	}
}

func hasModelPerm(user interface{}, meta *orm.ModelMeta, action string) bool {
	if checker, ok := user.(interface{ HasPerm(string) bool }); ok {
		return checker.HasPerm(fmt.Sprintf("%s.%s_%s", meta.AppLabel, action, strings.ToLower(meta.ModelName)))
	}
	return hasAdminAccess(user)
}

func key(meta *orm.ModelMeta) string {
	if meta == nil {
		return ""
	}
	return meta.AppLabel + "." + meta.ModelName
}
