package security

type Config struct {
	SecretKey                    string
	AllowedHosts                 []string
	Debug                        bool
	CSRFEnabled                  bool
	CSRFCookieName               string
	CSRFCookieSecure             bool
	CSRFCookieHTTPOnly           bool
	SecureBrowserXSSFilter       bool
	SecureContentTypeNoSniff     bool
	SecureSSLRedirect            bool
	SecureHSTSSeconds            int
	SecureHSTSIncludeSubDomains  bool
	SecureHSTSPreload            bool
	SecureReferrerPolicy         string
	SecureProxySSLHeader         string
	FrameDeny                    bool
	FrameOption                  string
	CrossOriginOpenerPolicy     string
	CrossOriginEmbedderPolicy    string
	CrossOriginResourcePolicy    string
}

func DefaultConfig() *Config {
	return &Config{
		CSRFEnabled:                true,
		CSRFCookieName:             "csrftoken",
		CSRFCookieSecure:           false,
		CSRFCookieHTTPOnly:         false,
		SecureBrowserXSSFilter:     true,
		SecureContentTypeNoSniff:    true,
		FrameDeny:                  false,
		FrameOption:                "DENY",
		SecureReferrerPolicy:       "same-origin",
		CrossOriginOpenerPolicy:    "same-origin",
	}
}