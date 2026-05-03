package views

import (
	"fmt"
	"net/http"
	"strings"

	jangohttp "github.com/iMerica/jango/http"
)

type TemplateView struct {
	TemplateName string
	ContextData  interface{}
	Engine       jangohttp.TemplateEngine
}

func (tv *TemplateView) DispatchGet(req *jangohttp.Request) jangohttp.Response {
	if tv.Engine == nil {
		return jangohttp.NewHTMLResponse(fmt.Sprintf("<html><body><h1>%s</h1></body></html>", tv.TemplateName))
	}
	return jangohttp.NewTemplateResponse(tv.Engine, tv.TemplateName, tv.ContextData)
}

func (tv *TemplateView) DispatchPost(req *jangohttp.Request) jangohttp.Response {
	return jangohttp.NewMethodNotAllowedResponse("POST not allowed")
}

func (tv *TemplateView) DispatchPut(req *jangohttp.Request) jangohttp.Response {
	return jangohttp.NewMethodNotAllowedResponse("PUT not allowed")
}

func (tv *TemplateView) DispatchPatch(req *jangohttp.Request) jangohttp.Response {
	return jangohttp.NewMethodNotAllowedResponse("PATCH not allowed")
}

func (tv *TemplateView) DispatchDelete(req *jangohttp.Request) jangohttp.Response {
	return jangohttp.NewMethodNotAllowedResponse("DELETE not allowed")
}

func (tv *TemplateView) DispatchHead(req *jangohttp.Request) jangohttp.Response {
	return jangohttp.NewTextResponseWithStatus("", http.StatusOK)
}

func (tv *TemplateView) DispatchOptions(req *jangohttp.Request) jangohttp.Response {
	resp := jangohttp.NewTextResponse("")
	resp.SetHeader("Allow", "GET, HEAD, OPTIONS")
	return resp
}

func NewTemplateView(engine jangohttp.TemplateEngine, templateName string, contextData interface{}) *TemplateView {
	return &TemplateView{
		TemplateName: templateName,
		ContextData:  contextData,
		Engine:       engine,
	}
}

type RedirectView struct {
	URL       string
	Permanent bool
}

func (rv *RedirectView) DispatchGet(req *jangohttp.Request) jangohttp.Response {
	url := rv.URL
	if url == "" {
		url = req.URL.Path
	}
	if rv.Permanent {
		return jangohttp.NewPermanentRedirectResponse(url)
	}
	return jangohttp.NewRedirectResponse(url)
}

func (rv *RedirectView) DispatchPost(req *jangohttp.Request) jangohttp.Response {
	return rv.DispatchGet(req)
}

func (rv *RedirectView) DispatchPut(req *jangohttp.Request) jangohttp.Response {
	return jangohttp.NewMethodNotAllowedResponse("PUT not allowed")
}

func (rv *RedirectView) DispatchPatch(req *jangohttp.Request) jangohttp.Response {
	return jangohttp.NewMethodNotAllowedResponse("PATCH not allowed")
}

func (rv *RedirectView) DispatchDelete(req *jangohttp.Request) jangohttp.Response {
	return jangohttp.NewMethodNotAllowedResponse("DELETE not allowed")
}

func (rv *RedirectView) DispatchHead(req *jangohttp.Request) jangohttp.Response {
	if rv.Permanent {
		return jangohttp.NewPermanentRedirectResponse(rv.URL)
	}
	return jangohttp.NewRedirectResponse(rv.URL)
}

func (rv *RedirectView) DispatchOptions(req *jangohttp.Request) jangohttp.Response {
	resp := jangohttp.NewTextResponse("")
	resp.SetHeader("Allow", "GET, POST, HEAD, OPTIONS")
	return resp
}

func NewRedirectView(url string, permanent bool) *RedirectView {
	return &RedirectView{
		URL:       url,
		Permanent: permanent,
	}
}

type ListView struct {
	TemplateName string
	Engine       jangohttp.TemplateEngine
	QuerysetFunc func() ([]interface{}, error)
	ObjectName   string
}

