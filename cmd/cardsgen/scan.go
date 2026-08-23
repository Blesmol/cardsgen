package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// imageExtensions are the sibling image types recognised for a card, in
// priority order when several exist for the same base name.
var imageExtensions = []string{".png", ".jpg", ".jpeg", ".webp", ".gif"}

// stripBOM removes a UTF-8 BOM that Windows editors sometimes prepend.
func stripBOM(s string) string { return strings.TrimPrefix(s, "\xEF\xBB\xBF") }

var kvKeyLine = regexp.MustCompile(`^([^\s:]+):\s*(.*)$`)
var templatePlaceholder = regexp.MustCompile(`\{([^}]+)\}`)
var blankLines = regexp.MustCompile(`\n{2,}`)
var mdInlineImage = regexp.MustCompile(`!\[([^\]]*)\]\(([^)]+)\)`)
var htmlH1 = regexp.MustCompile(`(?i)<h1[^>]*>(.*?)</h1>`)

type templateKind string

const (
	templateMarkdown templateKind = "markdown"
	templateHTML     templateKind = "html"
)

// templateEntry holds a loaded template and its format kind.
type templateEntry struct {
	content string
	kind    templateKind
}

// scanCategories walks root and groups items (markdown and txt) by their
// top-level directory (the category). Files directly in root are ignored.
func scanCategories(root string, cfg Config, global Grid) ([]Category, error) {
	// Accumulators: items grouped by category name, templates keyed by category, and txt files deferred for later.
	byName := map[string][]Item{}
	templates := map[string]templateEntry{} // absolute directory path → template

	type pendingTxt struct {
		path  string
		parts []string
	}
	var txtPending []pendingTxt

	// Walk the directory tree, classifying each file by extension.
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		// Skip hidden directories; descend into all others.
		if d.IsDir() {
			if path != root && strings.HasPrefix(d.Name(), ".") {
				return filepath.SkipDir
			}
			return nil
		}

		ext := strings.ToLower(filepath.Ext(d.Name()))

		// Build a slash-separated path relative to root and split into parts;
		// the first part is the category name.
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		parts := strings.Split(filepath.ToSlash(rel), "/")
		if len(parts) < 2 {
			return nil // file directly in root; not a card
		}

		// Dispatch by file type: store templates, parse markdown items, defer txt files.
		switch {
		case ext == ".md" && strings.EqualFold(d.Name(), "template.md"):
			if _, exists := templates[filepath.Dir(path)]; exists {
				return fmt.Errorf("%s: directory already contains a template file", rel)
			}
			data, err := os.ReadFile(path)
			if err != nil {
				return fmt.Errorf("%s: %w", rel, err)
			}
			templates[filepath.Dir(path)] = templateEntry{content: stripBOM(string(data)), kind: templateMarkdown}

		case ext == ".html" && strings.EqualFold(d.Name(), "template.html"):
			if _, exists := templates[filepath.Dir(path)]; exists {
				return fmt.Errorf("%s: directory already contains a template file", rel)
			}
			data, err := os.ReadFile(path)
			if err != nil {
				return fmt.Errorf("%s: %w", rel, err)
			}
			templates[filepath.Dir(path)] = templateEntry{content: stripBOM(string(data)), kind: templateHTML}

		case ext == ".md" || ext == ".html":
			item, err := parseItem(path, parts)
			if err != nil {
				return fmt.Errorf("%s: %w", rel, err)
			}
			byName[item.Category] = append(byName[item.Category], item)

		case ext == ".txt":
			txtPending = append(txtPending, pendingTxt{path, parts})
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	// Process txt files after the walk so that template.md is always loaded
	// before it is needed, regardless of filesystem ordering.
	for _, p := range txtPending {
		tmpl, ok := resolveTemplate(root, filepath.Dir(p.path), templates)
		if !ok {
			return nil, fmt.Errorf("%s: no template found in directory or any parent up to %s",
				filepath.Join(p.parts...), root)
		}
		item, err := parseTxtItem(p.path, p.parts, tmpl)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", filepath.Join(p.parts...), err)
		}
		byName[item.Category] = append(byName[item.Category], item)
	}

	// Collect and sort category names for deterministic output order.
	names := make([]string, 0, len(byName))
	for name := range byName {
		names = append(names, name)
	}
	sort.Strings(names)

	// Build the final Category slice, sorting items alphabetically within each category.
	categories := make([]Category, 0, len(names))
	for _, name := range names {
		items := byName[name]
		sort.Slice(items, func(i, j int) bool {
			if items[i].Subcategory != items[j].Subcategory {
				return items[i].Subcategory < items[j].Subcategory
			}
			return items[i].Title < items[j].Title
		})

		grid, err := gridForCategory(cfg, name, global)
		if err != nil {
			return nil, err
		}
		categories = append(categories, Category{
			Name:  nameForCategory(cfg, name),
			Color: colorForCategory(cfg, name),
			Grid:  grid,
			Items: items,
		})
	}
	return categories, nil
}

