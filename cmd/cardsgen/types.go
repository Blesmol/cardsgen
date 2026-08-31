package main

// Config is the on-disk TOML configuration. All fields are optional; missing
// values fall back to the built-in defaults in defaultConfig.
type Config struct {
	Paper    string  `toml:"paper"`     // "a4" or "letter"
	Grid     string  `toml:"grid"`      // global default grid, e.g. "2x3"
	MarginMM *float64 `toml:"margin_mm"` // page margin in millimetres; nil means "use default"
	GapMM    *float64 `toml:"gap_mm"`    // gap between cards in millimetres; nil means "use default"

	CategoryColors map[string]string `toml:"category_colors"` // category name -> CSS color

	// CategoryNames overrides the display name shown on cards per category.
	CategoryNames map[string]string `toml:"category_names"`

	// CategoryGrid overrides the grid per category (config-only, e.g. "3x4").
	CategoryGrid map[string]string `toml:"category_grid"`
}

// Grid is a number of columns and rows of cards on a page.
type Grid struct {
	Cols int
	Rows int
}

// PaperSize is a physical page size in millimetres.
type PaperSize struct {
	Name     string
	WidthMM  float64
	HeightMM float64
}

// Item is a single card sourced from one markdown file.
type Item struct {
	Title        string
	Category     string
	Subcategory  string // immediate parent folder when nested deeper than the category
	BodyMarkdown string // markdown after the leading H1
	ImagePath    string // absolute path to a same-named sibling image, or "" if none
}

// Category groups items sharing a top-level directory.
type Category struct {
	Name  string
	Color string
	Grid  Grid
	Items []Item
}
