// Package heraldmd converts Markdown into styled terminal output using herald.
//
// It parses Markdown with goldmark (CommonMark + GFM extensions) and maps each
// AST node to the corresponding herald typography method, preserving the
// document's theme and styling.
//
// Quick start:
//
//	ty := herald.New()
//	fmt.Println(heraldmd.Render(ty, []byte("# Hello\n\nSome **bold** text.")))
//
// For custom goldmark extensions, create a Renderer:
//
//	r := heraldmd.NewRenderer(goldmark.WithExtensions(extension.Footnote))
//	fmt.Println(r.Render(ty, source))
package heraldmd

import (
	"strings"

	"github.com/indaco/herald"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	east "github.com/yuin/goldmark/extension/ast"
	"github.com/yuin/goldmark/text"
)

// Renderer holds a configured goldmark parser and converts Markdown to
// herald-styled output. Use NewRenderer to create one with custom goldmark
// options, or use the package-level Render function for defaults.
type Renderer struct {
	parser goldmark.Markdown
}

// NewRenderer creates a Renderer with the given goldmark options appended
// after the default GFM extension. Pass additional goldmark.Option values
// to enable extra extensions (e.g. footnotes, math).
func NewRenderer(opts ...goldmark.Option) *Renderer {
	all := make([]goldmark.Option, 0, 1+len(opts))
	all = append(all, goldmark.WithExtensions(extension.GFM))
	all = append(all, opts...)
	return &Renderer{parser: goldmark.New(all...)}
}

// Render parses the Markdown source and returns a herald-styled string.
// The provided Typography instance controls the theme and styling of all
// rendered elements. If ty is nil, an empty string is returned.
func (mr *Renderer) Render(ty *herald.Typography, source []byte) string {
	if ty == nil {
		return ""
	}
	doc := mr.parser.Parser().Parse(text.NewReader(source))
	r := &walker{ty: ty, source: source}
	return r.renderDocument(doc)
}

// defaultRenderer is used by the package-level Render function.
var defaultRenderer = NewRenderer()

// Render parses the Markdown source and returns a herald-styled string
// using the default goldmark configuration (CommonMark + GFM).
// The provided Typography instance controls the theme and styling of all
// rendered elements. If ty is nil, an empty string is returned.
func Render(ty *herald.Typography, source []byte) string {
	return defaultRenderer.Render(ty, source)
}

// walker walks a goldmark AST and produces herald-styled output.
type walker struct {
	ty     *herald.Typography
	source []byte
}

// renderDocument renders all top-level block children and composes them.
func (r *walker) renderDocument(doc ast.Node) string {
	blocks := r.collectBlockChildren(doc)
	return r.ty.Compose(blocks...)
}

// collectBlockChildren renders each direct child block of a container node.
func (r *walker) collectBlockChildren(parent ast.Node) []string {
	var blocks []string
	for child := parent.FirstChild(); child != nil; child = child.NextSibling() {
		if rendered := r.renderBlock(child); rendered != "" {
			blocks = append(blocks, rendered)
		}
	}
	return blocks
}

// renderBlock dispatches a block-level node to its herald equivalent.
func (r *walker) renderBlock(node ast.Node) string {
	switch n := node.(type) {
	case *ast.Heading:
		return r.renderHeading(n)
	case *ast.Paragraph:
		return r.ty.P(r.renderInlineChildren(n))
	case *ast.Blockquote:
		return r.renderBlockquote(n)
	case *ast.FencedCodeBlock:
		return r.renderFencedCodeBlock(n)
	case *ast.CodeBlock:
		return r.ty.CodeBlock(r.collectLines(n))
	case *ast.ThematicBreak:
		return r.ty.HR()
	case *ast.List:
		return r.renderList(n)
	case *ast.HTMLBlock:
		return r.collectLines(n)
	case *east.Table:
		return r.renderTable(n)
	case *east.DefinitionList:
		return r.renderDefinitionList(n)
	case *east.FootnoteList:
		return r.renderFootnoteList(n)
	default:
		// Unknown block - render inline content as-is.
		return r.renderInlineChildren(node)
	}
}

// renderHeading maps heading level 1-6 to the corresponding herald method.
func (r *walker) renderHeading(n *ast.Heading) string {
	text := r.renderInlineChildren(n)
	switch n.Level {
	case 1:
		return r.ty.H1(text)
	case 2:
		return r.ty.H2(text)
	case 3:
		return r.ty.H3(text)
	case 4:
		return r.ty.H4(text)
	case 5:
		return r.ty.H5(text)
	case 6:
		return r.ty.H6(text)
	default:
		return r.ty.H6(text)
	}
}

