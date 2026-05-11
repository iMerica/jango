package rest

import (
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strings"

	jangohttp "github.com/iMerica/jango/http"
	"github.com/iMerica/jango/orm"
)

type APIHandler func(*APIRequest) jangohttp.Response

type Throttle interface {
	AllowRequest(req *APIRequest, view interface{}) bool
	Wait(req *APIRequest, view interface{}) int
}

type APIView struct {
	Authenticators []Authenticator
	Permissions    []Permission
	Throttles      []Throttle
	Parsers        []Parser
	Renderers      []Renderer
	Negotiator     ContentNegotiator
	Versioning     VersioningStrategy
	ThrottleScope  string
}

func (v APIView) AsView(handlers map[string]APIHandler) jangohttp.ViewFunc {
	if handlers == nil {
		handlers = make(map[string]APIHandler)
	}
	if handlers[http.MethodOptions] == nil && handlers[strings.ToLower(http.MethodOptions)] == nil {
		handlers[http.MethodOptions] = func(req *APIRequest) jangohttp.Response {
			resp := v.Options(req)
			if apiResp, ok := resp.(*APIResponse); ok {
				apiResp.SetHeader("Allow", joinMethods(allowedMethodsFromHandlers(normalizeHandlers(handlers))))
			}
			return resp
		}
	}
	normalized := normalizeHandlers(handlers)
	view := func(req *jangohttp.Request) jangohttp.Response {
		apiReq := NewAPIRequest(req)
		return v.Dispatch(apiReq, v, normalized)
	}
	registerRouteMetadata(view, RouteMetadata{
		Methods:      allowedMethodsFromHandlers(normalized),
		View:         v,
		Versioning:   v.Versioning,
		Throttled:    len(v.Throttles) > 0,
		AuthRequired: len(v.Authenticators) > 0,
	})
	return view
}

func (v APIView) Options(req *APIRequest) jangohttp.Response {
	return NewAPIResponse(SimpleMetadata{}.DetermineMetadata(req, v, nil), http.StatusOK)
}

func (v APIView) Dispatch(req *APIRequest, view interface{}, handlers map[string]APIHandler) jangohttp.Response {
	v = v.withDefaults()
	handlers = normalizeHandlers(handlers)
	if err := parseRequestData(req, v.Parsers); err != nil {
		return BadRequest(err.Error())
	}
	if v.Versioning != nil {
		version, err := v.Versioning.DetermineVersion(req, view)
		if err != nil {
			return BadRequest(err.Error())
		}
		req.Version = version
	}
	if authResp := authenticateRequest(req, v.Authenticators); authResp != nil {
		return v.finalize(req, authResp)
	}
	if permissionResp := checkPermissions(req, view, v.Permissions); permissionResp != nil {
		return v.finalize(req, permissionResp)
	}
	for _, throttle := range v.Throttles {
		if !throttle.AllowRequest(req, view) {
			resp := ErrorResponse("request was throttled", http.StatusTooManyRequests)
			if wait := throttle.Wait(req, view); wait > 0 {
				resp.SetHeader("Retry-After", fmt.Sprintf("%d", wait))
			}
			return v.finalize(req, resp)
		}
	}
	renderer, format, err := v.Negotiator.SelectRenderer(req, v.Renderers)
	if err != nil {
		return ErrorResponse(err.Error(), http.StatusNotAcceptable)
	}
	req.Format = format
	req.AcceptedRenderer = renderer
	method := strings.ToUpper(req.Method)
	handler := handlers[method]
	if handler == nil && req.Method == http.MethodHead {
		handler = handlers[http.MethodGet]
	}
	if handler == nil {
		allowed := allowedMethodsFromHandlers(handlers)
		return v.finalize(req, MethodNotAllowed(req.Method, allowed...))
	}
	return v.finalize(req, handler(req))
}

func normalizeHandlers(handlers map[string]APIHandler) map[string]APIHandler {
	normalized := make(map[string]APIHandler, len(handlers))
	for method, handler := range handlers {
		if handler == nil {
			continue
		}
		method = strings.ToUpper(strings.TrimSpace(method))
		if method == "" {
			continue
		}
		normalized[method] = handler
	}
	return normalized
}

