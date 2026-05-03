package server

import (
	"log"
	"net/http"

	jangohttp "github.com/iMerica/jango/http"
	"github.com/iMerica/jango/middleware"
	"github.com/iMerica/jango/urls"
)

type Server struct {
	Resolver    *urls.Resolver
	Middlewares []middleware.MiddlewareFunc
}

func NewServer(resolver *urls.Resolver, middlewares ...middleware.MiddlewareFunc) *Server {
	return &Server{
		Resolver:    resolver,
		Middlewares: middlewares,
	}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	frameworkReq := jangohttp.WrapRequest(r)

	resp := s.handleRequest(frameworkReq)

	if streamResp, ok := resp.(*jangohttp.StreamResponse); ok {
		if err := streamResp.WriteTo(w, r); err != nil {
			log.Printf("stream response error: %v", err)
		}
		return
	}

	if err := resp.WriteTo(w, r); err != nil {
		log.Printf("response write error: %v", err)
	}
}

func (s *Server) handleRequest(req *jangohttp.Request) jangohttp.Response {
	handler := s.resolveHandler(req)
	if handler == nil {
		return jangohttp.NewNotFoundResponse("Page not found")
	}

	finalHandler := jangohttp.ViewFunc(handler)

	chain := middleware.Chain(s.Middlewares...)
	return chain(finalHandler)(req)
}

func (s *Server) resolveHandler(req *jangohttp.Request) jangohttp.ViewFunc {
	match, err := s.Resolver.Resolve(req.Request.URL.Path)
	if err != nil {
		return nil
	}
	for k, v := range match.Params {
		req.Params[k] = v
	}
	if match.Handler != nil {
		if vf, ok := match.Handler.(jangohttp.ViewFunc); ok {
			return vf
		}
		if h, ok := match.Handler.(http.Handler); ok {
			return jangohttp.WrapHTTPHandler(h)
		}
	}
	return nil
}

func NewHandler(urlconf []urls.Pattern, middlewares ...middleware.MiddlewareFunc) http.Handler {
	resolver := urls.NewResolver(urlconf)
	return NewServer(resolver, middlewares...)
}
