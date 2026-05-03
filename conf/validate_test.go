package conf

import (
	"testing"
)

func TestValidateEmptyProjectName(t *testing.T) {
	s := DefaultSettings()
	s.ProjectName = ""
	errs := Validate(s)
	found := false
	for _, e := range errs {
		if e.Field == "ProjectName" {
			found = true
		}
	}
	if !found {
		t.Error("expected validation error for empty ProjectName")
	}
}

func TestValidateProductionWithoutSecretKey(t *testing.T) {
	s := DefaultSettings()
	s.Debug = false
	s.SecretKey = ""
	errs := Validate(s)
	found := false
	for _, e := range errs {
		if e.Field == "SecretKey" {
			found = true
		}
	}
	if !found {
		t.Error("expected validation error for empty SecretKey in production")
	}
}

func TestValidateProductionWithoutAllowedHosts(t *testing.T) {
	s := DefaultSettings()
	s.Debug = false
	s.AllowedHosts = []string{}
	errs := Validate(s)
	found := false
	for _, e := range errs {
		if e.Field == "AllowedHosts" {
			found = true
		}
	}
	if !found {
		t.Error("expected validation error for empty AllowedHosts in production")
	}
}

func TestValidateWildcardAllowedHosts(t *testing.T) {
	s := DefaultSettings()
	s.Debug = false
	s.AllowedHosts = []string{"*"}
	s.SecretKey = "a-long-enough-secret-key-for-validation-purposes-testing!!"
	errs := Validate(s)
	found := false
	for _, e := range errs {
		if e.Field == "AllowedHosts" && e.Message == "wildcard '*' is insecure in production" {
			found = true
		}
	}
	if !found {
		t.Error("expected validation error for wildcard AllowedHosts in production")
	}
}

func TestValidateDebugAllowsEmptySecretKey(t *testing.T) {
	s := DefaultSettings()
	s.Debug = true
	s.ProjectName = "test"
	s.SecretKey = ""
	errs := Validate(s)
	for _, e := range errs {
		if e.Field == "SecretKey" {
			t.Error("should not flag empty SecretKey when Debug=true")
		}
	}
}

func TestValidateShortSecretKey(t *testing.T) {
	s := DefaultSettings()
	s.ProjectName = "test"
	s.SecretKey = "short"
	errs := Validate(s)
	found := false
	for _, e := range errs {
		if e.Field == "SecretKey" && e.Message == "should be at least 50 characters for cryptographic safety" {
			found = true
		}
	}
	if !found {
		t.Error("expected validation warning for short secret key")
	}
}

func TestValidateEmptyDatabaseEngine(t *testing.T) {
	s := DefaultSettings()
	s.ProjectName = "test"
	s.Databases["default"].Engine = ""
	errs := Validate(s)
	found := false
	for _, e := range errs {
		if e.Field == "Databases[default].Engine" {
			found = true
		}
	}
	if !found {
		t.Error("expected validation error for empty database engine")
	}
}

func TestValidateTemplateNoBackend(t *testing.T) {
	s := DefaultSettings()
	s.ProjectName = "test"
	s.Templates = []TemplateConfig{{Backends: []string{}}}
	errs := Validate(s)
	found := false
	for _, e := range errs {
		if e.Field == "Templates[0].Backends" {
			found = true
		}
	}
	if !found {
		t.Error("expected validation error for template config with no backends")
	}
}

func TestValidateValidSettings(t *testing.T) {
	s := DefaultSettings()
	s.ProjectName = "myproject"
	s.SecretKey = "a-very-long-secret-key-that-is-at-least-50-characters-long-for-testing"
	s.AllowedHosts = []string{"example.com"}
	s.Debug = false
	errs := Validate(s)
	for _, e := range errs {
		if e.Field == "ProjectName" || e.Field == "SecretKey" || e.Field == "AllowedHosts" {
			t.Errorf("unexpected error for %s: %s", e.Field, e.Message)
		}
	}
}

func TestValidateDeployDebug(t *testing.T) {
	s := DefaultSettings()
	s.Debug = true
	errs := ValidateDeploy(s)
	found := false
	for _, e := range errs {
		if e.Field == "Debug" {
			found = true
		}
	}
	if !found {
		t.Error("expected deploy check error for Debug=true")
	}
}

func TestValidateDeployEmptySecretKey(t *testing.T) {
	s := DefaultSettings()
	s.SecretKey = ""
	errs := ValidateDeploy(s)
	found := false
	for _, e := range errs {
		if e.Field == "SecretKey" {
			found = true
		}
	}
	if !found {
		t.Error("expected deploy check error for empty SecretKey")
	}
}

func TestValidateDeployEmptyAllowedHosts(t *testing.T) {
	s := DefaultSettings()
	s.AllowedHosts = []string{}
	errs := ValidateDeploy(s)
	found := false
	for _, e := range errs {
		if e.Field == "AllowedHosts" {
			found = true
		}
	}
	if !found {
		t.Error("expected deploy check error for empty AllowedHosts")
	}
}

func TestValidateDeployNoSSL(t *testing.T) {
	s := DefaultSettings()
	s.Secure.SSLRedirect = false
	errs := ValidateDeploy(s)
	found := false
	for _, e := range errs {
		if e.Field == "Secure.SSLRedirect" {
			found = true
		}
	}
	if !found {
		t.Error("expected deploy check warning for no SSL redirect")
	}
}

func TestValidateDeployNoHSTS(t *testing.T) {
	s := DefaultSettings()
	s.Secure.HSTSSeconds = 0
	errs := ValidateDeploy(s)
	found := false
	for _, e := range errs {
		if e.Field == "Secure.HSTSSeconds" {
			found = true
		}
	}
	if !found {
		t.Error("expected deploy check warning for no HSTS")
	}
}

func TestValidateDeployWeakSecurity(t *testing.T) {
	s := DefaultSettings()
	s.Secure.BrowserXSSFilter = false
	s.Secure.ContentTypeNoSniff = false
	errs := ValidateDeploy(s)
	foundXSS := false
	foundNoSniff := false
	for _, e := range errs {
		if e.Field == "Secure.BrowserXSSFilter" {
			foundXSS = true
		}
		if e.Field == "Secure.ContentTypeNoSniff" {
			foundNoSniff = true
		}
	}
	if !foundXSS {
		t.Error("expected deploy check warning for no XSS filter")
	}
	if !foundNoSniff {
		t.Error("expected deploy check warning for no content type nosniff")
	}
}

func TestValidateDeployEmptyDBPassword(t *testing.T) {
	s := DefaultSettings()
	s.Databases["default"].Password = ""
	errs := ValidateDeploy(s)
	found := false
	for _, e := range errs {
		if e.Field == "Databases[default].Password" {
			found = true
		}
	}
	if !found {
		t.Error("expected deploy check warning for empty DB password")
	}
}

func TestValidationErrorsString(t *testing.T) {
	errs := ValidationErrors{
		{Field: "A", Message: "error a"},
		{Field: "B", Message: "error b"},
	}
	result := errs.Error()
	if result == "" {
		t.Error("expected non-empty error string")
	}
}

func TestValidationErrorsHasErrors(t *testing.T) {
	errs := ValidationErrors{
		{Field: "A", Message: "error a"},
	}
	if !errs.HasErrors() {
		t.Error("expected HasErrors=true")
	}
	emptyErrs := ValidationErrors{}
	if emptyErrs.HasErrors() {
		t.Error("expected HasErrors=false for empty slice")
	}
}