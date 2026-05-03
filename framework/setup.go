package framework

import (
	"fmt"
	"log/slog"

	"github.com/iMerica/jango/apps"
	"github.com/iMerica/jango/conf"
)

type SetupResult struct {
	Settings       *conf.Settings
	AppResult      *apps.BootstrapResult
	DeployWarnings []string
}

func Setup(settings *conf.Settings, appList []apps.App) (*SetupResult, error) {
	conf.ApplyEnv(settings)

	if errs := conf.Validate(settings); len(errs) > 0 {
		return nil, fmt.Errorf("conf: validation failed: %v", errs)
	}

	conf.ConfigureLogging(settings)

	slog.Info("jango: initializing settings", "project", settings.ProjectName, "debug", settings.Debug)

	conf.Init(settings)

	if settings.ScriptPrefix != "" {
		slog.Info("jango: script prefix set", "prefix", settings.ScriptPrefix)
	}

	appResult, err := apps.Bootstrap(settings, appList)
	if err != nil {
		return nil, fmt.Errorf("framework: bootstrap failed: %w", err)
	}

	conf.Freeze()

	var deployWarnings []string
	if !settings.Debug {
		deployErrs := conf.ValidateDeploy(settings)
		for _, e := range deployErrs {
			msg := fmt.Sprintf("DEPLOY WARNING: %s: %s", e.Field, e.Message)
			deployWarnings = append(deployWarnings, msg)
			slog.Warn(msg)
		}
	}

	slog.Info("jango: setup complete", "apps", appResult.AppsRegistered, "frozen", conf.IsFrozen())

	return &SetupResult{
		Settings:       settings,
		AppResult:      appResult,
		DeployWarnings: deployWarnings,
	}, nil
}
