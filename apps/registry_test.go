package apps

import (
	"sync"
	"testing"
)

type testApp struct {
	config *AppConfig
}

func (a *testApp) Config() *AppConfig {
	return a.config
}

func newTestApp(label, name, path string) *testApp {
	return &testApp{
		config: &AppConfig{
			Label:           label,
			Name:            name,
			Path:            path,
			VerboseName:     label,
			DefaultAutoField: "AutoField",
		},
	}
}

func TestRegisterAndRetrieve(t *testing.T) {
	ResetRegistry()
	app := newTestApp("polls", "myproject.apps.polls", "/apps/polls")
	if err := registry.Register(app); err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	cfg, err := registry.GetAppConfig("polls")
	if err != nil {
		t.Fatalf("GetAppConfig failed: %v", err)
	}
	if cfg.Label != "polls" {
		t.Errorf("expected Label=polls, got %s", cfg.Label)
	}
	if cfg.Name != "myproject.apps.polls" {
		t.Errorf("expected Name=myproject.apps.polls, got %s", cfg.Name)
	}
}

func TestGetAppConfigNotFound(t *testing.T) {
	ResetRegistry()
	_, err := registry.GetAppConfig("nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent app, got nil")
	}
}

func TestDuplicateRegistration(t *testing.T) {
	ResetRegistry()
	app1 := newTestApp("polls", "myproject.apps.polls", "/apps/polls")
	app2 := newTestApp("polls", "other.apps.polls", "/other/polls")
	if err := registry.Register(app1); err != nil {
		t.Fatalf("first Register failed: %v", err)
	}
	if err := registry.Register(app2); err == nil {
		t.Fatal("expected error for duplicate label, got nil")
	}
}

func TestGetAppConfigsDeterministicOrder(t *testing.T) {
	ResetRegistry()
	appA := newTestApp("alpha", "apps.alpha", "/apps/alpha")
	appB := newTestApp("beta", "apps.beta", "/apps/beta")
	appC := newTestApp("gamma", "apps.gamma", "/apps/gamma")

	if err := registry.Register(appA); err != nil {
		t.Fatalf("Register A failed: %v", err)
	}
	if err := registry.Register(appB); err != nil {
		t.Fatalf("Register B failed: %v", err)
	}
	if err := registry.Register(appC); err != nil {
		t.Fatalf("Register C failed: %v", err)
	}

	configs := registry.GetAppConfigs()
	if len(configs) != 3 {
		t.Fatalf("expected 3 configs, got %d", len(configs))
	}
	if configs[0].Label != "alpha" || configs[1].Label != "beta" || configs[2].Label != "gamma" {
		t.Errorf("expected deterministic order [alpha,beta,gamma], got %v",
			[]string{configs[0].Label, configs[1].Label, configs[2].Label})
	}
}

func TestRegisterAndRetrieveModel(t *testing.T) {
	ResetRegistry()
	app := newTestApp("polls", "apps.polls", "/apps/polls")
	if err := registry.Register(app); err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	model := ModelInfo{
		AppLabel:  "polls",
		ModelName: "Question",
		TableName: "polls_question",
		PKField:   "ID",
	}
	if err := registry.RegisterModel("polls", "Question", model); err != nil {
		t.Fatalf("RegisterModel failed: %v", err)
	}

	retrieved, err := registry.GetModel("polls", "Question")
	if err != nil {
		t.Fatalf("GetModel failed: %v", err)
	}
	if retrieved.ModelName != "Question" {
		t.Errorf("expected ModelName=Question, got %s", retrieved.ModelName)
	}
	if retrieved.TableName != "polls_question" {
		t.Errorf("expected TableName=polls_question, got %s", retrieved.TableName)
	}
}

func TestGetModelNotFound(t *testing.T) {
	ResetRegistry()
	_, err := registry.GetModel("nonexistent", "Foo")
	if err == nil {
		t.Fatal("expected error for nonexistent model, got nil")
	}
}

func TestBootstrapThreeStages(t *testing.T) {
	ResetRegistry()
	settings := &Settings{
		ProjectName:   "testproject",
		Debug:         true,
		InstalledApps: []string{"polls"},
	}

	appA := newTestApp("alpha", "apps.alpha", "/apps/alpha")
	appB := newTestApp("beta", "apps.beta", "/apps/beta")

	result, err := Bootstrap(settings, []App{appA, appB})
	if err != nil {
		t.Fatalf("Bootstrap failed: %v", err)
	}

	if result.AppsRegistered != 2 {
		t.Errorf("expected 2 apps registered, got %d", result.AppsRegistered)
	}

	if !registry.IsReady() {
		t.Error("expected registry to be ready after Bootstrap")
	}
	if registry.State() != "ready" {
		t.Errorf("expected state=ready, got %s", registry.State())
	}
}