func allowedMethodsFromHandlers(handlers map[string]APIHandler) []string {
	seen := make(map[string]bool, len(handlers)+2)
	for method := range handlers {
		seen[strings.ToUpper(method)] = true
	}
	if seen[http.MethodGet] {
		seen[http.MethodHead] = true
	}
	seen[http.MethodOptions] = true

	methods := make([]string, 0, len(seen))
	for method := range seen {
		methods = append(methods, method)
	}
	sort.Slice(methods, func(i, j int) bool {
		return methodOrder(methods[i]) < methodOrder(methods[j])
	})
	return methods
}

func methodOrder(method string) int {
	switch method {
	case http.MethodGet:
		return 0
	case http.MethodPost:
		return 1
	case http.MethodPut:
		return 2
	case http.MethodPatch:
		return 3
	case http.MethodDelete:
		return 4
	case http.MethodHead:
		return 5
	case http.MethodOptions:
		return 6
	default:
		if method == "" {
			return 1000
		}
		return 100 + int(method[0])
	}
}

func (v APIView) finalize(req *APIRequest, resp jangohttp.Response) jangohttp.Response {
	apiResp, ok := resp.(*APIResponse)
	if !ok {
		return resp
	}
	if renderer, ok := req.AcceptedRenderer.(Renderer); ok && apiResp.Renderer == nil {
		apiResp.Renderer = renderer
	} else if renderer, ok := req.AcceptedRenderer.(Renderer); ok {
		apiResp.Renderer = renderer
	}
	return apiResp
}

func (v APIView) withDefaults() APIView {
	if len(v.Parsers) == 0 {
		v.Parsers = []Parser{JSONParser{}, FormParser{}}
	}
	if len(v.Renderers) == 0 {
		v.Renderers = []Renderer{JSONRenderer{}, BrowsableAPIRenderer{}}
	}
	if v.Negotiator == nil {
		v.Negotiator = DefaultContentNegotiator{}
	}
	if len(v.Permissions) == 0 {
		v.Permissions = []Permission{AllowAny{}}
	}
	return v
}

type GenericAPIView[T any] struct {
	APIView
	QuerySet        *orm.QuerySet[T]
	Serializer      Serializer[T]
	LookupField     string
	LookupURLParam  string
	FilterFields    []string
	SearchFields    []string
	OrderingFields  []string
	Paginator       Paginator[T]
	DefaultPageSize int
}

func (v GenericAPIView[T]) getMeta() *orm.ModelMeta {
	if v.Serializer == nil {
		return nil
	}
	return v.Serializer.ModelMeta()
}

func (v GenericAPIView[T]) getObject(req *APIRequest) (*T, jangohttp.Response) {
	if v.QuerySet == nil {
		return nil, InternalServerError("queryset is required")
	}
	meta := v.getMeta()
	if meta == nil {
		return nil, InternalServerError("serializer model metadata is required")
	}
	lookupField := v.LookupField
	if lookupField == "" {
		lookupField = meta.PKField
	}
	field, ok := meta.FieldForNameOrColumn(lookupField)
	if !ok || field.FieldType == orm.ManyToManyType {
		return nil, InternalServerError(fmt.Sprintf("unknown lookup field %q", lookupField))
	}
	lookupURLParam := v.LookupURLParam
	if lookupURLParam == "" {
		lookupURLParam = meta.DBColumnForField(field.Name)
	}
	raw := req.URLParam(lookupURLParam)
	if raw == "" {
		return nil, BadRequest(fmt.Sprintf("missing URL parameter %q", lookupURLParam))
	}
	value, err := coerceQueryValue(field, "exact", raw)
	if err != nil {
		return nil, BadRequest(err.Error())
	}
	record, err := v.QuerySet.Get(req.Context(), orm.L(meta.DBColumnForField(field.Name), value))
	if err != nil {
		var missing *orm.DoesNotExist
		if errors.As(err, &missing) {
			return nil, NotFound("not found")
		}
		return nil, InternalServerError(err.Error())
	}
	if permissionResp := checkObjectPermissions(req, v, v.Permissions, record); permissionResp != nil {
		return nil, permissionResp
	}
	return record, nil
}

