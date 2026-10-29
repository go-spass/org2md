package org

// txt is the common case in these tables: a span of plain text with no markup.
func txt(s string) []Inline {
	return []Inline{&Text{Value: s}}
}

// textLines builds the Lines of a paragraph whose lines carry no markup.
func textLines(ss ...string) [][]Inline {
	out := make([][]Inline, len(ss))
	for i, s := range ss {
		out[i] = txt(s)
	}
	return out
}
