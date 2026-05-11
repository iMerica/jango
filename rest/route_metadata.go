package rest

import (
	"reflect"
	"sync"

	"github.com/iMerica/jango/orm"
)

type RouteAction struct {
	Method string
	Name   string
}

type RouteMetadata struct {
	Methods        []string
	Actions        []RouteAction
	View           interface{}
	Serializer     interface{}
	ModelMeta      *orm.ModelMeta
	FilterFields   []string
	SearchFields   []string
	OrderingFields []string
	Paginator      interface{}
	Versioning     VersioningStrategy
	Throttled      bool
	AuthRequired   bool
}

var routeMetadata sync.Map

func registerRouteMetadata(handler interface{}, metadata RouteMetadata) {
	key, ok := routeMetadataKey(handler)
	if !ok {
		return
	}
	routeMetadata.Store(key, metadata)
}

func lookupRouteMetadata(handler interface{}) (RouteMetadata, bool) {
	key, ok := routeMetadataKey(handler)
	if !ok {
		return RouteMetadata{}, false
	}
	value, ok := routeMetadata.Load(key)
	if !ok {
		return RouteMetadata{}, false
	}
	metadata, ok := value.(RouteMetadata)
	return metadata, ok
}

func routeMetadataKey(handler interface{}) (uintptr, bool) {
	if handler == nil {
		return 0, false
	}
	value := reflect.ValueOf(handler)
	if value.Kind() != reflect.Func {
		return 0, false
	}
	return value.Pointer(), true
}
