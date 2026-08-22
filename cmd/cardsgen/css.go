package main

import (
	"fmt"
	"os"
	"strings"
)

// loadCSSTemplate reads the base stylesheet from path.
func loadCSSTemplate(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("reading css template %q: %w", path, err)
	}
	return string(data), nil
}

// loadCSSOverride reads an optional per-working-directory CSS override.
// Returns an empty string (no error) when the file does not exist.
func loadCSSOverride(path string) (string, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("reading css override %q: %w", path, err)
	}
	return string(data), nil
}

// buildCSS assembles the print stylesheet from the base template plus the
// generated @page rule and per-category grid/color rules. Card dimensions are
// computed per category from the page size and that category's effective grid,
// so different categories can use different grids while keeping a fixed size.
// override is appended verbatim after the generated rules (empty string = no-op).
func buildCSS(cfg Config, paper PaperSize, categories []Category, template, override string) string {
	var b strings.Builder

	fmt.Fprintf(&b, "@page {\n  size: %.2fmm %.2fmm;\n  margin: %.2fmm;\n}\n\n",
		paper.WidthMM, paper.HeightMM, cfg.MarginMM)

	b.WriteString(strings.TrimRight(template, "\n"))
	b.WriteString("\n\n")

	fmt.Fprintf(&b, ".category {\n  display: grid;\n  gap: %.2fmm;\n  align-content: start;\n}\n\n", cfg.GapMM)

	for _, cat := range categories {
		if len(cat.Items) == 0 {
			continue
		}
		w, h := cardSize(paper, cfg.MarginMM, cfg.GapMM, cat.Grid)
		suffix := cssName(cat.Name)

		fmt.Fprintf(&b, ".category--%s {\n  grid-template-columns: repeat(%d, %.2fmm);\n  grid-auto-rows: %.2fmm;\n}\n",
			suffix, cat.Grid.Cols, w, h)
		fmt.Fprintf(&b, ".card--%s { --card-color: %s; }\n\n", suffix, cat.Color)
	}

	if override != "" {
		b.WriteString(strings.TrimRight(override, "\n"))
		b.WriteString("\n")
	}

	return b.String()
}

// cardSize returns the width and height in millimetres for a single card given
// the page, margins, gap, and grid. A 0.5 mm safety margin is subtracted from
// the available height so that the last row never lands exactly on the page
// boundary, which would cause sub-pixel rounding in the renderer to push it
// onto the next page.
func cardSize(paper PaperSize, margin, gap float64, grid Grid) (w, h float64) {
	availW := paper.WidthMM - 2*margin - float64(grid.Cols-1)*gap
	availH := paper.HeightMM - 2*margin - float64(grid.Rows-1)*gap - 0.5
	return availW / float64(grid.Cols), availH / float64(grid.Rows)
}