func (v GenericAPIView[T]) filteredQuerySet(req *APIRequest) (*orm.QuerySet[T], jangohttp.Response) {
	if v.QuerySet == nil {
		return nil, InternalServerError("queryset is required")
	}
	meta := v.getMeta()
	if meta == nil {
		return nil, InternalServerError("serializer model metadata is required")
	}
	listView := ListAPIView[T]{
		QuerySet:        v.QuerySet,
		Serializer:      v.Serializer,
		FilterFields:    v.FilterFields,
		SearchFields:    v.SearchFields,
		OrderingFields:  v.OrderingFields,
		DefaultPageSize: v.DefaultPageSize,
	}
	qs, err := listView.applyFilters(v.QuerySet, meta, req.QueryParams())
	if err != nil {
		return nil, BadRequest(err.Error())
	}
	qs, err = listView.applySearch(qs, meta, req.QueryParams())
	if err != nil {
		return nil, BadRequest(err.Error())
	}
	qs, err = listView.applyOrdering(qs, meta, req.QueryParams())
	if err != nil {
		return nil, BadRequest(err.Error())
	}
	return qs, nil
}

type ModelViewSet[T any] struct {
	GenericAPIView[T]
	ExtraActions []ExtraAction[T]
}

func (v ModelViewSet[T]) AsView(actions map[string]string) jangohttp.ViewFunc {
	handlers := make(map[string]APIHandler)
	for method, action := range actions {
		method = strings.ToUpper(method)
		switch strings.ToLower(action) {
		case "list":
			handlers[method] = v.List
		case "create":
			handlers[method] = v.Create
		case "retrieve":
			handlers[method] = v.Retrieve
		case "update":
			handlers[method] = v.Update
		case "partial_update":
			handlers[method] = v.PartialUpdate
		case "destroy":
			handlers[method] = v.Destroy
		default:
			if extra, ok := v.extraActionByHandlerName(action); ok {
				handlers[method] = extra.Handler
			}
		}
	}
	handlers[http.MethodOptions] = func(req *APIRequest) jangohttp.Response {
		return v.optionsWithAllowed(req, allowedMethodsFromHandlers(handlers))
	}
	api := v.APIView
	normalized := normalizeHandlers(handlers)
	view := func(req *jangohttp.Request) jangohttp.Response {
		apiReq := NewAPIRequest(req)
		return api.Dispatch(apiReq, v, normalized)
	}
	registerRouteMetadata(view, v.routeMetadata(actions, allowedMethodsFromHandlers(normalized)))
	return view
}

func (v ModelViewSet[T]) List(req *APIRequest) jangohttp.Response {
	if v.Serializer == nil {
		return InternalServerError("serializer is required")
	}
	qs, resp := v.filteredQuerySet(req)
	if resp != nil {
		return resp
	}
	meta := v.Serializer.ModelMeta()
	paginator := v.Paginator
	if paginator == nil {
		paginator = LimitOffsetPagination[T]{DefaultLimit: v.DefaultPageSize}
	}
	paged, pageData, err := paginator.Paginate(qs, req, meta)
	if err != nil {
		return BadRequest(err.Error())
	}
	records, err := paged.AllRecords(req.Context())
	if err != nil {
		return InternalServerError(err.Error())
	}
	results, err := v.Serializer.SerializeList(records)
	if err != nil {
		return InternalServerError(err.Error())
	}
	pageData["results"] = results
	return NewAPIResponse(pageData, http.StatusOK)
}

func (v ModelViewSet[T]) Retrieve(req *APIRequest) jangohttp.Response {
	if v.Serializer == nil {
		return InternalServerError("serializer is required")
	}
	record, resp := v.getObject(req)
	if resp != nil {
		return resp
	}
	data, err := v.Serializer.Serialize(record)
	if err != nil {
		return InternalServerError(err.Error())
	}
	return NewAPIResponse(data, http.StatusOK)
}

func (v ModelViewSet[T]) Create(req *APIRequest) jangohttp.Response {
	if v.Serializer == nil {
		return InternalServerError("serializer is required")
	}
	serializer := v.Serializer
	serializer.SetContext(SerializerContext{Request: req, View: v, Format: req.Format})
	if err := serializer.Bind(req.Data); err != nil {
		return NewAPIResponse(serializer.Errors(), http.StatusBadRequest)
	}
	record, err := serializer.Create(req.Context(), v.QuerySet)
	if err != nil {
		return InternalServerError(err.Error())
	}
	data, err := serializer.Serialize(record)
	if err != nil {
		return InternalServerError(err.Error())
	}
	return NewAPIResponse(data, http.StatusCreated)
}

