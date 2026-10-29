package md

import (
	"path"
	"strings"

	"github.com/go-spass/org2md/internal/org"
)

// inlines renders a span sequence. Code is written out verbatim, so nothing in
// here can reinterpret a marker the inline phase already decided was literal.
func inlines(ns []org.Inline) string {
	var b strings.Builder
	for _, n := range ns {
		switch n := n.(type) {
		case *org.Text:
			b.WriteString(n.Value)
		case *org.Bold:
			b.WriteString("**" + inlines(n.Children) + "**")
		case *org.Italic:
			b.WriteString("*" + inlines(n.Children) + "*")
		case *org.Underline:
			b.WriteString("<u>" + inlines(n.Children) + "</u>")
		case *org.Strike:
			b.WriteString("~~" + inlines(n.Children) + "~~")
		case *org.Code:
			b.WriteString("`" + n.Value + "`")
		case *org.Link:
			b.WriteString(link(n))
		case *org.Timestamp:
			b.WriteString(timestamp(n))
		case *org.FootnoteRef:
			b.WriteString("[^" + n.Label + "]")
		}
	}
	return b.String()
}

// link turns an org link into the closest Obsidian equivalent: a wikilink for
// a link to another org file or to a heading, and a Markdown link otherwise.
func link(l *org.Link) string {
	desc := inlines(l.Desc)

	if target, ok := wikiTarget(l.Target); ok {
		if desc == "" {
			return "[[" + target + "]]"
		}
		return "[[" + target + "|" + desc + "]]"
	}

	if desc == "" {
		return "<" + l.Target + ">"
	}
	return "[" + desc + "](" + l.Target + ")"
}

// wikiTarget reports the wikilink target for an org link, if it has one. A
// file: link to another org note becomes a link to the note of the same name,
// since the vault holds the converted file rather than the original.
func wikiTarget(target string) (string, bool) {
	if heading, ok := strings.CutPrefix(target, "*"); ok {
		return "#" + heading, true
	}
	if file, ok := strings.CutPrefix(target, "file:"); ok {
		if strings.HasSuffix(file, ".org") {
			return strings.TrimSuffix(path.Base(file), ".org"), true
		}
	}
	return "", false
}

func timestamp(t *org.Timestamp) string {
	if t.Rest == "" {
		return t.Date
	}
	return t.Date + " " + t.Rest
}
