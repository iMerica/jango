package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

type Response interface {
	StatusCode() int
	WriteTo(w http.ResponseWriter, r *http.Request) error
}

type BaseResponse struct {
	Status  int
	Headers http.Header
	Cookies []*http.Cookie
}

func newBaseResponse(status int) BaseResponse {
	return BaseResponse{
		Status:  status,
		Headers: make(http.Header),
	}
}

func (b *BaseResponse) StatusCode() int {
	return b.Status
}

func (b *BaseResponse) SetHeader(key, value string) {
	b.Headers.Set(key, value)
}

func (b *BaseResponse) AddHeader(key, value string) {
	b.Headers.Add(key, value)
}

func (b *BaseResponse) DelHeader(key string) {
	b.Headers.Del(key)
}

func (b *BaseResponse) SetCookie(cookie *http.Cookie) {
	b.Cookies = append(b.Cookies, cookie)
}

func (b *BaseResponse) writeHeadersAndCookies(w http.ResponseWriter) {
	for k, vv := range b.Headers {
		for _, v := range vv {
			w.Header().Add(k, v)
		}
	}
	for _, c := range b.Cookies {
		http.SetCookie(w, c)
	}
}

type TextResponse struct {
	BaseResponse
	Body string
}

func NewTextResponse(body string) *TextResponse {
	return &TextResponse{
		BaseResponse: newBaseResponse(http.StatusOK),
		Body:         body,
	}
}

func NewTextResponseWithStatus(body string, status int) *TextResponse {
	return &TextResponse{
		BaseResponse: newBaseResponse(status),
		Body:         body,
	}
}

func (r *TextResponse) WriteTo(w http.ResponseWriter, req *http.Request) error {
	r.Headers.Set("Content-Type", "text/plain; charset=utf-8")
	r.writeHeadersAndCookies(w)
	w.WriteHeader(r.Status)
	_, err := w.Write([]byte(r.Body))
	return err
}

type HTMLResponse struct {
	BaseResponse
	Body string
}

func NewHTMLResponse(body string) *HTMLResponse {
	return &HTMLResponse{
		BaseResponse: newBaseResponse(http.StatusOK),
		Body:         body,
	}
}

func NewHTMLResponseWithStatus(body string, status int) *HTMLResponse {
	return &HTMLResponse{
		BaseResponse: newBaseResponse(status),
		Body:         body,
	}
}

func (r *HTMLResponse) WriteTo(w http.ResponseWriter, req *http.Request) error {
	r.Headers.Set("Content-Type", "text/html; charset=utf-8")
	r.writeHeadersAndCookies(w)
	w.WriteHeader(r.Status)
	_, err := w.Write([]byte(r.Body))
	return err
}

type JSONResponse struct {
	BaseResponse
	Data interface{}
}

func NewJSONResponse(data interface{}) *JSONResponse {
	return &JSONResponse{
		BaseResponse: newBaseResponse(http.StatusOK),
		Data:         data,
	}
}

func NewJSONResponseWithStatus(data interface{}, status int) *JSONResponse {
	return &JSONResponse{
		BaseResponse: newBaseResponse(status),
		Data:         data,
	}
}

func (r *JSONResponse) WriteTo(w http.ResponseWriter, req *http.Request) error {
	body, err := json.Marshal(r.Data)
	if err != nil {
		return err
	}
	r.Headers.Set("Content-Type", "application/json")
	r.writeHeadersAndCookies(w)
	w.WriteHeader(r.Status)
	_, err = w.Write(body)
	return err
}

type RedirectResponse struct {
	BaseResponse
	URL       string
	Permanent bool
}

func NewRedirectResponse(url string) *RedirectResponse {
	return &RedirectResponse{
		BaseResponse: newBaseResponse(http.StatusFound),
		URL:          url,
		Permanent:    false,
	}
}

func NewPermanentRedirectResponse(url string) *RedirectResponse {
	return &RedirectResponse{
		BaseResponse: newBaseResponse(http.StatusMovedPermanently),
		URL:          url,
		Permanent:    true,
	}
}

func (r *RedirectResponse) WriteTo(w http.ResponseWriter, req *http.Request) error {
	r.Headers.Set("Location", r.URL)
	r.writeHeadersAndCookies(w)
	w.WriteHeader(r.Status)
	return nil
}

type TemplateResponse struct {
	BaseResponse
	TemplateName string
	ContextData  interface{}
	rendered     bool
	renderedBody *bytes.Buffer
	Engine       TemplateEngine
}

type TemplateEngine interface {
	ExecuteTemplate(wr io.Writer, name string, data interface{}) error
}

func NewTemplateResponse(engine TemplateEngine, templateName string, contextData interface{}) *TemplateResponse {
	return &TemplateResponse{
		BaseResponse: newBaseResponse(http.StatusOK),
		TemplateName: templateName,
		ContextData:  contextData,
		Engine:       engine,
	}
}

