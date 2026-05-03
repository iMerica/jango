package codegen

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

func CreateProject(name, target string) error {
	if target == "" {
		target = name
	}

	projectDir := target
	if _, err := os.Stat(projectDir); !os.IsNotExist(err) {
		return fmt.Errorf("startproject: directory %s already exists", projectDir)
	}

	if err := os.MkdirAll(projectDir, 0755); err != nil {
		return fmt.Errorf("startproject: cannot create directory %s: %w", projectDir, err)
	}

	modulePath := "github.com/example/" + name

	subdirs := []string{
		"settings",
		"apps",
		"templates",
		"static",
	}

	for _, dir := range subdirs {
		if err := os.MkdirAll(filepath.Join(projectDir, dir), 0755); err != nil {
			return fmt.Errorf("startproject: cannot create directory %s: %w", dir, err)
		}
	}

	type projectCtx struct {
		ProjectName string
		ModulePath  string
	}

	ctx := projectCtx{
		ProjectName: name,
		ModulePath:  modulePath,
	}

	files := map[string]string{
		"go.mod":               projectGoMod,
		"main.go":              projectMainGo,
		"settings/settings.go": projectSettingsGo,
		"urls.go":              projectUrlsGo,
		"middleware.go":        projectMiddlewareGo,
		"Makefile":             projectMakefile,
	}

	for fname, content := range files {
		tmpl, err := template.New(fname).Parse(content)
		if err != nil {
			return fmt.Errorf("startproject: template parse error for %s: %w", fname, err)
		}

		var buf strings.Builder
		if err := tmpl.Execute(&buf, ctx); err != nil {
			return fmt.Errorf("startproject: template execute error for %s: %w", fname, err)
		}

		fpath := filepath.Join(projectDir, fname)
		fdir := filepath.Dir(fpath)
		if err := os.MkdirAll(fdir, 0755); err != nil {
			return fmt.Errorf("startproject: cannot create directory %s: %w", fdir, err)
		}

		if err := os.WriteFile(fpath, []byte(buf.String()), 0644); err != nil {
			return fmt.Errorf("startproject: cannot write %s: %w", fpath, err)
		}
	}

	fmt.Printf("Project %q created in %s\n", name, projectDir)
	return nil
}

func CreateApp(name, target string) error {
	appDir := target
	if appDir == "" {
		appDir = filepath.Join("apps", name)
	}

	if _, err := os.Stat(appDir); !os.IsNotExist(err) {
		return fmt.Errorf("startapp: directory %s already exists", appDir)
	}

	parentDir := filepath.Dir(appDir)
	projectDir := findProjectRoot(parentDir)
	if projectDir == "" {
		projectDir = parentDir
	}

	modulePath, err := readModulePath(projectDir)
	if err != nil {
		return fmt.Errorf("startapp: cannot determine module path: %w", err)
	}

	subdirs := []string{
		"models",
		"views",
		"api",
		"migrations",
		filepath.Join("templates", name),
		filepath.Join("static", name),
		filepath.Join("management", "commands"),
		"admin",
		"signals",
		"checks",
	}

	for _, dir := range subdirs {
		if err := os.MkdirAll(filepath.Join(appDir, dir), 0755); err != nil {
			return fmt.Errorf("startapp: cannot create directory %s: %w", dir, err)
		}
	}

	type appCtx struct {
		AppName    string
		ModulePath string
		ImportPath string
	}

	ctx := appCtx{
		AppName:    name,
		ModulePath: modulePath,
		ImportPath: modulePath + "/apps/" + name,
	}

	files := map[string]string{
		"apps.go": appAppsGo,
		"urls.go": appUrlsGo,
	}

	for fname, content := range files {
		tmpl, err := template.New(fname).Parse(content)
		if err != nil {
			return fmt.Errorf("startapp: template parse error for %s: %w", fname, err)
		}

		var buf strings.Builder
		if err := tmpl.Execute(&buf, ctx); err != nil {
			return fmt.Errorf("startapp: template execute error for %s: %w", fname, err)
		}

		fpath := filepath.Join(appDir, fname)
		if err := os.WriteFile(fpath, []byte(buf.String()), 0644); err != nil {
			return fmt.Errorf("startapp: cannot write %s: %w", fpath, err)
		}
	}

	relPath, _ := filepath.Rel(projectDir, appDir)
	fmt.Printf("App %q created at %s\n", name, relPath)
	fmt.Printf("Remember to add %q to INSTALLED_APPS in your settings\n", name)
	return nil
}

func findProjectRoot(dir string) string {
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}

func readModulePath(projectDir string) (string, error) {
	data, err := os.ReadFile(filepath.Join(projectDir, "go.mod"))
	if err != nil {
		return "", err
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "module ")), nil
		}
	}
	return "", fmt.Errorf("no module directive found in go.mod")
}

const projectGoMod = `module {{.ModulePath}}

go 1.22
`

const projectMainGo = `package main

import (
	"os"

	"github.com/iMerica/jango/cli"
)

func main() {
	os.Exit(cli.Run())
}
`

const projectSettingsGo = `package settings

import "github.com/iMerica/jango/conf"

func Load() *conf.Settings {
	s := conf.DefaultSettings()
	s.ProjectName = "{{.ProjectName}}"
	s.InstalledApps = []string{}
	s.Debug = true
	s.RootURLConf = "{{.ModulePath}}/urls"
	s.Databases = map[string]*conf.DatabaseSettings{
		"default": {
			Engine: "postgresql",
			Host:   "localhost",
			Port:   5432,
			Name:   "{{.ProjectName}}",
			User:   "{{.ProjectName}}",
		},
	}
	s.SecretKey = "change-me-in-production"
	s.AllowedHosts = []string{"localhost", "127.0.0.1"}
	s.Middleware = []string{}
	return s
}
`

const projectUrlsGo = `package {{.ProjectName}}

import "github.com/iMerica/jango/urls"

func URLPatterns() []urls.Pattern {
	return []urls.Pattern{}
}
`

const projectMiddlewareGo = `package {{.ProjectName}}

import "github.com/iMerica/jango/middleware"

func MiddlewareStack() []middleware.Middleware {
	return []middleware.Middleware{}
}
`

const projectMakefile = `.PHONY: dev build migrate codegen

dev:
	go run main.go runserver

build:
	go build -o bin/{{.ProjectName}} main.go

migrate:
	go run main.go migrate

codegen:
	go run main.go codegen
`
