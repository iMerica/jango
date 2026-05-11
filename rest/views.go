package rest

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	jangohttp "github.com/iMerica/jango/http"
	"github.com/iMerica/jango/orm"
)

const (
	DefaultLimit = 100
	MaxLimit     = 1000
)

var readonlyMethods = []string{http.MethodGet, http.MethodHead, http.MethodOptions}

type ListAPIView[T any] struct {
	QuerySet        *orm.QuerySet[T]
	Serializer      Serializer[T]
	FilterFields    []string
	SearchFields    []string
	OrderingFields  []string
	DefaultPageSize int
}

type DetailAPIView[T any] struct {
	QuerySet       *orm.QuerySet[T]
	Serializer     Serializer[T]
	LookupField    string
	LookupURLParam string
}

func (v ListAPIView[T]) AsView() jangohttp.ViewFunc {
	return func(req *jangohttp.Request) jangohttp.Response {
		switch req.Method {
		case http.MethodGet, http.MethodHead:
			return v.DispatchGet(NewAPIRequest(req))
		case http.MethodOptions:
			return optionsResponse()
		default:
			return MethodNotAllowed(req.Method, readonlyMethods...)
		}
	}
}

func (v DetailAPIView[T]) AsView() jangohttp.ViewFunc {
	return func(req *jangohttp.Request) jangohttp.Response {
		switch req.Method {
		case http.MethodGet, http.MethodHead:
			return v.DispatchGet(NewAPIRequest(req))
		case http.MethodOptions:
			return optionsResponse()
		default:
			return MethodNotAllowed(req.Method, readonlyMethods...)
		}
	}
}

func (v ListAPIView[T]) DispatchGet(req *APIRequest) jangohttp.Response {
	if v.QuerySet == nil {
		return InternalServerError("queryset is required")
	}
	if v.Serializer == nil {
		return InternalServerError("serializer is required")
	}
	meta := v.Serializer.ModelMeta()
	if meta == nil {
		return InternalServerError("serializer model metadata is required")
	}

	qs, err := v.applyFilters(v.QuerySet, meta, req.QueryParams())
	if err != nil {
		return BadRequest(err.Error())
	}
	qs, err = v.applySearch(qs, meta, req.QueryParams())
	if err != nil {
		return BadRequest(err.Error())
	}
	qs, err = v.applyOrdering(qs, meta, req.QueryParams())
	if err != nil {
		return BadRequest(err.Error())
	}
	limit, offset, err := pagination(v.DefaultPageSize, req.QueryParams())
	if err != nil {
		return BadRequest(err.Error())
	}

	count, err := qs.Count(req.Context())
	if err != nil {
		return InternalServerError(err.Error())
	}
	records, err := qs.Limit(limit).Offset(offset).AllRecords(req.Context())
	if err != nil {
		return InternalServerError(err.Error())
	}
	results, err := v.Serializer.SerializeList(records)
	if err != nil {
		return InternalServerError(err.Error())
	}

	return NewAPIResponse(map[string]interface{}{
		"count":   count,
		"limit":   limit,
		"offset":  offset,
		"results": results,
	}, http.StatusOK)
}

func (v ListAPIView[T]) applySearch(qs *orm.QuerySet[T], meta *orm.ModelMeta, values url.Values) (*orm.QuerySet[T], error) {
	term := strings.TrimSpace(values.Get("search"))
	if term == "" {
		return qs, nil
	}
	if len(v.SearchFields) == 0 {
		return qs, nil
	}
	children := make([]orm.QNode, 0, len(v.SearchFields))
	for _, name := range v.SearchFields {
		field, ok := meta.FieldForNameOrColumn(name)
		if !ok || field.FieldType == orm.ManyToManyType {
			return nil, fmt.Errorf("unsupported search field %q", name)
		}
		column := meta.DBColumnForField(field.Name)
		children = append(children, orm.Q(orm.L(column+"__icontains", term)))
	}
	return qs.FilterQ(orm.QOr(children...)), nil
}