func (lv *ListView) DispatchGet(req *jangohttp.Request) jangohttp.Response {
	if lv.QuerysetFunc == nil {
		return jangohttp.NewInternalServerErrorResponse("no queryset function defined")
	}
	items, err := lv.QuerysetFunc()
	if err != nil {
		return jangohttp.NewInternalServerErrorResponse(err.Error())
	}
	objectName := lv.ObjectName
	if objectName == "" {
		objectName = "object_list"
	}
	contextData := map[string]interface{}{
		objectName: items,
	}
	if lv.Engine == nil {
		return jangohttp.NewJSONResponse(contextData)
	}
	return jangohttp.NewTemplateResponse(lv.Engine, lv.TemplateName, contextData)
}

func (lv *ListView) DispatchPost(req *jangohttp.Request) jangohttp.Response {
	return jangohttp.NewMethodNotAllowedResponse("POST not allowed")
}

func (lv *ListView) DispatchPut(req *jangohttp.Request) jangohttp.Response {
	return jangohttp.NewMethodNotAllowedResponse("PUT not allowed")
}

func (lv *ListView) DispatchPatch(req *jangohttp.Request) jangohttp.Response {
	return jangohttp.NewMethodNotAllowedResponse("PATCH not allowed")
}

func (lv *ListView) DispatchDelete(req *jangohttp.Request) jangohttp.Response {
	return jangohttp.NewMethodNotAllowedResponse("DELETE not allowed")
}

func (lv *ListView) DispatchHead(req *jangohttp.Request) jangohttp.Response {
	return jangohttp.NewTextResponseWithStatus("", http.StatusOK)
}

func (lv *ListView) DispatchOptions(req *jangohttp.Request) jangohttp.Response {
	resp := jangohttp.NewTextResponse("")
	resp.SetHeader("Allow", "GET, HEAD, OPTIONS")
	return resp
}

type DetailView struct {
	TemplateName string
	Engine       jangohttp.TemplateEngine
	GetObject    func(req *jangohttp.Request) (interface{}, error)
	ObjectName   string
}

func (dv *DetailView) DispatchGet(req *jangohttp.Request) jangohttp.Response {
	if dv.GetObject == nil {
		return jangohttp.NewInternalServerErrorResponse("no get function defined")
	}
	obj, err := dv.GetObject(req)
	if err != nil {
		return jangohttp.NewNotFoundResponse(err.Error())
	}
	objectName := dv.ObjectName
	if objectName == "" {
		objectName = "object"
	}
	contextData := map[string]interface{}{
		objectName: obj,
	}
	if dv.Engine == nil {
		return jangohttp.NewJSONResponse(contextData)
	}
	return jangohttp.NewTemplateResponse(dv.Engine, dv.TemplateName, contextData)
}

func (dv *DetailView) DispatchPost(req *jangohttp.Request) jangohttp.Response {
	return jangohttp.NewMethodNotAllowedResponse("POST not allowed")
}

func (dv *DetailView) DispatchPut(req *jangohttp.Request) jangohttp.Response {
	return jangohttp.NewMethodNotAllowedResponse("PUT not allowed")
}

func (dv *DetailView) DispatchPatch(req *jangohttp.Request) jangohttp.Response {
	return jangohttp.NewMethodNotAllowedResponse("PATCH not allowed")
}

func (dv *DetailView) DispatchDelete(req *jangohttp.Request) jangohttp.Response {
	return jangohttp.NewMethodNotAllowedResponse("DELETE not allowed")
}

func (dv *DetailView) DispatchHead(req *jangohttp.Request) jangohttp.Response {
	return jangohttp.NewTextResponseWithStatus("", http.StatusOK)
}

func (dv *DetailView) DispatchOptions(req *jangohttp.Request) jangohttp.Response {
	resp := jangohttp.NewTextResponse("")
	resp.SetHeader("Allow", "GET, HEAD, OPTIONS")
	return resp
}

type FormView struct {
	TemplateName string
	Engine       jangohttp.TemplateEngine
	FormClass    interface{}
	OnGet        func(req *jangohttp.Request) jangohttp.Response
	OnPost       func(req *jangohttp.Request) jangohttp.Response
}

