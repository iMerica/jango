package conf

import "time"

type Settings struct {
	BaseDir     string
	ProjectName string

	Debug                   bool
	AllowedHosts            []string
	InstalledApps           []string
	Middleware              []string
	RootURLConf             string
	AuthUserModel           string
	AuthenticationBackends []string

	Databases map[string]*DatabaseSettings
	Templates []TemplateConfig

	LanguageCode string
	TimeZone     string
	UseTZ        bool

	StaticURL  string
	StaticRoot string
	MediaURL   string
	MediaRoot  string

	SecretKey          string
	SecretKeyFallbacks []string
	Secure             SecureSettings

	Caches        map[string]*CacheSettings
	SessionEngine string

	RESTFramework RESTFrameworkSettings
	Logging        LoggingConfig

	ScriptPrefix string
}

type DatabaseSettings struct {
	Engine          string
	Host            string
	Port            int
	Name            string
	User            string
	Password        string
	Options         map[string]string
	ConnMaxLifetime time.Duration
	ConnMaxIdle     int
	ConnMaxOpen     int
}

type TemplateConfig struct {
	Backends          []string
	Dirs              []string
	AppDirs           bool
	ContextProcessors []string
	Options           map[string]interface{}
}

type CacheSettings struct {
	Backend   string
	Location  string
	Timeout   int
	KeyPrefix string
	Options   map[string]string
}

type SecureSettings struct {
	SSLRedirect           bool
	SSLHost                string
	HSTSSeconds            int
	HSTSIncludeSubdomains  bool
	HSTSPreload            bool
	ReferrerPolicy         string
	BrowserXSSFilter       bool
	ContentTypeNoSniff     bool
	ProxySSLHeader         []string
	RedirectExempt         []string
}

type RESTFrameworkSettings struct {
	DefaultAuthenticationClasses []string
	DefaultPermissionClasses     []string
	DefaultThrottleClasses       []string
	DefaultThrottleRates         map[string]string
	DefaultPaginationClass       string
	DefaultRendererClasses       []string
	DefaultParserClasses         []string
	DefaultFilterBackendClasses  []string
	DefaultMetadataClass         string
	DefaultSchemaClass           string
	DefaultVersioningClass       string
	SearchParam                  string
	OrderingParam                string
	PageSize                     int
	PageQueryParam               string
	URLInputOverridePrefix       string
	URLInputOverrideSuffix       string
}

type LoggingConfig struct {
	Level      string
	Format     string
	OutputPath string
	Loggers    map[string]LoggerConfig
}

type LoggerConfig struct {
	Level      string
	Format     string
	OutputPath string
}

func (s *Settings) DefaultDatabase() *DatabaseSettings {
	if s.Databases == nil {
		return nil
	}
	return s.Databases["default"]
}

func (s *Settings) DefaultCache() *CacheSettings {
	if s.Caches == nil {
		return nil
	}
	return s.Caches["default"]
}

func DefaultSettings() *Settings {
	return &Settings{
		Debug:         true,
		AllowedHosts:  []string{},
		InstalledApps: []string{},
		Middleware:    []string{},
		Databases: map[string]*DatabaseSettings{
			"default": {
				Engine:      "postgresql",
				Host:        "localhost",
				Port:         5432,
				Name:        "",
				User:        "",
				ConnMaxIdle: 2,
				ConnMaxOpen: 0,
			},
		},
		Templates: []TemplateConfig{
			{
				Backends:          []string{"html/template"},
				AppDirs:           true,
				ContextProcessors: []string{},
				Options:           map[string]interface{}{},
			},
		},
		LanguageCode:           "en-us",
		TimeZone:               "UTC",
		UseTZ:                  true,
		StaticURL:              "/static/",
		StaticRoot:             "",
		MediaURL:               "/media/",
		MediaRoot:              "",
		AuthUserModel:          "auth.User",
		AuthenticationBackends: []string{"auth.ModelBackend"},
		Secure: SecureSettings{
			BrowserXSSFilter:   false,
			ContentTypeNoSniff: false,
		},
		Caches: map[string]*CacheSettings{
			"default": {
				Backend:   "locmem",
				Timeout:   300,
				KeyPrefix: "",
			},
		},
		SessionEngine:  "db",
		RESTFramework:  defaultRESTFramework(),
		Logging: LoggingConfig{
			Level:      "INFO",
			Format:     "text",
			OutputPath: "stderr",
			Loggers:    map[string]LoggerConfig{},
		},
	}
}

func defaultRESTFramework() RESTFrameworkSettings {
	return RESTFrameworkSettings{
		DefaultAuthenticationClasses: []string{
			"rest.authentication.SessionAuthentication",
			"rest.authentication.BasicAuthentication",
		},
		DefaultPermissionClasses: []string{
			"rest.permissions.AllowAny",
		},
		DefaultRendererClasses: []string{
			"rest.renderers.JSONRenderer",
		},
		DefaultParserClasses: []string{
			"rest.parsers.JSONParser",
		},
		DefaultThrottleClasses:      []string{},
		DefaultThrottleRates:        map[string]string{},
		DefaultPaginationClass:      "",
		DefaultFilterBackendClasses: []string{},
		DefaultMetadataClass:         "rest.metadata.SimpleMetadata",
		DefaultSchemaClass:           "rest.schemas.AutoSchema",
		DefaultVersioningClass:       "",
		SearchParam:                  "search",
		OrderingParam:                "ordering",
		PageSize:                      0,
		PageQueryParam:               "page",
		URLInputOverridePrefix:       "",
		URLInputOverrideSuffix:       "",
	}
}