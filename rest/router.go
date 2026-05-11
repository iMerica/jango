package rest

import (
	"fmt"
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

type extraActionRoute struct {
	Name        string
	Methods     []string
	Detail      bool
	HandlerName string
}

func CollectionAction[T any](name string, handler APIHandler, methods ...string) ExtraAction[T] {
	return newExtraAction[T](name, false, handler, methods...)
}

func DetailAction[T any](name string, handler APIHandler, methods ...string) ExtraAction[T] {
	return newExtraAction[T](name, true, handler, methods...)
}

func newExtraAction[T any](name string, detail bool, handler APIHandler, methods ...string) ExtraAction[T] {
	action := ExtraAction[T]{
		Name:        strings.Trim(name, "/"),
		Methods:     methods,
		Detail:      detail,
		HandlerName: strings.Trim(name, "/"),
		Handler:     handler,
	}
	normalized, _ := normalizeExtraAction(action)
	return normalized
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
			collectionActions := map[string]string{
				http.MethodGet:  "list",
				http.MethodPost: "create",
			}
			collectionView := vs.AsView(collectionActions)
			collectionPattern := urls.Path(collection, collectionView, entry.Name+"-list")
			attachRouteMetadata(&collectionPattern, collectionView)
			detailActions := map[string]string{
				http.MethodGet:    "retrieve",
				http.MethodPut:    "update",
				http.MethodPatch:  "partial_update",
				http.MethodDelete: "destroy",
			}
			detailView := vs.AsView(detailActions)
			detailPattern := urls.Path(detail, detailView, entry.Name+"-detail")
			attachRouteMetadata(&detailPattern, detailView)
			patterns = append(patterns, collectionPattern, detailPattern)
			for _, action := range extraActionRoutes(entry.Handler) {
				actionPath := collection + action.Name + "/"
				actionName := entry.Name + "-" + action.Name
				if action.Detail {
					actionPath = detail + action.Name + "/"
					actionName = entry.Name + "-detail-" + action.Name
				}
				actions := make(map[string]string, len(action.Methods))
				for _, method := range action.Methods {
					actions[method] = action.HandlerName
				}
				view := vs.AsView(actions)
				pattern := urls.Path(actionPath, view, actionName)
				attachRouteMetadata(&pattern, view)
				patterns = append(patterns, pattern)
			}
			continue
		}
		patterns = append(patterns,
			urls.Path(collection, entry.Handler, entry.Name+"-list"),
			urls.Path(detail, entry.Handler, entry.Name+"-detail"),
		)
	}
	return patterns
}

func attachRouteMetadata(pattern *urls.Pattern, handler interface{}) {
	if metadata, ok := lookupRouteMetadata(handler); ok {
		pattern.Metadata = metadata
	}
}

func normalizeExtraAction[T any](action ExtraAction[T]) (ExtraAction[T], error) {
	action.Name = strings.Trim(action.Name, "/")
	if action.Name == "" {
		return action, fmt.Errorf("rest: extra action name is required")
	}
	if action.Handler == nil {
		return action, fmt.Errorf("rest: extra action %q handler is required", action.Name)
	}
	if len(action.Methods) == 0 {
		action.Methods = []string{http.MethodGet}
	}
	seen := make(map[string]bool, len(action.Methods))
	methods := make([]string, 0, len(action.Methods))
	for _, method := range action.Methods {
		method = strings.ToUpper(strings.TrimSpace(method))
		if method == "" || seen[method] {
			continue
		}
		seen[method] = true
		methods = append(methods, method)
	}
	if len(methods) == 0 {
		return action, fmt.Errorf("rest: extra action %q must define at least one method", action.Name)
	}
	action.Methods = methods
	action.HandlerName = strings.TrimSpace(action.HandlerName)
	if action.HandlerName == "" {
		action.HandlerName = action.Name
	}
	return action, nil
}

func extraActionRoutes(handler interface{}) []extraActionRoute {
	switch viewset := handler.(type) {
	case interface{ extraActionRoutes() []extraActionRoute }:
		return viewset.extraActionRoutes()
	default:
		return nil
	}
}

func (v ModelViewSet[T]) extraActionRoutes() []extraActionRoute {
	routes := make([]extraActionRoute, 0, len(v.ExtraActions))
	for _, action := range v.ExtraActions {
		normalized, err := normalizeExtraAction(action)
		if err != nil {
			continue
		}
		routes = append(routes, extraActionRoute{
			Name:        normalized.Name,
			Methods:     append([]string(nil), normalized.Methods...),
			Detail:      normalized.Detail,
			HandlerName: normalized.HandlerName,
		})
	}
	return routes
}

type DefaultRouter struct {
	*Router
}

func NewDefaultRouter() *DefaultRouter {
	return &DefaultRouter{Router: NewRouter()}
}