func (fv *FormView) DispatchGet(req *jangohttp.Request) jangohttp.Response {
	if fv.OnGet != nil {
		return fv.OnGet(req)
	}
	contextData := map[string]interface{}{
		"form": fv.FormClass,
	}
	if fv.Engine == nil {
		return jangohttp.NewJSONResponse(contextData)
	}
	return jangohttp.NewTemplateResponse(fv.Engine, fv.TemplateName, contextData)
}

func (fv *FormView) DispatchPost(req *jangohttp.Request) jangohttp.Response {
	if fv.OnPost != nil {
		return fv.OnPost(req)
	}
	return jangohttp.NewMethodNotAllowedResponse("POST not handled")
}

func (fv *FormView) DispatchPut(req *jangohttp.Request) jangohttp.Response {
	return jangohttp.NewMethodNotAllowedResponse("PUT not allowed")
}

func (fv *FormView) DispatchPatch(req *jangohttp.Request) jangohttp.Response {
	return jangohttp.NewMethodNotAllowedResponse("PATCH not allowed")
}

func (fv *FormView) DispatchDelete(req *jangohttp.Request) jangohttp.Response {
	return jangohttp.NewMethodNotAllowedResponse("DELETE not allowed")
}

func (fv *FormView) DispatchHead(req *jangohttp.Request) jangohttp.Response {
	return jangohttp.NewTextResponseWithStatus("", http.StatusOK)
}

func (fv *FormView) DispatchOptions(req *jangohttp.Request) jangohttp.Response {
	resp := jangohttp.NewTextResponse("")
	resp.SetHeader("Allow", "GET, POST, HEAD, OPTIONS")
	return resp
}

type CreateView struct {
	TemplateName string
	Engine       jangohttp.TemplateEngine
	OnCreate     func(req *jangohttp.Request) jangohttp.Response
}

func (cv *CreateView) DispatchGet(req *jangohttp.Request) jangohttp.Response {
	contextData := map[string]interface{}{
		"action": "create",
	}
	if cv.Engine == nil {
		return jangohttp.NewJSONResponse(contextData)
	}
	return jangohttp.NewTemplateResponse(cv.Engine, cv.TemplateName, contextData)
}

func (cv *CreateView) DispatchPost(req *jangohttp.Request) jangohttp.Response {
	if cv.OnCreate != nil {
		return cv.OnCreate(req)
	}
	return jangohttp.NewMethodNotAllowedResponse("POST not handled")
}

func (cv *CreateView) DispatchPut(req *jangohttp.Request) jangohttp.Response {
	return jangohttp.NewMethodNotAllowedResponse("PUT not allowed")
}

func (cv *CreateView) DispatchPatch(req *jangohttp.Request) jangohttp.Response {
	return jangohttp.NewMethodNotAllowedResponse("PATCH not allowed")
}

func (cv *CreateView) DispatchDelete(req *jangohttp.Request) jangohttp.Response {
	return jangohttp.NewMethodNotAllowedResponse("DELETE not allowed")
}

func (cv *CreateView) DispatchHead(req *jangohttp.Request) jangohttp.Response {
	return jangohttp.NewTextResponseWithStatus("", http.StatusOK)
}

func (cv *CreateView) DispatchOptions(req *jangohttp.Request) jangohttp.Response {
	resp := jangohttp.NewTextResponse("")
	resp.SetHeader("Allow", "GET, POST, HEAD, OPTIONS")
	return resp
}

type UpdateView struct {
	TemplateName string
	Engine       jangohttp.TemplateEngine
	GetObject    func(req *jangohttp.Request) (interface{}, error)
	OnUpdate     func(req *jangohttp.Request) jangohttp.Response
}

func (uv *UpdateView) DispatchGet(req *jangohttp.Request) jangohttp.Response {
	if uv.GetObject != nil {
		obj, err := uv.GetObject(req)
		if err != nil {
			return jangohttp.NewNotFoundResponse(err.Error())
		}
		contextData := map[string]interface{}{
			"object": obj,
			"action": "update",
		}
		if uv.Engine == nil {
			return jangohttp.NewJSONResponse(contextData)
		}
		return jangohttp.NewTemplateResponse(uv.Engine, uv.TemplateName, contextData)
	}
	return jangohttp.NewInternalServerErrorResponse("no get function defined")
}