func (v DetailAPIView[T]) DispatchGet(req *APIRequest) jangohttp.Response {
	if v.QuerySet == nil {
		return InternalServerError("queryset is required")
	}
	if v.Serializer == nil {
		return InternalServerError("serializer is required")
	}
	meta := v.Serializer.ModelMeta()
	if meta == nil {
		return InternalServerError("serializer model metadata is required")
	}
	lookupField := v.LookupField
	if lookupField == "" {
		lookupField = meta.PKField
	}
	if lookupField == "" {
		lookupField = "id"
	}
	field, ok := meta.FieldForNameOrColumn(lookupField)
	if !ok || field.FieldType == orm.ManyToManyType {
		return InternalServerError(fmt.Sprintf("unknown lookup field %q", lookupField))
	}
	lookupURLParam := v.LookupURLParam
	if lookupURLParam == "" {
		lookupURLParam = meta.DBColumnForField(field.Name)
	}
	raw := req.URLParam(lookupURLParam)
	if raw == "" {
		return BadRequest(fmt.Sprintf("missing URL parameter %q", lookupURLParam))
	}
	value, err := coerceQueryValue(field, "exact", raw)
	if err != nil {
		return BadRequest(err.Error())
	}

	record, err := v.QuerySet.Get(req.Context(), orm.L(meta.DBColumnForField(field.Name), value))
	if err != nil {
		var missing *orm.DoesNotExist
		if errors.As(err, &missing) {
			return NotFound("not found")
		}
		return InternalServerError(err.Error())
	}
	serialized, err := v.Serializer.Serialize(record)
	if err != nil {
		return InternalServerError(err.Error())
	}
	return NewAPIResponse(serialized, http.StatusOK)
}

func (v ListAPIView[T]) applyFilters(qs *orm.QuerySet[T], meta *orm.ModelMeta, values url.Values) (*orm.QuerySet[T], error) {
	allowed, err := allowedFieldSet(meta, v.FilterFields)
	if err != nil {
		return nil, err
	}
	for name, rawValues := range values {
		if isReservedQueryParam(name) {
			continue
		}
		baseName, op, err := splitLookupName(name)
		if err != nil {
			return nil, err
		}
		field, ok := meta.FieldForNameOrColumn(baseName)
		if !ok || field.FieldType == orm.ManyToManyType {
			return nil, fmt.Errorf("unsupported filter field %q", name)
		}
		column := meta.DBColumnForField(field.Name)
		if !allowed[column] && !allowed[name] {
			return nil, fmt.Errorf("unsupported filter field %q", name)
		}
		raw := ""
		if len(rawValues) > 0 {
			raw = rawValues[0]
		}
		value, err := coerceQueryValue(field, op, raw)
		if err != nil {
			return nil, err
		}
		lookup := column
		if op != "exact" {
			lookup += "__" + op
		}
		qs = qs.Filter(orm.L(lookup, value))
	}
	return qs, nil
}

func (v ListAPIView[T]) applyOrdering(qs *orm.QuerySet[T], meta *orm.ModelMeta, values url.Values) (*orm.QuerySet[T], error) {
	ordering := strings.TrimSpace(values.Get("ordering"))
	if ordering == "" {
		return qs, nil
	}
	allowed, err := allowedFieldSet(meta, v.OrderingFields)
	if err != nil {
		return nil, err
	}
	var fields []string
	for _, part := range strings.Split(ordering, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			return nil, fmt.Errorf("ordering contains an empty field")
		}
		desc := strings.HasPrefix(part, "-")
		name := strings.TrimPrefix(part, "-")
		field, ok := meta.FieldForNameOrColumn(name)
		if !ok || field.FieldType == orm.ManyToManyType {
			return nil, fmt.Errorf("unsupported ordering field %q", name)
		}
		column := meta.DBColumnForField(field.Name)
		if !allowed[column] {
			return nil, fmt.Errorf("unsupported ordering field %q", name)
		}
		if desc {
			fields = append(fields, "-"+column)
		} else {
			fields = append(fields, column)
		}
	}
	return qs.OrderBy(fields...), nil
}

func allowedFieldSet(meta *orm.ModelMeta, fields []string) (map[string]bool, error) {
	allowed := make(map[string]bool)
	if len(fields) == 0 {
		for _, field := range meta.ConcreteFields() {
			if field.FieldType == orm.ManyToManyType {
				continue
			}
			allowed[meta.DBColumnForField(field.Name)] = true
		}
		return allowed, nil
	}
	for _, name := range fields {
		baseName, _, err := splitLookupName(name)
		if err != nil {
			return nil, err
		}
		field, ok := meta.FieldForNameOrColumn(baseName)
		if !ok || field.FieldType == orm.ManyToManyType {
			return nil, fmt.Errorf("unknown allowed field %q", name)
		}
		allowed[meta.DBColumnForField(field.Name)] = true
		allowed[name] = true
	}
	return allowed, nil
}

func pagination(defaultPageSize int, values url.Values) (int, int, error) {
	limit := defaultPageSize
	if limit <= 0 {
		limit = DefaultLimit
	}
	if raw := values.Get("limit"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil {
			return 0, 0, fmt.Errorf("limit must be an integer")
		}
		limit = parsed
	}
	if limit < 0 {
		return 0, 0, fmt.Errorf("limit must be greater than or equal to 0")
	}
	if limit > MaxLimit {
		return 0, 0, fmt.Errorf("limit must be less than or equal to %d", MaxLimit)
	}

	offset := 0
	if raw := values.Get("offset"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil {
			return 0, 0, fmt.Errorf("offset must be an integer")
		}
		offset = parsed
	}
	if offset < 0 {
		return 0, 0, fmt.Errorf("offset must be greater than or equal to 0")
	}
	return limit, offset, nil
}

