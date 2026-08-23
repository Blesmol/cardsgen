# CardsGen

`cardsgen` is a small app written in golang that scans directories for content files and generates cards out of this.
The result is stored as ready-to-print PDF file.
Internally, it generates a file in Markdown format and then uses [pandoc](https://pandoc.org/) and [weasyprint](https://weasyprint.org/) to convert this first to HTML and from there to PDF.

Cards are organized into **categories** (one subdirectory = one category), each rendered on its own page(s) in a configurable grid (e.g. 2×3). You write card content as Markdown, HTML, YAML, CSV, or plain text.

Goal of this project was to provide an easy option to generate printable cards for TTRPGs with reasoable means of formatting.
There are other projects out there, like e.g. [nanDECK](https://nandeck.com/) that provide more functionality and options, but at the cost of higher complexity.

Ah, and this here is a command-line app.
No fancy GUI, sorry.

Have a look at the [example project](example/) for more insights on how to use this.

## Requirements

- [`pandoc`](https://pandoc.org/) and [`weasyprint`](https://weasyprint.org/) must be available on `PATH`

## Usage

```
cardsgen <directory> [flags]
```

`<directory>` is the root folder containing your card categories. Each top-level subdirectory becomes a category; files placed directly in the root are ignored.

**Flags:**

| Flag | Default | Description |
|---|---|---|
| `--config <path>` | `<dir>/cardsgen.cfg` | Path to the TOML config file |
| `--output <path>` | `<dir>/cards.pdf` | Output PDF path |
| `--css-template <path>` | `defaults.css` next to the binary | Base CSS template |
| `--paper <a4\|letter>` | `a4` | Paper size |
| `--grid <COLxROW>` | `2x3` | Card grid, e.g. `3x4` |
| `--margin <float>` | `10` | Page margin in mm |
| `--gap <float>` | `5` | Gap between cards in mm |
| `--html` | `false` | Also write `generated.html` for browser preview |
| `--md` | `false` | Keep intermediate `generated.md` and `generated.css` |

**Example:**

```
cardsgen ./mycards --grid 2x4 --paper letter --output mycards.pdf
```

## Config file

On the first run, `cardsgen` writes a commented-out `cardsgen.cfg` template into the target directory. All settings are optional; CLI flags override config values.

```toml
# Paper size: "a4" or "letter"
paper = "a4"

# Global card grid (columns x rows per page)
grid = "2x3"

# Page margin and card gap in mm
margin_mm = 10.0
gap_mm    = 5.0

# Human-readable display names for category directories
[category_names]
system_swade = "Savage Worlds"
system_dnd5e = "5e"

# Per-category grid override
[category_grid]
spells = "2x4"
chars  = "2x2"

# Accent color for each category's header bar (any CSS color value)
[category_colors]
spells = "#4a90d9"
chars  = "#e67e22"
```

Categories not listed in `category_colors` receive a deterministic color derived from their directory name.

## Input formats

Each top-level subdirectory is a category. Files inside (and in any nested subdirectories) become cards. The immediate parent folder of a file is used as a **subcategory** label shown in the card's header.

| File | Behavior |
|---|---|
| `*.md` / `*.html` | One card per file. The first `# H1` / `<h1>` is the card title; everything after is the body. |
| `template.md` / `template.html` | Template for the directory. Use `{placeholder}` variables substituted from data files in the same directory. Cardsgen walks up to ancestor directories to find a template. |
| `*.txt` | Key-value data (`key: value` lines). Requires a `template.md` or `template.html`. |
| `*.yml` / `*.yaml` | YAML data. A top-level list produces one card per element; a top-level map produces one card. Requires a template. |
| `*.csv` | CSV data. The first row defines column headers used as placeholder keys. Produces one card per row. Requires a template. |

To add a watermark image to a card, place an image file (`.png`, `.jpg`, `.jpeg`, `.webp`, `.gif`) with the same base name next to the `.md` file (e.g. `fireball.md` + `fireball.webp`).

## Customizing CSS

Card styling is driven by `defaults.css`, which ships alongside the binary. To override or extend styles without modifying that file, create a `cardsgen.css` file next to `cardsgen.cfg` in the target directory. Its contents are appended after the generated stylesheet, so any rules there take precedence over the defaults.

Useful things to override:

- **`.card-bar`** — the colored header bar at the top of each card
- **`.card-body`** — the content area (font, padding, spacing)
- **`.card`** — the card container itself (border, border-radius, background)
- Category-specific rules via `.category--<name>` and `.card--<name>`
