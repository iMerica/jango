package middleware

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	jangohttp "github.com/iMerica/jango/http"
)

func TestChainEmpty(t *testing.T) {
	chain := Chain()

	called := false
	finalHandler := func(req *jangohttp.Request) jangohttp.Response {
		called = true
		return jangohttp.NewTextResponse("ok")
	}

	resp := chain(finalHandler)(jangohttp.WrapRequest(httptest.NewRequest("GET", "/", nil)))
	if !called {
		t.Error("expected final handler to be called")
	}
	if resp.StatusCode() != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode())
	}
}

func TestChainSingleMiddleware(t *testing.T) {
	middlewareCalled := false
	mw := func(next jangohttp.ViewFunc) jangohttp.ViewFunc {
		return func(req *jangohttp.Request) jangohttp.Response {
			middlewareCalled = true
			return next(req)
		}
	}

	chain := Chain(mw)

	called := false
	finalHandler := func(req *jangohttp.Request) jangohttp.Response {
		called = true
		return jangohttp.NewTextResponse("ok")
	}

	resp := chain(finalHandler)(jangohttp.WrapRequest(httptest.NewRequest("GET", "/", nil)))
	if !middlewareCalled {
		t.Error("expected middleware to be called")
	}
	if !called {
		t.Error("expected final handler to be called")
	}
	if resp.StatusCode() != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode())
	}
}

func TestChainMultipleMiddlewares(t *testing.T) {
	var order []string

	mw1 := func(next jangohttp.ViewFunc) jangohttp.ViewFunc {
		return func(req *jangohttp.Request) jangohttp.Response {
			order = append(order, "mw1-before")
			resp := next(req)
			order = append(order, "mw1-after")
			return resp
		}
	}

	mw2 := func(next jangohttp.ViewFunc) jangohttp.ViewFunc {
		return func(req *jangohttp.Request) jangohttp.Response {
			order = append(order, "mw2-before")
			resp := next(req)
			order = append(order, "mw2-after")
			return resp
		}
	}

	chain := Chain(mw1, mw2)

	finalHandler := func(req *jangohttp.Request) jangohttp.Response {
		order = append(order, "handler")
		return jangohttp.NewTextResponse("ok")
	}

	chain(finalHandler)(jangohttp.WrapRequest(httptest.NewRequest("GET", "/", nil)))

	expected := []string{"mw1-before", "mw2-before", "handler", "mw2-after", "mw1-after"}
	if len(order) != len(expected) {
		t.Fatalf("expected %d calls, got %d: %v", len(expected), len(order), order)
	}
	for i, v := range expected {
		if order[i] != v {
			t.Errorf("at position %d: expected %s, got %s", i, v, order[i])
		}
	}
}

func TestChainShortCircuit(t *testing.T) {
	viewCalled := false
	mw := func(next jangohttp.ViewFunc) jangohttp.ViewFunc {
		return func(req *jangohttp.Request) jangohttp.Response {
			return jangohttp.NewForbiddenResponse("denied")
		}
	}

	chain := Chain(mw)

	finalHandler := func(req *jangohttp.Request) jangohttp.Response {
		viewCalled = true
		return jangohttp.NewTextResponse("ok")
	}

	resp := chain(finalHandler)(jangohttp.WrapRequest(httptest.NewRequest("GET", "/", nil)))
	if viewCalled {
		t.Error("expected handler NOT to be called when middleware short-circuits")
	}
	if resp.StatusCode() != http.StatusForbidden {
		t.Errorf("expected 403, got %d", resp.StatusCode())
	}
}

func TestChainResponseModification(t *testing.T) {
	mw := func(next jangohttp.ViewFunc) jangohttp.ViewFunc {
		return func(req *jangohttp.Request) jangohttp.Response {
			resp := next(req)
			if tr, ok := resp.(*jangohttp.TextResponse); ok {
				tr.SetHeader("X-Custom", "added-by-middleware")
			}
			return resp
		}
	}

	chain := Chain(mw)

	finalHandler := func(req *jangohttp.Request) jangohttp.Response {
		return jangohttp.NewTextResponse("ok")
	}

	resp := chain(finalHandler)(jangohttp.WrapRequest(httptest.NewRequest("GET", "/", nil)))

	rec := httptest.NewRecorder()
	resp.WriteTo(rec, httptest.NewRequest("GET", "/", nil))

	if rec.Header().Get("X-Custom") != "added-by-middleware" {
		t.Error("expected middleware to add custom header")
	}
}

func TestAdaptHooksOnRequest(t *testing.T) {
	requestModified := false
	hooks := Hooks{
		OnRequest: func(req *Request) (*Request, Response) {
			requestModified = true
			req.SetParam("hook-called", "true")
			return req, nil
		},
	}

	mw := AdaptHooks(hooks)
	viewCalled := false
	handler := mw(func(req *jangohttp.Request) jangohttp.Response {
		viewCalled = true
		if req.Param("hook-called") != "true" {
			t.Error("expected param to be set by OnRequest hook")
		}
		return jangohttp.NewTextResponse("ok")
	})

	req := jangohttp.WrapRequest(httptest.NewRequest("GET", "/", nil))
	handler(req)

	if !requestModified {
		t.Error("expected OnRequest hook to be called")
	}
	if !viewCalled {
		t.Error("expected view to be called")
	}
}

