package apps

import (
	"fmt"
	"sync"
)

type AppConfig struct {
	Name             string
	Label            string
	VerboseName      string
	Path             string
	DefaultAutoField string

	readyOnce sync.Once
	readyErr  error
}

func (ac *AppConfig) Ready() error {
	ac.readyOnce.Do(func() {
		if registry.state != stageReady {
			ac.readyErr = fmt.Errorf("apps: Ready() called before registry reached ready state (current stage: %d); DB queries during bootstrap are forbidden", registry.state)
			return
		}
	})
	return ac.readyErr
}

func (ac *AppConfig) String() string {
	return fmt.Sprintf("<AppConfig: %s>", ac.Label)
}