type readyTrackingApp struct {
	config     *AppConfig
	readyCalls int
	mu         sync.Mutex
}

func (a *readyTrackingApp) Config() *AppConfig {
	return a.config
}

func TestBootstrapWithoutApps(t *testing.T) {
	ResetRegistry()
	settings := &Settings{
		ProjectName: "empty",
		Debug:       true,
	}

	result, err := Bootstrap(settings, []App{})
	if err != nil {
		t.Fatalf("Bootstrap with no apps failed: %v", err)
	}
	if result.AppsRegistered != 0 {
		t.Errorf("expected 0 apps, got %d", result.AppsRegistered)
	}
	if !registry.IsReady() {
		t.Error("expected registry to be ready even with no apps")
	}
}

func TestRegisterCommand(t *testing.T) {
	ResetRegistry()
	cmd := CommandInfo{
		Name:        "runserver",
		AppLabel:    "core",
		Description: "Start the development server",
	}
	if err := registry.RegisterCommand("runserver", cmd); err != nil {
		t.Fatalf("RegisterCommand failed: %v", err)
	}

	commands := registry.GetCommands()
	if len(commands) != 1 {
		t.Fatalf("expected 1 command, got %d", len(commands))
	}
	if commands["runserver"].Description != "Start the development server" {
		t.Errorf("expected runserver command, got %v", commands["runserver"])
	}
}

func TestEmptyConfigLabelRejected(t *testing.T) {
	ResetRegistry()
	app := &testApp{
		config: &AppConfig{Label: "", Name: "test"},
	}
	if err := registry.Register(app); err == nil {
		t.Fatal("expected error for empty label, got nil")
	}
}

func TestAppConfigString(t *testing.T) {
	cfg := &AppConfig{Label: "polls", Name: "apps.polls"}
	if cfg.String() != "<AppConfig: polls>" {
		t.Errorf("expected <AppConfig: polls>, got %s", cfg.String())
	}
}

func TestRegisterModelForUnregisteredApp(t *testing.T) {
	ResetRegistry()
	model := ModelInfo{
		AppLabel:  "nonexistent",
		ModelName: "Foo",
		TableName: "nonexistent_foo",
	}
	if err := registry.RegisterModel("nonexistent", "Foo", model); err == nil {
		t.Fatal("expected error for registering model on unregistered app, got nil")
	}
}

func TestGetModels(t *testing.T) {
	ResetRegistry()
	app := newTestApp("polls", "apps.polls", "/apps/polls")
	registry.Register(app)

	registry.RegisterModel("polls", "Question", ModelInfo{AppLabel: "polls", ModelName: "Question"})
	registry.RegisterModel("polls", "Choice", ModelInfo{AppLabel: "polls", ModelName: "Choice"})

	models, err := registry.GetModels("polls")
	if err != nil {
		t.Fatalf("GetModels failed: %v", err)
	}
	if len(models) != 2 {
		t.Errorf("expected 2 models, got %d", len(models))
	}
}

func TestGetApp(t *testing.T) {
	ResetRegistry()
	app := newTestApp("polls", "apps.polls", "/apps/polls")
	registry.Register(app)

	retrieved, err := registry.GetApp("polls")
	if err != nil {
		t.Fatalf("GetApp failed: %v", err)
	}
	if retrieved.Config().Label != "polls" {
		t.Errorf("expected Label=polls, got %s", retrieved.Config().Label)
	}
}

func TestDuplicateModel(t *testing.T) {
	ResetRegistry()
	app := newTestApp("polls", "apps.polls", "/apps/polls")
	registry.Register(app)

	model := ModelInfo{AppLabel: "polls", ModelName: "Question", TableName: "polls_question"}
	if err := registry.RegisterModel("polls", "Question", model); err != nil {
		t.Fatalf("first RegisterModel failed: %v", err)
	}
	if err := registry.RegisterModel("polls", "Question", model); err == nil {
		t.Fatal("expected error for duplicate model, got nil")
	}
}