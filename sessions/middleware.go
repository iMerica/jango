package sessions

import (
	"net/http"
	"time"

	"github.com/iMerica/jango/conf"
	jangohttp "github.com/iMerica/jango/http"
)

var defaultEngine SessionEngine

func SetDefaultEngine(engine SessionEngine) {
	defaultEngine = engine
}

func GetDefaultEngine() SessionEngine {
	return defaultEngine
}

func InitEngine(settings *conf.Settings) SessionEngine {
	switch settings.SessionEngine {
	case "db":
		driver := NewORMSessionDriver(settings)
		defaultEngine = NewDBSessionEngine(driver)
	case "cache":
		driver := NewCacheDriverFromSettings(settings)
		defaultEngine = NewCacheSessionEngine(driver)
	case "cookie":
		secretKey := settings.SecretKey
		defaultEngine = NewSignedCookieSessionEngine(secretKey)
	default:
		driver := NewORMSessionDriver(settings)
		defaultEngine = NewDBSessionEngine(driver)
	}
	return defaultEngine
}

func SessionMiddleware(next jangohttp.ViewFunc) jangohttp.ViewFunc {
	return func(req *jangohttp.Request) jangohttp.Response {
		engine := GetDefaultEngine()
		if engine == nil {
			req.SetSession(emptySession{})
			return next(req)
		}

		config := DefaultConfig()

		cookie, err := req.Cookie(config.CookieName)
		var sessionKey string
		if err == nil && cookie != nil {
			sessionKey = cookie.Value
		}

		session, err := engine.Load(sessionKey)
		if err != nil || session == nil {
			session, err = engine.Create("")
			if err != nil {
				session = emptySession{}
			}
		}

		req.SetSession(session)

		resp := next(req)

		if session.IsModified() {
			if err := session.Save(); err != nil {
			}

			if bs, ok := session.(*baseSession); ok {
				cookie := &http.Cookie{
					Name:     config.CookieName,
					Value:    bs.SessionKey(),
					Path:     config.CookiePath,
					Domain:   config.CookieDomain,
					MaxAge:   config.CookieAge,
					Secure:   config.CookieSecure,
					HttpOnly: config.CookieHTTPOnly,
					SameSite: parseSameSite(config.CookieSameSite),
				}
				if resp != nil {
					type cookieSetter interface {
						SetCookie(*http.Cookie)
					}
					if cs, ok := resp.(cookieSetter); ok {
						cs.SetCookie(cookie)
					}
				}
			}
		}

		return resp
	}
}

type emptySession struct{}

func (e emptySession) Get(key string) (interface{}, bool) { return nil, false }
func (e emptySession) Set(key string, value interface{})  {}
func (e emptySession) Delete(key string)                  {}
func (e emptySession) Has(key string) bool                { return false }
func (e emptySession) Keys() []string                     { return nil }
func (e emptySession) Save() error                        { return nil }
func (e emptySession) IsModified() bool                   { return false }
func (e emptySession) Accessed() bool                     { return false }
func (e emptySession) SessionKey() string                 { return "" }
func (e emptySession) ExpireDate() time.Time              { return time.Time{} }
func (e emptySession) SetExpiry(duration time.Duration)   {}

func parseSameSite(s string) http.SameSite {
	switch s {
	case "Strict":
		return http.SameSiteStrictMode
	case "Lax":
		return http.SameSiteLaxMode
	case "None":
		return http.SameSiteNoneMode
	default:
		return http.SameSiteLaxMode
	}
}
