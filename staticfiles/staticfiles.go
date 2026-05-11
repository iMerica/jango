package staticfiles

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/iMerica/jango/apps"
	"github.com/iMerica/jango/management"
)

type Finder struct {
	dirs    []string
	AppDirs bool
}

func init() {
	management.RegisterCommand(&management.Command{
		Name:        "collectstatic",
		Description: "Collect static files into a single directory",
		Handler: func(args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("collectstatic: destination directory is required")
			}
			_, err := NewCollector(args[0], &Finder{AppDirs: true}).Collect()
			return err
		},
	})
}

func NewFinder(dirs ...string) *Finder {
	return &Finder{dirs: dirs}
}

func (f *Finder) Find(path string) (string, bool) {
	for _, dir := range f.Dirs() {
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

func (f *Finder) Dirs() []string {
	dirs := append([]string(nil), f.dirs...)
	if f.AppDirs {
		for _, cfg := range apps.GlobalRegistry().GetAppConfigs() {
			if cfg.Path == "" {
				continue
			}
			dirs = append(dirs, filepath.Join(cfg.Path, "static"))
		}
	}
	return dirs
}

type CollectedFile struct {
	Source      string
	Destination string
	Duplicate   bool
}

type Collector struct {
	Finder *Finder
	Dest   string
}

func NewCollector(dest string, finder *Finder) *Collector {
	if finder == nil {
		finder = NewFinder()
	}
	return &Collector{Dest: dest, Finder: finder}
}

func (c *Collector) Collect() ([]CollectedFile, error) {
	if c.Dest == "" {
		return nil, fmt.Errorf("staticfiles: destination is required")
	}
	seen := make(map[string]bool)
	var collected []CollectedFile
	for _, dir := range c.Finder.Dirs() {
		err := filepath.WalkDir(dir, func(path string, entry os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if entry.IsDir() {
				return nil
			}
			rel, err := filepath.Rel(dir, path)
			if err != nil {
				return err
			}
			dest := filepath.Join(c.Dest, rel)
			item := CollectedFile{Source: path, Destination: dest, Duplicate: seen[rel]}
			collected = append(collected, item)
			if item.Duplicate {
				return nil
			}
			seen[rel] = true
			if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
				return err
			}
			return copyFile(path, dest)
		})
		if err != nil && !os.IsNotExist(err) {
			return collected, err
		}
	}
	return collected, nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	return err
}
