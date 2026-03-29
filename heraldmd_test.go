package heraldmd

import (
	"regexp"
	"strings"
	"testing"

	"github.com/indaco/herald"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	east "github.com/yuin/goldmark/extension/ast"
	"github.com/yuin/goldmark/text"
)

var ansiRe = regexp.MustCompile(`\x1b\[[0-9;]*m`)

func stripANSI(s string) string {
	return ansiRe.ReplaceAllString(s, "")
}

func newTestTypography() *herald.Typography {
	return herald.New()
}

func TestRenderHeadings(t *testing.T) {
	ty := newTestTypography()

	tests := []struct {
		name     string
		md       string
		contains string
	}{
		{"H1", "# Title", "Title"},
		{"H2", "## Section", "Section"},
		{"H3", "### Subsection", "Subsection"},
		{"H4", "#### Minor", "Minor"},
		{"H5", "##### Small", "Small"},
		{"H6", "###### Tiny", "Tiny"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := stripANSI(Render(ty, []byte(tc.md)))
			if !strings.Contains(result, tc.contains) {
				t.Errorf("Render(%q) missing %q in:\n%s", tc.md, tc.contains, result)
			}
		})
	}
}

func TestRenderParagraph(t *testing.T) {
	ty := newTestTypography()
	result := stripANSI(Render(ty, []byte("Hello world.")))
	if !strings.Contains(result, "Hello world.") {
		t.Errorf("missing paragraph text in %q", result)
	}
}

func TestRenderInlineStyles(t *testing.T) {
	ty := newTestTypography()

	tests := []struct {
		name     string
		md       string
		contains string
	}{
		{"bold", "**bold text**", "bold text"},
		{"italic", "*italic text*", "italic text"},
		{"code", "`inline code`", "inline code"},
		{"strikethrough", "~~deleted~~", "deleted"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := stripANSI(Render(ty, []byte(tc.md)))
			if !strings.Contains(result, tc.contains) {
				t.Errorf("Render(%q) missing %q in:\n%s", tc.md, tc.contains, result)
			}
		})
	}
}

func TestRenderLink(t *testing.T) {
	ty := newTestTypography()

	t.Run("labeled link", func(t *testing.T) {
		result := stripANSI(Render(ty, []byte("[Go](https://go.dev)")))
		if !strings.Contains(result, "Go") || !strings.Contains(result, "https://go.dev") {
			t.Errorf("missing link content in %q", result)
		}
	})

	t.Run("autolink", func(t *testing.T) {
		result := stripANSI(Render(ty, []byte("<https://go.dev>")))
		if !strings.Contains(result, "https://go.dev") {
			t.Errorf("missing autolink in %q", result)
		}
	})
}

func TestRenderBlockquote(t *testing.T) {
	ty := newTestTypography()
	result := stripANSI(Render(ty, []byte("> A wise quote")))
	if !strings.Contains(result, "A wise quote") {
		t.Errorf("missing blockquote text in %q", result)
	}
}

func TestRenderCodeBlock(t *testing.T) {
	ty := newTestTypography()

	t.Run("fenced", func(t *testing.T) {
		md := "```go\nfmt.Println(\"hello\")\n```"
		result := stripANSI(Render(ty, []byte(md)))
		if !strings.Contains(result, "fmt.Println") {
			t.Errorf("missing code block content in %q", result)
		}
	})

	t.Run("indented", func(t *testing.T) {
		md := "    func main() {\n    }\n"
		result := stripANSI(Render(ty, []byte(md)))
		if !strings.Contains(result, "func main()") {
			t.Errorf("missing indented code block in %q", result)
		}
	})
}

func TestRenderHR(t *testing.T) {
	ty := newTestTypography()
	result := stripANSI(Render(ty, []byte("---")))
	// HR renders as repeated characters.
	if len(result) == 0 {
		t.Error("HR should produce output")
	}
}

func TestRenderUnorderedList(t *testing.T) {
	ty := newTestTypography()
	md := "- Apples\n- Bananas\n- Cherries"
	result := stripANSI(Render(ty, []byte(md)))

	for _, item := range []string{"Apples", "Bananas", "Cherries"} {
		if !strings.Contains(result, item) {
			t.Errorf("missing list item %q in %q", item, result)
		}
	}
}

