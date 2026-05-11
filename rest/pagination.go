package rest

import (
	"encoding/base64"
	"fmt"
	"net/url"
	"strconv"

	"github.com/iMerica/jango/orm"
)

type Paginator[T any] interface {
	Paginate(qs *orm.QuerySet[T], req *APIRequest, meta *orm.ModelMeta) (*orm.QuerySet[T], map[string]interface{}, error)
}

type PageNumberPagination[T any] struct {
	PageSize      int
	PageParam     string
	PageSizeParam string
	MaxPageSize   int
}

func (p PageNumberPagination[T]) Paginate(qs *orm.QuerySet[T], req *APIRequest, meta *orm.ModelMeta) (*orm.QuerySet[T], map[string]interface{}, error) {
	pageParam := p.PageParam
	if pageParam == "" {
		pageParam = "page"
	}
	sizeParam := p.PageSizeParam
	if sizeParam == "" {
		sizeParam = "page_size"
	}
	page := 1
	if raw := req.QueryParams().Get(pageParam); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed < 1 {
			return nil, nil, fmt.Errorf("page must be a positive integer")
		}
		page = parsed
	}
	size := p.PageSize
	if size <= 0 {
		size = DefaultLimit
	}
	if raw := req.QueryParams().Get(sizeParam); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed < 0 {
			return nil, nil, fmt.Errorf("page_size must be a non-negative integer")
		}
		size = parsed
	}
	maxSize := p.MaxPageSize
	if maxSize <= 0 {
		maxSize = MaxLimit
	}
	if size > maxSize {
		return nil, nil, fmt.Errorf("page_size must be less than or equal to %d", maxSize)
	}
	count, err := qs.Count(req.Context())
	if err != nil {
		return nil, nil, err
	}
	offset := (page - 1) * size
	return qs.Limit(size).Offset(offset), map[string]interface{}{
		"count":     count,
		"page":      page,
		"page_size": size,
	}, nil
}

type LimitOffsetPagination[T any] struct {
	DefaultLimit int
	MaxLimit     int
}

func (p LimitOffsetPagination[T]) Paginate(qs *orm.QuerySet[T], req *APIRequest, meta *orm.ModelMeta) (*orm.QuerySet[T], map[string]interface{}, error) {
	limit, offset, err := pagination(p.DefaultLimit, req.QueryParams())
	if err != nil {
		return nil, nil, err
	}
	if p.MaxLimit > 0 && limit > p.MaxLimit {
		return nil, nil, fmt.Errorf("limit must be less than or equal to %d", p.MaxLimit)
	}
	count, err := qs.Count(req.Context())
	if err != nil {
		return nil, nil, err
	}
	return qs.Limit(limit).Offset(offset), map[string]interface{}{
		"count":  count,
		"limit":  limit,
		"offset": offset,
	}, nil
}

type CursorPagination[T any] struct {
	PageSize    int
	CursorParam string
}

func (p CursorPagination[T]) Paginate(qs *orm.QuerySet[T], req *APIRequest, meta *orm.ModelMeta) (*orm.QuerySet[T], map[string]interface{}, error) {
	if meta == nil {
		return nil, nil, fmt.Errorf("model metadata is required")
	}
	cursorParam := p.CursorParam
	if cursorParam == "" {
		cursorParam = "cursor"
	}
	size := p.PageSize
	if size <= 0 {
		size = DefaultLimit
	}
	if raw := req.QueryParams().Get(cursorParam); raw != "" {
		decoded, err := base64.RawURLEncoding.DecodeString(raw)
		if err != nil {
			return nil, nil, fmt.Errorf("cursor is invalid")
		}
		qs = qs.Filter(orm.L(meta.PKColumn()+"__gt", string(decoded)))
	}
	return qs.OrderBy(meta.PKColumn()).Limit(size), map[string]interface{}{
		"page_size": size,
		"cursor":    req.QueryParams().Get(cursorParam),
	}, nil
}

func queryWith(values url.Values, key, value string) string {
	next := url.Values{}
	for k, vals := range values {
		for _, v := range vals {
			next.Add(k, v)
		}
	}
	next.Set(key, value)
	return "?" + next.Encode()
}
