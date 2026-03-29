<h1 align="center">
  herald-md
</h1>

<h2 align="center" style="font-size: 1.5rem;">
    Render Markdown to the terminal with herald's full theming and composability.
</h2>

<p align="center">
  <a href="https://github.com/indaco/herald-md/actions/workflows/ci.yml" target="_blank">
    <img src="https://github.com/indaco/herald-md/actions/workflows/ci.yml/badge.svg" alt="CI" />
  </a>
  <a href="https://codecov.io/gh/indaco/herald-md" target="_blank">
    <img src="https://codecov.io/gh/indaco/herald-md/branch/main/graph/badge.svg" alt="Code coverage" />
  </a>
  <a href="https://goreportcard.com/report/github.com/indaco/herald-md" target="_blank">
    <img src="https://goreportcard.com/badge/github.com/indaco/herald-md" alt="Go Report Card" />
  </a>
  <a href="https://github.com/indaco/herald-md/actions/workflows/security.yml" target="_blank">
    <img src="https://github.com/indaco/herald-md/actions/workflows/security.yml/badge.svg" alt="Security Scan" />
  </a>
  <a href="https://github.com/indaco/herald-md/releases" target="_blank">
    <img src="https://img.shields.io/github/v/tag/indaco/herald-md?label=version&sort=semver&color=4c1" alt="version">
  </a>
  <a href="https://pkg.go.dev/github.com/indaco/herald-md" target="_blank">
    <img src="https://pkg.go.dev/badge/github.com/indaco/herald-md.svg" alt="Go Reference" />
  </a>
  <a href="LICENSE" target="_blank">
    <img src="https://img.shields.io/badge/license-mit-blue?style=flat-square" alt="License" />
  </a>
</p>

<p align="center">
  <b><a href="#when-to-use-herald-md">When to use</a></b> |
  <b><a href="#installation">Installation</a></b> |
  <b><a href="#quick-start">Quick Start</a></b> |
  <b><a href="#supported-elements">Elements</a></b> |
  <b><a href="#theming">Theming</a></b> |
  <b><a href="#custom-goldmark-extensions">Extensions</a></b> |
  <b><a href="#composing-with-herald">Composing</a></b> |
  <b><a href="#examples">Examples</a></b>
</p>

