package rest

import (
	"net/http"
	"net/url"

	jangohttp "github.com/iMerica/jango/http"
)

type APIRequest struct {
	*jangohttp.Request
}

func NewAPIRequest(req *jangohttp.Request) *APIRequest {
	return &APIRequest{Request: req}
}

func WrapRequest(req *http.Request) *APIRequest {
	return NewAPIRequest(jangohttp.WrapRequest(req))
}

func (r *APIRequest) QueryParams() url.Values {
	return r.Query()
}

func (r *APIRequest) URLParam(name string) string {
	return r.Param(name)
}
