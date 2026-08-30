package main

import (
	"os"
	"path/filepath"
	"strings"
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

func TestDetectCSVSeparator(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    rune
	}{
		{"semicolons win", "a;b;c\n1;2;3\n", ';'},
		{"commas win", "a,b,c\n1,2,3\n", ','},
		{"tie defaults to comma", "a,b;c\n", ','},
		{"no separator defaults to comma", "abc\n", ','},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := detectCSVSeparator(tc.content); got != tc.want {
				t.Errorf("detectCSVSeparator(%q) = %q, want %q", tc.content, got, tc.want)
			}
		})
	}
}

func TestParseCSVItems_comma(t *testing.T) {
	content := "name,size\nFoo,42\nBar,99\n"
	path := writeTemp(t, "items.csv", content)
	items, err := parseCSVItems(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 2 {
		t.Fatalf("want 2 items, got %d", len(items))
	}
	if got := items[0]["name"]; got != "Foo" {
		t.Errorf("items[0].name = %q, want %q", got, "Foo")
	}
	if got := items[1]["size"]; got != "99" {
		t.Errorf("items[1].size = %q, want %q", got, "99")
	}
}

func TestParseCSVItems_semicolon(t *testing.T) {
	content := "name;size\nAlpha;1\nBeta;2\n"
	path := writeTemp(t, "items.csv", content)
	items, err := parseCSVItems(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 2 {
		t.Fatalf("want 2 items, got %d", len(items))
	}
	if got := items[0]["name"]; got != "Alpha" {
		t.Errorf("items[0].name = %q, want %q", got, "Alpha")
	}
}

func TestParseCSVItems_newlineEscape(t *testing.T) {
	content := "name,desc\nFoo,\"line one\\nline two\"\n"
	path := writeTemp(t, "items.csv", content)
	items, err := parseCSVItems(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 {
		t.Fatalf("want 1 item, got %d", len(items))
	}
	desc := items[0]["desc"]
	if !strings.Contains(desc, "\n") {
		t.Errorf("expected actual newline in desc, got %q", desc)
	}
}

func TestParseCSVItems_quotedFields(t *testing.T) {
	content := "name,desc\n\"Foo, Bar\",\"hello world\"\n"
	path := writeTemp(t, "items.csv", content)
	items, err := parseCSVItems(path)
	if err != nil {
		t.Fatal(err)
	}
	if got := items[0]["name"]; got != "Foo, Bar" {
		t.Errorf("name = %q, want %q", got, "Foo, Bar")
	}
}

func TestParseCSVItems_emptyFile(t *testing.T) {
	path := writeTemp(t, "empty.csv", "")
	items, err := parseCSVItems(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 0 {
		t.Errorf("want 0 items, got %d", len(items))
	}
}

func TestResolveIncludes_noIncludes(t *testing.T) {
	content := "plain content, no directives\n"
	got, err := resolveIncludes(content, filepath.Join("testdata", "includes"))
	if err != nil {
		t.Fatal(err)
	}
	if got != content {
		t.Errorf("got %q, want %q", got, content)
	}
}

func TestResolveIncludes_singleInclude(t *testing.T) {
	dir := filepath.Join("testdata", "includes")
	data, err := os.ReadFile(filepath.Join(dir, "simple_include.md"))
	if err != nil {
		t.Fatal(err)
	}
	got, err := resolveIncludes(string(data), dir)
	if err != nil {
		t.Fatal(err)
	}
	want := "before\nincluded content\n\nafter\n"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestResolveIncludes_multipleIncludes(t *testing.T) {
	dir := filepath.Join("testdata", "includes")
	data, err := os.ReadFile(filepath.Join(dir, "two_includes.md"))
	if err != nil {
		t.Fatal(err)
	}
	got, err := resolveIncludes(string(data), dir)
	if err != nil {
		t.Fatal(err)
	}
	want := "included content\n\nmiddle\nincluded content\n\n"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestResolveIncludes_nestedIncludes(t *testing.T) {
	dir := filepath.Join("testdata", "includes", "deep")
	data, err := os.ReadFile(filepath.Join(dir, "root.md"))
	if err != nil {
		t.Fatal(err)
	}
	got, err := resolveIncludes(string(data), dir)
	if err != nil {
		t.Fatal(err)
	}
	want := "root\nL1\nL2\nL3 content\n\n/L2\n\n/L1\n\nend\n"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestResolveIncludes_missingFile(t *testing.T) {
	_, err := resolveIncludes("{include:does_not_exist.md}", filepath.Join("testdata", "includes"))
	if err == nil {
		t.Error("expected error for missing include file")
	}
}

