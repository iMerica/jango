package templates

import (
	"html/template"
	"io"
)

type Engine struct {
	Dirs     []string
	funcMap  template.FuncMap
	templates *template.Template
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

func (e *Engine) Render(w io.Writer, name string, data interface{}) error {
	if e.templates == nil {
		return nil
	}
	return e.templates.ExecuteTemplate(w, name, data)
}

func (e *Engine) Load() error {
	var err error
	e.templates = template.New("").Funcs(e.funcMap)
	for _, dir := range e.Dirs {
		e.templates, err = e.templates.ParseGlob(dir + "/*.html")
		if err != nil {
			return err
		}
	}
	return nil
}