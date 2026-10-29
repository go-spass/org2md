package md

import "github.com/go-spass/org2md/internal/org"

func txt(s string) []org.Inline {
	return []org.Inline{&org.Text{Value: s}}
}

func textLines(ss ...string) [][]org.Inline {
	out := make([][]org.Inline, len(ss))
	for i, s := range ss {
		out[i] = txt(s)
	}
	return out
}