// resolveTemplate walks up from dir toward root (inclusive) returning the first
// template entry found. Returns zero value, false if none exists in any ancestor.
func resolveTemplate(root, dir string, templates map[string]templateEntry) (templateEntry, bool) {
	for {
		if tmpl, ok := templates[dir]; ok {
			return tmpl, true
		}
		if dir == root {
			break
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return templateEntry{}, false
}

// parseItem reads a markdown file and builds an Item. parts is the file path
// relative to the scan root, split on "/".
func parseItem(path string, parts []string) (Item, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Item{}, err
	}
	// Strip UTF-8 BOM that Windows editors sometimes prepend.
	content := stripBOM(string(data))
	var title, body string
	if strings.ToLower(filepath.Ext(path)) == ".html" {
		title, body = splitHTMLTitle(content)
	} else {
		title, body = splitMarkdownTitle(content)
	}
	if title == "" {
		title = strings.TrimSuffix(parts[len(parts)-1], filepath.Ext(parts[len(parts)-1]))
	}

	category := parts[0]
	var subcategory string
	if len(parts) > 2 {
		subcategory = parts[len(parts)-2] // immediate parent folder when nested
	}

	return Item{
		Title:        title,
		Category:     category,
		Subcategory:  subcategory,
		BodyMarkdown: rewriteBodyImagePaths(body, filepath.Dir(path)),
		ImagePath:    findSiblingImage(path),
	}, nil
}

// parseTxtItem reads a key-value txt file, applies the category template, and
// returns the resulting Item.
func parseTxtItem(path string, parts []string, tmpl templateEntry) (Item, error) {
	kv, err := parseKVFile(path)
	if err != nil {
		return Item{}, err
	}

	rendered := applyTemplate(tmpl.content, kv, filepath.Join(parts...))

	var title, body string
	if tmpl.kind == templateHTML {
		title, body = splitHTMLTitle(rendered)
	} else {
		title, body = splitMarkdownTitle(rendered)
	}
	if title == "" {
		title = strings.TrimSuffix(parts[len(parts)-1], filepath.Ext(parts[len(parts)-1]))
	}

	category := parts[0]
	var subcategory string
	if len(parts) > 2 {
		subcategory = parts[len(parts)-2]
	}

	return Item{
		Title:        title,
		Category:     category,
		Subcategory:  subcategory,
		BodyMarkdown: rewriteBodyImagePaths(body, filepath.Dir(path)),
		ImagePath:    findSiblingImage(path),
	}, nil
}

// parseKVFile reads a UTF-8 txt file and returns a map of key→value pairs.
// Keys are bare words (no whitespace) at the start of a line followed by a
// colon. Values may span multiple lines; internal whitespace is preserved, but
// leading and trailing whitespace/newlines are trimmed.
func parseKVFile(path string) (map[string]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	content := stripBOM(string(data))
	content = strings.ReplaceAll(content, "\r\n", "\n")
	content = strings.ReplaceAll(content, "\r", "\n")

	kv := make(map[string]string)
	var currentKey string
	var valueParts []string

	save := func() {
		if currentKey != "" {
			kv[currentKey] = strings.TrimSpace(strings.Join(valueParts, "\n"))
		}
	}

	for _, line := range strings.Split(content, "\n") {
		if m := kvKeyLine.FindStringSubmatch(line); m != nil {
			save()
			currentKey = m[1]
			valueParts = []string{m[2]}
		} else if currentKey != "" {
			valueParts = append(valueParts, line)
		}
	}
	save()

	return kv, nil
}

// applyTemplate replaces {key} placeholders in tmpl with values from kv.
// Missing keys produce a warning on stderr and are replaced with an empty string.
// Blank lines within a value are converted to pandoc hard line breaks (a
// backslash at the end of a line) so that multi-line values remain compatible
// with inline markdown formatting such as *{description}*.
func applyTemplate(tmpl string, kv map[string]string, srcPath string) string {
	return templatePlaceholder.ReplaceAllStringFunc(tmpl, func(match string) string {
		key := match[1 : len(match)-1]
		if val, ok := kv[key]; ok {
			return blankLines.ReplaceAllString(val, "\\\n")
		}
		fmt.Fprintf(os.Stderr, "warning: %s: no value for placeholder {%s}\n", srcPath, key)
		return ""
	})
}

// splitMarkdownTitle extracts the first level-1 heading as the title and returns the
// remaining markdown as the body.
func splitMarkdownTitle(content string) (title, body string) {
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "# ") {
			title = strings.TrimSpace(strings.TrimPrefix(trimmed, "# "))
			body = strings.Join(lines[i+1:], "\n")
			return title, strings.TrimSpace(body)
		}
	}
	return "", strings.TrimSpace(content)
}

