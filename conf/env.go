package conf

import (
	"os"
	"strconv"
	"strings"
)

func ApplyEnv(s *Settings) {
	if s == nil {
		return
	}

	applyEnvString("JANGO_SECRET_KEY", func(v string) { s.SecretKey = v })
	applyEnvString("JANGO_PROJECT_NAME", func(v string) { s.ProjectName = v })
	applyEnvString("JANGO_BASE_DIR", func(v string) { s.BaseDir = v })
	applyEnvString("JANGO_ROOT_URLCONF", func(v string) { s.RootURLConf = v })
	applyEnvString("JANGO_AUTH_USER_MODEL", func(v string) { s.AuthUserModel = v })
	applyEnvString("JANGO_LANGUAGE_CODE", func(v string) { s.LanguageCode = v })
	applyEnvString("JANGO_TIME_ZONE", func(v string) { s.TimeZone = v })
	applyEnvString("JANGO_STATIC_URL", func(v string) { s.StaticURL = v })
	applyEnvString("JANGO_STATIC_ROOT", func(v string) { s.StaticRoot = v })
	applyEnvString("JANGO_MEDIA_URL", func(v string) { s.MediaURL = v })
	applyEnvString("JANGO_MEDIA_ROOT", func(v string) { s.MediaRoot = v })
	applyEnvString("JANGO_SESSION_ENGINE", func(v string) { s.SessionEngine = v })
	applyEnvString("JANGO_SCRIPT_PREFIX", func(v string) { s.ScriptPrefix = v })

	applyEnvBool("JANGO_DEBUG", func(v bool) { s.Debug = v })
	applyEnvBool("JANGO_USE_TZ", func(v bool) { s.UseTZ = v })

	applyEnvStringSlice("JANGO_ALLOWED_HOSTS", func(v []string) { s.AllowedHosts = v })
	applyEnvStringSlice("JANGO_INSTALLED_APPS", func(v []string) { s.InstalledApps = v })
	applyEnvStringSlice("JANGO_MIDDLEWARE", func(v []string) { s.Middleware = v })
	applyEnvStringSlice("JANGO_AUTHENTICATION_BACKENDS", func(v []string) { s.AuthenticationBackends = v })
	applyEnvStringSlice("JANGO_SECRET_KEY_FALLBACKS", func(v []string) { s.SecretKeyFallbacks = v })

	applyEnvSecure(s)
	applyEnvDatabase(s)
	applyEnvRESTFramework(s)
}

func applyEnvString(key string, set func(string)) {
	if v := os.Getenv(key); v != "" {
		set(v)
	}
}

func applyEnvBool(key string, set func(bool)) {
	if v := os.Getenv(key); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			set(b)
		}
	}
}

func applyEnvInt(key string, set func(int)) {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			set(n)
		}
	}
}

func applyEnvStringSlice(key string, set func([]string)) {
	if v := os.Getenv(key); v != "" {
		set(strings.Split(v, ","))
	}
}

func applyEnvSecure(s *Settings) {
	applyEnvBool("JANGO_SECURE_SSL_REDIRECT", func(v bool) { s.Secure.SSLRedirect = v })
	applyEnvString("JANGO_SECURE_SSL_HOST", func(v string) { s.Secure.SSLHost = v })
	applyEnvInt("JANGO_SECURE_HSTS_SECONDS", func(v int) { s.Secure.HSTSSeconds = v })
	applyEnvBool("JANGO_SECURE_HSTS_INCLUDE_SUBDOMAINS", func(v bool) { s.Secure.HSTSIncludeSubdomains = v })
	applyEnvBool("JANGO_SECURE_HSTS_PRELOAD", func(v bool) { s.Secure.HSTSPreload = v })
	applyEnvString("JANGO_SECURE_REFERRER_POLICY", func(v string) { s.Secure.ReferrerPolicy = v })
	applyEnvBool("JANGO_SECURE_BROWSER_XSS_FILTER", func(v bool) { s.Secure.BrowserXSSFilter = v })
	applyEnvBool("JANGO_SECURE_CONTENT_TYPE_NOSNIFF", func(v bool) { s.Secure.ContentTypeNoSniff = v })
}

func applyEnvDatabase(s *Settings) {
	if s.Databases == nil {
		s.Databases = make(map[string]*DatabaseSettings)
	}
	db, ok := s.Databases["default"]
	if !ok {
		db = &DatabaseSettings{}
		s.Databases["default"] = db
	}

	applyEnvString("JANGO_DATABASE_ENGINE", func(v string) { db.Engine = v })
	applyEnvString("JANGO_DATABASE_HOST", func(v string) { db.Host = v })
	applyEnvString("JANGO_DATABASE_NAME", func(v string) { db.Name = v })
	applyEnvString("JANGO_DATABASE_USER", func(v string) { db.User = v })
	applyEnvString("JANGO_DATABASE_PASSWORD", func(v string) { db.Password = v })
	if v := os.Getenv("JANGO_DATABASE_PORT"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			db.Port = n
		}
	}
}

func applyEnvRESTFramework(s *Settings) {
	applyEnvString("JANGO_REST_DEFAULT_PAGINATION_CLASS", func(v string) {
		s.RESTFramework.DefaultPaginationClass = v
	})
	applyEnvString("JANGO_REST_SEARCH_PARAM", func(v string) {
		s.RESTFramework.SearchParam = v
	})
	applyEnvString("JANGO_REST_ORDERING_PARAM", func(v string) {
		s.RESTFramework.OrderingParam = v
	})
	applyEnvString("JANGO_REST_DEFAULT_VERSIONING_CLASS", func(v string) {
		s.RESTFramework.DefaultVersioningClass = v
	})
	applyEnvInt("JANGO_REST_PAGE_SIZE", func(v int) {
		s.RESTFramework.PageSize = v
	})
	applyEnvStringSlice("JANGO_REST_DEFAULT_AUTHENTICATION_CLASSES", func(v []string) {
		s.RESTFramework.DefaultAuthenticationClasses = v
	})
	applyEnvStringSlice("JANGO_REST_DEFAULT_PERMISSION_CLASSES", func(v []string) {
		s.RESTFramework.DefaultPermissionClasses = v
	})
	applyEnvStringSlice("JANGO_REST_DEFAULT_RENDERER_CLASSES", func(v []string) {
		s.RESTFramework.DefaultRendererClasses = v
	})
	applyEnvStringSlice("JANGO_REST_DEFAULT_PARSER_CLASSES", func(v []string) {
		s.RESTFramework.DefaultParserClasses = v
	})
	applyEnvStringSlice("JANGO_REST_DEFAULT_THROTTLE_CLASSES", func(v []string) {
		s.RESTFramework.DefaultThrottleClasses = v
	})
	applyEnvStringSlice("JANGO_REST_DEFAULT_FILTER_BACKEND_CLASSES", func(v []string) {
		s.RESTFramework.DefaultFilterBackendClasses = v
	})
}