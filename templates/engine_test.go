package templates_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/iMerica/jango/apps"
	"github.com/iMerica/jango/templates"
)

type testApp struct{ cfg *apps.AppConfig }

func (a testApp) Config() *apps.AppConfig { return a.cfg }

func TestEngineLoadsProjectAndAppTemplates(t *testing.T) {
	apps.ResetRegistry()
	projectDir := t.TempDir()
	appDir := t.TempDir()
	t.Cleanup(apps.ResetRegistry)

	if err := os.WriteFile(filepath.Join(projectDir, "base.html"), []byte(`Hello {{.name}} {{.site}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	appTemplates := filepath.Join(appDir, "templates", "blog")
	if err := os.MkdirAll(appTemplates, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(appTemplates, "detail.html"), []byte(`{{.object}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := apps.GlobalRegistry().Register(testApp{cfg: &apps.AppConfig{Label: "blog", Path: appDir}}); err != nil {
		t.Fatal(err)
	}

	engine := templates.NewEngine(projectDir)
	engine.AppDirs = true
	engine.AddContextProcessor(func() map[string]interface{} {
		return map[string]interface{}{"site": "JanGO"}
	})
	var out bytes.Buffer
	if err := engine.Render(&out, "base.html", map[string]interface{}{"name": "Templates"}); err != nil {
		t.Fatalf("Render returned error: %v", err)
	}
	if out.String() != "Hello Templates JanGO" {
		t.Fatalf("unexpected render: %q", out.String())
	}

	out.Reset()
	if err := engine.Render(&out, "blog/detail.html", "post"); err != nil {
		t.Fatalf("namespaced render returned error: %v", err)
	}
	if out.String() != "post" {
		t.Fatalf("unexpected namespaced render: %q", out.String())
	}
}
