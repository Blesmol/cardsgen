package main

import (
	"strings"
	"testing"
)

func TestFileURL_plainPath(t *testing.T) {
	got := fileURL("/some/path/image.png")
	if got != "file:///some/path/image.png" {
		t.Errorf("fileURL plain = %q, want file:///some/path/image.png", got)
	}
}

func TestFileURL_spaceInPath(t *testing.T) {
	got := fileURL("/some/path with spaces/image.png")
	if strings.Contains(got, " ") {
		t.Errorf("fileURL with spaces should not contain a literal space, got %q", got)
	}
	if !strings.Contains(got, "%20") {
		t.Errorf("fileURL with spaces should percent-encode space as %%20, got %q", got)
	}
}

func TestFileURL_singleQuoteInPath(t *testing.T) {
	// A single quote in the path must be percent-encoded so that it is safe
	// inside a CSS url('...') value.
	got := fileURL("/some/path/it's an image.png")
	if strings.Contains(got, "'") {
		t.Errorf("fileURL with single quote should not contain a literal quote, got %q", got)
	}
	if !strings.Contains(got, "%27") {
		t.Errorf("fileURL with single quote should percent-encode ' as %%27, got %q", got)
	}
}

func TestFileURL_scheme(t *testing.T) {
	got := fileURL("/any/path.png")
	if !strings.HasPrefix(got, "file://") {
		t.Errorf("fileURL should start with file://, got %q", got)
	}
}
