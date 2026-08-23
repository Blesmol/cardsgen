package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestStripBOM(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"no BOM", "hello", "hello"},
		{"with BOM", "\xEF\xBB\xBFhello", "hello"},
		{"BOM only", "\xEF\xBB\xBF", ""},
		{"empty string", "", ""},
		{"BOM in middle", "hel\xEF\xBB\xBFlo", "hel\xEF\xBB\xBFlo"},
		{"double BOM", "\xEF\xBB\xBF\xEF\xBB\xBFhello", "\xEF\xBB\xBFhello"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := stripBOM(tc.input)
			if got != tc.want {
				t.Errorf("stripBOM(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

func writeTemp(t *testing.T, name, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestParseYAMLItems_singleMap(t *testing.T) {
	path := writeTemp(t, "item.yml", "name: Foo\nsize: 42\n")
	items, err := parseYAMLItems(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 {
		t.Fatalf("want 1 item, got %d", len(items))
	}
	if got := items[0]["name"]; got != "Foo" {
		t.Errorf("name = %q, want %q", got, "Foo")
	}
	if got := items[0]["size"]; got != "42" {
		t.Errorf("size = %q, want %q", got, "42")
	}
}

func TestParseYAMLItems_list(t *testing.T) {
	yaml := "- name: Alpha\n  val: 1\n- name: Beta\n  val: 2\n"
	path := writeTemp(t, "items.yml", yaml)
	items, err := parseYAMLItems(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 2 {
		t.Fatalf("want 2 items, got %d", len(items))
	}
	if got := items[0]["name"]; got != "Alpha" {
		t.Errorf("items[0].name = %q, want %q", got, "Alpha")
	}
	if got := items[1]["name"]; got != "Beta" {
		t.Errorf("items[1].name = %q, want %q", got, "Beta")
	}
}

func TestParseYAMLItems_multilineValue(t *testing.T) {
	yaml := "name: Test\ndesc: |\n  line one\n\n  line two\n"
	path := writeTemp(t, "item.yaml", yaml)
	items, err := parseYAMLItems(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 {
		t.Fatalf("want 1 item, got %d", len(items))
	}
	// yaml literal block scalar preserves internal newlines
	desc := items[0]["desc"]
	if desc == "" {
		t.Error("desc should not be empty")
	}
}

func TestParseYAMLItems_nullValue(t *testing.T) {
	path := writeTemp(t, "item.yml", "name: Test\noptional:\n")
	items, err := parseYAMLItems(path)
	if err != nil {
		t.Fatal(err)
	}
	if got := items[0]["optional"]; got != "" {
		t.Errorf("null value should map to empty string, got %q", got)
	}
}

func TestParseYAMLItems_invalidYAML(t *testing.T) {
	path := writeTemp(t, "bad.yml", "key: [unclosed\n")
	if _, err := parseYAMLItems(path); err == nil {
		t.Error("expected error for invalid YAML")
	}
}

