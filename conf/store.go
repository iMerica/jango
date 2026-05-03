package conf

import (
	"fmt"
	"sync"
)

var (
	global   *Settings
	globalMu sync.RWMutex
	frozen   bool
)

func Init(settings *Settings) {
	globalMu.Lock()
	defer globalMu.Unlock()
	if frozen {
		panic("conf: settings are frozen; use conf.OverrideForTest() in tests")
	}
	global = settings
}

func Get() *Settings {
	globalMu.RLock()
	defer globalMu.RUnlock()
	if global == nil {
		panic("conf: settings not initialized; call conf.Init() or framework.Setup() first")
	}
	return global
}

func Set(settings *Settings) {
	globalMu.Lock()
	defer globalMu.Unlock()
	if frozen {
		panic("conf: settings are frozen; use conf.OverrideForTest() in tests")
	}
	global = settings
}

func Freeze() {
	globalMu.Lock()
	defer globalMu.Unlock()
	if global == nil {
		panic("conf: cannot freeze nil settings")
	}
	frozen = true
}

func IsFrozen() bool {
	globalMu.RLock()
	defer globalMu.RUnlock()
	return frozen
}

func OverrideForTest(settings *Settings, fn func()) {
	globalMu.Lock()
	prev := global
	prevFrozen := frozen
	global = settings
	frozen = false
	globalMu.Unlock()

	defer func() {
		globalMu.Lock()
		global = prev
		frozen = prevFrozen
		globalMu.Unlock()
	}()

	fn()
}

func Reset() {
	globalMu.Lock()
	defer globalMu.Unlock()
	global = nil
	frozen = false
}

func MustGet() *Settings {
	s, err := TryGet()
	if err != nil {
		panic(err)
	}
	return s
}

func TryGet() (*Settings, error) {
	globalMu.RLock()
	defer globalMu.RUnlock()
	if global == nil {
		return nil, fmt.Errorf("conf: settings not initialized; call conf.Init() or framework.Setup() first")
	}
	return global, nil
}