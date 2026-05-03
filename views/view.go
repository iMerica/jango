package views

import (
	"net/http"

	jangohttp "github.com/iMerica/jango/http"
	"github.com/iMerica/jango/urls"
)

type ViewDispatcher interface {
	DispatchGet(req *jangohttp.Request) jangohttp.Response
	DispatchPost(req *jangohttp.Request) jangohttp.Response
	DispatchPut(req *jangohttp.Request) jangohttp.Response
	DispatchPatch(req *jangohttp.Request) jangohttp.Response
	DispatchDelete(req *jangohttp.Request) jangohttp.Response
	DispatchHead(req *jangohttp.Request) jangohttp.Response
	DispatchOptions(req *jangohttp.Request) jangohttp.Response
}

type View struct {
	TemplateName string
	ContentType  string
}

func (v *View) DispatchGet(req *jangohttp.Request) jangohttp.Response {
	return jangohttp.NewMethodNotAllowedResponse("GET not allowed")
}

func (v *View) DispatchPost(req *jangohttp.Request) jangohttp.Response {
	return jangohttp.NewMethodNotAllowedResponse("POST not allowed")
}

func (v *View) DispatchPut(req *jangohttp.Request) jangohttp.Response {
	return jangohttp.NewMethodNotAllowedResponse("PUT not allowed")
}

func (v *View) DispatchPatch(req *jangohttp.Request) jangohttp.Response {
	return jangohttp.NewMethodNotAllowedResponse("PATCH not allowed")
}

func (v *View) DispatchDelete(req *jangohttp.Request) jangohttp.Response {
	return jangohttp.NewMethodNotAllowedResponse("DELETE not allowed")
}

func (v *View) DispatchHead(req *jangohttp.Request) jangohttp.Response {
	return jangohttp.NewTextResponseWithStatus("", http.StatusOK)
}

func (v *View) DispatchOptions(req *jangohttp.Request) jangohttp.Response {
	resp := jangohttp.NewTextResponse("")
	resp.SetHeader("Allow", "GET, POST, PUT, PATCH, DELETE, HEAD, OPTIONS")
	return resp
}

func AsView(d ViewDispatcher) jangohttp.ViewFunc {
	return func(req *jangohttp.Request) jangohttp.Response {
		method := req.Method
		switch method {
		case http.MethodGet:
			return d.DispatchGet(req)
		case http.MethodPost:
			return d.DispatchPost(req)
		case http.MethodPut:
			return d.DispatchPut(req)
		case http.MethodPatch:
			return d.DispatchPatch(req)
		case http.MethodDelete:
			return d.DispatchDelete(req)
		case http.MethodHead:
			return d.DispatchHead(req)
		case http.MethodOptions:
			return d.DispatchOptions(req)
		default:
			return jangohttp.NewMethodNotAllowedResponse("Method not allowed")
		}
	}
}

type FuncView func(req *jangohttp.Request) jangohttp.Response

func (fn FuncView) AsView() jangohttp.ViewFunc {
	return jangohttp.ViewFunc(fn)
}

type Viewable interface {
	AsView() jangohttp.ViewFunc
}

func RegisterView(urlpatterns []urls.Pattern, pattern string, viewable Viewable, name string) []urls.Pattern {
	return append(urlpatterns, urls.Path(pattern, viewable.AsView(), name))
}

func RegisterFuncView(urlpatterns []urls.Pattern, pattern string, fn jangohttp.ViewFunc, name string) []urls.Pattern {
	return append(urlpatterns, urls.Path(pattern, fn, name))
}
