package main

import "testing"

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
