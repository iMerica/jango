package http

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestTextResponse(t *testing.T) {
	resp := NewTextResponse("hello world")
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)

	if resp.StatusCode() != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode())
	}

	err := resp.WriteTo(rec, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
	if rec.Body.String() != "hello world" {
		t.Errorf("expected body 'hello world', got %q", rec.Body.String())
	}
	ct := rec.Header().Get("Content-Type")
	if ct != "text/plain; charset=utf-8" {
		t.Errorf("expected text/plain content type, got %s", ct)
	}
}

func TestTextResponseWithStatus(t *testing.T) {
	resp := NewTextResponseWithStatus("not found", http.StatusNotFound)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)

	err := resp.WriteTo(rec, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", rec.Code)
	}
}

func TestHTMLResponse(t *testing.T) {
	html := "<html><body><h1>Hello</h1></body></html>"
	resp := NewHTMLResponse(html)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)

	err := resp.WriteTo(rec, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Body.String() != html {
		t.Errorf("expected html body, got %q", rec.Body.String())
	}
	ct := rec.Header().Get("Content-Type")
	if ct != "text/html; charset=utf-8" {
		t.Errorf("expected text/html content type, got %s", ct)
	}
}

func TestJSONResponse(t *testing.T) {
	data := map[string]interface{}{"name": "test", "count": 42}
	resp := NewJSONResponse(data)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)

	err := resp.WriteTo(rec, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ct := rec.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("expected application/json content type, got %s", ct)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}
	if result["name"] != "test" {
		t.Errorf("expected name=test, got %v", result["name"])
	}
}

func TestJSONResponseWithStatus(t *testing.T) {
	data := map[string]interface{}{"error": "bad request"}
	resp := NewJSONResponseWithStatus(data, http.StatusBadRequest)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)

	err := resp.WriteTo(rec, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestRedirectResponse(t *testing.T) {
	resp := NewRedirectResponse("/new-url")
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/old-url", nil)

	err := resp.WriteTo(rec, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusFound {
		t.Errorf("expected status 302, got %d", rec.Code)
	}
	if rec.Header().Get("Location") != "/new-url" {
		t.Errorf("expected Location /new-url, got %s", rec.Header().Get("Location"))
	}
}

func TestPermanentRedirectResponse(t *testing.T) {
	resp := NewPermanentRedirectResponse("/permanent")
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/old", nil)

	err := resp.WriteTo(rec, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusMovedPermanently {
		t.Errorf("expected status 301, got %d", rec.Code)
	}
	if rec.Header().Get("Location") != "/permanent" {
		t.Errorf("expected Location /permanent, got %s", rec.Header().Get("Location"))
	}
}

func TestErrorResponse(t *testing.T) {
	resp := NewNotFoundResponse("Page not found")
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/missing", nil)

	err := resp.WriteTo(rec, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", rec.Code)
	}
	if rec.Body.String() != "Page not found" {
		t.Errorf("expected 'Page not found', got %q", rec.Body.String())
	}
}

func TestErrorResponseHelpers(t *testing.T) {
	tests := []struct {
		name     string
		resp     *ErrorResponse
		status   int
		message  string
	}{
		{"bad request", NewBadRequestResponse("bad"), 400, "bad"},
		{"unauthorized", NewUnauthorizedResponse("no auth"), 401, "no auth"},
		{"forbidden", NewForbiddenResponse("denied"), 403, "denied"},
		{"not found", NewNotFoundResponse("missing"), 404, "missing"},
		{"method not allowed", NewMethodNotAllowedResponse("nope"), 405, "nope"},
		{"internal error", NewInternalServerErrorResponse("boom"), 500, "boom"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.resp.StatusCode() != tt.status {
				t.Errorf("expected status %d, got %d", tt.status, tt.resp.StatusCode())
			}
			if tt.resp.Message != tt.message {
				t.Errorf("expected message %q, got %q", tt.message, tt.resp.Message)
			}
		})
	}
}

func TestBaseResponseHeaders(t *testing.T) {
	resp := NewTextResponse("test")
	resp.SetHeader("X-Custom", "value")
	resp.AddHeader("X-Multi", "v1")
	resp.AddHeader("X-Multi", "v2")
	resp.DelHeader("X-Multi")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	resp.WriteTo(rec, req)

	if rec.Header().Get("X-Custom") != "value" {
		t.Errorf("expected X-Custom=value, got %s", rec.Header().Get("X-Custom"))
	}
	if rec.Header().Get("X-Multi") != "" {
		t.Error("expected X-Multi to be deleted")
	}
}

func TestBaseResponseCookies(t *testing.T) {
	resp := NewTextResponse("test")
	resp.SetCookie(&http.Cookie{Name: "session", Value: "abc123"})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	resp.WriteTo(rec, req)

	cookies := rec.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("expected 1 cookie, got %d", len(cookies))
	}
	if cookies[0].Name != "session" || cookies[0].Value != "abc123" {
		t.Errorf("expected session=abc123, got %s=%s", cookies[0].Name, cookies[0].Value)
	}
}

type mockTemplateEngine struct {
	templates map[string]string
}

func (e *mockTemplateEngine) ExecuteTemplate(w io.Writer, name string, data interface{}) error {
	tmpl, ok := e.templates[name]
	if !ok {
		return io.EOF
	}
	_, err := w.Write([]byte(tmpl))
	return err
}

func TestTemplateResponseLazyRendering(t *testing.T) {
	engine := &mockTemplateEngine{
		templates: map[string]string{
			"test.html": "<h1>Hello, World!</h1>",
		},
	}

	resp := NewTemplateResponse(engine, "test.html", nil)

	if resp.IsRendered() {
		t.Error("expected template to not be rendered yet")
	}

	err := resp.Render()
	if err != nil {
		t.Fatalf("unexpected render error: %v", err)
	}

	if !resp.IsRendered() {
		t.Error("expected template to be rendered")
	}

	err = resp.Render()
	if err != nil {
		t.Fatalf("unexpected error on second render: %v", err)
	}
}

func TestTemplateResponseWriteTo(t *testing.T) {
	engine := &mockTemplateEngine{
		templates: map[string]string{
			"test.html": "<h1>Hello!</h1>",
		},
	}

	resp := NewTemplateResponse(engine, "test.html", nil)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)

	err := resp.WriteTo(rec, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
	if rec.Body.String() != "<h1>Hello!</h1>" {
		t.Errorf("expected rendered template body, got %q", rec.Body.String())
	}
	ct := rec.Header().Get("Content-Type")
	if ct != "text/html; charset=utf-8" {
		t.Errorf("expected text/html content type, got %s", ct)
	}
}

func TestTemplateResponseMissingTemplate(t *testing.T) {
	engine := &mockTemplateEngine{
		templates: map[string]string{},
	}

	resp := NewTemplateResponse(engine, "missing.html", nil)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)

	err := resp.WriteTo(rec, req)
	if err == nil {
		t.Error("expected error for missing template")
	}
}