func (uv *UpdateView) DispatchPost(req *jangohttp.Request) jangohttp.Response {
	if uv.OnUpdate != nil {
		return uv.OnUpdate(req)
	}
	return jangohttp.NewMethodNotAllowedResponse("POST not handled")
}

func (uv *UpdateView) DispatchPut(req *jangohttp.Request) jangohttp.Response {
	return jangohttp.NewMethodNotAllowedResponse("PUT not allowed")
}

func (uv *UpdateView) DispatchPatch(req *jangohttp.Request) jangohttp.Response {
	return jangohttp.NewMethodNotAllowedResponse("PATCH not allowed")
}

func (uv *UpdateView) DispatchDelete(req *jangohttp.Request) jangohttp.Response {
	return jangohttp.NewMethodNotAllowedResponse("DELETE not allowed")
}

func (uv *UpdateView) DispatchHead(req *jangohttp.Request) jangohttp.Response {
	return jangohttp.NewTextResponseWithStatus("", http.StatusOK)
}

func (uv *UpdateView) DispatchOptions(req *jangohttp.Request) jangohttp.Response {
	resp := jangohttp.NewTextResponse("")
	resp.SetHeader("Allow", "GET, POST, HEAD, OPTIONS")
	return resp
}

type DeleteView struct {
	TemplateName string
	Engine       jangohttp.TemplateEngine
	GetObject    func(req *jangohttp.Request) (interface{}, error)
	OnDelete     func(req *jangohttp.Request) jangohttp.Response
	SuccessURL   string
}

func (dv *DeleteView) DispatchGet(req *jangohttp.Request) jangohttp.Response {
	if dv.GetObject != nil {
		obj, err := dv.GetObject(req)
		if err != nil {
			return jangohttp.NewNotFoundResponse(err.Error())
		}
		contextData := map[string]interface{}{
			"object": obj,
			"action": "delete",
		}
		if dv.Engine == nil {
			return jangohttp.NewJSONResponse(contextData)
		}
		return jangohttp.NewTemplateResponse(dv.Engine, dv.TemplateName, contextData)
	}
	return jangohttp.NewInternalServerErrorResponse("no get function defined")
}

func (dv *DeleteView) DispatchPost(req *jangohttp.Request) jangohttp.Response {
	if dv.OnDelete != nil {
		return dv.OnDelete(req)
	}
	if dv.SuccessURL != "" {
		return jangohttp.NewRedirectResponse(dv.SuccessURL)
	}
	return jangohttp.NewMethodNotAllowedResponse("POST not handled")
}

func (dv *DeleteView) DispatchPut(req *jangohttp.Request) jangohttp.Response {
	return jangohttp.NewMethodNotAllowedResponse("PUT not allowed")
}

func (dv *DeleteView) DispatchPatch(req *jangohttp.Request) jangohttp.Response {
	return jangohttp.NewMethodNotAllowedResponse("PATCH not allowed")
}

func (dv *DeleteView) DispatchDelete(req *jangohttp.Request) jangohttp.Response {
	return jangohttp.NewMethodNotAllowedResponse("DELETE not allowed")
}

func (dv *DeleteView) DispatchHead(req *jangohttp.Request) jangohttp.Response {
	return jangohttp.NewTextResponseWithStatus("", http.StatusOK)
}

func (dv *DeleteView) DispatchOptions(req *jangohttp.Request) jangohttp.Response {
	resp := jangohttp.NewTextResponse("")
	resp.SetHeader("Allow", "GET, POST, HEAD, OPTIONS")
	return resp
}

func AllowedMethods(methods ...string) jangohttp.ViewFunc {
	methodSet := make(map[string]bool, len(methods))
	for _, m := range methods {
		methodSet[strings.ToUpper(m)] = true
	}
	allowHeader := strings.Join(methods, ", ")

	return func(req *jangohttp.Request) jangohttp.Response {
		if !methodSet[req.Method] {
			resp := jangohttp.NewMethodNotAllowedResponse("Method not allowed")
			resp.SetHeader("Allow", allowHeader)
			return resp
		}
		return nil
	}
}
