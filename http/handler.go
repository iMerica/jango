package http

import (
	"bytes"
	"net/http"
)

type ViewFunc func(*Request) Response

func (fn ViewFunc) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	req := WrapRequest(r)
	resp := fn(req)
	if err := resp.WriteTo(w, r); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func Adapt(fn ViewFunc) http.Handler {
	return fn
}

func WrapHTTPHandler(h http.Handler) ViewFunc {
	return func(req *Request) Response {
		rec := newResponseRecorder()
		h.ServeHTTP(rec, req.Request)
		return &recorderResponse{
			statusCode: rec.code,
			header:     rec.HeaderMap,
			body:       rec.body.Bytes(),
		}
	}
}

type recorderResponse struct {
	statusCode int
	header     http.Header
	body       []byte
}

func (r *recorderResponse) StatusCode() int {
	if r.statusCode == 0 {
		return http.StatusOK
	}
	return r.statusCode
}

func (r *recorderResponse) WriteTo(w http.ResponseWriter, _ *http.Request) error {
	for k, vv := range r.header {
		for _, v := range vv {
			w.Header().Add(k, v)
		}
	}
	if r.statusCode > 0 {
		w.WriteHeader(r.statusCode)
	}
	_, err := w.Write(r.body)
	return err
}

type responseRecorder struct {
	code       int
	HeaderMap  http.Header
	body       *bytes.Buffer
	wroteHeader bool
}

func newResponseRecorder() *responseRecorder {
	return &responseRecorder{
		HeaderMap: make(http.Header),
		body:      new(bytes.Buffer),
	}
}

func (r *responseRecorder) Header() http.Header {
	return r.HeaderMap
}

func (r *responseRecorder) Write(buf []byte) (int, error) {
	if !r.wroteHeader {
		r.WriteHeader(http.StatusOK)
	}
	return r.body.Write(buf)
}

func (r *responseRecorder) WriteHeader(code int) {
	if r.wroteHeader {
		return
	}
	r.code = code
	r.wroteHeader = true
}