func (v ModelViewSet[T]) Update(req *APIRequest) jangohttp.Response {
	return v.update(req, false)
}

func (v ModelViewSet[T]) PartialUpdate(req *APIRequest) jangohttp.Response {
	return v.update(req, true)
}

func (v ModelViewSet[T]) update(req *APIRequest, partial bool) jangohttp.Response {
	if v.Serializer == nil {
		return InternalServerError("serializer is required")
	}
	record, resp := v.getObject(req)
	if resp != nil {
		return resp
	}
	serializer := v.Serializer
	serializer.SetContext(SerializerContext{Request: req, View: v, Format: req.Format})
	var err error
	if partial {
		err = serializer.BindPartial(req.Data)
	} else {
		err = serializer.Bind(req.Data)
	}
	if err != nil {
		return NewAPIResponse(serializer.Errors(), http.StatusBadRequest)
	}
	updated, err := serializer.Update(req.Context(), v.QuerySet, record)
	if err != nil {
		return InternalServerError(err.Error())
	}
	data, err := serializer.Serialize(updated)
	if err != nil {
		return InternalServerError(err.Error())
	}
	return NewAPIResponse(data, http.StatusOK)
}

func (v ModelViewSet[T]) Destroy(req *APIRequest) jangohttp.Response {
	record, resp := v.getObject(req)
	if resp != nil {
		return resp
	}
	meta := v.Serializer.ModelMeta()
	pk, ok := readPKValue(record, meta)
	if !ok {
		return InternalServerError("primary key is required")
	}
	affected, err := v.QuerySet.Filter(orm.L(meta.PKColumn(), pk)).BaseQuerySet.Delete(req.Context())
	if err != nil {
		return InternalServerError(err.Error())
	}
	if affected == 0 {
		return NotFound("not found")
	}
	return NewAPIResponse(nil, http.StatusNoContent)
}

func (v ModelViewSet[T]) Options(req *APIRequest) jangohttp.Response {
	return v.optionsWithAllowed(req, []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"})
}

func (v ModelViewSet[T]) optionsWithAllowed(req *APIRequest, allowed []string) jangohttp.Response {
	var meta *orm.ModelMeta
	if v.Serializer != nil {
		meta = v.Serializer.ModelMeta()
	}
	data := SimpleMetadata{}.DetermineMetadata(req, v, meta)
	resp := NewAPIResponse(data, http.StatusOK)
	resp.SetHeader("Allow", joinMethods(allowed))
	return resp
}

func modelMetaFromView(view interface{}) *orm.ModelMeta {
	if getter, ok := view.(interface{ getMeta() *orm.ModelMeta }); ok {
		return getter.getMeta()
	}
	return nil
}

func (v ModelViewSet[T]) extraActionByHandlerName(name string) (ExtraAction[T], bool) {
	for _, action := range v.ExtraActions {
		normalized, err := normalizeExtraAction(action)
		if err != nil {
			continue
		}
		if strings.EqualFold(normalized.HandlerName, name) || strings.EqualFold(normalized.Name, name) {
			return normalized, true
		}
	}
	return ExtraAction[T]{}, false
}

func (v ModelViewSet[T]) routeMetadata(actions map[string]string, methods []string) RouteMetadata {
	meta := RouteMetadata{
		Methods:        methods,
		View:           v,
		Serializer:     v.Serializer,
		ModelMeta:      v.getMeta(),
		FilterFields:   append([]string(nil), v.FilterFields...),
		SearchFields:   append([]string(nil), v.SearchFields...),
		OrderingFields: append([]string(nil), v.OrderingFields...),
		Paginator:      v.Paginator,
		Versioning:     v.Versioning,
		Throttled:      len(v.Throttles) > 0,
		AuthRequired:   len(v.Authenticators) > 0,
	}
	for method, action := range actions {
		meta.Actions = append(meta.Actions, RouteAction{
			Method: strings.ToUpper(method),
			Name:   strings.ToLower(action),
		})
	}
	return meta
}