func TestRenderOrderedList(t *testing.T) {
	ty := newTestTypography()
	md := "1. First\n2. Second\n3. Third"
	result := stripANSI(Render(ty, []byte(md)))

	for _, item := range []string{"First", "Second", "Third"} {
		if !strings.Contains(result, item) {
			t.Errorf("missing list item %q in %q", item, result)
		}
	}
	// Should have numbered markers.
	if !strings.Contains(result, "1.") {
		t.Errorf("missing ordered marker in %q", result)
	}
}

func TestRenderTable(t *testing.T) {
	ty := newTestTypography()
	md := "| Name | Role |\n| --- | --- |\n| Alice | Admin |"
	result := stripANSI(Render(ty, []byte(md)))

	for _, cell := range []string{"Name", "Role", "Alice", "Admin"} {
		if !strings.Contains(result, cell) {
			t.Errorf("missing table cell %q in %q", cell, result)
		}
	}
}

func TestRenderTableAlignment(t *testing.T) {
	ty := newTestTypography()
	md := "| Left | Center | Right |\n| :--- | :---: | ---: |\n| a | b | c |"
	result := stripANSI(Render(ty, []byte(md)))

	// Just verify it renders without error and contains content.
	for _, cell := range []string{"Left", "Center", "Right", "a", "b", "c"} {
		if !strings.Contains(result, cell) {
			t.Errorf("missing table cell %q in %q", cell, result)
		}
	}
}

func TestRenderImage(t *testing.T) {
	ty := newTestTypography()
	md := "![alt text](https://example.com/img.png)"
	result := stripANSI(Render(ty, []byte(md)))

	if !strings.Contains(result, "alt text") {
		t.Errorf("missing image alt text in %q", result)
	}
}

func TestRenderMixed(t *testing.T) {
	ty := newTestTypography()
	md := `# Welcome

This is a **bold** statement with *emphasis*.

- One
- Two

` + "```go\nfmt.Println()\n```" + `

---

> Quote here
`
	result := stripANSI(Render(ty, []byte(md)))

	checks := []string{"Welcome", "bold", "emphasis", "One", "Two", "fmt.Println", "Quote here"}
	for _, want := range checks {
		if !strings.Contains(result, want) {
			t.Errorf("mixed render missing %q", want)
		}
	}
}

func TestRenderEmpty(t *testing.T) {
	ty := newTestTypography()

	t.Run("empty input", func(t *testing.T) {
		result := Render(ty, []byte(""))
		if result != "" {
			t.Errorf("expected empty output, got %q", result)
		}
	})

	t.Run("whitespace only", func(t *testing.T) {
		result := Render(ty, []byte("   \n\n   "))
		if strings.TrimSpace(stripANSI(result)) != "" {
			t.Errorf("expected empty output for whitespace, got %q", result)
		}
	})
}

func TestRenderMultipleParagraphs(t *testing.T) {
	ty := newTestTypography()
	md := "First paragraph.\n\nSecond paragraph."
	result := stripANSI(Render(ty, []byte(md)))

	if !strings.Contains(result, "First paragraph.") {
		t.Errorf("missing first paragraph")
	}
	if !strings.Contains(result, "Second paragraph.") {
		t.Errorf("missing second paragraph")
	}
}

func TestRenderGitHubAlerts(t *testing.T) {
	ty := newTestTypography()

	tests := []struct {
		name     string
		md       string
		contains []string
	}{
		{
			"note",
			"> [!NOTE]\n> Useful information.",
			[]string{"Note", "Useful information."},
		},
		{
			"tip",
			"> [!TIP]\n> Helpful advice.",
			[]string{"Tip", "Helpful advice."},
		},
		{
			"important",
			"> [!IMPORTANT]\n> Key information.",
			[]string{"Important", "Key information."},
		},
		{
			"warning",
			"> [!WARNING]\n> Urgent info.",
			[]string{"Warning", "Urgent info."},
		},
		{
			"caution",
			"> [!CAUTION]\n> Risk ahead.",
			[]string{"Caution", "Risk ahead."},
		},
		{
			"regular blockquote",
			"> Just a normal quote.",
			[]string{"Just a normal quote."},
		},
		{
			"multiline alert",
			"> [!NOTE]\n> First line.\n> Second line.",
			[]string{"Note", "First line.", "Second line."},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := stripANSI(Render(ty, []byte(tc.md)))
			for _, want := range tc.contains {
				if !strings.Contains(result, want) {
					t.Errorf("missing %q in:\n%s", want, result)
				}
			}
		})
	}
}

