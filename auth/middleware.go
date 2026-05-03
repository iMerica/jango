package auth

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/iMerica/jango/conf"
	jangohttp "github.com/iMerica/jango/http"
)

func AuthMiddleware(next jangohttp.ViewFunc) jangohttp.ViewFunc {
	return func(req *jangohttp.Request) jangohttp.Response {
		chain := NewDefaultBackendChain()

		session := req.Session
		if session != nil {
			if userID, ok := sessionGetUserID(session, "_auth_user_id"); ok {
				id, err := strconv.ParseInt(userID, 10, 64)
				if err == nil {
					user, err := chain.GetUser(req.Context(), id)
					if err == nil && user != nil {
						req.SetUser(user)
						return next(req)
					}
				}
			}
		}

		req.SetUser(Anonymous)
		return next(req)
	}
}

func sessionGetUserID(session interface{}, key string) (string, bool) {
	type getter interface {
		Get(key string) (interface{}, bool)
	}
	if s, ok := session.(getter); ok {
		val, ok := s.Get(key)
		if !ok {
			return "", false
		}
		if str, ok := val.(string); ok {
			return str, true
		}
		return fmt.Sprintf("%v", val), true
	}
	return "", false
}

func BasicAuthMiddleware(next jangohttp.ViewFunc) jangohttp.ViewFunc {
	return func(req *jangohttp.Request) jangohttp.Response {
		username, password, ok := req.BasicAuth()
		if !ok {
			resp := jangohttp.NewUnauthorizedResponse("Authentication required")
			resp.SetHeader("WWW-Authenticate", `Basic realm="Restricted"`)
			return resp
		}

		chain := NewDefaultBackendChain()
		user, err := chain.Authenticate(req.Context(), username, password)
		if err != nil || user == nil {
			resp := jangohttp.NewUnauthorizedResponse("Invalid credentials")
			resp.SetHeader("WWW-Authenticate", `Basic realm="Restricted"`)
			return resp
		}

		req.SetUser(user)
		return next(req)
	}
}

func HostValidationMiddleware(allowedHosts []string) func(jangohttp.ViewFunc) jangohttp.ViewFunc {
	return func(next jangohttp.ViewFunc) jangohttp.ViewFunc {
		return func(req *jangohttp.Request) jangohttp.Response {
			host := req.Host
			if strings.Contains(host, ":") {
				host = strings.Split(host, ":")[0]
			}

			if !isAllowedHost(host, allowedHosts) {
				return jangohttp.NewForbiddenResponse("Invalid Host header")
			}

			return next(req)
		}
	}
}

func isAllowedHost(host string, allowedHosts []string) bool {
	for _, allowed := range allowedHosts {
		if allowed == "*" {
			return true
		}
		if strings.HasPrefix(allowed, ".") {
			if host == allowed[1:] || strings.HasSuffix(host, allowed) {
				return true
			}
		}
		if host == allowed {
			return true
		}
	}
	return false
}

func SecurityMiddleware(settings *conf.Settings) func(jangohttp.ViewFunc) jangohttp.ViewFunc {
	return func(next jangohttp.ViewFunc) jangohttp.ViewFunc {
		return func(req *jangohttp.Request) jangohttp.Response {
			resp := next(req)

			if resp == nil {
				return resp
			}

			if settings.Secure.ContentTypeNoSniff {
				setResponseHeader(resp, "X-Content-Type-Options", "nosniff")
			}

			if settings.Secure.BrowserXSSFilter {
				setResponseHeader(resp, "X-XSS-Protection", "1; mode=block")
			}

			if settings.Secure.ReferrerPolicy != "" {
				setResponseHeader(resp, "Referrer-Policy", settings.Secure.ReferrerPolicy)
			}

			if isSecureRequest(req) && settings.Secure.HSTSSeconds > 0 {
				value := "max-age=" + strconv.Itoa(settings.Secure.HSTSSeconds)
				if settings.Secure.HSTSIncludeSubdomains {
					value += "; includeSubDomains"
				}
				if settings.Secure.HSTSPreload {
					value += "; preload"
				}
				setResponseHeader(resp, "Strict-Transport-Security", value)
			}

			if settings.Secure.SSLRedirect && !isSecureRequest(req) {
				host := req.Host
				if settings.Secure.SSLHost != "" {
					host = settings.Secure.SSLHost
				}
				target := "https://" + host + req.URL.Path
				if req.URL.RawQuery != "" {
					target += "?" + req.URL.RawQuery
				}
				return jangohttp.NewRedirectResponse(target)
			}

			return resp
		}
	}
}

func XFrameOptionsMiddleware(frameOption string) func(jangohttp.ViewFunc) jangohttp.ViewFunc {
	if frameOption == "" {
		frameOption = "DENY"
	}
	return func(next jangohttp.ViewFunc) jangohttp.ViewFunc {
		return func(req *jangohttp.Request) jangohttp.Response {
			resp := next(req)
			if resp != nil {
				setResponseHeader(resp, "X-Frame-Options", frameOption)
			}
			return resp
		}
	}
}

func isSecureRequest(req *jangohttp.Request) bool {
	if req.URL.Scheme == "https" {
		return true
	}
	if req.Header.Get("X-Forwarded-Proto") == "https" {
		return true
	}
	return false
}

func setResponseHeader(resp jangohttp.Response, key, value string) {
	type headerSetter interface {
		SetHeader(string, string)
	}
	if h, ok := resp.(headerSetter); ok {
		h.SetHeader(key, value)
	}
}
