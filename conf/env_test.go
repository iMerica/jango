package conf

import (
	"os"
	"testing"
)

func TestApplyEnvString(t *testing.T) {
	s := DefaultSettings()
	s.ProjectName = "original"
	os.Setenv("JANGO_PROJECT_NAME", "fromenv")
	defer os.Unsetenv("JANGO_PROJECT_NAME")

	ApplyEnv(s)

	if s.ProjectName != "fromenv" {
		t.Errorf("expected ProjectName=fromenv, got %s", s.ProjectName)
	}
}

func TestApplyEnvBool(t *testing.T) {
	s := DefaultSettings()
	s.Debug = true
	os.Setenv("JANGO_DEBUG", "false")
	defer os.Unsetenv("JANGO_DEBUG")

	ApplyEnv(s)

	if s.Debug != false {
		t.Errorf("expected Debug=false, got %v", s.Debug)
	}
}

func TestApplyEnvSecretKey(t *testing.T) {
	s := DefaultSettings()
	s.SecretKey = ""
	os.Setenv("JANGO_SECRET_KEY", "env-secret-key-that-is-at-least-50-characters-long-for-safety!!")
	defer os.Unsetenv("JANGO_SECRET_KEY")

	ApplyEnv(s)

	if s.SecretKey != "env-secret-key-that-is-at-least-50-characters-long-for-safety!!" {
		t.Errorf("expected SecretKey from env, got %s", s.SecretKey)
	}
}

func TestApplyEnvStringSlice(t *testing.T) {
	s := DefaultSettings()
	os.Setenv("JANGO_ALLOWED_HOSTS", "example.com,localhost,127.0.0.1")
	defer os.Unsetenv("JANGO_ALLOWED_HOSTS")

	ApplyEnv(s)

	if len(s.AllowedHosts) != 3 {
		t.Fatalf("expected 3 AllowedHosts, got %d", len(s.AllowedHosts))
	}
	if s.AllowedHosts[0] != "example.com" {
		t.Errorf("expected example.com, got %s", s.AllowedHosts[0])
	}
}

func TestApplyEnvDatabase(t *testing.T) {
	s := DefaultSettings()
	os.Setenv("JANGO_DATABASE_HOST", "prod-db.example.com")
	os.Setenv("JANGO_DATABASE_NAME", "prod_db")
	os.Setenv("JANGO_DATABASE_USER", "prod_user")
	os.Setenv("JANGO_DATABASE_PASSWORD", "prod_pass")
	defer os.Unsetenv("JANGO_DATABASE_HOST")
	defer os.Unsetenv("JANGO_DATABASE_NAME")
	defer os.Unsetenv("JANGO_DATABASE_USER")
	defer os.Unsetenv("JANGO_DATABASE_PASSWORD")

	ApplyEnv(s)

	db := s.DefaultDatabase()
	if db == nil {
		t.Fatal("expected default database config")
	}
	if db.Host != "prod-db.example.com" {
		t.Errorf("expected prod-db.example.com, got %s", db.Host)
	}
	if db.Name != "prod_db" {
		t.Errorf("expected prod_db, got %s", db.Name)
	}
}

func TestApplyEnvDatabasePort(t *testing.T) {
	s := DefaultSettings()
	os.Setenv("JANGO_DATABASE_PORT", "5433")
	defer os.Unsetenv("JANGO_DATABASE_PORT")

	ApplyEnv(s)

	db := s.DefaultDatabase()
	if db.Port != 5433 {
		t.Errorf("expected port 5433, got %d", db.Port)
	}
}

func TestApplyEnvSecure(t *testing.T) {
	s := DefaultSettings()
	os.Setenv("JANGO_SECURE_SSL_REDIRECT", "true")
	os.Setenv("JANGO_SECURE_HSTS_SECONDS", "31536000")
	os.Setenv("JANGO_SECURE_BROWSER_XSS_FILTER", "true")
	os.Setenv("JANGO_SECURE_CONTENT_TYPE_NOSNIFF", "true")
	os.Setenv("JANGO_SECURE_REFERRER_POLICY", "same-origin")
	defer os.Unsetenv("JANGO_SECURE_SSL_REDIRECT")
	defer os.Unsetenv("JANGO_SECURE_HSTS_SECONDS")
	defer os.Unsetenv("JANGO_SECURE_BROWSER_XSS_FILTER")
	defer os.Unsetenv("JANGO_SECURE_CONTENT_TYPE_NOSNIFF")
	defer os.Unsetenv("JANGO_SECURE_REFERRER_POLICY")

	ApplyEnv(s)

	if !s.Secure.SSLRedirect {
		t.Error("expected SSLRedirect=true")
	}
	if s.Secure.HSTSSeconds != 31536000 {
		t.Errorf("expected 31536000, got %d", s.Secure.HSTSSeconds)
	}
	if !s.Secure.BrowserXSSFilter {
		t.Error("expected BrowserXSSFilter=true")
	}
	if !s.Secure.ContentTypeNoSniff {
		t.Error("expected ContentTypeNoSniff=true")
	}
	if s.Secure.ReferrerPolicy != "same-origin" {
		t.Errorf("expected same-origin, got %s", s.Secure.ReferrerPolicy)
	}
}