func TestTryParseAlertEdgeCases(t *testing.T) {
	ty := newTestTypography()

	t.Run("empty blockquote", func(t *testing.T) {
		r := &walker{ty: ty, source: nil}
		bq := ast.NewBlockquote()
		_, _, ok := r.tryParseAlert(bq)
		if ok {
			t.Error("empty blockquote should not match alert")
		}
	})

	t.Run("non-paragraph first child", func(t *testing.T) {
		r := &walker{ty: ty, source: nil}
		bq := ast.NewBlockquote()
		bq.AppendChild(bq, ast.NewThematicBreak())
		_, _, ok := r.tryParseAlert(bq)
		if ok {
			t.Error("non-paragraph first child should not match alert")
		}
	})

	t.Run("non-text first inline", func(t *testing.T) {
		r := &walker{ty: ty, source: nil}
		bq := ast.NewBlockquote()
		para := ast.NewParagraph()
		para.AppendChild(para, ast.NewCodeSpan())
		bq.AppendChild(bq, para)
		_, _, ok := r.tryParseAlert(bq)
		if ok {
			t.Error("non-text first inline should not match alert")
		}
	})

	t.Run("empty paragraph", func(t *testing.T) {
		r := &walker{ty: ty, source: nil}
		bq := ast.NewBlockquote()
		bq.AppendChild(bq, ast.NewParagraph())
		_, _, ok := r.tryParseAlert(bq)
		if ok {
			t.Error("empty paragraph should not match alert")
		}
	})

	t.Run("invalid marker", func(t *testing.T) {
		_, ok := matchAlertMarker("[!INVALID]")
		if ok {
			t.Error("[!INVALID] should not match")
		}
	})

	t.Run("short string", func(t *testing.T) {
		_, ok := matchAlertMarker("[!")
		if ok {
			t.Error("[! should not match")
		}
	})

	t.Run("no bracket", func(t *testing.T) {
		_, ok := matchAlertMarker("just text")
		if ok {
			t.Error("plain text should not match")
		}
	})

	t.Run("multi-paragraph alert", func(t *testing.T) {
		md := "> [!WARNING]\n> First paragraph.\n>\n> Second paragraph."
		result := stripANSI(Render(ty, []byte(md)))
		if !strings.Contains(result, "Warning") {
			t.Errorf("missing Warning label in:\n%s", result)
		}
		if !strings.Contains(result, "First paragraph.") || !strings.Contains(result, "Second paragraph.") {
			t.Errorf("missing multi-paragraph content in:\n%s", result)
		}
	})
}

func TestRenderBlockquoteMultiline(t *testing.T) {
	ty := newTestTypography()
	md := "> Line one\n> Line two"
	result := stripANSI(Render(ty, []byte(md)))

	if !strings.Contains(result, "Line one") || !strings.Contains(result, "Line two") {
		t.Errorf("missing multiline blockquote content in %q", result)
	}
}

func TestRenderBoldItalicNested(t *testing.T) {
	ty := newTestTypography()
	md := "***bold and italic***"
	result := stripANSI(Render(ty, []byte(md)))

	if !strings.Contains(result, "bold and italic") {
		t.Errorf("missing nested bold+italic text in %q", result)
	}
}

func TestRenderFencedCodeBlockNoLang(t *testing.T) {
	ty := newTestTypography()
	md := "```\nplain code\n```"
	result := stripANSI(Render(ty, []byte(md)))
	if !strings.Contains(result, "plain code") {
		t.Errorf("missing fenced code block without lang in %q", result)
	}
}

