package codegen

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCreateProject(t *testing.T) {
	tmpDir := t.TempDir()
	target := filepath.Join(tmpDir, "myproject")

	if err := CreateProject("myproject", target); err != nil {
		t.Fatalf("CreateProject failed: %v", err)
	}

	expectedDirs := []string{
		"settings",
		"apps",
		"templates",
		"static",
	}
	for _, dir := range expectedDirs {
		if _, err := os.Stat(filepath.Join(target, dir)); os.IsNotExist(err) {
			t.Errorf("expected directory %s to exist", dir)
		}
	}

	expectedFiles := []string{
		"go.mod",
		"main.go",
		"settings/settings.go",
		"urls.go",
		"middleware.go",
		"Makefile",
	}
	for _, f := range expectedFiles {
		if _, err := os.Stat(filepath.Join(target, f)); os.IsNotExist(err) {
			t.Errorf("expected file %s to exist", f)
		}
	}
}

func TestCreateProjectDuplicateDir(t *testing.T) {
	tmpDir := t.TempDir()
	target := filepath.Join(tmpDir, "existing")
	os.MkdirAll(target, 0755)

	if err := CreateProject("existing", target); err == nil {
		t.Fatal("expected error for existing directory, got nil")
	}
}

func TestCreateApp(t *testing.T) {
	tmpDir := t.TempDir()
	projectDir := filepath.Join(tmpDir, "myproject")
	os.MkdirAll(projectDir, 0755)

	goMod := "module github.com/example/myproject\n\ngo 1.22\n"
	os.WriteFile(filepath.Join(projectDir, "go.mod"), []byte(goMod), 0644)

	appDir := filepath.Join(projectDir, "apps", "polls")
	if err := CreateApp("polls", appDir); err != nil {
		t.Fatalf("CreateApp failed: %v", err)
	}

	expectedSubdirs := []string{
		"models",
		"views",
		"api",
		"migrations",
		filepath.Join("templates", "polls"),
		filepath.Join("static", "polls"),
		filepath.Join("management", "commands"),
		"admin",
		"signals",
		"checks",
	}
	for _, dir := range expectedSubdirs {
		if _, err := os.Stat(filepath.Join(appDir, dir)); os.IsNotExist(err) {
			t.Errorf("expected subdirectory %s to exist", dir)
		}
	}

	expectedFiles := []string{
		"apps.go",
		"urls.go",
	}
	for _, f := range expectedFiles {
		if _, err := os.Stat(filepath.Join(appDir, f)); os.IsNotExist(err) {
			t.Errorf("expected file %s to exist", f)
		}
	}
}

func TestFindProjectInfo(t *testing.T) {
	tmpDir := t.TempDir()
	goMod := "module github.com/test/myproject\n\ngo 1.22\n"
	os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(goMod), 0644)

	appsDir := filepath.Join(tmpDir, "apps", "polls")
	os.MkdirAll(appsDir, 0755)
	appGoCode := `package polls

import "github.com/iMerica/jango/apps"

var App = &pollsApp{}

type pollsApp struct{}

func (a *pollsApp) Config() *apps.AppConfig {
	return &apps.AppConfig{
		Label: "polls",
		Name:  "github.com/test/myproject/apps/polls",
	}
}
`
	os.WriteFile(filepath.Join(appsDir, "apps.go"), []byte(appGoCode), 0644)

	info, err := FindProjectInfo(tmpDir)
	if err != nil {
		t.Fatalf("FindProjectInfo failed: %v", err)
	}
	if info.ModulePath != "github.com/test/myproject" {
		t.Errorf("expected module github.com/test/myproject, got %s", info.ModulePath)
	}

	if len(info.Apps) < 1 {
		t.Fatalf("expected at least 1 app, got %d", len(info.Apps))
	}
	found := false
	for _, app := range info.Apps {
		if app.Label == "polls" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected to find polls app")
	}
}

func TestGenerateRegistry(t *testing.T) {
	tmpDir := t.TempDir()
	info := &ProjectInfo{
		ModulePath: "github.com/test/myproject",
		ProjectDir: tmpDir,
		Apps: []AppEntry{
			{ImportPath: "github.com/test/myproject/apps/polls", Label: "polls", Name: "Polls"},
			{ImportPath: "github.com/test/myproject/apps/blogapi", Label: "blogapi", Name: "Blogapi"},
		},
	}

	if err := GenerateRegistry(info, tmpDir); err != nil {
		t.Fatalf("GenerateRegistry failed: %v", err)
	}

	outputPath := filepath.Join(tmpDir, "zz_generated_registry.go")
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Fatal("expected zz_generated_registry.go to exist")
	}

	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("failed to read generated file: %v", err)
	}

	contentStr := string(content)
	if !contains(contentStr, "polls") {
		t.Error("expected polls import in generated registry")
	}
	if !contains(contentStr, "blogapi") {
		t.Error("expected blogapi import in generated registry")
	}
	if !contains(contentStr, "RegisterInstalledApps") {
		t.Error("expected RegisterInstalledApps function in generated registry")
	}
	if !contains(contentStr, "Bootstrap") {
		t.Error("expected Bootstrap function in generated registry")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsStr(s, substr))
}

func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
