package main

import (
	"fmt"
	"path/filepath"
	"strings"
)

// buildMarkdown accumulates all categories into a single markdown document
// using pandoc fenced divs. Each category is wrapped in its own grid container
// so the CSS can page-break between categories and size cards per category.
func buildMarkdown(categories []Category) string {
	var b strings.Builder
	for _, cat := range categories {
		if len(cat.Items) == 0 {
			continue
		}
		fmt.Fprintf(&b, "::::: {.category .category--%s}\n\n", cssName(cat.Name))
		for _, item := range cat.Items {
			writeCard(&b, cat, item)
		}
		b.WriteString(":::::\n\n")
	}
	return b.String()
}

// writeCard emits a single card div: a colored category bar, an optional
// background watermark, and the item body.
func writeCard(b *strings.Builder, cat Category, item Item) {
	fmt.Fprintf(b, ":::: {.card .card--%s}\n\n", cssName(cat.Name))

	fmt.Fprintf(b, "::: {.card-bar}\n%s\n:::\n\n", barLabel(item))

	if item.ImagePath != "" {
		fmt.Fprintf(b, "::: {.card-watermark style=\"background-image: url('%s')\"}\n:::\n\n", fileURL(item.ImagePath))
	}

	b.WriteString("::: {.card-body}\n\n")
	fmt.Fprintf(b, "# %s\n\n", item.Title)
	if body := strings.TrimSpace(item.BodyMarkdown); body != "" {
		b.WriteString(body)
		b.WriteString("\n\n")
	}
	b.WriteString(":::\n\n")

	b.WriteString("::::\n\n")
}

// barLabel is the text shown in the card's colored top bar.
func barLabel(item Item) string {
	label := strings.ToUpper(item.Category)
	if item.Subcategory != "" {
		label += " · " + item.Subcategory
	}
	return label
}

// cssName normalises a category name into a value safe for a CSS class suffix.
func cssName(s string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(s) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
		default:
			b.WriteRune('-')
		}
	}
	return b.String()
}

// fileURL converts a local filesystem path into a file:// URL that weasyprint
// can resolve on any platform.
func fileURL(p string) string {
	abs, err := filepath.Abs(p)
	if err != nil {
		abs = p
	}
	s := filepath.ToSlash(abs)
	if !strings.HasPrefix(s, "/") {
		s = "/" + s // Windows drive paths like C:/...
	}
	return "file://" + s
}