func TestRenderNestedList(t *testing.T) {
	ty := newTestTypography()

	t.Run("unordered nested", func(t *testing.T) {
		md := "- Parent\n  - Child A\n  - Child B"
		result := stripANSI(Render(ty, []byte(md)))
		for _, want := range []string{"Parent", "Child A", "Child B"} {
			if !strings.Contains(result, want) {
				t.Errorf("missing %q in:\n%s", want, result)
			}
		}
	})

	t.Run("ordered nested", func(t *testing.T) {
		md := "1. First\n   1. Sub one\n   2. Sub two\n2. Second"
		result := stripANSI(Render(ty, []byte(md)))
		for _, want := range []string{"First", "Sub one", "Sub two", "Second"} {
			if !strings.Contains(result, want) {
				t.Errorf("missing %q in:\n%s", want, result)
			}
		}
	})

	t.Run("mixed nesting", func(t *testing.T) {
		md := "- Parent\n  1. Ordered child\n  2. Another"
		result := stripANSI(Render(ty, []byte(md)))
		for _, want := range []string{"Parent", "Ordered child", "Another"} {
			if !strings.Contains(result, want) {
				t.Errorf("missing %q in:\n%s", want, result)
			}
		}
	})

	t.Run("flat list no nesting", func(t *testing.T) {
		md := "- A\n- B\n- C"
		result := stripANSI(Render(ty, []byte(md)))
		for _, want := range []string{"A", "B", "C"} {
			if !strings.Contains(result, want) {
				t.Errorf("missing %q in:\n%s", want, result)
			}
		}
	})
}

func TestRenderHardLineBreak(t *testing.T) {
	ty := newTestTypography()
	// Two trailing spaces create a hard line break in CommonMark.
	md := "Line one  \nLine two"
	result := stripANSI(Render(ty, []byte(md)))
	if !strings.Contains(result, "Line one") || !strings.Contains(result, "Line two") {
		t.Errorf("missing hard line break content in %q", result)
	}
}

func TestRenderImageNoAlt(t *testing.T) {
	ty := newTestTypography()
	md := "![](https://example.com/img.png)"
	result := stripANSI(Render(ty, []byte(md)))
	if !strings.Contains(result, "https://example.com/img.png") {
		t.Errorf("missing image URL fallback in %q", result)
	}
}

func TestRenderLinkSameAsURL(t *testing.T) {
	ty := newTestTypography()
	md := "[https://go.dev](https://go.dev)"
	result := stripANSI(Render(ty, []byte(md)))
	if !strings.Contains(result, "https://go.dev") {
		t.Errorf("missing link in %q", result)
	}
}

func TestRenderRawHTML(t *testing.T) {
	ty := newTestTypography()

	t.Run("inline raw HTML", func(t *testing.T) {
		md := "Text with <br> break"
		result := stripANSI(Render(ty, []byte(md)))
		if !strings.Contains(result, "<br>") {
			t.Errorf("missing inline raw HTML in %q", result)
		}
	})

	t.Run("inline html tag", func(t *testing.T) {
		md := "Some <em>emphasis</em> here"
		result := stripANSI(Render(ty, []byte(md)))
		if !strings.Contains(result, "<em>") {
			t.Errorf("missing inline HTML tag in %q", result)
		}
	})
}

func TestRenderHTMLBlock(t *testing.T) {
	ty := newTestTypography()
	md := "<table>\n<tr><td>cell</td></tr>\n</table>"
	result := stripANSI(Render(ty, []byte(md)))
	if !strings.Contains(result, "cell") {
		t.Errorf("missing HTML block content in %q", result)
	}
}

func TestRenderTableNoAlignment(t *testing.T) {
	ty := newTestTypography()
	md := "| A | B |\n| --- | --- |\n| 1 | 2 |"
	result := stripANSI(Render(ty, []byte(md)))
	for _, cell := range []string{"A", "B", "1", "2"} {
		if !strings.Contains(result, cell) {
			t.Errorf("missing table cell %q in %q", cell, result)
		}
	}
}

func TestRenderSoftLineBreak(t *testing.T) {
	ty := newTestTypography()
	md := "Line one\nLine two"
	result := stripANSI(Render(ty, []byte(md)))
	if !strings.Contains(result, "Line one") || !strings.Contains(result, "Line two") {
		t.Errorf("missing soft line break content in %q", result)
	}
}

// ---------------------------------------------------------------------------
// Direct renderer method tests for hard-to-reach branches
// ---------------------------------------------------------------------------

func TestRenderInlineStringNode(t *testing.T) {
	ty := newTestTypography()
	r := &walker{ty: ty, source: nil}
	node := ast.NewString([]byte("hello string"))
	result := r.renderInline(node)
	if result != "hello string" {
		t.Errorf("renderInline(String) = %q, want %q", result, "hello string")
	}
}

