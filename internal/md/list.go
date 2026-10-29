package md

import (
	"strconv"
	"strings"

	"github.com/go-spass/org2md/internal/org"
)

// checkboxes maps org's checkbox states onto Markdown task syntax. A partial
// box has no Markdown equivalent; [/] is the convention Obsidian understands.
var checkboxes = map[org.Checkbox]string{
	org.Unchecked:        "[ ]",
	org.Checked:          "[x]",
	org.PartiallyChecked: "[/]",
}

func list(l *org.List, indent string) string {
	return strings.Join(listLines(l, indent), "\n")
}

// listLines renders a list and everything nested under it. Continuation lines
// and nested lists are indented to the parent item's text column, which is
// what Markdown needs to keep them inside the item.
func listLines(l *org.List, indent string) []string {
	var out []string

	for i, item := range l.Items {
		marker := "-"
		if l.Ordered {
			marker = strconv.Itoa(i+1) + "."
		}
		inner := indent + strings.Repeat(" ", len(marker)+1)

		text := itemLines(item)
		out = append(out, strings.TrimRight(indent+marker+" "+text[0], " \t"))
		for _, line := range text[1:] {
			out = append(out, strings.TrimRight(inner+line, " \t"))
		}

		for _, child := range item.Children {
			if nested, ok := child.(*org.List); ok {
				out = append(out, listLines(nested, inner)...)
			}
		}
	}

	return out
}

// itemLines renders an item's own text, checkbox and definition term. The
// result always holds at least one line, so the marker has something to sit
// beside even for an empty item.
func itemLines(item *org.ListItem) []string {
	lines := make([]string, 0, len(item.Lines)+1)
	for _, line := range item.Lines {
		lines = append(lines, inlines(line))
	}
	if len(lines) == 0 {
		lines = append(lines, "")
	}

	var prefix string
	if box, ok := checkboxes[item.Checkbox]; ok {
		prefix = box + " "
	}
	if len(item.Term) > 0 {
		prefix += "**" + inlines(item.Term) + "**: "
	}
	lines[0] = prefix + lines[0]

	return lines
}