func TestAdaptHooksOnRequestShortCircuit(t *testing.T) {
	hooks := Hooks{
		OnRequest: func(req *Request) (*Request, Response) {
			resp := jangohttp.NewForbiddenResponse("blocked")
			return nil, resp
		},
	}

	mw := AdaptHooks(hooks)
	viewCalled := false
	handler := mw(func(req *jangohttp.Request) jangohttp.Response {
		viewCalled = true
		return jangohttp.NewTextResponse("ok")
	})

	req := jangohttp.WrapRequest(httptest.NewRequest("GET", "/", nil))
	resp := handler(req)

	if viewCalled {
		t.Error("expected view NOT to be called on short-circuit")
	}
	if resp.StatusCode() != http.StatusForbidden {
		t.Errorf("expected 403, got %d", resp.StatusCode())
	}
}

func TestAdaptHooksOnResponse(t *testing.T) {
	responseModified := false
	hooks := Hooks{
		OnResponse: func(req *Request, resp Response) Response {
			responseModified = true
			if tr, ok := resp.(*jangohttp.TextResponse); ok {
				tr.SetHeader("X-Modified", "yes")
			}
			return resp
		},
	}

	mw := AdaptHooks(hooks)
	handler := mw(func(req *jangohttp.Request) jangohttp.Response {
		return jangohttp.NewTextResponse("ok")
	})

	req := jangohttp.WrapRequest(httptest.NewRequest("GET", "/", nil))
	resp := handler(req)

	if !responseModified {
		t.Error("expected OnResponse hook to be called")
	}
	if resp == nil {
		t.Fatal("expected non-nil response")
	}

	rec := httptest.NewRecorder()
	resp.WriteTo(rec, httptest.NewRequest("GET", "/", nil))
	if rec.Header().Get("X-Modified") != "yes" {
		t.Error("expected response header to be modified")
	}
}

func TestAdaptHooksOnException(t *testing.T) {
	exceptionHandled := false
	hooks := Hooks{
		OnException: func(req *Request, err error) Response {
			exceptionHandled = true
			return jangohttp.NewErrorResponse(err.Error(), http.StatusBadRequest)
		},
	}

	hookMw := AdaptHooks(hooks)

	panicFn := func(req *jangohttp.Request) jangohttp.Response {
		panic("test panic")
	}

	chain := Chain(hookMw)
	resp := chain(panicFn)(jangohttp.WrapRequest(httptest.NewRequest("GET", "/", nil)))

	if !exceptionHandled {
		t.Error("expected OnException to be called after panic")
	}
	if resp == nil {
		t.Fatal("expected non-nil response")
	}
	if resp.StatusCode() != http.StatusBadRequest {
		t.Errorf("expected 400 from OnException handler, got %d", resp.StatusCode())
	}
}

func TestAdaptHooksOnTemplateResponse(t *testing.T) {
	templateModified := false
	hooks := Hooks{
		OnTemplateResponse: func(req *Request, resp *jangohttp.TemplateResponse) *jangohttp.TemplateResponse {
			templateModified = true
			resp.Status = http.StatusAccepted
			return resp
		},
	}

	engine := &mockTemplateEngine{
		templates: map[string]string{
			"test.html": "<h1>Hello!</h1>",
		},
	}

	mw := AdaptHooks(hooks)
	handler := mw(func(req *jangohttp.Request) jangohttp.Response {
		return jangohttp.NewTemplateResponse(engine, "test.html", nil)
	})

	req := jangohttp.WrapRequest(httptest.NewRequest("GET", "/", nil))
	resp := handler(req)

	if !templateModified {
		t.Error("expected OnTemplateResponse to be called")
	}
	if resp.StatusCode() != http.StatusAccepted {
		t.Errorf("expected 202, got %d", resp.StatusCode())
	}
}

func TestRecoveryMiddleware(t *testing.T) {
	panicFn := func(req *jangohttp.Request) jangohttp.Response {
		panic("test panic")
	}

	chain := Chain(RecoveryMiddleware)
	resp := chain(panicFn)(jangohttp.WrapRequest(httptest.NewRequest("GET", "/", nil)))

	if resp == nil {
		t.Fatal("expected non-nil response from RecoveryMiddleware")
	}
	if resp.StatusCode() != http.StatusInternalServerError {
		t.Errorf("expected 500 from RecoveryMiddleware, got %d", resp.StatusCode())
	}
}

func TestLoggingMiddleware(t *testing.T) {
	handlerCalled := false
	handler := func(req *jangohttp.Request) jangohttp.Response {
		handlerCalled = true
		return jangohttp.NewTextResponse("ok")
	}

	chain := Chain(LoggingMiddleware)
	resp := chain(jangohttp.ViewFunc(handler))(jangohttp.WrapRequest(httptest.NewRequest("GET", "/test", nil)))

	if !handlerCalled {
		t.Error("expected handler to be called")
	}
	if resp.StatusCode() != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode())
	}
}

func TestCommonMiddleware(t *testing.T) {
	handler := func(req *jangohttp.Request) jangohttp.Response {
		return jangohttp.NewTextResponse("ok")
	}

	chain := Chain(CommonMiddleware)
	resp := chain(jangohttp.ViewFunc(handler))(jangohttp.WrapRequest(httptest.NewRequest("GET", "/", nil)))

	rec := httptest.NewRecorder()
	resp.WriteTo(rec, httptest.NewRequest("GET", "/", nil))

	if rec.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Error("expected X-Content-Type-Options header")
	}
	if rec.Header().Get("X-Frame-Options") != "DENY" {
		t.Error("expected X-Frame-Options header")
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
