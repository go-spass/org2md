package org

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func cells(ss ...string) [][]Inline {
	out := make([][]Inline, len(ss))
	for i, s := range ss {
		if s == "" {
			continue
		}
		out[i] = txt(s)
	}
	return out
}

func TestParseTables(t *testing.T) {
	tests := []struct {
		name string
		src  string
		want *Document
	}{
		{
			name: "header, separator and body",
			src:  "| Name | Qty |\n|------+-----|\n| Apple | 3 |\n",
			want: &Document{Children: []Node{
				&Table{Rows: []*TableRow{
					{Cells: cells("Name", "Qty")},
					{Separator: true},
					{Cells: cells("Apple", "3")},
				}},
			}},
		},
		{
			name: "a separator of pipes and dashes",
			src:  "| a |\n|---|\n| b |\n",
			want: &Document{Children: []Node{
				&Table{Rows: []*TableRow{
					{Cells: cells("a")},
					{Separator: true},
					{Cells: cells("b")},
				}},
			}},
		},
		{
			name: "the short separator form",
			src:  "| a |\n|-\n| b |\n",
			want: &Document{Children: []Node{
				&Table{Rows: []*TableRow{
					{Cells: cells("a")},
					{Separator: true},
					{Cells: cells("b")},
				}},
			}},
		},
		{
			name: "no trailing pipe",
			src:  "| a | b\n",
			want: &Document{Children: []Node{
				&Table{Rows: []*TableRow{{Cells: cells("a", "b")}}},
			}},
		},
		{
			name: "an empty cell",
			src:  "| a | | c |\n",
			want: &Document{Children: []Node{
				&Table{Rows: []*TableRow{{Cells: cells("a", "", "c")}}},
			}},
		},
		{
			name: "markup inside a cell",
			src:  "| *bold* | [[file:n.org][link]] |\n",
			want: &Document{Children: []Node{
				&Table{Rows: []*TableRow{{Cells: [][]Inline{
					{&Bold{Children: txt("bold")}},
					{&Link{Target: "file:n.org", Desc: txt("link")}},
				}}}},
			}},
		},
		{
			name: "alignment cookies set the column alignment and are not a row",
			src:  "| Name | Qty | Mid |\n| <l> | <r> | <c> |\n| Apple | 3 | x |\n",
			want: &Document{Children: []Node{
				&Table{
					Align: []Alignment{AlignLeft, AlignRight, AlignCenter},
					Rows: []*TableRow{
						{Cells: cells("Name", "Qty", "Mid")},
						{Cells: cells("Apple", "3", "x")},
					},
				},
			}},
		},
		{
			name: "width cookies carry no alignment",
			src:  "| a |\n| <5> |\n",
			want: &Document{Children: []Node{
				&Table{
					Align: []Alignment{AlignDefault},
					Rows:  []*TableRow{{Cells: cells("a")}},
				},
			}},
		},
		{
			name: "an entirely empty row is a row, not a cookie row",
			src:  "| a | b |\n| | |\n",
			want: &Document{Children: []Node{
				&Table{Rows: []*TableRow{
					{Cells: cells("a", "b")},
					{Cells: cells("", "")},
				}},
			}},
		},
		{
			name: "a blank line ends the table",
			src:  "| a |\n\n| b |\n",
			want: &Document{Children: []Node{
				&Table{Rows: []*TableRow{{Cells: cells("a")}}},
				&Table{Rows: []*TableRow{{Cells: cells("b")}}},
			}},
		},
		{
			name: "prose ends the table",
			src:  "| a |\nAfter.\n",
			want: &Document{Children: []Node{
				&Table{Rows: []*TableRow{{Cells: cells("a")}}},
				&Paragraph{Lines: textLines("After.")},
			}},
		},
		{
			name: "a table inside a source block is literal",
			src:  "#+BEGIN_EXAMPLE\n| a |\n#+END_EXAMPLE\n",
			want: &Document{Children: []Node{
				&Block{Lines: []string{"| a |"}},
			}},
		},
		{
			name: "a table under a heading",
			src:  "* Heading\n| a |\n",
			want: &Document{Children: []Node{
				&Heading{Level: 1, Title: txt("Heading"), Children: []Node{
					&Table{Rows: []*TableRow{{Cells: cells("a")}}},
				}},
			}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(strings.NewReader(tt.src))
			if err != nil {
				t.Fatalf("Parse() error = %v, want nil", err)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("Parse() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestParseFootnotes(t *testing.T) {
	tests := []struct {
		name string
		src  string
		want *Document
	}{
		{
			name: "reference and definition",
			src:  "Prose with a note[fn:1].\n\n[fn:1] The note itself.\n",
			want: &Document{Children: []Node{
				&Paragraph{Lines: [][]Inline{{
					&Text{Value: "Prose with a note"},
					&FootnoteRef{Label: "1"},
					&Text{Value: "."},
				}}},
				&FootnoteDef{Label: "1", Lines: textLines("The note itself.")},
			}},
		},
		{
			name: "named label",
			src:  "[fn:why] Because.\n",
			want: &Document{Children: []Node{
				&FootnoteDef{Label: "why", Lines: textLines("Because.")},
			}},
		},
		{
			name: "definition spanning lines",
			src:  "[fn:1] First line.\nSecond line.\n",
			want: &Document{Children: []Node{
				&FootnoteDef{Label: "1", Lines: textLines("First line.", "Second line.")},
			}},
		},
		{
			name: "definition with no text",
			src:  "[fn:1]\n",
			want: &Document{Children: []Node{
				&FootnoteDef{Label: "1"},
			}},
		},
		{
			name: "markup inside a definition",
			src:  "[fn:1] A *bold* note.\n",
			want: &Document{Children: []Node{
				&FootnoteDef{Label: "1", Lines: [][]Inline{{
					&Text{Value: "A "},
					&Bold{Children: txt("bold")},
					&Text{Value: " note."},
				}}},
			}},
		},
		{
			name: "a heading ends the definition",
			src:  "[fn:1] The note.\n* Heading\n",
			want: &Document{Children: []Node{
				&FootnoteDef{Label: "1", Lines: textLines("The note.")},
				&Heading{Level: 1, Title: txt("Heading")},
			}},
		},
		{
			name: "a bracketed word is not a footnote",
			src:  "See [not a footnote] here.\n",
			want: &Document{Children: []Node{
				&Paragraph{Lines: textLines("See [not a footnote] here.")},
			}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(strings.NewReader(tt.src))
			if err != nil {
				t.Fatalf("Parse() error = %v, want nil", err)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("Parse() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
