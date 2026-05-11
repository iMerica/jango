package rest

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	jangohttp "github.com/iMerica/jango/http"
)

type Renderer interface {
	Render(data interface{}) ([]byte, error)
	ContentType() string
}

type JSONRenderer struct{}

func (r JSONRenderer) Render(data interface{}) ([]byte, error) {
	if data == nil {
		return []byte("null"), nil
	}
	return json.Marshal(data)
}

func (r JSONRenderer) ContentType() string {
	return "application/json"
}

type BrowsableAPIRenderer struct{}

func (r BrowsableAPIRenderer) Render(data interface{}) ([]byte, error) {
	body, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return nil, err
	}
	return []byte("<!doctype html><html><head><meta charset=\"utf-8\"><title>JanGO API</title></head><body><pre>" + escapeHTML(string(body)) + "</pre></body></html>"), nil
}

func (r BrowsableAPIRenderer) ContentType() string {
	return "text/html; charset=utf-8"
}

type ContentNegotiator interface {
	SelectRenderer(req *APIRequest, renderers []Renderer) (Renderer, string, error)
}

type DefaultContentNegotiator struct{}

func (n DefaultContentNegotiator) SelectRenderer(req *APIRequest, renderers []Renderer) (Renderer, string, error) {
	if len(renderers) == 0 {
		renderers = []Renderer{JSONRenderer{}}
	}
	format := formatFromPath(req.URL.Path)
	if format != "" {
		for _, renderer := range renderers {
			if rendererMatchesFormat(renderer, format) {
				return renderer, format, nil
			}
		}
		return nil, "", fmt.Errorf("unsupported format %q", format)
	}
	accept := req.Header.Get("Accept")
	if accept == "" || accept == "*/*" {
		return renderers[0], formatForRenderer(renderers[0]), nil
	}
	for _, part := range strings.Split(accept, ",") {
		media := strings.TrimSpace(strings.Split(part, ";")[0])
		for _, renderer := range renderers {
			if media == "*/*" || media == renderer.ContentType() || strings.HasPrefix(renderer.ContentType(), media) {
				return renderer, formatForRenderer(renderer), nil
			}
		}
	}
	return renderers[0], formatForRenderer(renderers[0]), nil
}

type APIResponse struct {
	Status   int
	Data     interface{}
	Headers  http.Header
	Cookies  []*http.Cookie
	Renderer Renderer
}

func NewAPIResponse(data interface{}, status int) *APIResponse {
	if status == 0 {
		status = http.StatusOK
	}
	return &APIResponse{
		Status:   status,
		Data:     data,
		Headers:  make(http.Header),
		Renderer: JSONRenderer{},
	}
}

func (r *APIResponse) StatusCode() int {
	return r.Status
}

func (r *APIResponse) SetHeader(key, value string) {
	r.ensureHeaders()
	r.Headers.Set(key, value)
}

func (r *APIResponse) AddHeader(key, value string) {
	r.ensureHeaders()
	r.Headers.Add(key, value)
}

func (r *APIResponse) DelHeader(key string) {
	r.ensureHeaders()
	r.Headers.Del(key)
}

func (r *APIResponse) SetCookie(cookie *http.Cookie) {
	r.Cookies = append(r.Cookies, cookie)
}

func (r *APIResponse) WriteTo(w http.ResponseWriter, req *http.Request) error {
	renderer := r.Renderer
	if renderer == nil {
		renderer = JSONRenderer{}
	}
	body, err := renderer.Render(r.Data)
	if err != nil {
		return err
	}
	r.ensureHeaders()
	r.Headers.Set("Content-Type", renderer.ContentType())
	for k, vv := range r.Headers {
		for _, v := range vv {
			w.Header().Add(k, v)
		}
	}
	for _, c := range r.Cookies {
		http.SetCookie(w, c)
	}
	w.WriteHeader(r.Status)
	if req != nil && req.Method == http.MethodHead {
		return nil
	}
	_, err = w.Write(body)
	return err
}

func formatFromPath(path string) string {
	lastSlash := strings.LastIndex(path, "/")
	lastDot := strings.LastIndex(path, ".")
	if lastDot <= lastSlash {
		return ""
	}
	return strings.Trim(path[lastDot+1:], "/")
}

func formatForRenderer(renderer Renderer) string {
	switch renderer.(type) {
	case BrowsableAPIRenderer:
		return "api"
	default:
		if strings.HasPrefix(renderer.ContentType(), "application/json") {
			return "json"
		}
		return strings.Split(renderer.ContentType(), "/")[0]
	}
}

func rendererMatchesFormat(renderer Renderer, format string) bool {
	switch format {
	case "json":
		return strings.HasPrefix(renderer.ContentType(), "application/json")
	case "api", "html":
		return strings.HasPrefix(renderer.ContentType(), "text/html")
	default:
		return false
	}
}

func escapeHTML(s string) string {
	replacer := strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;", "\"", "&#34;", "'", "&#39;")
	return replacer.Replace(s)
}

func (r *APIResponse) ensureHeaders() {
	if r.Headers == nil {
		r.Headers = make(http.Header)
	}
}

func ErrorResponse(detail string, status int) *APIResponse {
	return NewAPIResponse(map[string]interface{}{"detail": detail}, status)
}

func BadRequest(detail string) *APIResponse {
	return ErrorResponse(detail, http.StatusBadRequest)
}

func NotFound(detail string) *APIResponse {
	return ErrorResponse(detail, http.StatusNotFound)
}

func MethodNotAllowed(method string, allowed ...string) *APIResponse {
	resp := ErrorResponse("method not allowed: "+method, http.StatusMethodNotAllowed)
	if len(allowed) > 0 {
		resp.SetHeader("Allow", joinMethods(allowed))
	}
	return resp
}

func InternalServerError(detail string) *APIResponse {
	return ErrorResponse(detail, http.StatusInternalServerError)
}

func joinMethods(methods []string) string {
	if len(methods) == 0 {
		return ""
	}
	n := 0
	for _, method := range methods {
		n += len(method)
	}
	n += 2 * (len(methods) - 1)
	buf := make([]byte, 0, n)
	for i, method := range methods {
		if i > 0 {
			buf = append(buf, ',', ' ')
		}
		buf = append(buf, method...)
	}
	return string(buf)
}

var _ jangohttp.Response = (*APIResponse)(nil)