// alertTypes maps GitHub-style alert markers to herald AlertType values.
var alertTypes = map[string]herald.AlertType{
	"NOTE":      herald.AlertNote,
	"TIP":       herald.AlertTip,
	"IMPORTANT": herald.AlertImportant,
	"WARNING":   herald.AlertWarning,
	"CAUTION":   herald.AlertCaution,
}

// renderBlockquote renders nested paragraphs as a blockquote. If the first
// paragraph starts with a GitHub-style alert marker (e.g. [!NOTE]), it is
// rendered as a herald Alert instead.
func (r *walker) renderBlockquote(n *ast.Blockquote) string {
	if alertType, body, ok := r.tryParseAlert(n); ok {
		return r.ty.Alert(alertType, body)
	}

	var lines []string
	for child := n.FirstChild(); child != nil; child = child.NextSibling() {
		lines = append(lines, r.renderInlineChildren(child))
	}
	return r.ty.Blockquote(strings.Join(lines, "\n"))
}

// tryParseAlert checks whether a blockquote is a GitHub-style alert.
// Goldmark splits "[!NOTE]" into multiple Text nodes (e.g. "[", "!NOTE", "]"),
// so tryParseAlert collect the raw text from the leading inline nodes of the first
// paragraph and look for the "[!TYPE]" pattern in the combined string.
func (r *walker) tryParseAlert(n *ast.Blockquote) (herald.AlertType, string, bool) {
	first := n.FirstChild()
	if first == nil {
		return 0, "", false
	}
	if _, ok := first.(*ast.Paragraph); !ok {
		return 0, "", false
	}

	// Collect raw text from leading Text nodes until the closing "]".
	var prefix string
	var afterMarker ast.Node // first node after the marker
	for child := first.FirstChild(); child != nil; child = child.NextSibling() {
		t, ok := child.(*ast.Text)
		if !ok {
			break
		}
		prefix += string(t.Segment.Value(r.source))
		if strings.Contains(prefix, "]") {
			afterMarker = child.NextSibling()
			break
		}
	}

	alertType, ok := matchAlertMarker(prefix)
	if !ok {
		return 0, "", false
	}

	// Build the body from remaining inline nodes + remaining block children.
	var sb strings.Builder
	for sib := afterMarker; sib != nil; sib = sib.NextSibling() {
		sb.WriteString(r.renderInline(sib))
	}
	body := strings.TrimSpace(sb.String())

	// Remaining block children (second paragraph onward).
	for child := first.NextSibling(); child != nil; child = child.NextSibling() {
		text := r.renderInlineChildren(child)
		if body != "" && text != "" {
			body += "\n"
		}
		body += text
	}

	return alertType, body, true
}

// matchAlertMarker checks if a string contains a "[!TYPE]" pattern and returns
// the corresponding AlertType.
func matchAlertMarker(s string) (herald.AlertType, bool) {
	start := strings.Index(s, "[!")
	if start < 0 {
		return 0, false
	}
	end := strings.IndexByte(s[start:], ']')
	if end < 3 {
		return 0, false
	}
	label := s[start+2 : start+end]
	at, ok := alertTypes[label]
	return at, ok
}

// renderFencedCodeBlock renders a fenced code block with optional language.
func (r *walker) renderFencedCodeBlock(n *ast.FencedCodeBlock) string {
	code := r.collectLines(n)
	lang := string(n.Language(r.source))
	if lang != "" {
		return r.ty.CodeBlock(code, lang)
	}
	return r.ty.CodeBlock(code)
}

// ---------------------------------------------------------------------------
// Lists
// ---------------------------------------------------------------------------

// renderList renders a list. If any item has nested sub-lists, it uses
// herald's NestUL/NestOL with ListItem trees. Otherwise it falls back to
// the flat UL/OL for simpler output.
func (r *walker) renderList(n *ast.List) string {
	if r.isTaskList(n) {
		return r.renderTaskList(n)
	}

	if r.listHasNesting(n) {
		items := r.buildListItems(n)
		if n.IsOrdered() {
			return r.ty.NestOL(items...)
		}
		return r.ty.NestUL(items...)
	}

	var items []string
	for child := n.FirstChild(); child != nil; child = child.NextSibling() {
		if _, ok := child.(*ast.ListItem); ok {
			items = append(items, r.renderListItemText(child))
		}
	}
	if n.IsOrdered() {
		return r.ty.OL(items...)
	}
	return r.ty.UL(items...)
}