herald-md parses Markdown with [goldmark](https://github.com/yuin/goldmark) (CommonMark + GFM) and renders it to the terminal through [herald](https://github.com/indaco/herald)'s typography system - themed headings, styled code blocks, tables, alerts, and more, all controlled by a single `Typography` instance.

## When to use herald-md

[glamour](https://github.com/charmbracelet/glamour) is an excellent library for rendering Markdown in the terminal.

herald-md takes a different approach. Paired with [herald](https://github.com/indaco/herald), it gives you a full typography layer where Markdown is one input format among many. Use it when:

- You want **one theme for your entire CLI or TUI** - pick a built-in theme (Dracula, Catppuccin, Base16, Charm) or define your own, and both Markdown and programmatic output use it automatically - no separate style configuration needed.
- You need **fine-grained customization** - functional options and `ColorPalette` give you full control over every element.
- You are **mixing Markdown with programmatic elements** - `Render` returns a plain string you can pass directly to herald's [`Compose`](https://github.com/indaco/herald#composing-elements) alongside other herald calls.

## Installation

Requires Go 1.25+.

```sh
go get github.com/indaco/herald-md@latest
```

## Quick Start

> [!NOTE]
> The Go module path is `github.com/indaco/herald-md` but the package name is `heraldmd`. Use an import alias: `heraldmd "github.com/indaco/herald-md"`.

```go
package main

import (
    "fmt"
    "os"

    "github.com/indaco/herald"
    heraldmd "github.com/indaco/herald-md"
)

func main() {
    source, _ := os.ReadFile("doc.md")
    ty := herald.New()
    fmt.Println(heraldmd.Render(ty, source))
}
```

## Supported elements

| Element             | Markdown syntax      | Herald method                                     |
| ------------------- | -------------------- | ------------------------------------------------- |
| Headings H1–H6      | `# H1` … `###### H6` | `H1()` – `H6()`                                   |
| Paragraph           | plain text           | `P()`                                             |
| Blockquote          | `> text`             | `Blockquote()`                                    |
| Fenced code block   | ` ```lang `          | `CodeBlock(code, lang)`                           |
| Indented code block | 4-space indent       | `CodeBlock(code)`                                 |
| Horizontal rule     | `---`                | `HR()`                                            |
| Unordered list      | `- item`             | `UL()` / `NestUL()`                               |
| Ordered list        | `1. item`            | `OL()` / `NestOL()`                               |
| Bold                | `**text**`           | `Bold()`                                          |
| Italic              | `*text*`             | `Italic()`                                        |
| Strikethrough       | `~~text~~`           | `Strikethrough()`                                 |
| Inline code         | `` `code` ``         | `Code()`                                          |
| Link                | `[label](url)`       | `Link(label, url)`                                |
| Autolink            | `<https://...>`      | `Link(url)`                                       |
| Image               | `![alt](url)`        | `Link(alt, url)` (terminals cannot render images) |
| GFM table           | pipe table syntax    | `Table()` / `TableWithOpts()`                     |
| GitHub alert        | `> [!NOTE]` etc.     | `Alert()` / `Note()` / `Tip()` etc.               |
| Definition list \*  | `Term` + `:  desc`   | `DL()`                                            |
| Footnote ref \*     | `[^1]`               | `FootnoteRef()`                                   |
| Footnote section \* | `[^1]: text`         | `FootnoteSection()`                               |

Elements marked with **\*** require enabling the corresponding goldmark extension via `NewRenderer`. All other elements work out of the box with the default GFM configuration.

- Nested lists are automatically rendered with `NestUL`/`NestOL` when sub-lists are detected, falling back to flat `UL`/`OL` for simple lists.
- Column alignment specified in GFM tables (`:-`, `:-:`, `-:`) is preserved via `WithColumnAlign`.

## Theming

Pass any herald theme to `herald.New()` and all Markdown output will use it:

```go
ty := herald.New(herald.WithTheme(herald.DraculaTheme()))
fmt.Println(heraldmd.Render(ty, source))
```

Custom `ColorPalette` values and individual `With*` options work the same way - configure the `Typography` instance once and `Render` inherits everything.

## Custom goldmark extensions

Use `NewRenderer` to add goldmark extensions beyond the default GFM set:

```go
r := heraldmd.NewRenderer(
    goldmark.WithExtensions(extension.Footnote),
)
fmt.Println(r.Render(ty, source))
```

The package-level `Render` function uses a default renderer with GFM enabled.

## Composing with herald

Because `Render` returns a string, it composes directly with other herald output via `ty.Compose`:

```go
ty := herald.New()

page := ty.Compose(
    ty.H1("My App"),
    heraldmd.Render(ty, readmeBytes), // readmeBytes loaded via os.ReadFile(...)
    ty.HR(),
    ty.KVGroup([][2]string{
        {"Version", "1.0.0"},
        {"License", "MIT"},
    }),
)

fmt.Println(page)
```

## Examples

Runnable examples are in the [`examples/`](examples/) directory:

### Core (`0xx`)

| Example                                      | Description                                                                  | Run                                  |
| -------------------------------------------- | ---------------------------------------------------------------------------- | ------------------------------------ |
| [000_render-file](examples/000_render-file/) | Render a Markdown file with the default theme; accepts any `.md` as argument | `go run ./examples/000_render-file/` |

### Extensions (`1xx`)

Examples in this range are separate Go modules with their own `go.mod` to keep extra dependencies out of herald-md's core. Run them with `cd` into the example directory first.

| Example                                              | Description                                                         | Run                                           |
| ---------------------------------------------------- | ------------------------------------------------------------------- | --------------------------------------------- |
| [100_definition-list](examples/100_definition-list/) | Custom goldmark extension: definition lists rendered as herald `DL` | `cd examples/100_definition-list && go run .` |

## License

This project is licensed under the MIT License - see the [LICENSE](./LICENSE) file for details.
