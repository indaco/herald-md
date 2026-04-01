# Herald MD

Convert Markdown to styled terminal output using **herald** typography.

## Features

- Full CommonMark support via [goldmark](https://github.com/yuin/goldmark)
- GFM extensions: ~~strikethrough~~, tables, autolinks
- Themed output with _any_ herald theme
- Nested lists
  - Unordered sub-items
  - Mixed nesting with ordered children
    1. First
    2. Second

## Quick Start

```go
ty := herald.New()
fmt.Println(heraldmd.Render(ty, source))
```

> herald-md is a bridge between Markdown input and herald's typography system.
> It parses the AST and maps each node to the corresponding herald method.

## Supported Elements

| Element   | Markdown     | Herald Method   |
| --------- | ------------ | --------------- |
| Heading   | `# Title`    | `H1()` - `H6()` |
| Paragraph | plain text   | `P()`           |
| Bold      | `**text**`   | `Bold()`        |
| Italic    | `*text*`     | `Italic()`      |
| Code      | `` `code` `` | `Code()`        |
| List      | `- item`     | `UL()` / `OL()` |

## Task List

- [x] CommonMark parsing
- [x] GFM extensions
- [ ] Custom theme support
- [ ] Plugin system

## Inline HTML Elements

As <cite>The Go Programming Language</cite> notes, <q>simplicity is the key to good software</q>.

Set the <var>PORT</var> environment variable, then check the output: <samp>listening on :8080</samp>.

Press <kbd>Ctrl</kbd> + <kbd>C</kbd> to stop the server. This is <mark>important</mark>.

Use <abbr>CSS</abbr> for styling, <small>though terminal support varies</small>.

Text can be <ins>inserted</ins> or <del>removed</del>, and even rendered as <sub>subscript</sub> or <sup>superscript</sup>.

## Alerts

> [!NOTE]
> Useful information that users should know.

> [!TIP]
> Helpful advice for doing things better or more easily.

> [!WARNING]
> Urgent info that needs immediate user attention.

---

Built with `herald` and `goldmark`.
