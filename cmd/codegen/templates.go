package codegen

const appAppsGo = `package {{.AppName}}

import "github.com/iMerica/jango/apps"

// App is the framework registration entry point for this app.
var App AppRegistration

type AppRegistration struct{}

func (a *AppRegistration) Config() *apps.AppConfig {
	return &apps.AppConfig{
		Name:        "{{.ImportPath}}",
		Label:       "{{.AppName}}",
		VerboseName: "{{.AppName}}",
		Path:        "{{.ImportPath}}",
	}
}
`

const appUrlsGo = `package {{.AppName}}

import "github.com/iMerica/jango/urls"

func URLPatterns() []urls.Pattern {
	return []urls.Pattern{}
}
`
