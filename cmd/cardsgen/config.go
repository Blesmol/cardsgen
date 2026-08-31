package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/BurntSushi/toml"
)

// defaultConfig returns the built-in defaults used when no config file (or no
// value) is provided.
func defaultConfig() Config {
	margin, gap := 10.0, 5.0
	return Config{
		Paper:          "a4",
		Grid:           "2x3",
		MarginMM:       &margin,
		GapMM:          &gap,
		CategoryColors: map[string]string{},
		CategoryNames:  map[string]string{},
		CategoryGrid:   map[string]string{},
	}
}

// paperSizes holds the supported physical page sizes in millimetres.
var paperSizes = map[string]PaperSize{
	"a4":     {Name: "A4", WidthMM: 210, HeightMM: 297},
	"letter": {Name: "letter", WidthMM: 215.9, HeightMM: 279.4},
}

// loadConfig loads the TOML config at path. If the file does not exist it
// writes a template (all values commented out) and returns the defaults.
func loadConfig(path string) (Config, error) {
	def := defaultConfig()

	if _, err := os.Stat(path); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "no config found; writing commented template to %s\n", path)
		if werr := writeConfigTemplate(path); werr != nil {
			return def, fmt.Errorf("writing config template %q: %w", path, werr)
		}
		return def, nil
	} else if err != nil {
		return def, fmt.Errorf("checking config %q: %w", path, err)
	}

	var fileCfg Config
	meta, err := toml.DecodeFile(path, &fileCfg)
	if err != nil {
		return def, fmt.Errorf("parsing config %q: %w", path, err)
	}
	if keys := meta.Undecoded(); len(keys) > 0 {
		names := make([]string, len(keys))
		for i, k := range keys {
			names[i] = k.String()
		}
		return def, fmt.Errorf("config %q: unrecognized keys: %s", path, strings.Join(names, ", "))
	}
	return mergeDefaults(fileCfg, def), nil
}

// mergeDefaults fills unset scalar fields from def and merges map entries so a
// partial [category_colors] table keeps the built-in colors for untouched categories.
func mergeDefaults(cfg, def Config) Config {
	if strings.TrimSpace(cfg.Paper) == "" {
		cfg.Paper = def.Paper
	}
	if strings.TrimSpace(cfg.Grid) == "" {
		cfg.Grid = def.Grid
	}
	if cfg.MarginMM == nil {
		cfg.MarginMM = def.MarginMM
	}
	if cfg.GapMM == nil {
		cfg.GapMM = def.GapMM
	}
	if cfg.CategoryColors == nil {
		cfg.CategoryColors = map[string]string{}
	}
	if cfg.CategoryNames == nil {
		cfg.CategoryNames = map[string]string{}
	}
	if cfg.CategoryGrid == nil {
		cfg.CategoryGrid = map[string]string{}
	}
	return cfg
}

// parseGrid parses a "CxR" grid specification such as "2x3" into a Grid.
func parseGrid(s string) (Grid, error) {
	parts := strings.Split(strings.ToLower(strings.TrimSpace(s)), "x")
	if len(parts) != 2 {
		return Grid{}, fmt.Errorf("invalid grid %q: expected form CxR like 2x3", s)
	}
	cols, err := strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil || cols < 1 {
		return Grid{}, fmt.Errorf("invalid grid columns in %q", s)
	}
	rows, err := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err != nil || rows < 1 {
		return Grid{}, fmt.Errorf("invalid grid rows in %q", s)
	}
	return Grid{Cols: cols, Rows: rows}, nil
}

// resolvePaper returns the PaperSize for the configured paper name.
func resolvePaper(name string) (PaperSize, error) {
	p, ok := paperSizes[strings.ToLower(strings.TrimSpace(name))]
	if !ok {
		return PaperSize{}, fmt.Errorf("unknown paper %q: use a4 or letter", name)
	}
	return p, nil
}

// gridForCategory returns the effective grid for a category, honouring a
// per-category override from the config and falling back to the global grid.
func gridForCategory(cfg Config, category string, global Grid) (Grid, error) {
	if spec, ok := cfg.CategoryGrid[category]; ok && strings.TrimSpace(spec) != "" {
		g, err := parseGrid(spec)
		if err != nil {
			return Grid{}, fmt.Errorf("category %q: %w", category, err)
		}
		return g, nil
	}
	return global, nil
}

// colorForCategory returns the configured color for a category, or a
// deterministic color derived from the category name via FNV-1a hashing.
func colorForCategory(cfg Config, category string) string {
	if c, ok := cfg.CategoryColors[category]; ok && strings.TrimSpace(c) != "" {
		return c
	}
	return categoryColorFromName(category)
}

// nameForCategory returns the configured display name for a category, or the
// directory name unchanged when no override is set.
func nameForCategory(cfg Config, category string) string {
	if n, ok := cfg.CategoryNames[category]; ok && strings.TrimSpace(n) != "" {
		return n
	}
	return category
}

// categoryColorFromName derives a stable color from a category name by hashing
// it with FNV-1a to select a hue, keeping saturation and lightness fixed.
func categoryColorFromName(name string) string {
	var h uint32 = 2166136261
	for i := 0; i < len(name); i++ {
		h ^= uint32(name[i])
		h *= 16777619
	}
	return fmt.Sprintf("hsl(%d,55%%,35%%)", h%360)
}

// configTemplate documents every option with its default, fully commented out.
const configTemplate = `# ttrpgcards configuration.
# Every value below is optional and shown with its default, commented out.
# Uncomment and edit the lines you want to change.

# Paper size: "a4" or "letter".
# paper = "a4"

# Global card grid as COLUMNSxROWS. Card size is derived automatically.
# grid = "2x3"

# Page margin and gap between cards, in millimetres.
# margin_mm = 10.0
# gap_mm = 5.0

# Per-category colors (used for the card's top bar). Keys are top-level
# directory names. Each unlisted category automatically gets a color derived
# from its name, so this section is only needed to override specific categories.
# [category_colors]
# weapons = "#8c2f2f"
# spells  = "#7b2f8c"

# Per-category display name overrides. Keys are top-level directory names;
# values are the label shown on each card. Useful when directory names must
# be filesystem-safe but you want a nicer label (e.g. spaces, mixed case).
# [category_names]
# weapons         = "Weapons & Armour"
# savage_worlds   = "Savage Worlds"

# Per-category grid overrides (config-only, not available as a CLI flag).
# Each listed category uses its own grid; others use the global grid above.
# Format is COLUMNSxROWS, e.g. "3x4" for 3 columns and 4 rows.
# [category_grid]
# weapons = "3x4"
# armor = "2x2"
`

// writeConfigTemplate writes the commented default config to path.
func writeConfigTemplate(path string) error {
	return os.WriteFile(path, []byte(configTemplate), 0o644)
}