// isTaskList returns true if the first list item starts with a TaskCheckBox.
func (r *walker) isTaskList(n *ast.List) bool {
	first := n.FirstChild()
	if first == nil {
		return false
	}
	para := first.FirstChild()
	if para == nil {
		return false
	}
	_, ok := para.FirstChild().(*east.TaskCheckBox)
	return ok
}

// renderTaskList renders a task list as plain lines without bullet markers.
func (r *walker) renderTaskList(n *ast.List) string {
	var lines []string
	for child := n.FirstChild(); child != nil; child = child.NextSibling() {
		if _, ok := child.(*ast.ListItem); ok {
			lines = append(lines, r.renderListItemText(child))
		}
	}
	return strings.Join(lines, "\n")
}

// listHasNesting returns true if any list item contains a nested list.
func (r *walker) listHasNesting(n *ast.List) bool {
	for li := n.FirstChild(); li != nil; li = li.NextSibling() {
		for child := li.FirstChild(); child != nil; child = child.NextSibling() {
			if _, ok := child.(*ast.List); ok {
				return true
			}
		}
	}
	return false
}

// buildListItems recursively builds herald.ListItem trees from a goldmark List.
func (r *walker) buildListItems(list *ast.List) []herald.ListItem {
	var items []herald.ListItem
	for child := list.FirstChild(); child != nil; child = child.NextSibling() {
		if _, ok := child.(*ast.ListItem); !ok {
			continue
		}
		item := herald.ListItem{Text: r.renderListItemText(child)}
		for gc := child.FirstChild(); gc != nil; gc = gc.NextSibling() {
			if nested, ok := gc.(*ast.List); ok {
				item.Children = r.buildListItems(nested)
				if nested.IsOrdered() {
					item.Kind = herald.Ordered
				}
			}
		}
		items = append(items, item)
	}
	return items
}

// renderListItemText renders the inline text of a list item, skipping
// any nested list children.
func (r *walker) renderListItemText(node ast.Node) string {
	var parts []string
	for child := node.FirstChild(); child != nil; child = child.NextSibling() {
		if _, ok := child.(*ast.List); ok {
			continue
		}
		parts = append(parts, r.renderInlineChildren(child))
	}
	return strings.Join(parts, " ")
}

// ---------------------------------------------------------------------------
// Definition lists
// ---------------------------------------------------------------------------

// renderDefinitionList converts a goldmark DefinitionList (from the
// extension.DefinitionList extension) into herald's DL method. Terms and
// descriptions are paired in order; unpaired trailing terms are rendered
// with an empty description.
func (r *walker) renderDefinitionList(n *east.DefinitionList) string {
	var pairs [][2]string
	var currentTerm string
	for child := n.FirstChild(); child != nil; child = child.NextSibling() {
		switch child.Kind() {
		case east.KindDefinitionTerm:
			currentTerm = r.renderInlineChildren(child)
		case east.KindDefinitionDescription:
			desc := r.renderInlineChildren(child)
			pairs = append(pairs, [2]string{currentTerm, desc})
			currentTerm = ""
		}
	}
	if currentTerm != "" {
		pairs = append(pairs, [2]string{currentTerm, ""})
	}
	return r.ty.DL(pairs)
}

// ---------------------------------------------------------------------------
// Footnotes
// ---------------------------------------------------------------------------

// renderFootnoteList converts a goldmark FootnoteList (from the
// extension.Footnote extension) into herald's FootnoteSection method.
// Each Footnote child's inline content becomes a note string.
func (r *walker) renderFootnoteList(n *east.FootnoteList) string {
	var notes []string
	for child := n.FirstChild(); child != nil; child = child.NextSibling() {
		if _, ok := child.(*east.Footnote); !ok {
			continue
		}
		notes = append(notes, r.renderInlineChildren(child))
	}
	return r.ty.FootnoteSection(notes)
}

// ---------------------------------------------------------------------------
// Tables
// ---------------------------------------------------------------------------

