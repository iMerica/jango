package conf

import (
	"testing"
)

func TestDefaultSettings(t *testing.T) {
	s := DefaultSettings()
	if s.Debug != true {
		t.Error("expected Debug=true by default")
	}
	if s.Databases["default"].Engine != "postgresql" {
		t.Errorf("expected postgresql engine, got %s", s.Databases["default"].Engine)
	}
	if s.Databases["default"].Port != 5432 {
		t.Errorf("expected port 5432, got %d", s.Databases["default"].Port)
	}
	if s.StaticURL != "/static/" {
		t.Errorf("expected /static/, got %s", s.StaticURL)
	}
	if s.LanguageCode != "en-us" {
		t.Errorf("expected en-us, got %s", s.LanguageCode)
	}
	if s.UseTZ != true {
		t.Error("expected UseTZ=true by default")
	}
	if s.AuthUserModel != "auth.User" {
		t.Errorf("expected auth.User, got %s", s.AuthUserModel)
	}
	if len(s.AuthenticationBackends) != 1 || s.AuthenticationBackends[0] != "auth.ModelBackend" {
		t.Errorf("expected [auth.ModelBackend], got %v", s.AuthenticationBackends)
	}
	if len(s.Databases) != 1 {
		t.Errorf("expected 1 database config, got %d", len(s.Databases))
	}
	if s.Databases["default"] == nil {
		t.Error("expected default database config")
	}
	if s.Databases["default"].Engine != "postgresql" {
		t.Errorf("expected default database engine=postgresql, got %s", s.Databases["default"].Engine)
	}
	if len(s.Caches) != 1 {
		t.Errorf("expected 1 cache config, got %d", len(s.Caches))
	}
	if s.Caches["default"] == nil {
		t.Error("expected default cache config")
	}
	if s.Caches["default"].Backend != "locmem" {
		t.Errorf("expected default cache backend=locmem, got %s", s.Caches["default"].Backend)
	}
	if s.SessionEngine != "db" {
		t.Errorf("expected db session engine, got %s", s.SessionEngine)
	}
	if s.Logging.Level != "INFO" {
		t.Errorf("expected INFO log level, got %s", s.Logging.Level)
	}
	if s.MediaURL != "/media/" {
		t.Errorf("expected /media/, got %s", s.MediaURL)
	}
}

func TestDefaultRESTFramework(t *testing.T) {
	rf := DefaultSettings().RESTFramework
	if len(rf.DefaultAuthenticationClasses) != 2 {
		t.Errorf("expected 2 default auth classes, got %d", len(rf.DefaultAuthenticationClasses))
	}
	if len(rf.DefaultPermissionClasses) != 1 {
		t.Errorf("expected 1 default permission class, got %d", len(rf.DefaultPermissionClasses))
	}
	if rf.SearchParam != "search" {
		t.Errorf("expected search param 'search', got %s", rf.SearchParam)
	}
	if rf.OrderingParam != "ordering" {
		t.Errorf("expected ordering param 'ordering', got %s", rf.OrderingParam)
	}
	if rf.DefaultMetadataClass != "rest.metadata.SimpleMetadata" {
		t.Errorf("expected SimpleMetadata, got %s", rf.DefaultMetadataClass)
	}
	if rf.DefaultSchemaClass != "rest.schemas.AutoSchema" {
		t.Errorf("expected AutoSchema, got %s", rf.DefaultSchemaClass)
	}
}

func TestDefaultDatabase(t *testing.T) {
	s := DefaultSettings()
	db := s.DefaultDatabase()
	if db == nil {
		t.Fatal("expected default database config")
	}
	if db.Engine != "postgresql" {
		t.Errorf("expected postgresql, got %s", db.Engine)
	}
	if db.Host != "localhost" {
		t.Errorf("expected localhost, got %s", db.Host)
	}
	if db.Port != 5432 {
		t.Errorf("expected 5432, got %d", db.Port)
	}
}

func TestDefaultCache(t *testing.T) {
	s := DefaultSettings()
	cache := s.DefaultCache()
	if cache == nil {
		t.Fatal("expected default cache config")
	}
	if cache.Backend != "locmem" {
		t.Errorf("expected locmem, got %s", cache.Backend)
	}
	if cache.Timeout != 300 {
		t.Errorf("expected 300s timeout, got %d", cache.Timeout)
	}
}

