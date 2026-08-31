package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWarnRemove_existingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "tmp.txt")
	if err := os.WriteFile(path, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	var w strings.Builder
	warnRemove(path, &w)
	if w.Len() != 0 {
		t.Errorf("expected no warning for successful removal, got %q", w.String())
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Error("file should have been removed")
	}
}

func TestWarnRemove_nonExistentFile(t *testing.T) {
	var w strings.Builder
	warnRemove(filepath.Join(t.TempDir(), "does_not_exist.txt"), &w)
	if w.Len() != 0 {
		t.Errorf("expected no warning for missing file, got %q", w.String())
	}
}

func TestWarnRemove_unremovableFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "locked.txt")
	if err := os.WriteFile(path, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	// Make the parent directory read-only so Remove fails with a permission error.
	if err := os.Chmod(dir, 0o500); err != nil {
		t.Skip("cannot change directory permissions on this platform")
	}
	t.Cleanup(func() { os.Chmod(dir, 0o700) })

	var w strings.Builder
	warnRemove(path, &w)
	if !strings.Contains(w.String(), "warning:") {
		t.Errorf("expected a warning message, got %q", w.String())
	}
}
