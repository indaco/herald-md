// Render Markdown with definition lists and footnotes using custom goldmark
// extensions.
//
// This example uses NewRenderer to add goldmark's DefinitionList and Footnote
// extensions beyond the default GFM set. Definition lists are rendered as
// herald DL elements; footnotes use FootnoteRef and FootnoteSection.
//
// This example is a separate Go module with its own go.mod to demonstrate
// custom goldmark extensions without adding them to herald-md's core.
//
// Run from the repository root:
//
//	cd examples/100_definition-list && go run .
package main

import (
	"fmt"

	"github.com/indaco/herald"
	heraldmd "github.com/indaco/herald-md"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
)

const sample = `# Programming Languages

A quick reference of popular languages[^1] and their strengths.

Go
:   A statically typed, compiled language designed for simplicity and concurrency.

Rust
:   A systems programming language focused on safety, speed, and memory management.

Python
:   A dynamic, interpreted language known for readability and a vast ecosystem.

## Why Definition Lists?

Definition lists are useful for:

- Glossaries and terminology
- Configuration option documentation[^2]
- API parameter descriptions

> Definition lists use the PHP Markdown Extra syntax:
> a term on one line, followed by ": " and the description on the next.

[^1]: This list is not exhaustive.
[^2]: See the goldmark extension docs for the full syntax.
`

func main() {
	ty := herald.New()

	// Create a renderer with DefinitionList and Footnote extensions enabled.
	r := heraldmd.NewRenderer(
		goldmark.WithExtensions(
			extension.DefinitionList,
			extension.Footnote,
		),
	)

	fmt.Println(r.Render(ty, []byte(sample)))
}
