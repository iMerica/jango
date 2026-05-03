package security

import (
	"net"
	"strings"

	"github.com/iMerica/jango/checks"
	"github.com/iMerica/jango/conf"
)

func ValidateHost(host string, allowedHosts []string) bool {
	if len(allowedHosts) == 0 {
		return true
	}

	host = strings.ToLower(host)
	if strings.Contains(host, ":") {
		h, _, err := net.SplitHostPort(host)
		if err == nil {
			host = h
		}
	}

	for _, allowed := range allowedHosts {
		allowed = strings.ToLower(allowed)
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

	if ip := net.ParseIP(host); ip != nil && ip.IsLoopback() {
		for _, allowed := range allowedHosts {
			if allowed == "127.0.0.1" || allowed == "::1" || allowed == "localhost" {
				return true
			}
		}
	}

	return false
}

func init() {
	checks.RegisterDeploy(func() []checks.CheckResult {
		var results []checks.CheckResult
		s, err := conf.TryGet()
		if err != nil {
			return results
		}

		if s.SecretKey == "" {
			results = append(results, checks.Error("security.E001", "SecretKey is empty", "Generate a secret key with: openssl rand -hex 50"))
		} else if len(s.SecretKey) < 50 {
			results = append(results, checks.Warn("security.W001", "SecretKey is shorter than 50 characters", "Generate a longer secret key"))
		}

		if s.Debug {
			results = append(results, checks.Error("security.E002", "Debug is true in production", "Set Debug=false in production"))
		}

		if len(s.AllowedHosts) == 0 {
			results = append(results, checks.Error("security.E003", "AllowedHosts is empty in production", "Set AllowedHosts to permitted host values"))
		}

		for _, host := range s.AllowedHosts {
			if host == "*" {
				results = append(results, checks.Error("security.E004", "AllowedHosts contains wildcard '*'", "Use specific hostnames instead"))
			}
		}

		if !s.Secure.SSLRedirect {
			results = append(results, checks.Warn("security.W002", "Secure.SSLRedirect is false", "Enable SSL redirect in production"))
		}

		if s.Secure.HSTSSeconds == 0 {
			results = append(results, checks.Warn("security.W003", "Secure.HSTSSeconds is 0", "Set HSTS to at least 31536000 seconds"))
		}

		if !s.Secure.BrowserXSSFilter {
			results = append(results, checks.Warn("security.W004", "Secure.BrowserXSSFilter is false", "Enable X-XSS-Protection header"))
		}

		if !s.Secure.ContentTypeNoSniff {
			results = append(results, checks.Warn("security.W005", "Secure.ContentTypeNoSniff is false", "Enable X-Content-Type-Options: nosniff"))
		}

		if s.Secure.ReferrerPolicy == "" {
			results = append(results, checks.Warn("security.W006", "Secure.ReferrerPolicy is empty", "Set Referrer-Policy to 'same-origin' or 'strict-origin-when-cross-origin'"))
		}

		if s.SessionEngine == "" {
			results = append(results, checks.Error("security.E005", "SessionEngine is empty", "Set SessionEngine to 'db', 'cache', or 'cookie'"))
		}

		return results
	})
}
