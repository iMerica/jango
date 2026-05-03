package framework

import (
	"testing"

	"github.com/iMerica/jango/apps"
	"github.com/iMerica/jango/conf"
)

type testApp struct {
	config *apps.AppConfig
}

func (a *testApp) Config() *apps.AppConfig {
	return a.config
}

func TestSetupBasic(t *testing.T) {
	conf.Reset()
	apps.ResetRegistry()

	s := conf.DefaultSettings()
	s.ProjectName = "testproject"
	s.RootURLConf = "testproject/urls"
	s.SecretKey = "a-very-long-secret-key-that-is-at-least-50-characters-long-for-testing!!"

	app := &testApp{
		config: &apps.AppConfig{
			Label: "testapp",
			Name:  "testapp",
		},
	}

	result, err := Setup(s, []apps.App{app})
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	if result.Settings.ProjectName != "testproject" {
		t.Errorf("expected testproject, got %s", result.Settings.ProjectName)
	}

	if result.AppResult.AppsRegistered != 1 {
		t.Errorf("expected 1 app registered, got %d", result.AppResult.AppsRegistered)
	}

	if !conf.IsFrozen() {
		t.Error("expected settings to be frozen after Setup")
	}

	conf.Reset()
	apps.ResetRegistry()
}

func TestSetupNoApps(t *testing.T) {
	conf.Reset()
	apps.ResetRegistry()

	s := conf.DefaultSettings()
	s.ProjectName = "emptyproject"
	s.RootURLConf = "emptyproject/urls"
	s.SecretKey = "a-very-long-secret-key-that-is-at-least-50-characters-long-for-testing!!"

	result, err := Setup(s, []apps.App{})
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	if result.AppResult.AppsRegistered != 0 {
		t.Errorf("expected 0 apps, got %d", result.AppResult.AppsRegistered)
	}

	conf.Reset()
	apps.ResetRegistry()
}

func TestSetupValidationError(t *testing.T) {
	conf.Reset()
	apps.ResetRegistry()

	s := conf.DefaultSettings()
	s.ProjectName = ""
	s.Debug = false
	s.SecretKey = ""

	_, err := Setup(s, []apps.App{})
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}

	conf.Reset()
	apps.ResetRegistry()
}

func TestSetupDeployWarnings(t *testing.T) {
	conf.Reset()
	apps.ResetRegistry()

	s := conf.DefaultSettings()
	s.ProjectName = "prodproject"
	s.RootURLConf = "prodproject/urls"
	s.SecretKey = "a-very-long-secret-key-that-is-at-least-50-characters-long-for-testing!!"
	s.Debug = false
	s.AllowedHosts = []string{"example.com"}
	s.Secure.SSLRedirect = false
	s.Secure.HSTSSeconds = 0
	s.Secure.BrowserXSSFilter = false
	s.Secure.ContentTypeNoSniff = false

	result, err := Setup(s, []apps.App{})
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	if len(result.DeployWarnings) == 0 {
		t.Error("expected deploy warnings for insecure production settings")
	}

	conf.Reset()
	apps.ResetRegistry()
}

func TestSetupFrozenSettings(t *testing.T) {
	conf.Reset()
	apps.ResetRegistry()

	s := conf.DefaultSettings()
	s.ProjectName = "frozenproject"
	s.RootURLConf = "frozenproject/urls"
	s.SecretKey = "a-very-long-secret-key-that-is-at-least-50-characters-long-for-testing!!"

	_, err := Setup(s, []apps.App{})
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	if !conf.IsFrozen() {
		t.Error("expected settings to be frozen after Setup")
	}

	defer func() {
		r := recover()
		if r == nil {
			t.Error("expected panic when modifying frozen settings")
		}
	}()
	conf.Init(conf.DefaultSettings())

	conf.Reset()
	apps.ResetRegistry()
}

func TestSetupOverrideForTest(t *testing.T) {
	conf.Reset()
	apps.ResetRegistry()

	s := conf.DefaultSettings()
	s.ProjectName = "original"
	s.RootURLConf = "original/urls"
	s.SecretKey = "a-very-long-secret-key-that-is-at-least-50-characters-long-for-testing!!"

	_, err := Setup(s, []apps.App{})
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	testSettings := conf.DefaultSettings()
	testSettings.ProjectName = "testoverride"
	testSettings.SecretKey = "a-very-long-secret-key-that-is-at-least-50-characters-long-for-testing!!"

	conf.OverrideForTest(testSettings, func() {
		retrieved := conf.Get()
		if retrieved.ProjectName != "testoverride" {
			t.Errorf("expected testoverride inside OverrideForTest, got %s", retrieved.ProjectName)
		}
	})

	retrieved := conf.Get()
	if retrieved.ProjectName != "original" {
		t.Errorf("expected original after OverrideForTest, got %s", retrieved.ProjectName)
	}

	conf.Reset()
	apps.ResetRegistry()
}
