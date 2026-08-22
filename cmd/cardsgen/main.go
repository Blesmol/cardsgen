package main

import (
	"fmt"
	"os"
	"path/filepath"

	flag "github.com/spf13/pflag"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run() error {
	configPath := flag.String("config", "", "path to the TOML config file (default: <dir>/cardsgen.cfg)")
	output := flag.String("output", "", "output PDF path (default: <dir>/cards.pdf)")
	mdOut := flag.String("md-out", "", "path for the accumulated markdown (default: <dir>/cards.md)")
	cssOut := flag.String("css-out", "", "path for the generated stylesheet (default: <dir>/cards.css)")
	cssTemplate := flag.String("css-template", "", "path to the base CSS template (default: <exe-dir>/defaults.css)")
	htmlOut := flag.String("html-out", "", "path for the debug HTML (default: <dir>/cards.html)")
	paper := flag.String("paper", "", "paper size override: a4 or letter")
	grid := flag.String("grid", "", "global grid override as COLUMNSxROWS, e.g. 2x3")
	margin := flag.Float64("margin", -1, "page margin override in millimetres")
	gap := flag.Float64("gap", -1, "gap between cards override in millimetres")
	doPandoc := flag.Bool("pandoc", false, "run pandoc to produce the PDF")
	doHTML := flag.Bool("html", false, "also produce a standalone HTML file for debugging")
	flag.Parse()

	args := flag.Args()
	if len(args) == 0 {
		return fmt.Errorf("usage: cardsgen <directory> [flags]")
	}

	absDir, err := filepath.Abs(args[0])
	if err != nil {
		return fmt.Errorf("resolving dir %q: %w", args[0], err)
	}

	if *configPath == "" {
		*configPath = filepath.Join(absDir, "cardsgen.cfg")
	}
	if *output == "" {
		*output = filepath.Join(absDir, "cards.pdf")
	}
	if *mdOut == "" {
		*mdOut = filepath.Join(absDir, "cards.md")
	}
	if *cssOut == "" {
		*cssOut = filepath.Join(absDir, "cards.css")
	}
	if *htmlOut == "" {
		*htmlOut = filepath.Join(absDir, "cards.html")
	}
	if *cssTemplate == "" {
		exeDir, err := executableDir()
		if err != nil {
			return err
		}
		*cssTemplate = filepath.Join(exeDir, "defaults.css")
	}

	cfg, err := loadConfig(*configPath)
	if err != nil {
		return err
	}
	applyOverrides(&cfg, *paper, *grid, *margin, *gap)

	paperSize, err := resolvePaper(cfg.Paper)
	if err != nil {
		return err
	}
	globalGrid, err := parseGrid(cfg.Grid)
	if err != nil {
		return err
	}

	categories, err := scanCategories(absDir, cfg, globalGrid)
	if err != nil {
		return err
	}
	if len(categories) == 0 {
		return fmt.Errorf("no category folders with markdown items found under %q", absDir)
	}

	markdown := buildMarkdown(categories)
	if err := os.WriteFile(*mdOut, []byte(markdown), 0o644); err != nil {
		return fmt.Errorf("writing markdown %q: %w", *mdOut, err)
	}

	baseCSS, err := loadCSSTemplate(*cssTemplate)
	if err != nil {
		return err
	}
	overrideCSS, err := loadCSSOverride(filepath.Join(absDir, "cardsgen.css"))
	if err != nil {
		return err
	}
	css := buildCSS(cfg, paperSize, categories, baseCSS, overrideCSS)
	if err := os.WriteFile(*cssOut, []byte(css), 0o644); err != nil {
		return fmt.Errorf("writing stylesheet %q: %w", *cssOut, err)
	}

	fmt.Printf("scanned %d categories, %d cards; wrote %s and %s\n",
		len(categories), countItems(categories), *mdOut, *cssOut)

	if *doPandoc {
		if err := checkTools(true); err != nil {
			return err
		}
		if err := runPandocPDF(*mdOut, *cssOut, *output); err != nil {
			return err
		}
		fmt.Printf("wrote %s\n", *output)
	}

	if *doHTML {
		if err := checkTools(false); err != nil {
			return err
		}
		if err := runPandocHTML(*mdOut, *cssOut, *htmlOut); err != nil {
			return err
		}
		fmt.Printf("wrote %s\n", *htmlOut)
	}

	return nil
}

// executableDir returns the directory containing the running executable,
// following symlinks so it always points to the real location.
func executableDir() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("resolving executable path: %w", err)
	}
	resolved, err := filepath.EvalSymlinks(exe)
	if err != nil {
		return "", fmt.Errorf("resolving executable symlinks: %w", err)
	}
	return filepath.Dir(resolved), nil
}

// applyOverrides mutates cfg with any CLI flags that were explicitly set.
func applyOverrides(cfg *Config, paper, grid string, margin, gap float64) {
	if paper != "" {
		cfg.Paper = paper
	}
	if grid != "" {
		cfg.Grid = grid
	}
	if margin >= 0 {
		cfg.MarginMM = margin
	}
	if gap >= 0 {
		cfg.GapMM = gap
	}
}

// countItems totals the cards across all categories.
func countItems(categories []Category) int {
	n := 0
	for _, c := range categories {
		n += len(c.Items)
	}
	return n
}
