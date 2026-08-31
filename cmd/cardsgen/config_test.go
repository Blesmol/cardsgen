package main

import (
	"testing"
)

func fp(f float64) *float64 { return &f }

// TestMergeDefaults_zeroMarginPreserved verifies that an explicit zero value for
// margin_mm / gap_mm in the config is not silently replaced by the built-in default.
func TestMergeDefaults_zeroMarginPreserved(t *testing.T) {
	cfg := Config{MarginMM: fp(0), GapMM: fp(0)}
	merged := mergeDefaults(cfg, defaultConfig())
	if *merged.MarginMM != 0 {
		t.Errorf("MarginMM = %v, want 0 (explicit zero should not be replaced by default)", *merged.MarginMM)
	}
	if *merged.GapMM != 0 {
		t.Errorf("GapMM = %v, want 0 (explicit zero should not be replaced by default)", *merged.GapMM)
	}
}

// TestMergeDefaults_nilUsesDefault verifies that absent margin_mm / gap_mm fields
// (nil pointer = not present in TOML) fall back to the built-in defaults.
func TestMergeDefaults_nilUsesDefault(t *testing.T) {
	merged := mergeDefaults(Config{}, defaultConfig())
	if *merged.MarginMM != 10 {
		t.Errorf("MarginMM = %v, want 10", *merged.MarginMM)
	}
	if *merged.GapMM != 5 {
		t.Errorf("GapMM = %v, want 5", *merged.GapMM)
	}
}

// TestMergeDefaults_nonZeroPreserved verifies that an explicit non-zero value
// is kept as-is and not overwritten by defaults.
func TestMergeDefaults_nonZeroPreserved(t *testing.T) {
	merged := mergeDefaults(Config{MarginMM: fp(3.5), GapMM: fp(1.25)}, defaultConfig())
	if *merged.MarginMM != 3.5 {
		t.Errorf("MarginMM = %v, want 3.5", *merged.MarginMM)
	}
	if *merged.GapMM != 1.25 {
		t.Errorf("GapMM = %v, want 1.25", *merged.GapMM)
	}
}

// TestApplyOverrides_zeroMarginApplied verifies that passing margin=0 via CLI
// sets the field to 0 (not skipped because 0 >= 0 is true).
func TestApplyOverrides_zeroMarginApplied(t *testing.T) {
	cfg := defaultConfig()
	applyOverrides(&cfg, "", "", 0, 0)
	if *cfg.MarginMM != 0 {
		t.Errorf("MarginMM = %v, want 0 after override", *cfg.MarginMM)
	}
	if *cfg.GapMM != 0 {
		t.Errorf("GapMM = %v, want 0 after override", *cfg.GapMM)
	}
}

// TestApplyOverrides_negativeSkipped verifies that the sentinel -1 (CLI default)
// does not overwrite the config value.
func TestApplyOverrides_negativeSkipped(t *testing.T) {
	cfg := defaultConfig()
	applyOverrides(&cfg, "", "", -1, -1)
	if *cfg.MarginMM != 10 {
		t.Errorf("MarginMM = %v, want 10 (negative sentinel should not override)", *cfg.MarginMM)
	}
	if *cfg.GapMM != 5 {
		t.Errorf("GapMM = %v, want 5 (negative sentinel should not override)", *cfg.GapMM)
	}
}
