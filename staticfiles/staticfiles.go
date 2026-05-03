package staticfiles

import (
	"os"
	"path/filepath"
)

type Finder struct {
	dirs []string
}

func NewFinder(dirs ...string) *Finder {
	return &Finder{dirs: dirs}
}

func (f *Finder) Find(path string) (string, bool) {
	for _, dir := range f.dirs {
		fullPath := filepath.Join(dir, path)
		if _, err := os.Stat(fullPath); err == nil {
			return fullPath, true
		}
	}
	return "", false
}

func (f *Finder) AddDir(dir string) {
	f.dirs = append(f.dirs, dir)
}