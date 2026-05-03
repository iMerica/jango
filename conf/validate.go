package conf

import (
	"fmt"
	"strings"
)

type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("conf: %s: %s", e.Field, e.Message)
}

type ValidationErrors []ValidationError

func (errs ValidationErrors) Error() string {
	var msgs []string
	for _, e := range errs {
		msgs = append(msgs, e.Error())
	}
	return strings.Join(msgs, "; ")
}

func (errs ValidationErrors) HasErrors() bool {
	return len(errs) > 0
}

func Validate(s *Settings) ValidationErrors {
	var errs ValidationErrors

	if s.ProjectName == "" {
		errs = append(errs, ValidationError{Field: "ProjectName", Message: "must not be empty"})
	}

	if !s.Debug {
		if s.SecretKey == "" {
			errs = append(errs, ValidationError{Field: "SecretKey", Message: "must not be empty in production"})
		}
		if len(s.AllowedHosts) == 0 {
			errs = append(errs, ValidationError{Field: "AllowedHosts", Message: "must not be empty in production"})
		}
		for _, host := range s.AllowedHosts {
			if host == "*" {
				errs = append(errs, ValidationError{Field: "AllowedHosts", Message: "wildcard '*' is insecure in production"})
			}
		}
	}

	if s.SecretKey != "" && len(s.SecretKey) < 50 {
		errs = append(errs, ValidationError{Field: "SecretKey", Message: "should be at least 50 characters for cryptographic safety"})
	}

	if db := s.DefaultDatabase(); db != nil {
		if db.Engine == "" {
			errs = append(errs, ValidationError{Field: "Databases[default].Engine", Message: "must not be empty"})
		}
	}

	if len(s.AllowedHosts) > 0 {
		for _, host := range s.AllowedHosts {
			if host == "" {
				errs = append(errs, ValidationError{Field: "AllowedHosts", Message: "must not contain empty strings"})
			}
		}
	}

	for i, tc := range s.Templates {
		if len(tc.Backends) == 0 {
			errs = append(errs, ValidationError{
				Field:   fmt.Sprintf("Templates[%d].Backends", i),
				Message: "must specify at least one backend",
			})
		}
	}

	return errs
}

func ValidateDeploy(s *Settings) ValidationErrors {
	var errs ValidationErrors

	if s.Debug {
		errs = append(errs, ValidationError{
			Field:   "Debug",
			Message: "must be false in production",
		})
	}

	if s.SecretKey == "" {
		errs = append(errs, ValidationError{
			Field:   "SecretKey",
			Message: "must not be empty",
		})
	}

	if len(s.AllowedHosts) == 0 {
		errs = append(errs, ValidationError{
			Field:   "AllowedHosts",
			Message: "must not be empty in production",
		})
	}

	for _, host := range s.AllowedHosts {
		if host == "*" {
			errs = append(errs, ValidationError{
				Field:   "AllowedHosts",
				Message: "wildcard '*' is insecure in production",
			})
		}
	}

	if !s.Secure.SSLRedirect {
		errs = append(errs, ValidationError{
			Field:   "Secure.SSLRedirect",
			Message: "should be true in production",
		})
	}

	if s.Secure.HSTSSeconds == 0 {
		errs = append(errs, ValidationError{
			Field:   "Secure.HSTSSeconds",
			Message: "should be set to a nonzero value in production",
		})
	}

	if !s.Secure.BrowserXSSFilter {
		errs = append(errs, ValidationError{
			Field:   "Secure.BrowserXSSFilter",
			Message: "should be true in production",
		})
	}

	if !s.Secure.ContentTypeNoSniff {
		errs = append(errs, ValidationError{
			Field:   "Secure.ContentTypeNoSniff",
			Message: "should be true in production",
		})
	}

	if db := s.DefaultDatabase(); db != nil {
		if db.Password == "" {
			errs = append(errs, ValidationError{
				Field:   "Databases[default].Password",
				Message: "should not be empty in production",
			})
		}
	}

	if s.SessionEngine == "" {
		errs = append(errs, ValidationError{
			Field:   "SessionEngine",
			Message: "must not be empty",
		})
	}

	return errs
}