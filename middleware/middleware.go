package middleware

import (
	"fmt"
	"log"

	jangohttp "github.com/iMerica/jango/http"
)

type ViewFunc = jangohttp.ViewFunc

type Response = jangohttp.Response

type Request = jangohttp.Request

type MiddlewareFunc func(next ViewFunc) ViewFunc

func Chain(middlewares ...MiddlewareFunc) MiddlewareFunc {
	return func(final ViewFunc) ViewFunc {
		handler := final
		for i := len(middlewares) - 1; i >= 0; i-- {
			handler = middlewares[i](handler)
		}
		return handler
	}
}

type Hooks struct {
	OnRequest          func(*Request) (*Request, Response)
	OnView             func(*Request, ViewFunc) Response
	OnResponse         func(*Request, Response) Response
	OnException        func(*Request, error) Response
	OnTemplateResponse func(*Request, *jangohttp.TemplateResponse) *jangohttp.TemplateResponse
}

func AdaptHooks(h Hooks) MiddlewareFunc {
	return func(next ViewFunc) ViewFunc {
		return func(req *Request) Response {
			if h.OnRequest != nil {
				updatedReq, shortCircuitResp := h.OnRequest(req)
				if shortCircuitResp != nil {
					return shortCircuitResp
				}
				if updatedReq != nil {
					req = updatedReq
				}
			}

			var resp Response
			var err error

			func() {
				defer func() {
					if r := recover(); r != nil {
						switch v := r.(type) {
						case error:
							err = v
						case string:
							err = fmt.Errorf("%s", v)
						default:
							err = fmt.Errorf("%v", v)
						}
					}
				}()

				if h.OnView != nil {
					resp = h.OnView(req, next)
				} else {
					resp = next(req)
				}
			}()

			if err != nil {
				if h.OnException != nil {
					recoveryResp := h.OnException(req, err)
					if recoveryResp != nil {
						resp = recoveryResp
						err = nil
					} else {
						resp = jangohttp.NewInternalServerErrorResponse(err.Error())
					}
				} else {
					resp = jangohttp.NewInternalServerErrorResponse(err.Error())
				}
			}

			if tmplResp, ok := resp.(*jangohttp.TemplateResponse); ok && h.OnTemplateResponse != nil {
				tmplResp = h.OnTemplateResponse(req, tmplResp)
				if tmplResp != nil {
					resp = tmplResp
				}
			}

			if h.OnResponse != nil {
				resp = h.OnResponse(req, resp)
			}

			return resp
		}
	}
}

func RecoveryMiddleware(next ViewFunc) ViewFunc {
	return func(req *Request) (resp Response) {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("[jango] panic recovered: %v", r)
				resp = jangohttp.NewInternalServerErrorResponse(fmt.Sprintf("Internal Server Error: %v", r))
			}
		}()
		return next(req)
	}
}

func LoggingMiddleware(next ViewFunc) ViewFunc {
	return func(req *Request) Response {
		log.Printf("[jango] %s %s", req.Method, req.URL.Path)
		resp := next(req)
		log.Printf("[jango] %s %s -> %d", req.Method, req.URL.Path, resp.StatusCode())
		return resp
	}
}

func CommonMiddleware(next ViewFunc) ViewFunc {
	return func(req *Request) Response {
		resp := next(req)

		if resp != nil {
			baseResp, ok := resp.(interface{ SetHeader(string, string) })
			if ok {
				baseResp.SetHeader("X-Content-Type-Options", "nosniff")
				baseResp.SetHeader("X-Frame-Options", "DENY")
			}
		}

		return resp
	}
}
