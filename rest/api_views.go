package rest

import (
	"errors"
	"fmt"
	"net/http"
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
}

func (v APIView) AsView(handlers map[string]APIHandler) jangohttp.ViewFunc {
	return func(req *jangohttp.Request) jangohttp.Response {
		apiReq := NewAPIRequest(req)
		return v.Dispatch(apiReq, v, handlers)
	}
}

func (v APIView) Dispatch(req *APIRequest, view interface{}, handlers map[string]APIHandler) jangohttp.Response {
	v = v.withDefaults()
	if err := parseRequestData(req, v.Parsers); err != nil {
		return BadRequest(err.Error())
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
	handler := handlers[req.Method]
	if handler == nil && req.Method == http.MethodHead {
		handler = handlers[http.MethodGet]
	}
	if handler == nil {
		allowed := make([]string, 0, len(handlers))
		for method := range handlers {
			allowed = append(allowed, method)
		}
		allowed = append(allowed, http.MethodOptions)
		return v.finalize(req, MethodNotAllowed(req.Method, allowed...))
	}
	return v.finalize(req, handler(req))
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
		}
	}
	handlers[http.MethodOptions] = v.Options
	api := v.APIView
	return func(req *jangohttp.Request) jangohttp.Response {
		apiReq := NewAPIRequest(req)
		return api.Dispatch(apiReq, v, handlers)
	}
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
	meta := v.Serializer.ModelMeta()
	data := SimpleMetadata{}.DetermineMetadata(req, v, meta)
	resp := NewAPIResponse(data, http.StatusOK)
	resp.SetHeader("Allow", "GET, POST, PUT, PATCH, DELETE, HEAD, OPTIONS")
	return resp
}

func modelMetaFromView(view interface{}) *orm.ModelMeta {
	if getter, ok := view.(interface{ getMeta() *orm.ModelMeta }); ok {
		return getter.getMeta()
	}
	return nil
}
