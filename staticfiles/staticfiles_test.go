package staticfiles_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/iMerica/jango/staticfiles"
)

func TestFinderAndCollector(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()
	if err := os.MkdirAll(filepath.Join(src, "css"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(src, "css", "app.css"), []byte("body{}"), 0o644); err != nil {
		t.Fatal(err)
	}
	finder := staticfiles.NewFinder(src)
	found, ok := finder.Find("css/app.css")
	if !ok || found == "" {
		t.Fatalf("expected static file to be found")
	}
	collected, err := staticfiles.NewCollector(dst, finder).Collect()
	if err != nil {
		t.Fatalf("Collect returned error: %v", err)
	}
	if len(collected) != 1 {
		t.Fatalf("expected one collected file, got %d", len(collected))
	}
	if _, err := os.Stat(filepath.Join(dst, "css", "app.css")); err != nil {
		t.Fatalf("expected collected file: %v", err)
	}
}