// splitHTMLTitle extracts the first <h1> tag as the title and returns the
// remaining HTML as the body, mirroring splitMarkdownTitle for markdown files.
func splitHTMLTitle(content string) (title, body string) {
	m := htmlH1.FindStringSubmatchIndex(content)
	if m == nil {
		return "", strings.TrimSpace(content)
	}
	title = strings.TrimSpace(content[m[2]:m[3]])
	body = strings.TrimSpace(content[:m[0]] + content[m[1]:])
	return title, body
}

// rewriteBodyImagePaths rewrites relative image paths in markdown body text so
// they still resolve after the body is embedded into cards.md. Absolute paths and URLs are
// left unchanged. Relative paths are resolved against sourceDir and written
// back using the CommonMark angle-bracket destination syntax
// (![alt](<absolute path>)) so that spaces in directory names are handled
// correctly without URL-encoding.
func rewriteBodyImagePaths(body, sourceDir string) string {
	return mdInlineImage.ReplaceAllStringFunc(body, func(match string) string {
		m := mdInlineImage.FindStringSubmatch(match)
		alt, rawPath := m[1], m[2]
		if strings.HasPrefix(rawPath, "http://") || strings.HasPrefix(rawPath, "https://") ||
			strings.HasPrefix(rawPath, "file://") || strings.HasPrefix(rawPath, "/") {
			return match
		}
		abs, err := filepath.Abs(filepath.Join(sourceDir, rawPath))
		if err != nil {
			return match
		}
		return fmt.Sprintf("![%s](<%s>)", alt, filepath.ToSlash(abs))
	})
}

// findSiblingImage returns the path to an image sharing the file's base name,
// or "" when none exists.
func findSiblingImage(mdPath string) string {
	base := strings.TrimSuffix(mdPath, filepath.Ext(mdPath))
	for _, ext := range imageExtensions {
		candidate := base + ext
		if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
			return candidate
		}
	}
	return ""
}
