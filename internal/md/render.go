// Package md renders a parsed org document as Obsidian-flavored Markdown.
package md

import (
	"fmt"
	"strings"

	"github.com/go-spass/org2md/internal/org"
)

// maxHeadingLevel is the deepest heading Markdown can express. Org outlines can
// nest further, and seven or more hashes is a paragraph rather than a heading,
// so deeper levels are clamped.
const maxHeadingLevel = 6

// renderer carries the warnings a conversion accumulates. Nothing here fails,
// but org constructs with no Markdown equivalent are reported rather than
// dropped in silence.
type renderer struct {
	warnings []string
}

func (r *renderer) warnf(format string, args ...any) {
	r.warnings = append(r.warnings, fmt.Sprintf(format, args...))
}

// Render converts a parsed org document to Markdown, returning the result and
// any warnings in document order. The result ends with a single newline, or is
// empty when the document holds nothing.
func Render(doc *org.Document) (string, []string) {
	r := &renderer{}

	front := r.frontmatter(doc)
	body := strings.TrimRight(r.blocks(doc.Children), "\n")

	switch {
	case front == "" && body == "":
		return "", r.warnings
	case front == "":
		return body + "\n", r.warnings
	case body == "":
		return front, r.warnings
	default:
		return front + "\n" + body + "\n", r.warnings
	}
}

// blocks renders a sequence of block nodes, separated by blank lines.
func (r *renderer) blocks(nodes []org.Node) string {
	var b strings.Builder
	r.renderNodes(&b, nodes)
	return b.String()
}

func (r *renderer) renderNodes(b *strings.Builder, nodes []org.Node) {
	for _, n := range nodes {
		switch n := n.(type) {
		case *org.Heading:
			writeBlock(b, heading(n))
			if dates := planning(n); dates != "" {
				writeBlock(b, dates)
			}
			r.renderNodes(b, n.Children)
		case *org.Paragraph:
			writeBlock(b, paragraph(n))
		case *org.List:
			writeBlock(b, list(n, ""))
		case *org.Block:
			writeBlock(b, block(n))
		case *org.Quote:
			writeBlock(b, r.quote(n))
		case *org.Drawer:
			// A logbook is clock data with nothing to say in Markdown; any
			// other drawer is content the reader loses.
			if n.Name != "LOGBOOK" {
				r.warnf("dropped :%s: drawer", n.Name)
			}
		case *org.Rule:
			writeBlock(b, "---")
		case *org.Table:
			if s := table(n); s != "" {
				writeBlock(b, s)
			}
		case *org.FootnoteDef:
			writeBlock(b, footnoteDef(n))
		}
	}
}

// writeBlock appends one block, separated from the previous one by a blank line.
func writeBlock(b *strings.Builder, s string) {
	if b.Len() > 0 {
		b.WriteString("\n")
	}
	b.WriteString(s)
	b.WriteString("\n")
}

func heading(h *org.Heading) string {
	hashes := strings.Repeat("#", min(h.Level, maxHeadingLevel))

	var parts []string
	if title := inlines(h.Title); title != "" {
		parts = append(parts, title)
	}
	parts = append(parts, headingTags(h)...)

	if len(parts) == 0 {
		return hashes
	}
	return hashes + " " + strings.Join(parts, " ")
}

// paragraph trims trailing whitespace from each line. Org treats it as
// insignificant, but two trailing spaces are a hard line break in Markdown.
func paragraph(p *org.Paragraph) string {
	lines := make([]string, len(p.Lines))
	for i, line := range p.Lines {
		lines[i] = strings.TrimRight(inlines(line), " \t")
	}
	return strings.Join(lines, "\n")
}
