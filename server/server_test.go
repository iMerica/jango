package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	jangohttp "github.com/iMerica/jango/http"
	"github.com/iMerica/jango/middleware"
	"github.com/iMerica/jango/urls"
)

func TestNewHandler(t *testing.T) {
	patterns := []urls.Pattern{
		urls.Path("/", jangohttp.ViewFunc(func(req *jangohttp.Request) jangohttp.Response {
			return jangohttp.NewTextResponse("hello")
		}), "index"),
		urls.Path("/about/", jangohttp.ViewFunc(func(req *jangohttp.Request) jangohttp.Response {
			return jangohttp.NewTextResponse("about")
		}), "about"),
	}

	handler := NewHandler(patterns)
	if handler == nil {
		t.Fatal("expected non-nil handler")
	}
}

func TestServerBasicRouting(t *testing.T) {
	patterns := []urls.Pattern{
		urls.Path("/", jangohttp.ViewFunc(func(req *jangohttp.Request) jangohttp.Response {
			return jangohttp.NewTextResponse("index")
		}), "index"),
		urls.Path("/hello/", jangohttp.ViewFunc(func(req *jangohttp.Request) jangohttp.Response {
			return jangohttp.NewTextResponse("hello")
		}), "hello"),
	}

	handler := NewHandler(patterns)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	if rec.Body.String() != "index" {
		t.Errorf("expected 'index', got %q", rec.Body.String())
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/hello/", nil)
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	if rec.Body.String() != "hello" {
		t.Errorf("expected 'hello', got %q", rec.Body.String())
	}
}

func TestServerNotFound(t *testing.T) {
	patterns := []urls.Pattern{
		urls.Path("/", jangohttp.ViewFunc(func(req *jangohttp.Request) jangohttp.Response {
			return jangohttp.NewTextResponse("index")
		}), "index"),
	}

	handler := NewHandler(patterns)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/nonexistent/", nil)
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}

func TestServerWithPathParams(t *testing.T) {
	patterns := []urls.Pattern{
		urls.Path("/articles/<int:id>/", jangohttp.ViewFunc(func(req *jangohttp.Request) jangohttp.Response {
			return jangohttp.NewTextResponse("article:" + req.Param("id"))
		}), "article-detail"),
	}

	handler := NewHandler(patterns)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/articles/42/", nil)
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	if rec.Body.String() != "article:42" {
		t.Errorf("expected 'article:42', got %q", rec.Body.String())
	}
}

func TestServerWithMiddleware(t *testing.T) {
	headerAdded := false
	mw := func(next jangohttp.ViewFunc) jangohttp.ViewFunc {
		return func(req *jangohttp.Request) jangohttp.Response {
			headerAdded = true
			resp := next(req)
			resp.(*jangohttp.TextResponse).SetHeader("X-Middleware", "yes")
			return resp
		}
	}

	patterns := []urls.Pattern{
		urls.Path("/", jangohttp.ViewFunc(func(req *jangohttp.Request) jangohttp.Response {
			return jangohttp.NewTextResponse("ok")
		}), "index"),
	}

	handler := NewHandler(patterns, middleware.MiddlewareFunc(mw))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	handler.ServeHTTP(rec, req)

	if !headerAdded {
		t.Error("expected middleware to be called")
	}
	if rec.Header().Get("X-Middleware") != "yes" {
		t.Error("expected X-Middleware header to be set")
	}
}

func TestServerMiddlewareShortCircuit(t *testing.T) {
	viewCalled := false
	mw := func(next jangohttp.ViewFunc) jangohttp.ViewFunc {
		return func(req *jangohttp.Request) jangohttp.Response {
			return jangohttp.NewForbiddenResponse("denied")
		}
	}

	patterns := []urls.Pattern{
		urls.Path("/", jangohttp.ViewFunc(func(req *jangohttp.Request) jangohttp.Response {
			viewCalled = true
			return jangohttp.NewTextResponse("ok")
		}), "index"),
	}

	handler := NewHandler(patterns, middleware.MiddlewareFunc(mw))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	handler.ServeHTTP(rec, req)

	if viewCalled {
		t.Error("expected view NOT to be called when middleware short-circuits")
	}
	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rec.Code)
	}
}

func TestServerWithInclude(t *testing.T) {
	subPatterns := []urls.Pattern{
		urls.Path("/", jangohttp.ViewFunc(func(req *jangohttp.Request) jangohttp.Response {
			return jangohttp.NewTextResponse("api-root")
		}), "api-root"),
		urls.Path("/items/", jangohttp.ViewFunc(func(req *jangohttp.Request) jangohttp.Response {
			return jangohttp.NewTextResponse("api-items")
		}), "api-items"),
	}

	patterns := []urls.Pattern{
		urls.Include("/api", subPatterns, "api"),
		urls.Path("/", jangohttp.ViewFunc(func(req *jangohttp.Request) jangohttp.Response {
			return jangohttp.NewTextResponse("index")
		}), "index"),
	}

	handler := NewHandler(patterns)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/items/", nil)
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	if rec.Body.String() != "api-items" {
		t.Errorf("expected 'api-items', got %q", rec.Body.String())
	}
}

func TestServerHTTPHandler(t *testing.T) {
	stdHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("std handler"))
	})

	patterns := []urls.Pattern{
		urls.Path("/std/", jangohttp.WrapHTTPHandler(stdHandler), "std"),
		urls.Path("/view/", jangohttp.ViewFunc(func(req *jangohttp.Request) jangohttp.Response {
			return jangohttp.NewTextResponse("view handler")
		}), "view"),
	}

	handler := NewHandler(patterns)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/std/", nil)
	handler.ServeHTTP(rec, req)

	if rec.Body.String() != "std handler" {
		t.Errorf("expected 'std handler', got %q", rec.Body.String())
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/view/", nil)
	handler.ServeHTTP(rec, req)

	if rec.Body.String() != "view handler" {
		t.Errorf("expected 'view handler', got %q", rec.Body.String())
	}
}
