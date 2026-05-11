package templates

import (
	"html/template"
	"io"
	"os"
	"path/filepath"

	"github.com/iMerica/jango/apps"
)

type ContextProcessor func() map[string]interface{}

type Engine struct {
	Dirs              []string
	AppDirs           bool
	funcMap           template.FuncMap
	contextProcessors []ContextProcessor
	templates         *template.Template
}

func NewEngine(dirs ...string) *Engine {
	return &Engine{
		Dirs:    dirs,
		funcMap: make(template.FuncMap),
	}
}

func (e *Engine) AddFunc(name string, fn interface{}) {
	e.funcMap[name] = fn
}

func (e *Engine) AddContextProcessor(processor ContextProcessor) {
	e.contextProcessors = append(e.contextProcessors, processor)
}

func (e *Engine) Render(w io.Writer, name string, data interface{}) error {
	if e.templates == nil {
		if err := e.Load(); err != nil {
			return err
		}
	}
	return e.templates.ExecuteTemplate(w, name, e.context(data))
}

func (e *Engine) ExecuteTemplate(w io.Writer, name string, data interface{}) error {
	return e.Render(w, name, data)
}

func (e *Engine) Load() error {
	e.templates = template.New("").Funcs(e.funcMap)
	for _, dir := range e.templateDirs() {
		if err := e.parseDir(dir); err != nil {
			return err
		}
	}
	return nil
}

func (e *Engine) templateDirs() []string {
	dirs := append([]string(nil), e.Dirs...)
	if e.AppDirs {
		for _, cfg := range apps.GlobalRegistry().GetAppConfigs() {
			if cfg.Path == "" {
				continue
			}
			dirs = append(dirs, filepath.Join(cfg.Path, "templates"))
		}
	}
	return dirs
}

func (e *Engine) parseDir(dir string) error {
	return filepath.WalkDir(dir, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return err
		}
		if entry.IsDir() || filepath.Ext(path) != ".html" {
			return nil
		}
		body, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		name := filepath.ToSlash(rel)
		if _, err := e.templates.New(name).Parse(string(body)); err != nil {
			return err
		}
		base := filepath.Base(path)
		if e.templates.Lookup(base) == nil {
			if _, err := e.templates.New(base).Parse(string(body)); err != nil {
				return err
			}
		}
		return nil
	})
}

func (e *Engine) context(data interface{}) interface{} {
	base := make(map[string]interface{})
	for _, processor := range e.contextProcessors {
		for k, v := range processor() {
			base[k] = v
		}
	}
	switch values := data.(type) {
	case map[string]interface{}:
		for k, v := range values {
			base[k] = v
		}
		return base
	default:
		if len(base) == 0 {
			return data
		}
		base["object"] = data
		return base
	}
}
