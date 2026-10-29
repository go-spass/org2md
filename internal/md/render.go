// Package md renders a parsed org document as Obsidian-flavored Markdown.
package md

import (
	"strings"

	"github.com/go-spass/org2md/internal/org"
)

// maxHeadingLevel is the deepest heading Markdown can express. Org outlines can
// nest further, and seven or more hashes is a paragraph rather than a heading,
// so deeper levels are clamped.
const maxHeadingLevel = 6

// Render converts a parsed org document to Markdown. The result ends with a
// single newline, or is empty when the document holds nothing.
func Render(doc *org.Document) string {
	var b strings.Builder
	renderNodes(&b, doc.Children)

	out := strings.TrimRight(b.String(), "\n")
	if out == "" {
		return ""
	}
	return out + "\n"
}

func renderNodes(b *strings.Builder, nodes []org.Node) {
	for _, n := range nodes {
		switch n := n.(type) {
		case *org.Heading:
			writeBlock(b, heading(n))
			renderNodes(b, n.Children)
		case *org.Paragraph:
			writeBlock(b, paragraph(n))
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
	title := inlines(h.Title)
	if title == "" {
		return hashes
	}
	return hashes + " " + title
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