func splitLookupName(name string) (string, string, error) {
	for _, lookup := range supportedLookupSuffixes {
		suffix := "__" + lookup
		if strings.HasSuffix(name, suffix) && len(name) > len(suffix) {
			return strings.TrimSuffix(name, suffix), lookup, nil
		}
	}
	if strings.Contains(name, "__") {
		return "", "", fmt.Errorf("unsupported lookup %q", name)
	}
	return name, "exact", nil
}

var supportedLookupSuffixes = []string{
	"exact", "iexact", "contains", "icontains", "startswith", "istartswith",
	"endswith", "iendswith", "regex", "iregex", "gt", "gte", "lt", "lte",
	"in", "isnull", "range", "year", "month", "day", "hour",
	"minute", "second", "date", "search",
}

func coerceQueryValue(field orm.FieldDef, op string, raw string) (interface{}, error) {
	switch op {
	case "isnull":
		value, err := strconv.ParseBool(raw)
		if err != nil {
			return nil, fmt.Errorf("%s__isnull must be a boolean", field.DBColumn)
		}
		return value, nil
	case "in":
		if raw == "" {
			return []interface{}{}, nil
		}
		parts := strings.Split(raw, ",")
		values := make([]interface{}, 0, len(parts))
		for _, part := range parts {
			value, err := coerceQueryValue(field, "exact", strings.TrimSpace(part))
			if err != nil {
				return nil, err
			}
			values = append(values, value)
		}
		return values, nil
	case "range":
		parts := strings.Split(raw, ",")
		if len(parts) != 2 {
			return nil, fmt.Errorf("%s__range must contain two comma-separated values", field.DBColumn)
		}
		start, err := coerceQueryValue(field, "exact", strings.TrimSpace(parts[0]))
		if err != nil {
			return nil, err
		}
		end, err := coerceQueryValue(field, "exact", strings.TrimSpace(parts[1]))
		if err != nil {
			return nil, err
		}
		return [2]interface{}{start, end}, nil
	case "year", "month", "day", "hour", "minute", "second":
		value, err := strconv.Atoi(raw)
		if err != nil {
			return nil, fmt.Errorf("%s__%s must be an integer", field.DBColumn, op)
		}
		return value, nil
	}

	switch field.FieldType {
	case orm.AutoFieldType, orm.BigAutoFieldType, orm.SmallAutoFieldType,
		orm.IntFieldType, orm.BigIntFieldType, orm.SmallIntFieldType,
		orm.PositiveIntFieldType, orm.PositiveBigIntFieldType, orm.PositiveSmallIntFieldType,
		orm.ForeignKeyType, orm.OneToOneType:
		value, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("%s must be an integer", field.DBColumn)
		}
		return value, nil
	case orm.BooleanFieldType, orm.NullBooleanFieldType:
		value, err := strconv.ParseBool(raw)
		if err != nil {
			return nil, fmt.Errorf("%s must be a boolean", field.DBColumn)
		}
		return value, nil
	case orm.FloatFieldType, orm.DoubleFieldType, orm.DecimalFieldType:
		value, err := strconv.ParseFloat(raw, 64)
		if err != nil {
			return nil, fmt.Errorf("%s must be a number", field.DBColumn)
		}
		return value, nil
	case orm.DateTimeFieldType:
		value, err := time.Parse(time.RFC3339, raw)
		if err != nil {
			return nil, fmt.Errorf("%s must be an RFC3339 timestamp", field.DBColumn)
		}
		return value, nil
	case orm.DateFieldType:
		value, err := time.Parse("2006-01-02", raw)
		if err != nil {
			return nil, fmt.Errorf("%s must be a date in YYYY-MM-DD format", field.DBColumn)
		}
		return value, nil
	case orm.TimeFieldType:
		value, err := time.Parse("15:04:05", raw)
		if err != nil {
			return nil, fmt.Errorf("%s must be a time in HH:MM:SS format", field.DBColumn)
		}
		return value, nil
	default:
		return raw, nil
	}
}

func isReservedQueryParam(name string) bool {
	return name == "ordering" || name == "limit" || name == "offset" || name == "search" || name == "page" || name == "cursor" || name == "version"
}

func optionsResponse() *APIResponse {
	resp := NewAPIResponse(map[string]interface{}{"allowed_methods": readonlyMethods}, http.StatusOK)
	resp.SetHeader("Allow", joinMethods(readonlyMethods))
	return resp
}
