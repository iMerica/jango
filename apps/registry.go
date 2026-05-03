package apps

import (
	"fmt"
	"sync"
)

type startupStage int

const (
	stageInitializing startupStage = iota
	stageModelsRegistered
	stageReady
)

type Registry struct {
	mu       sync.RWMutex
	configs  map[string]*AppConfig
	apps     map[string]App
	ordered  []string
	models   map[string]map[string]ModelInfo
	state    startupStage
	commands map[string]CommandInfo
}

var registry = &Registry{
	configs:  make(map[string]*AppConfig),
	apps:     make(map[string]App),
	models:   make(map[string]map[string]ModelInfo),
	commands: make(map[string]CommandInfo),
}

func GlobalRegistry() *Registry {
	return registry
}

func (r *Registry) Register(app App) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	cfg := app.Config()
	if cfg.Label == "" {
		return fmt.Errorf("apps: app config must have a non-empty Label")
	}
	if _, exists := r.configs[cfg.Label]; exists {
		return fmt.Errorf("apps: duplicate app label %q", cfg.Label)
	}

	r.configs[cfg.Label] = cfg
	r.apps[cfg.Label] = app
	r.ordered = append(r.ordered, cfg.Label)
	r.models[cfg.Label] = make(map[string]ModelInfo)
	return nil
}

func (r *Registry) RegisterModel(appLabel, modelName string, info ModelInfo) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.configs[appLabel]; !exists {
		return fmt.Errorf("apps: no app registered with label %q", appLabel)
	}
	if _, exists := r.models[appLabel][modelName]; exists {
		return fmt.Errorf("apps: duplicate model %q in app %q", modelName, appLabel)
	}
	r.models[appLabel][modelName] = info
	return nil
}

func (r *Registry) RegisterCommand(name string, info CommandInfo) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.commands[name]; exists {
		return fmt.Errorf("apps: duplicate command %q", name)
	}
	r.commands[name] = info
	return nil
}

func (r *Registry) GetAppConfig(label string) (*AppConfig, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	cfg, ok := r.configs[label]
	if !ok {
		return nil, fmt.Errorf("apps: no app with label %q", label)
	}
	return cfg, nil
}

func (r *Registry) GetAppConfigs() []*AppConfig {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]*AppConfig, 0, len(r.ordered))
	for _, label := range r.ordered {
		result = append(result, r.configs[label])
	}
	return result
}

func (r *Registry) GetApp(label string) (App, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	app, ok := r.apps[label]
	if !ok {
		return nil, fmt.Errorf("apps: no app with label %q", label)
	}
	return app, nil
}

func (r *Registry) GetModel(appLabel, modelName string) (ModelInfo, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	appModels, ok := r.models[appLabel]
	if !ok {
		return ModelInfo{}, fmt.Errorf("apps: no app with label %q", appLabel)
	}
	info, ok := appModels[modelName]
	if !ok {
		return ModelInfo{}, fmt.Errorf("apps: no model %q in app %q", modelName, appLabel)
	}
	return info, nil
}

func (r *Registry) GetModels(appLabel string) ([]ModelInfo, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	appModels, ok := r.models[appLabel]
	if !ok {
		return nil, fmt.Errorf("apps: no app with label %q", appLabel)
	}
	result := make([]ModelInfo, 0, len(appModels))
	for _, m := range appModels {
		result = append(result, m)
	}
	return result, nil
}

func (r *Registry) GetCommands() map[string]CommandInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make(map[string]CommandInfo, len(r.commands))
	for k, v := range r.commands {
		result[k] = v
	}
	return result
}

func (r *Registry) IsReady() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.state == stageReady
}

func (r *Registry) State() string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	switch r.state {
	case stageInitializing:
		return "initializing"
	case stageModelsRegistered:
		return "models_registered"
	case stageReady:
		return "ready"
	default:
		return "unknown"
	}
}

func (r *Registry) RunReadyHooks() error {
	r.mu.Lock()
	r.state = stageReady
	r.mu.Unlock()

	for _, label := range r.ordered {
		app := r.apps[label]
		if err := app.Config().Ready(); err != nil {
			return fmt.Errorf("apps: Ready() hook failed for app %q: %w", label, err)
		}
	}
	return nil
}

func (r *Registry) SetModelsRegistered() {
	r.mu.Lock()
	r.state = stageModelsRegistered
	r.mu.Unlock()
}

func ResetRegistry() {
	registry = &Registry{
		configs:  make(map[string]*AppConfig),
		apps:     make(map[string]App),
		models:   make(map[string]map[string]ModelInfo),
		commands: make(map[string]CommandInfo),
	}
}