func TestRenderInlineUnknownWithChildren(t *testing.T) {
	ty := newTestTypography()
	source := []byte("child text")
	r := &walker{ty: ty, source: source}

	parent := ast.NewParagraph()
	seg := text.NewSegment(0, len(source))
	textNode := ast.NewTextSegment(seg)
	parent.AppendChild(parent, textNode)

	result := r.renderInline(parent)
	if !strings.Contains(result, "child text") {
		t.Errorf("renderInline(unknown with children) = %q, want content", result)
	}
}

func TestRenderInlineUnknownLeaf(t *testing.T) {
	ty := newTestTypography()
	r := &walker{ty: ty, source: nil}

	node := ast.NewThematicBreak()
	result := r.renderInline(node)
	if result != "" {
		t.Errorf("renderInline(unknown leaf) = %q, want empty", result)
	}
}

func TestRenderBlockUnknown(t *testing.T) {
	ty := newTestTypography()
	source := []byte("fallback text")
	r := &walker{ty: ty, source: source}

	node := ast.NewTextBlock()
	seg := text.NewSegment(0, len(source))
	textNode := ast.NewTextSegment(seg)
	node.AppendChild(node, textNode)

	result := stripANSI(r.renderBlock(node))
	if !strings.Contains(result, "fallback text") {
		t.Errorf("renderBlock(unknown) = %q, want content", result)
	}
}

func TestCollectInlineTextRecursive(t *testing.T) {
	ty := newTestTypography()
	source := []byte("nested text")
	r := &walker{ty: ty, source: source}

	em := ast.NewEmphasis(1)
	seg := text.NewSegment(0, len(source))
	textNode := ast.NewTextSegment(seg)
	em.AppendChild(em, textNode)

	parent := ast.NewEmphasis(2)
	parent.AppendChild(parent, em)

	result := r.collectInlineText(parent)
	if result != "nested text" {
		t.Errorf("collectInlineText(recursive) = %q, want %q", result, "nested text")
	}
}

func TestRenderHeadingLevelAbove6(t *testing.T) {
	ty := newTestTypography()
	r := &walker{ty: ty, source: []byte("deep")}

	heading := ast.NewHeading(7)
	seg := text.NewSegment(0, 4)
	textNode := ast.NewTextSegment(seg)
	heading.AppendChild(heading, textNode)

	result := stripANSI(r.renderHeading(heading))
	if !strings.Contains(result, "deep") {
		t.Errorf("renderHeading(level 7) = %q, want content", result)
	}
}

func TestRenderTableEmpty(t *testing.T) {
	ty := newTestTypography()
	r := &walker{ty: ty, source: nil}

	table := east.NewTable()
	result := r.renderTable(table)
	if result != "" {
		t.Errorf("renderTable(empty) = %q, want empty", result)
	}
}

func TestCollectInlineTextStringNode(t *testing.T) {
	ty := newTestTypography()
	r := &walker{ty: ty, source: nil}

	parent := ast.NewEmphasis(1)
	strNode := ast.NewString([]byte("string value"))
	parent.AppendChild(parent, strNode)

	result := r.collectInlineText(parent)
	if result != "string value" {
		t.Errorf("collectInlineText(String) = %q, want %q", result, "string value")
	}
}

// ---------------------------------------------------------------------------
// Nil Typography guard
// ---------------------------------------------------------------------------

func TestRenderFootnotes(t *testing.T) {
	ty := newTestTypography()
	r := NewRenderer(
		goldmark.WithExtensions(extension.Footnote),
	)
	md := "Text with a footnote[^1] and another[^2].\n\n[^1]: First note.\n[^2]: Second note."
	result := stripANSI(r.Render(ty, []byte(md)))

	// Should contain inline refs.
	if !strings.Contains(result, "[1]") {
		t.Errorf("missing footnote ref [1] in:\n%s", result)
	}
	if !strings.Contains(result, "[2]") {
		t.Errorf("missing footnote ref [2] in:\n%s", result)
	}
	// Should contain note text.
	if !strings.Contains(result, "First note.") {
		t.Errorf("missing footnote text 'First note.' in:\n%s", result)
	}
	if !strings.Contains(result, "Second note.") {
		t.Errorf("missing footnote text 'Second note.' in:\n%s", result)
	}
}