func TestStreamResponse(t *testing.T) {
	data := strings.NewReader("streaming data here")
	resp := NewStreamResponse(data)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)

	err := resp.WriteTo(rec, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
	if rec.Body.String() != "streaming data here" {
		t.Errorf("expected streaming body, got %q", rec.Body.String())
	}
}

func TestStreamResponseClose(t *testing.T) {
	data := strings.NewReader("data")
	resp := NewStreamResponse(data)

	resp.Close()

	select {
	case <-resp.Closed:
	default:
		t.Error("expected stream to be closed")
	}
}

func TestFileResponse(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "testfile-*.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.WriteString("file content here")
	tmpFile.Close()

	resp := NewFileResponse(tmpFile.Name())
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/file", nil)

	err = resp.WriteTo(rec, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Body.String() != "file content here" {
		t.Errorf("expected file content, got %q", rec.Body.String())
	}
}

func TestDetectContentType(t *testing.T) {
	tests := []struct {
		ext      string
		expected string
	}{
		{".html", "text/html; charset=utf-8"},
		{".css", "text/css; charset=utf-8"},
		{".js", "application/javascript"},
		{".json", "application/json"},
		{".png", "image/png"},
		{".txt", "text/plain; charset=utf-8"},
		{".unknown", "application/octet-stream"},
	}
	for _, tt := range tests {
		t.Run(tt.ext, func(t *testing.T) {
			result := detectContentType("file" + tt.ext)
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestAdaptViewFunc(t *testing.T) {
	handler := ViewFunc(func(req *Request) Response {
		return NewTextResponse("adapted")
	})

	rec := httptest.NewRecorder()
	stdReq := httptest.NewRequest("GET", "/test", nil)
	handler.ServeHTTP(rec, stdReq)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
	if rec.Body.String() != "adapted" {
		t.Errorf("expected 'adapted', got %q", rec.Body.String())
	}
}

func TestWrapHTTPHandler(t *testing.T) {
	stdHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Custom", "from-std")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("std handler response"))
	})

	adapted := WrapHTTPHandler(stdHandler)
	req := WrapRequest(httptest.NewRequest("GET", "/test", nil))
	resp := adapted(req)

	if resp.StatusCode() != http.StatusCreated {
		t.Errorf("expected 201, got %d", resp.StatusCode())
	}
}

func TestWrapHTTPHandlerAndAdapt(t *testing.T) {
	viewFn := ViewFunc(func(req *Request) Response {
		return NewTextResponse("view func")
	})

	handler := Adapt(viewFn)
	rec := httptest.NewRecorder()
	stdReq := httptest.NewRequest("GET", "/test", nil)
	handler.ServeHTTP(rec, stdReq)

	if rec.Body.String() != "view func" {
		t.Errorf("expected 'view func', got %q", rec.Body.String())
	}
}