func (r *TemplateResponse) IsRendered() bool {
	return r.rendered
}

func (r *TemplateResponse) Render() error {
	if r.rendered {
		return nil
	}
	buf := new(bytes.Buffer)
	if err := r.Engine.ExecuteTemplate(buf, r.TemplateName, r.ContextData); err != nil {
		return err
	}
	r.renderedBody = buf
	r.rendered = true
	return nil
}

func (r *TemplateResponse) WriteTo(w http.ResponseWriter, req *http.Request) error {
	if !r.rendered {
		if err := r.Render(); err != nil {
			return err
		}
	}
	r.Headers.Set("Content-Type", "text/html; charset=utf-8")
	r.writeHeadersAndCookies(w)
	w.WriteHeader(r.Status)
	_, err := io.Copy(w, r.renderedBody)
	return err
}

type FileResponse struct {
	BaseResponse
	FilePath    string
	ContentType string
}

func NewFileResponse(filePath string) *FileResponse {
	return &FileResponse{
		BaseResponse: newBaseResponse(http.StatusOK),
		FilePath:     filePath,
	}
}

func (r *FileResponse) WriteTo(w http.ResponseWriter, req *http.Request) error {
	f, err := os.Open(r.FilePath)
	if err != nil {
		return err
	}
	defer f.Close()

	stat, err := f.Stat()
	if err != nil {
		return err
	}

	if r.ContentType != "" {
		r.Headers.Set("Content-Type", r.ContentType)
	} else {
		r.Headers.Set("Content-Type", detectContentType(r.FilePath))
	}
	r.Headers.Set("Content-Length", fmt.Sprintf("%d", stat.Size()))
	r.writeHeadersAndCookies(w)
	w.WriteHeader(r.Status)
	_, err = io.Copy(w, f)
	return err
}

type StreamResponse struct {
	BaseResponse
	Stream  io.Reader
	Flusher http.Flusher
	Closed  chan struct{}
}

func NewStreamResponse(stream io.Reader) *StreamResponse {
	return &StreamResponse{
		BaseResponse: newBaseResponse(http.StatusOK),
		Stream:       stream,
		Closed:       make(chan struct{}),
	}
}

func (r *StreamResponse) WriteTo(w http.ResponseWriter, req *http.Request) error {
	r.writeHeadersAndCookies(w)
	w.WriteHeader(r.Status)
	if flusher, ok := w.(http.Flusher); ok {
		r.Flusher = flusher
	}

	buf := make([]byte, 4096)
	for {
		select {
		case <-r.Closed:
			return nil
		case <-req.Context().Done():
			return req.Context().Err()
		default:
		}

		n, err := r.Stream.Read(buf)
		if n > 0 {
			if _, werr := w.Write(buf[:n]); werr != nil {
				return werr
			}
			if r.Flusher != nil {
				r.Flusher.Flush()
			}
		}
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
	}
}

func (r *StreamResponse) Close() {
	select {
	case <-r.Closed:
	default:
		close(r.Closed)
	}
}

type ErrorResponse struct {
	BaseResponse
	Message string
}

func NewErrorResponse(message string, status int) *ErrorResponse {
	return &ErrorResponse{
		BaseResponse: newBaseResponse(status),
		Message:      message,
	}
}

func NewBadRequestResponse(message string) *ErrorResponse {
	return NewErrorResponse(message, http.StatusBadRequest)
}

func NewUnauthorizedResponse(message string) *ErrorResponse {
	return NewErrorResponse(message, http.StatusUnauthorized)
}

func NewForbiddenResponse(message string) *ErrorResponse {
	return NewErrorResponse(message, http.StatusForbidden)
}

func NewNotFoundResponse(message string) *ErrorResponse {
	return NewErrorResponse(message, http.StatusNotFound)
}

func NewMethodNotAllowedResponse(message string) *ErrorResponse {
	return NewErrorResponse(message, http.StatusMethodNotAllowed)
}

func NewInternalServerErrorResponse(message string) *ErrorResponse {
	return NewErrorResponse(message, http.StatusInternalServerError)
}

func (r *ErrorResponse) WriteTo(w http.ResponseWriter, req *http.Request) error {
	r.Headers.Set("Content-Type", "text/plain; charset=utf-8")
	r.writeHeadersAndCookies(w)
	w.WriteHeader(r.Status)
	_, err := w.Write([]byte(r.Message))
	return err
}

func detectContentType(path string) string {
	ext := filepath.Ext(path)
	switch ext {
	case ".html", ".htm":
		return "text/html; charset=utf-8"
	case ".css":
		return "text/css; charset=utf-8"
	case ".js":
		return "application/javascript"
	case ".json":
		return "application/json"
	case ".pdf":
		return "application/pdf"
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".gif":
		return "image/gif"
	case ".svg":
		return "image/svg+xml"
	case ".txt":
		return "text/plain; charset=utf-8"
	case ".xml":
		return "application/xml"
	default:
		return "application/octet-stream"
	}
}