package checks

import (
	"fmt"

	"github.com/iMerica/jango/conf"
)

func init() {
	Register(func() []CheckResult {
		s, err := conf.TryGet()
		if err != nil {
			return nil
		}
		var results []CheckResult

		if s.SecretKey == "" {
			results = append(results, Error(
				"security.W001",
				"SecretKey is empty; this is insecure",
				"Set a strong secret key in production settings",
			))
		} else if len(s.SecretKey) < 50 {
			results = append(results, Warn(
				"security.W002",
				fmt.Sprintf("SecretKey is %d characters; recommended minimum is 50", len(s.SecretKey)),
				"Use a longer secret key for better cryptographic strength",
			))
		}

		if s.Debug {
			results = append(results, Warn(
				"security.W003",
				"DEBUG is enabled; this should be disabled in production",
				"Set Debug=false in production settings",
			))
		}

		if len(s.AllowedHosts) == 0 {
			results = append(results, Warn(
				"security.W004",
				"AllowedHosts is empty; any host will be accepted",
				"Configure AllowedHosts with valid hostnames for production",
			))
		}

		for _, host := range s.AllowedHosts {
			if host == "*" {
				results = append(results, Warn(
					"security.W005",
					"AllowedHosts contains '*' wildcard; this is insecure in production",
					"Replace wildcard with specific hostnames",
				))
				break
			}
		}

		if !s.Secure.SSLRedirect {
			results = append(results, Warn(
				"security.W006",
				"Secure.SSLRedirect is not enabled",
				"Enable SSL redirect for production",
			))
		}

		if s.Secure.HSTSSeconds == 0 {
			results = append(results, Warn(
				"security.W007",
				"Secure.HSTSSeconds is 0; HSTS is not enabled",
				"Set HSTSSeconds to at least 31536000 for one year",
			))
		}

		if !s.Secure.BrowserXSSFilter {
			results = append(results, Warn(
				"security.W008",
				"Secure.BrowserXSSFilter is not enabled",
				"Enable BrowserXSSFilter for production",
			))
		}

		if !s.Secure.ContentTypeNoSniff {
			results = append(results, Warn(
				"security.W009",
				"Secure.ContentTypeNoSniff is not enabled",
				"Enable ContentTypeNoSniff for production",
			))
		}

		if s.Secure.ReferrerPolicy == "" {
			results = append(results, Warn(
				"security.W010",
				"Secure.ReferrerPolicy is not set",
				"Set ReferrerPolicy to 'same-origin' or 'no-referrer'",
			))
		}

		return results
	})
}