// renderTable converts a GFM table into herald's Table method.
func (r *walker) renderTable(n *east.Table) string {
	var rows [][]string
	for child := n.FirstChild(); child != nil; child = child.NextSibling() {
		switch child.Kind() {
		case east.KindTableHeader, east.KindTableRow:
			row := r.renderTableRow(child)
			rows = append(rows, row)
		}
	}
	if len(rows) == 0 {
		return ""
	}

	// Map GFM alignments to herald alignments.
	var opts []herald.TableOption
	for i, align := range n.Alignments {
		switch align {
		case east.AlignCenter:
			opts = append(opts, herald.WithColumnAlign(i, herald.AlignCenter))
		case east.AlignRight:
			opts = append(opts, herald.WithColumnAlign(i, herald.AlignRight))
		}
	}

	if len(opts) > 0 {
		return r.ty.TableWithOpts(rows, opts...)
	}
	return r.ty.Table(rows)
}

// renderTableRow collects cell text from a table row or header.
func (r *walker) renderTableRow(node ast.Node) []string {
	var cells []string
	for cell := node.FirstChild(); cell != nil; cell = cell.NextSibling() {
		cells = append(cells, r.renderInlineChildren(cell))
	}
	return cells
}

// ---------------------------------------------------------------------------
// Inline elements
// ---------------------------------------------------------------------------

// htmlTagStyles maps HTML tag names to herald inline methods.
var htmlTagStyles = map[string]func(*walker, string) string{
	"q":      func(w *walker, s string) string { return w.ty.Q(s) },
	"cite":   func(w *walker, s string) string { return w.ty.Cite(s) },
	"samp":   func(w *walker, s string) string { return w.ty.Samp(s) },
	"var":    func(w *walker, s string) string { return w.ty.Var(s) },
	"kbd":    func(w *walker, s string) string { return w.ty.Kbd(s) },
	"mark":   func(w *walker, s string) string { return w.ty.Mark(s) },
	"ins":    func(w *walker, s string) string { return w.ty.Ins(s) },
	"del":    func(w *walker, s string) string { return w.ty.Del(s) },
	"sub":    func(w *walker, s string) string { return w.ty.Sub(s) },
	"sup":    func(w *walker, s string) string { return w.ty.Sup(s) },
	"abbr":   func(w *walker, s string) string { return w.ty.Abbr(s) },
	"b":      func(w *walker, s string) string { return w.ty.Bold(s) },
	"i":      func(w *walker, s string) string { return w.ty.Italic(s) },
	"u":      func(w *walker, s string) string { return w.ty.Underline(s) },
	"s":      func(w *walker, s string) string { return w.ty.Strikethrough(s) },
	"em":     func(w *walker, s string) string { return w.ty.Italic(s) },
	"strong": func(w *walker, s string) string { return w.ty.Bold(s) },
	"small":  func(w *walker, s string) string { return w.ty.Small(s) },
	"code":   func(w *walker, s string) string { return w.ty.Code(s) },
}

// parseOpenTag extracts the tag name from an opening HTML tag like "<q>" or
// "<abbr>". Returns the tag name and true, or "" and false if not an opening tag.
func parseOpenTag(raw string) (string, bool) {
	if len(raw) < 3 || raw[0] != '<' || raw[1] == '/' || raw[len(raw)-1] != '>' {
		return "", false
	}
	tag := raw[1 : len(raw)-1]
	// Handle self-closing tags: <br/>, <hr/>.
	tag = strings.TrimSuffix(tag, "/")
	// Handle tags with attributes: <abbr title="...">.
	if i := strings.IndexByte(tag, ' '); i > 0 {
		tag = tag[:i]
	}
	return strings.ToLower(tag), true
}

// voidTags lists HTML void elements that have no closing tag and no content.
var voidTags = map[string]func(*walker) string{
	"br": func(w *walker) string { return w.ty.BR() },
}

// renderInlineChildren walks inline children of a node and concatenates
// their styled text. When an opening HTML tag for a known element (e.g.
// <q>, <cite>, <samp>, <var>) is encountered, the content between the
// opening and closing tags is collected and rendered via the corresponding
// herald method.
func (r *walker) renderInlineChildren(node ast.Node) string {
	var sb strings.Builder
	child := node.FirstChild()
	for child != nil {
		rh, isRaw := child.(*ast.RawHTML)
		if isRaw {
			raw := r.collectSegments(rh)
			if tag, ok := parseOpenTag(raw); ok {
				// Handle void (self-closing) elements like <br>.
				if voidFn, isVoid := voidTags[tag]; isVoid {
					sb.WriteString(voidFn(r))
					child = child.NextSibling()
					continue
				}
				if styleFn, known := htmlTagStyles[tag]; known {
					content, next := r.collectHTMLTagContent(child.NextSibling(), tag)
					sb.WriteString(styleFn(r, content))
					child = next
					continue
				}
			}
		}
		sb.WriteString(r.renderInline(child))
		child = child.NextSibling()
	}
	return sb.String()
}