func TestRenderDefinitionList(t *testing.T) {
	ty := newTestTypography()

	// Definition list syntax (PHP Markdown Extra):
	// Term
	// : Description
	r := NewRenderer(
		goldmark.WithExtensions(extension.DefinitionList),
	)
	md := "Go\n:   A compiled language\n\nRust\n:   A systems language"
	result := stripANSI(r.Render(ty, []byte(md)))

	for _, want := range []string{"Go", "A compiled language", "Rust", "A systems language"} {
		if !strings.Contains(result, want) {
			t.Errorf("definition list missing %q in:\n%s", want, result)
		}
	}
}

func TestRenderFootnoteListSkipsNonFootnote(t *testing.T) {
	ty := newTestTypography()
	r := &walker{ty: ty, source: nil}

	fnList := east.NewFootnoteList()
	// Append a non-Footnote child - should be skipped.
	para := ast.NewParagraph()
	fnList.AppendChild(fnList, para)

	result := r.renderFootnoteList(fnList)
	if result != "" {
		t.Errorf("renderFootnoteList with no footnotes should be empty, got %q", result)
	}
}

func TestRenderDefinitionListTrailingTerm(t *testing.T) {
	ty := newTestTypography()
	source := []byte("orphan")
	r := &walker{ty: ty, source: source}

	// Build a DefinitionList with a term but no description - triggers
	// the trailing-term guard.
	dl := east.NewDefinitionList(0, nil)
	term := east.NewDefinitionTerm()
	seg := text.NewSegment(0, len(source))
	textNode := ast.NewTextSegment(seg)
	term.AppendChild(term, textNode)
	dl.AppendChild(dl, term)

	result := stripANSI(r.renderDefinitionList(dl))
	if !strings.Contains(result, "orphan") {
		t.Errorf("trailing term missing in:\n%s", result)
	}
}

func TestBuildListItemsSkipsNonListItem(t *testing.T) {
	ty := newTestTypography()
	source := []byte("text")
	r := &walker{ty: ty, source: source}

	list := ast.NewList('-')
	// Append a non-ListItem child (paragraph) - should be skipped.
	para := ast.NewParagraph()
	seg := text.NewSegment(0, len(source))
	textNode := ast.NewTextSegment(seg)
	para.AppendChild(para, textNode)
	list.AppendChild(list, para)

	items := r.buildListItems(list)
	if len(items) != 0 {
		t.Errorf("buildListItems should skip non-ListItem children, got %d items", len(items))
	}
}

func TestRenderNilTypography(t *testing.T) {
	result := Render(nil, []byte("# Title"))
	if result != "" {
		t.Errorf("Render(nil, ...) = %q, want empty", result)
	}
}

func TestRendererRenderNilTypography(t *testing.T) {
	r := NewRenderer()
	result := r.Render(nil, []byte("# Title"))
	if result != "" {
		t.Errorf("Renderer.Render(nil, ...) = %q, want empty", result)
	}
}

// ---------------------------------------------------------------------------
// NewRenderer
// ---------------------------------------------------------------------------

func TestNewRenderer(t *testing.T) {
	ty := newTestTypography()
	r := NewRenderer()

	md := "# Hello\n\n**bold** and ~~strike~~"
	result := stripANSI(r.Render(ty, []byte(md)))
	for _, want := range []string{"Hello", "bold", "strike"} {
		if !strings.Contains(result, want) {
			t.Errorf("NewRenderer.Render missing %q in %q", want, result)
		}
	}
}

// ---------------------------------------------------------------------------
// Benchmark
// ---------------------------------------------------------------------------

func BenchmarkRender(b *testing.B) {
	ty := herald.New()
	source := []byte(`# Title

Paragraph with **bold** and *italic* text.

- Item one
- Item two
  - Nested item

1. First
2. Second

> A blockquote here.

` + "```go\nfmt.Println(\"hello\")\n```" + `

---

| Name  | Role  |
| ----- | ----- |
| Alice | Admin |
| Bob   | User  |

[Go](https://go.dev) and ~~deleted~~.
`)

	for b.Loop() {
		_ = Render(ty, source)
	}
}