func TestDataSettings(t *testing.T) {
	s := DefaultSettings()
	s.Databases = map[string]*DatabaseSettings{
		"default": {
			Engine:   "postgresql",
			Host:     "db.example.com",
			Port:     5432,
			Name:     "mydb",
			User:     "myuser",
			Password: "mypass",
			Options:  map[string]string{"sslmode": "require"},
		},
	}
	db := s.DefaultDatabase()
	if db.Engine != "postgresql" {
		t.Errorf("expected postgresql, got %s", db.Engine)
	}
	if db.Options["sslmode"] != "require" {
		t.Error("expected sslmode=require in options")
	}
}

func TestGlobalSettingsGet(t *testing.T) {
	Reset()
	defer func() {
		r := recover()
		if r == nil {
			t.Error("expected panic when settings not initialized")
		}
	}()
	Get()
}

func TestGlobalSettingsInitAndGet(t *testing.T) {
	Reset()
	s := DefaultSettings()
	s.ProjectName = "testproject"
	Init(s)

	retrieved := Get()
	if retrieved.ProjectName != "testproject" {
		t.Errorf("expected testproject, got %s", retrieved.ProjectName)
	}
	Reset()
}

func TestGlobalSettingsSet(t *testing.T) {
	Reset()
	s1 := DefaultSettings()
	s1.ProjectName = "first"
	Init(s1)

	s2 := DefaultSettings()
	s2.ProjectName = "second"
	Set(s2)

	retrieved := Get()
	if retrieved.ProjectName != "second" {
		t.Errorf("expected second, got %s", retrieved.ProjectName)
	}
	Reset()
}

func TestFreeze(t *testing.T) {
	Reset()
	s := DefaultSettings()
	s.ProjectName = "frozen"
	Init(s)
	Freeze()

	if !IsFrozen() {
		t.Error("expected settings to be frozen")
	}

	defer func() {
		r := recover()
		if r == nil {
			t.Error("expected panic when modifying frozen settings")
		}
	}()
	Init(DefaultSettings())
}

func TestOverrideForTest(t *testing.T) {
	Reset()
	s1 := DefaultSettings()
	s1.ProjectName = "production"
	Init(s1)

	s2 := DefaultSettings()
	s2.ProjectName = "test"

	OverrideForTest(s2, func() {
		retrieved := Get()
		if retrieved.ProjectName != "test" {
			t.Errorf("expected test inside OverrideForTest, got %s", retrieved.ProjectName)
		}
	})

	retrieved := Get()
	if retrieved.ProjectName != "production" {
		t.Errorf("expected production after OverrideForTest, got %s", retrieved.ProjectName)
	}
	Reset()
}

func TestMustGet(t *testing.T) {
	Reset()
	defer func() {
		r := recover()
		if r == nil {
			t.Error("expected panic when settings not initialized")
		}
	}()
	MustGet()
}

func TestMustGetSuccess(t *testing.T) {
	Reset()
	s := DefaultSettings()
	s.ProjectName = "mustget"
	Init(s)
	retrieved := MustGet()
	if retrieved.ProjectName != "mustget" {
		t.Errorf("expected mustget, got %s", retrieved.ProjectName)
	}
	Reset()
}

func TestTryGet(t *testing.T) {
	Reset()
	_, err := TryGet()
	if err == nil {
		t.Error("expected error when settings not initialized")
	}

	s := DefaultSettings()
	s.ProjectName = "tryget"
	Init(s)
	retrieved, err := TryGet()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if retrieved.ProjectName != "tryget" {
		t.Errorf("expected tryget, got %s", retrieved.ProjectName)
	}
	Reset()
}

func TestSecureSettings(t *testing.T) {
	s := DefaultSettings()
	s.Secure = SecureSettings{
		SSLRedirect:          true,
		SSLHost:               "example.com",
		HSTSSeconds:           31536000,
		HSTSIncludeSubdomains: true,
		HSTSPreload:           true,
		BrowserXSSFilter:      true,
		ContentTypeNoSniff:    true,
		ReferrerPolicy:        "same-origin",
	}
	if !s.Secure.SSLRedirect {
		t.Error("expected SSLRedirect=true")
	}
	if s.Secure.HSTSSeconds != 31536000 {
		t.Errorf("expected HSTSSeconds=31536000, got %d", s.Secure.HSTSSeconds)
	}
}

func TestTemplateConfig(t *testing.T) {
	s := DefaultSettings()
	if len(s.Templates) != 1 {
		t.Fatalf("expected 1 template config, got %d", len(s.Templates))
	}
	tc := s.Templates[0]
	if tc.AppDirs != true {
		t.Error("expected AppDirs=true by default")
	}
	if len(tc.Backends) != 1 || tc.Backends[0] != "html/template" {
		t.Errorf("expected html/template backend, got %v", tc.Backends)
	}
}