// collectHTMLTagContent collects inline content after an opening HTML tag
// until the matching closing tag is found. Returns the collected text and
// the node after the closing tag (to resume iteration).
func (r *walker) collectHTMLTagContent(start ast.Node, tag string) (string, ast.Node) {
	closeTag := "</" + tag + ">"
	var content strings.Builder
	node := start
	for node != nil {
		if rh, ok := node.(*ast.RawHTML); ok {
			raw := r.collectSegments(rh)
			if strings.EqualFold(raw, closeTag) {
				return content.String(), node.NextSibling()
			}
			content.WriteString(raw)
		} else {
			content.WriteString(r.renderInline(node))
		}
		node = node.NextSibling()
	}
	return content.String(), nil
}

// renderInline dispatches an inline node to its herald equivalent.
func (r *walker) renderInline(node ast.Node) string {
	switch n := node.(type) {
	case *ast.Text:
		return r.renderText(n)
	case *ast.String:
		return string(n.Value)
	case *ast.CodeSpan:
		return r.ty.Code(r.collectInlineText(n))
	case *ast.Emphasis:
		return r.renderEmphasis(n)
	case *ast.Link:
		return r.renderLink(n)
	case *ast.AutoLink:
		return r.ty.Link(string(n.URL(r.source)))
	case *ast.Image:
		return r.renderImage(n)
	case *ast.RawHTML:
		return r.collectSegments(n)
	case *east.Strikethrough:
		return r.ty.Strikethrough(r.renderInlineChildren(n))
	case *east.FootnoteLink:
		return r.ty.FootnoteRef(n.Index)
	case *east.FootnoteBacklink:
		return "" // terminal can't link back to the reference
	case *east.TaskCheckBox:
		if n.IsChecked {
			return "[x] "
		}
		return "[ ] "
	default:
		if node.HasChildren() {
			return r.renderInlineChildren(node)
		}
		return ""
	}
}

// renderText extracts text content and appends line breaks when present.
func (r *walker) renderText(n *ast.Text) string {
	text := string(n.Segment.Value(r.source))
	if n.SoftLineBreak() {
		text += "\n"
	}
	if n.HardLineBreak() {
		text += "\n"
	}
	return text
}

// renderEmphasis renders bold (level 2) or italic (level 1) text.
func (r *walker) renderEmphasis(n *ast.Emphasis) string {
	inner := r.renderInlineChildren(n)
	if n.Level == 2 {
		return r.ty.Bold(inner)
	}
	return r.ty.Italic(inner)
}

// renderLink renders a labeled or bare link.
func (r *walker) renderLink(n *ast.Link) string {
	label := r.renderInlineChildren(n)
	url := string(n.Destination)
	if label == url || label == "" {
		return r.ty.Link(url)
	}
	return r.ty.Link(label, url)
}

// renderImage renders an image as a link (terminal fallback).
func (r *walker) renderImage(n *ast.Image) string {
	alt := r.renderInlineChildren(n)
	url := string(n.Destination)
	if alt == "" {
		alt = url
	}
	return r.ty.Link(alt, url)
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// collectLines concatenates all lines of a block node from the source.
func (r *walker) collectLines(node ast.Node) string {
	var sb strings.Builder
	lines := node.Lines()
	for i := range lines.Len() {
		line := lines.At(i)
		sb.Write(line.Value(r.source))
	}
	return strings.TrimRight(sb.String(), "\n")
}

// collectInlineText extracts raw text from inline children (ignoring styles).
func (r *walker) collectInlineText(node ast.Node) string {
	var sb strings.Builder
	for child := node.FirstChild(); child != nil; child = child.NextSibling() {
		switch t := child.(type) {
		case *ast.Text:
			sb.Write(t.Segment.Value(r.source))
		case *ast.String:
			sb.Write(t.Value)
		default:
			if child.HasChildren() {
				sb.WriteString(r.collectInlineText(child))
			}
		}
	}
	return sb.String()
}

// collectSegments reads raw HTML segments.
func (r *walker) collectSegments(n *ast.RawHTML) string {
	var sb strings.Builder
	for i := range n.Segments.Len() {
		seg := n.Segments.At(i)
		sb.Write(seg.Value(r.source))
	}
	return sb.String()
}
