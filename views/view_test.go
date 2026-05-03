package views

import (
	"net/http"
	"net/http/httptest"
	"testing"

	jangohttp "github.com/iMerica/jango/http"
	"github.com/iMerica/jango/urls"
)

func TestViewAsView(t *testing.T) {
	v := &View{}
	handler := AsView(v)

	req := jangohttp.WrapRequest(httptest.NewRequest("GET", "/", nil))
	resp := handler(req)
	if resp == nil {
		t.Fatal("expected non-nil response")
	}
	if resp.StatusCode() != http.StatusMethodNotAllowed {
		t.Errorf("expected 405 for GET on default View, got %d", resp.StatusCode())
	}
}

func TestViewMethodDispatch(t *testing.T) {
	v := &View{}
	handler := AsView(v)

	methods := []string{"GET", "POST", "PUT", "PATCH", "DELETE"}
	for _, method := range methods {
		req := jangohttp.WrapRequest(httptest.NewRequest(method, "/", nil))
		resp := handler(req)
		if resp.StatusCode() != http.StatusMethodNotAllowed {
			t.Errorf("expected 405 for %s on default View, got %d", method, resp.StatusCode())
		}
	}
}

func TestViewOPTIONSMethod(t *testing.T) {
	v := &View{}
	handler := AsView(v)

	req := jangohttp.WrapRequest(httptest.NewRequest("OPTIONS", "/", nil))
	resp := handler(req)

	rec := httptest.NewRecorder()
	resp.WriteTo(rec, httptest.NewRequest("OPTIONS", "/", nil))
	allow := rec.Header().Get("Allow")
	if allow == "" {
		t.Error("expected Allow header in OPTIONS response")
	}
}

func TestFuncView(t *testing.T) {
	fv := FuncView(func(req *jangohttp.Request) jangohttp.Response {
		return jangohttp.NewTextResponse("funcView response")
	})

	handler := fv.AsView()
	req := jangohttp.WrapRequest(httptest.NewRequest("GET", "/", nil))
	resp := handler(req)

	if resp.StatusCode() != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode())
	}

	rec := httptest.NewRecorder()
	resp.WriteTo(rec, httptest.NewRequest("GET", "/", nil))
	if rec.Body.String() != "funcView response" {
		t.Errorf("expected 'funcView response', got %q", rec.Body.String())
	}
}

func TestRegisterView(t *testing.T) {
	tv := &TemplateView{
		TemplateName: "test.html",
	}
	var patterns []urls.Pattern
	patterns = RegisterFuncView(patterns, "/test/", AsView(tv), "test-view")

	if len(patterns) != 1 {
		t.Fatalf("expected 1 pattern, got %d", len(patterns))
	}
	if patterns[0].Name != "test-view" {
		t.Errorf("expected name=test-view, got %s", patterns[0].Name)
	}
}

func TestRegisterFuncView(t *testing.T) {
	handler := jangohttp.ViewFunc(func(req *jangohttp.Request) jangohttp.Response {
		return jangohttp.NewTextResponse("hello")
	})
	var patterns []urls.Pattern
	patterns = RegisterFuncView(patterns, "/hello/", handler, "hello")

	if len(patterns) != 1 {
		t.Fatalf("expected 1 pattern, got %d", len(patterns))
	}
	if patterns[0].Name != "hello" {
		t.Errorf("expected name=hello, got %s", patterns[0].Name)
	}
}

func TestViewableViewableInterface(t *testing.T) {
	// Verify TemplateView satisfies ViewDispatcher
	var _ ViewDispatcher = &TemplateView{}
	var _ ViewDispatcher = &RedirectView{}
	var _ ViewDispatcher = &ListView{}
	var _ ViewDispatcher = &DetailView{}
	var _ ViewDispatcher = &FormView{}
	var _ ViewDispatcher = &CreateView{}
	var _ ViewDispatcher = &UpdateView{}
	var _ ViewDispatcher = &DeleteView{}
}
