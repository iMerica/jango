package rest

import (
	"net/http"
	"strings"

	jangohttp "github.com/iMerica/jango/http"
	"github.com/iMerica/jango/urls"
)

type ViewSet interface {
	AsView(actions map[string]string) jangohttp.ViewFunc
}

type ExtraAction[T any] struct {
	Name        string
	Methods     []string
	Detail      bool
	HandlerName string
	Handler     APIHandler
}

type Router struct {
	prefixes []routerEntry
}

type routerEntry struct {
	Prefix  string
	Name    string
	Handler interface{}
}

func NewRouter() *Router {
	return &Router{}
}

func (r *Router) Register(prefix, basename string, handler interface{}) {
	r.prefixes = append(r.prefixes, routerEntry{
		Prefix:  strings.Trim(prefix, "/"),
		Name:    basename,
		Handler: handler,
	})
}

func (r *Router) URLPatterns() []urls.Pattern {
	patterns := make([]urls.Pattern, 0, len(r.prefixes)*2)
	for _, entry := range r.prefixes {
		collection := "/" + entry.Prefix + "/"
		detail := "/" + entry.Prefix + "/<int:id>/"
		if vs, ok := entry.Handler.(ViewSet); ok {
			patterns = append(patterns,
				urls.Path(collection, vs.AsView(map[string]string{
					http.MethodGet:  "list",
					http.MethodPost: "create",
				}), entry.Name+"-list"),
				urls.Path(detail, vs.AsView(map[string]string{
					http.MethodGet:    "retrieve",
					http.MethodPut:    "update",
					http.MethodPatch:  "partial_update",
					http.MethodDelete: "destroy",
				}), entry.Name+"-detail"),
			)
			continue
		}
		patterns = append(patterns,
			urls.Path(collection, entry.Handler, entry.Name+"-list"),
			urls.Path(detail, entry.Handler, entry.Name+"-detail"),
		)
	}
	return patterns
}

type DefaultRouter struct {
	*Router
}

func NewDefaultRouter() *DefaultRouter {
	return &DefaultRouter{Router: NewRouter()}
}
