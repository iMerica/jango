package rest

import (
	"encoding/json"
	"net/http"

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
