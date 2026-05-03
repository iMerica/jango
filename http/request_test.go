package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestWrapRequest(t *testing.T) {
	stdReq := httptest.NewRequest("GET", "/test?foo=bar", nil)
	req := WrapRequest(stdReq)

	if req.Request != stdReq {
		t.Error("expected embedded *http.Request to match")
	}
	if req.Params == nil {
		t.Error("expected Params to be initialized")
	}
	if req.User != nil {
		t.Error("expected User to be nil initially")
	}
	if req.Session != nil {
		t.Error("expected Session to be nil initially")
	}
}

func TestRequestContext(t *testing.T) {
	stdReq := httptest.NewRequest("GET", "/test", nil)
	req := WrapRequest(stdReq)

	ctx := req.Context()
	if ctx == nil {
		t.Error("expected context to be non-nil")
	}

	ctx2 := context.WithValue(ctx, "testkey", "testval")
	req2 := req.WithContext(ctx2)
	if req2.Context().Value("testkey") != "testval" {
		t.Error("expected context value to propagate")
	}
}

func TestRequestQuery(t *testing.T) {
	stdReq := httptest.NewRequest("GET", "/test?foo=bar&baz=qux", nil)
	req := WrapRequest(stdReq)

	q := req.Query()
	if q.Get("foo") != "bar" {
		t.Errorf("expected foo=bar, got %s", q.Get("foo"))
	}
	if q.Get("baz") != "qux" {
		t.Errorf("expected baz=qux, got %s", q.Get("baz"))
	}

	q2 := req.Query()
	if &q == &q2 {
		t.Error("expected different pointer on repeated call (but same values)")
	}
}

func TestRequestQueryParams(t *testing.T) {
	stdReq := httptest.NewRequest("GET", "/search?q=jango&lang=en&page=2", nil)
	req := WrapRequest(stdReq)

	q := req.Query()
	if q.Get("q") != "jango" {
		t.Errorf("expected q=jango, got %s", q.Get("q"))
	}
	if q.Get("lang") != "en" {
		t.Errorf("expected lang=en, got %s", q.Get("lang"))
	}
	if q.Get("page") != "2" {
		t.Errorf("expected page=2, got %s", q.Get("page"))
	}
}

func TestRequestForm(t *testing.T) {
	form := url.Values{}
	form.Set("username", "testuser")
	form.Set("password", "testpass")
	stdReq := httptest.NewRequest("POST", "/login", strings.NewReader(form.Encode()))
	stdReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req := WrapRequest(stdReq)

	f, err := req.Form()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f.Get("username") != "testuser" {
		t.Errorf("expected username=testuser, got %s", f.Get("username"))
	}
	if f.Get("password") != "testpass" {
		t.Errorf("expected password=testpass, got %s", f.Get("password"))
	}
}

func TestRequestBody(t *testing.T) {
	body := `{"name": "test", "value": 42}`
	stdReq := httptest.NewRequest("POST", "/api", strings.NewReader(body))
	stdReq.Header.Set("Content-Type", "application/json")
	req := WrapRequest(stdReq)

	data, err := req.Body()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(data) != body {
		t.Errorf("expected body %q, got %q", body, string(data))
	}

	data2, err := req.Body()
	if err != nil {
		t.Fatalf("unexpected error on second call: %v", err)
	}
	if string(data2) != body {
		t.Error("expected body to be reusable after first read")
	}
}

func TestRequestParseBodyJSON(t *testing.T) {
	body := `{"name": "test", "value": 42}`
	stdReq := httptest.NewRequest("POST", "/api", strings.NewReader(body))
	stdReq.Header.Set("Content-Type", "application/json")
	req := WrapRequest(stdReq)

	var result map[string]interface{}
	err := req.ParseBody(&result)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result["name"] != "test" {
		t.Errorf("expected name=test, got %v", result["name"])
	}
}

func TestRequestParam(t *testing.T) {
	stdReq := httptest.NewRequest("GET", "/test", nil)
	req := WrapRequest(stdReq)
	req.Params["id"] = "123"
	req.Params["slug"] = "hello-world"

	if req.Param("id") != "123" {
		t.Errorf("expected id=123, got %s", req.Param("id"))
	}
	if req.Param("slug") != "hello-world" {
		t.Errorf("expected slug=hello-world, got %s", req.Param("slug"))
	}
	if req.Param("nonexistent") != "" {
		t.Error("expected empty string for nonexistent param")
	}
}

func TestRequestSetParam(t *testing.T) {
	stdReq := httptest.NewRequest("GET", "/test", nil)
	req := WrapRequest(stdReq)
	req.SetParam("id", "456")

	if req.Param("id") != "456" {
		t.Errorf("expected id=456, got %s", req.Param("id"))
	}
}

func TestRequestSetUserSession(t *testing.T) {
	stdReq := httptest.NewRequest("GET", "/test", nil)
	req := WrapRequest(stdReq)

	user := struct{ Name string }{Name: "admin"}
	req.SetUser(user)
	if req.User.(struct{ Name string }).Name != "admin" {
		t.Errorf("expected user Name=admin")
	}

	session := map[string]interface{}{"user_id": 1}
	req.SetSession(session)
	if req.Session.(map[string]interface{})["user_id"] != 1 {
		t.Error("expected session user_id=1")
	}
}

func TestRequestCookies(t *testing.T) {
	stdReq := httptest.NewRequest("GET", "/test", nil)
	stdReq.AddCookie(&http.Cookie{Name: "sessionid", Value: "abc123"})
	req := WrapRequest(stdReq)

	cookies := req.Cookies()
	if len(cookies) != 1 {
		t.Fatalf("expected 1 cookie, got %d", len(cookies))
	}
	if cookies[0].Name != "sessionid" || cookies[0].Value != "abc123" {
		t.Errorf("expected sessionid=abc123, got %s=%s", cookies[0].Name, cookies[0].Value)
	}
}