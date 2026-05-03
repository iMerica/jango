package security

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"net/http"
	"strings"

	jangohttp "github.com/iMerica/jango/http"
	"github.com/iMerica/jango/signing"
)

const (
	csrfTokenLength    = 32
	csrfTokenCookieKey = "csrftoken"
	csrfFormFieldKey   = "csrfmiddlewaretoken"
	csrfHeaderKey      = "X-CSRFToken"
)

var csrfSafeMethods = map[string]bool{
	"GET":     true,
	"HEAD":    true,
	"OPTIONS": true,
	"TRACE":   true,
}

type CSRFConfig struct {
	SecretKey      string
	CookieName     string
	CookieSecure   bool
	CookieHTTPOnly bool
	CookiePath     string
	CookieDomain   string
	CookieSameSite string
	FallbackKeys   []string
	TrustedOrigins []string
}

func DefaultCSRFConfig(secretKey string) *CSRFConfig {
	return &CSRFConfig{
		SecretKey:      secretKey,
		CookieName:     csrfTokenCookieKey,
		CookieSecure:   false,
		CookieHTTPOnly: false,
		CookiePath:     "/",
		CookieSameSite: "Lax",
	}
}

func CSRFMiddleware(config *CSRFConfig) func(jangohttp.ViewFunc) jangohttp.ViewFunc {
	if config == nil {
		config = DefaultCSRFConfig("")
	}
	return func(next jangohttp.ViewFunc) jangohttp.ViewFunc {
		return func(req *jangohttp.Request) jangohttp.Response {
			token := getOrGenerateCSRFToken(req, config)

			req.SetParam("_csrf_token", token)

			if csrfSafeMethods[req.Method] {
				resp := next(req)
				setCSRFCookie(resp, token, config)
				return resp
			}

			if !validateCSRFToken(req, token, config) {
				return jangohttp.NewForbiddenResponse("CSRF verification failed. Request aborted.")
			}

			resp := next(req)
			setCSRFCookie(resp, token, config)
			return resp
		}
	}
}

func getOrGenerateCSRFToken(req *jangohttp.Request, config *CSRFConfig) string {
	cookie, err := req.Cookie(config.CookieName)
	if err == nil && cookie != nil && cookie.Value != "" {
		return cookie.Value
	}
	return generateCSRFToken()
}

func generateCSRFToken() string {
	b := make([]byte, csrfTokenLength)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func validateCSRFToken(req *jangohttp.Request, expected string, config *CSRFConfig) bool {
	token := ""

	form, err := req.Form()
	if err == nil {
		if val, ok := form[csrfFormFieldKey]; ok && len(val) > 0 {
			token = val[0]
		}
	}

	if token == "" {
		token = req.Header.Get(csrfHeaderKey)
	}

	if token == "" {
		token = req.Header.Get("X-CSRFTOKEN")
	}

	if token == "" {
		return false
	}

	if subtle.ConstantTimeCompare([]byte(token), []byte(expected)) == 1 {
		return true
	}

	signer := signing.NewSigner(config.SecretKey, signing.WithSalt("csrf"))
	if _, err := signer.Unsign(token); err == nil {
		return true
	}

	for _, key := range config.FallbackKeys {
		fallbackSigner := signing.NewSigner(key, signing.WithSalt("csrf"))
		if _, err := fallbackSigner.Unsign(token); err == nil {
			return true
		}
	}

	return false
}

func maskCSRFToken(token, secretKey string) string {
	h := sha256.New()
	h.Write([]byte(secretKey))
	h.Write([]byte(token))
	mask := hex.EncodeToString(h.Sum(nil))[:8]
	return mask + token
}

func setCSRFCookie(resp jangohttp.Response, token string, config *CSRFConfig) {
	if resp == nil {
		return
	}
	cookie := &http.Cookie{
		Name:     config.CookieName,
		Value:    token,
		Path:     config.CookiePath,
		Domain:   config.CookieDomain,
		Secure:   config.CookieSecure,
		HttpOnly: config.CookieHTTPOnly,
		SameSite: parseCSRFCSameSite(config.CookieSameSite),
	}
	type cookieSetter interface {
		SetCookie(*http.Cookie)
	}
	if cs, ok := resp.(cookieSetter); ok {
		cs.SetCookie(cookie)
	}
}

func parseCSRFCSameSite(s string) http.SameSite {
	switch strings.ToLower(s) {
	case "strict":
		return http.SameSiteStrictMode
	case "none":
		return http.SameSiteNoneMode
	default:
		return http.SameSiteLaxMode
	}
}

func GetCSRFToken(req *jangohttp.Request) string {
	return req.Param("_csrf_token")
}

func IsCSRFSafeMethod(method string) bool {
	return csrfSafeMethods[method]
}

var (
	ErrCSRFMissingToken = &CSRFError{Message: "CSRF token missing"}
	ErrCSRFInvalidToken = &CSRFError{Message: "CSRF token invalid"}
)

type CSRFError struct {
	Message string
}

func (e *CSRFError) Error() string { return e.Message }