func TestApplyEnvRESTFramework(t *testing.T) {
	s := DefaultSettings()
	os.Setenv("JANGO_REST_PAGE_SIZE", "25")
	os.Setenv("JANGO_REST_SEARCH_PARAM", "q")
	defer os.Unsetenv("JANGO_REST_PAGE_SIZE")
	defer os.Unsetenv("JANGO_REST_SEARCH_PARAM")

	ApplyEnv(s)

	if s.RESTFramework.PageSize != 25 {
		t.Errorf("expected PageSize=25, got %d", s.RESTFramework.PageSize)
	}
	if s.RESTFramework.SearchParam != "q" {
		t.Errorf("expected SearchParam=q, got %s", s.RESTFramework.SearchParam)
	}
}

func TestApplyEnvUseTZ(t *testing.T) {
	s := DefaultSettings()
	s.UseTZ = true
	os.Setenv("JANGO_USE_TZ", "false")
	defer os.Unsetenv("JANGO_USE_TZ")

	ApplyEnv(s)

	if s.UseTZ != false {
		t.Errorf("expected UseTZ=false, got %v", s.UseTZ)
	}
}

func TestApplyEnvNilSettings(t *testing.T) {
	ApplyEnv(nil)
}

func TestApplyEnvEmptyValuesIgnored(t *testing.T) {
	s := DefaultSettings()
	s.ProjectName = "original"
	os.Setenv("JANGO_PROJECT_NAME", "")
	defer os.Unsetenv("JANGO_PROJECT_NAME")

	ApplyEnv(s)

	if s.ProjectName != "original" {
		t.Errorf("expected original (empty env should not override), got %s", s.ProjectName)
	}
}

func TestApplyEnvDatabaseCreatesDefault(t *testing.T) {
	s := DefaultSettings()
	s.Databases = nil

	os.Setenv("JANGO_DATABASE_HOST", "newhost")
	defer os.Unsetenv("JANGO_DATABASE_HOST")

	ApplyEnv(s)

	if s.Databases == nil {
		t.Fatal("expected Databases map to be created")
	}
	if s.Databases["default"] == nil {
		t.Fatal("expected default database entry to be created")
	}
	if s.Databases["default"].Host != "newhost" {
		t.Errorf("expected newhost, got %s", s.Databases["default"].Host)
	}
}

func TestApplyEnvMiddlewareSlice(t *testing.T) {
	s := DefaultSettings()
	os.Setenv("JANGO_MIDDLEWARE", "middleware1,middleware2")
	defer os.Unsetenv("JANGO_MIDDLEWARE")

	ApplyEnv(s)

	if len(s.Middleware) != 2 {
		t.Fatalf("expected 2 middleware entries, got %d", len(s.Middleware))
	}
	if s.Middleware[0] != "middleware1" {
		t.Errorf("expected middleware1, got %s", s.Middleware[0])
	}
}

func TestApplyEnvRESTFrameworkRendererClasses(t *testing.T) {
	s := DefaultSettings()
	os.Setenv("JANGO_REST_DEFAULT_RENDERER_CLASSES", "rest.renderers.JSONRenderer,rest.renderers.TemplateRenderer")
	defer os.Unsetenv("JANGO_REST_DEFAULT_RENDERER_CLASSES")

	ApplyEnv(s)

	if len(s.RESTFramework.DefaultRendererClasses) != 2 {
		t.Fatalf("expected 2 renderer classes, got %d", len(s.RESTFramework.DefaultRendererClasses))
	}
	if s.RESTFramework.DefaultRendererClasses[0] != "rest.renderers.JSONRenderer" {
		t.Errorf("expected first renderer to be JSONRenderer, got %s", s.RESTFramework.DefaultRendererClasses[0])
	}
}