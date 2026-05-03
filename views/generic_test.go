package views

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	jangohttp "github.com/iMerica/jango/http"
)

type mockGenericTemplateEngine struct {
	templates map[string]string
}

func (e *mockGenericTemplateEngine) ExecuteTemplate(w io.Writer, name string, data interface{}) error {
	tmpl, ok := e.templates[name]
	if !ok {
		return io.EOF
	}
	_, err := w.Write([]byte(tmpl))
	return err
}

func TestTemplateViewGet(t *testing.T) {
	engine := &mockGenericTemplateEngine{
		templates: map[string]string{
			"test.html": "<h1>Hello, World!</h1>",
		},
	}

	tv := &TemplateView{
		TemplateName: "test.html",
		Engine:       engine,
		ContextData:  map[string]interface{}{"Name": "World"},
	}

	handler := AsView(tv)
	req := jangohttp.WrapRequest(httptest.NewRequest("GET", "/", nil))
	resp := handler(req)

	if resp.StatusCode() != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode())
	}

	rec := httptest.NewRecorder()
	resp.WriteTo(rec, httptest.NewRequest("GET", "/", nil))
	if rec.Body.String() != "<h1>Hello, World!</h1>" {
		t.Errorf("expected rendered template, got %q", rec.Body.String())
	}
}

func TestTemplateViewNoEngine(t *testing.T) {
	tv := &TemplateView{
		TemplateName: "test.html",
	}

	handler := AsView(tv)
	req := jangohttp.WrapRequest(httptest.NewRequest("GET", "/", nil))
	resp := handler(req)

	if resp.StatusCode() != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode())
	}

	rec := httptest.NewRecorder()
	resp.WriteTo(rec, httptest.NewRequest("GET", "/", nil))
	body := rec.Body.String()
	if body != "<html><body><h1>test.html</h1></body></html>" {
		t.Errorf("expected fallback HTML, got %q", body)
	}
}

func TestNewTemplateView(t *testing.T) {
	engine := &mockGenericTemplateEngine{
		templates: map[string]string{
			"greet.html": "Hello!",
		},
	}
	tv := NewTemplateView(engine, "greet.html", nil)
	if tv.TemplateName != "greet.html" {
		t.Errorf("expected template name 'greet.html', got %s", tv.TemplateName)
	}
}

func TestRedirectViewTemporary(t *testing.T) {
	rv := &RedirectView{
		URL:       "/new-location/",
		Permanent: false,
	}

	handler := AsView(rv)
	req := jangohttp.WrapRequest(httptest.NewRequest("GET", "/", nil))
	resp := handler(req)

	if resp.StatusCode() != http.StatusFound {
		t.Errorf("expected 302, got %d", resp.StatusCode())
	}

	rec := httptest.NewRecorder()
	resp.WriteTo(rec, httptest.NewRequest("GET", "/", nil))
	if rec.Header().Get("Location") != "/new-location/" {
		t.Errorf("expected Location /new-location/, got %s", rec.Header().Get("Location"))
	}
}

func TestRedirectViewPermanent(t *testing.T) {
	rv := &RedirectView{
		URL:       "/permanent/",
		Permanent: true,
	}

	handler := AsView(rv)
	req := jangohttp.WrapRequest(httptest.NewRequest("GET", "/", nil))
	resp := handler(req)

	if resp.StatusCode() != http.StatusMovedPermanently {
		t.Errorf("expected 301, got %d", resp.StatusCode())
	}
}

func TestNewRedirectView(t *testing.T) {
	rv := NewRedirectView("/go-here/", false)
	if rv.URL != "/go-here/" {
		t.Errorf("expected URL /go-here/, got %s", rv.URL)
	}
	if rv.Permanent != false {
		t.Error("expected Permanent=false")
	}

	rvPerm := NewRedirectView("/permanent/", true)
	if rvPerm.Permanent != true {
		t.Error("expected Permanent=true")
	}
}

func TestListViewGet(t *testing.T) {
	items := []interface{}{"item1", "item2", "item3"}
	lv := &ListView{
		TemplateName: "list.html",
		QuerysetFunc: func() ([]interface{}, error) {
			return items, nil
		},
		ObjectName: "items",
	}

	handler := AsView(lv)
	req := jangohttp.WrapRequest(httptest.NewRequest("GET", "/", nil))
	resp := handler(req)

	if resp.StatusCode() != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode())
	}

	rec := httptest.NewRecorder()
	resp.WriteTo(rec, httptest.NewRequest("GET", "/", nil))
	if rec.Header().Get("Content-Type") != "application/json" {
		t.Errorf("expected application/json, got %s", rec.Header().Get("Content-Type"))
	}
}

func TestListViewNoQueryset(t *testing.T) {
	lv := &ListView{
		TemplateName: "list.html",
	}

	handler := AsView(lv)
	req := jangohttp.WrapRequest(httptest.NewRequest("GET", "/", nil))
	resp := handler(req)

	if resp.StatusCode() != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", resp.StatusCode())
	}
}

func TestListViewWithTemplateEngine(t *testing.T) {
	engine := &mockGenericTemplateEngine{
		templates: map[string]string{
			"list.html": "<ul><li>items</li></ul>",
		},
	}

	lv := &ListView{
		TemplateName: "list.html",
		Engine:       engine,
		QuerysetFunc: func() ([]interface{}, error) {
			return []interface{}{"a", "b"}, nil
		},
		ObjectName: "items",
	}

	handler := AsView(lv)
	req := jangohttp.WrapRequest(httptest.NewRequest("GET", "/", nil))
	resp := handler(req)

	if resp.StatusCode() != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode())
	}

	rec := httptest.NewRecorder()
	resp.WriteTo(rec, httptest.NewRequest("GET", "/", nil))
	if rec.Body.String() != "<ul><li>items</li></ul>" {
		t.Errorf("expected rendered template, got %q", rec.Body.String())
	}
}

