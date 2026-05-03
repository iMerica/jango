package apps

import (
	"fmt"

	"github.com/iMerica/jango/conf"
)

type BootstrapResult struct {
	AppsRegistered   int
	ModelsRegistered int
	ReadyHooksRun    int
}

func Bootstrap(settings *Settings, appList []App) (*BootstrapResult, error) {
	registry.mu.Lock()
	registry.state = stageInitializing
	registry.mu.Unlock()

	conf.Init(settings)

	for _, app := range appList {
		if err := registry.Register(app); err != nil {
			return nil, fmt.Errorf("apps: bootstrap failed during app registration: %w", err)
		}
	}

	registry.SetModelsRegistered()

	readyCount := 0
	if err := registry.RunReadyHooks(); err != nil {
		return nil, err
	}
	readyCount = len(registry.ordered)

	return &BootstrapResult{
		AppsRegistered:   len(registry.ordered),
		ModelsRegistered: countModels(),
		ReadyHooksRun:    readyCount,
	}, nil
}

func countModels() int {
	count := 0
	for _, appModels := range registry.models {
		count += len(appModels)
	}
	return count
}

type Settings = conf.Settings