func TestDetailViewGet(t *testing.T) {
	dv := &DetailView{
		TemplateName: "detail.html",
		GetObject: func(req *jangohttp.Request) (interface{}, error) {
			return map[string]string{"name": "test"}, nil
		},
		ObjectName: "object",
	}

	handler := AsView(dv)
	req := jangohttp.WrapRequest(httptest.NewRequest("GET", "/", nil))
	resp := handler(req)

	if resp.StatusCode() != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode())
	}
}

func TestDetailViewNotFound(t *testing.T) {
	dv := &DetailView{
		GetObject: func(req *jangohttp.Request) (interface{}, error) {
			return nil, http.ErrNoCookie
		},
	}

	handler := AsView(dv)
	req := jangohttp.WrapRequest(httptest.NewRequest("GET", "/", nil))
	resp := handler(req)

	if resp.StatusCode() != http.StatusNotFound {
		t.Errorf("expected 404, got %d", resp.StatusCode())
	}
}

func TestFormViewGet(t *testing.T) {
	fv := &FormView{
		TemplateName: "form.html",
		FormClass:    nil,
	}

	handler := AsView(fv)
	req := jangohttp.WrapRequest(httptest.NewRequest("GET", "/", nil))
	resp := handler(req)

	if resp.StatusCode() != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode())
	}
}

func TestFormViewCustomGet(t *testing.T) {
	fv := &FormView{
		OnGet: func(req *jangohttp.Request) jangohttp.Response {
			return jangohttp.NewTextResponse("custom form get")
		},
	}

	handler := AsView(fv)
	req := jangohttp.WrapRequest(httptest.NewRequest("GET", "/", nil))
	resp := handler(req)

	rec := httptest.NewRecorder()
	resp.WriteTo(rec, httptest.NewRequest("GET", "/", nil))
	if rec.Body.String() != "custom form get" {
		t.Errorf("expected 'custom form get', got %q", rec.Body.String())
	}
}

func TestFormViewPostNotHandled(t *testing.T) {
	fv := &FormView{}

	handler := AsView(fv)
	req := jangohttp.WrapRequest(httptest.NewRequest("POST", "/", nil))
	resp := handler(req)

	if resp.StatusCode() != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", resp.StatusCode())
	}
}

func TestCreateViewGet(t *testing.T) {
	cv := &CreateView{
		TemplateName: "create.html",
	}

	handler := AsView(cv)
	req := jangohttp.WrapRequest(httptest.NewRequest("GET", "/", nil))
	resp := handler(req)

	if resp.StatusCode() != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode())
	}
}

func TestCreateViewPost(t *testing.T) {
	cv := &CreateView{
		OnCreate: func(req *jangohttp.Request) jangohttp.Response {
			return jangohttp.NewRedirectResponse("/created/")
		},
	}

	handler := AsView(cv)
	req := jangohttp.WrapRequest(httptest.NewRequest("POST", "/", nil))
	resp := handler(req)

	rec := httptest.NewRecorder()
	resp.WriteTo(rec, httptest.NewRequest("POST", "/", nil))
	if rec.Code != http.StatusFound {
		t.Errorf("expected 302, got %d", rec.Code)
	}
}

func TestUpdateViewGet(t *testing.T) {
	uv := &UpdateView{
		GetObject: func(req *jangohttp.Request) (interface{}, error) {
			return map[string]string{"name": "item"}, nil
		},
		TemplateName: "update.html",
	}

	handler := AsView(uv)
	req := jangohttp.WrapRequest(httptest.NewRequest("GET", "/", nil))
	resp := handler(req)

	if resp.StatusCode() != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode())
	}
}

func TestUpdateViewPost(t *testing.T) {
	uv := &UpdateView{
		OnUpdate: func(req *jangohttp.Request) jangohttp.Response {
			return jangohttp.NewRedirectResponse("/updated/")
		},
	}

	handler := AsView(uv)
	req := jangohttp.WrapRequest(httptest.NewRequest("POST", "/", nil))
	resp := handler(req)

	if resp.StatusCode() != http.StatusFound {
		t.Errorf("expected 302, got %d", resp.StatusCode())
	}
}

func TestDeleteViewGet(t *testing.T) {
	dv := &DeleteView{
		GetObject: func(req *jangohttp.Request) (interface{}, error) {
			return map[string]string{"name": "item"}, nil
		},
		TemplateName: "delete.html",
	}

	handler := AsView(dv)
	req := jangohttp.WrapRequest(httptest.NewRequest("GET", "/", nil))
	resp := handler(req)

	if resp.StatusCode() != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode())
	}
}

func TestDeleteViewPost(t *testing.T) {
	dv := &DeleteView{
		SuccessURL: "/deleted/",
	}

	handler := AsView(dv)
	req := jangohttp.WrapRequest(httptest.NewRequest("POST", "/", nil))
	resp := handler(req)

	if resp.StatusCode() != http.StatusFound {
		t.Errorf("expected 302, got %d", resp.StatusCode())
	}
}

func TestAllowedMethods(t *testing.T) {
	handler := AllowedMethods("GET", "POST")

	req := jangohttp.WrapRequest(httptest.NewRequest("DELETE", "/", nil))
	resp := handler(req)
	if resp == nil {
		t.Fatal("expected non-nil response for disallowed method")
	}
	if resp.StatusCode() != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", resp.StatusCode())
	}
}
