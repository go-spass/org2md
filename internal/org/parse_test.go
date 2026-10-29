package org

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name string
		src  string
		want *Document
	}{
		{
			name: "empty input",
			src:  "",
			want: &Document{},
		},
		{
			name: "blank lines only",
			src:  "\n\n\n",
			want: &Document{},
		},
		{
			name: "one paragraph keeps its source line breaks",
			src:  "Just some text.\nOn two lines.\n",
			want: &Document{Children: []Node{
				&Paragraph{Lines: textLines("Just some text.", "On two lines.")},
			}},
		},
		{
			name: "a blank line separates paragraphs",
			src:  "First.\n\n\nSecond.\n",
			want: &Document{Children: []Node{
				&Paragraph{Lines: textLines("First.")},
				&Paragraph{Lines: textLines("Second.")},
			}},
		},
		{
			name: "sibling headings",
			src:  "* One\n* Two\n",
			want: &Document{Children: []Node{
				&Heading{Level: 1, Title: txt("One")},
				&Heading{Level: 1, Title: txt("Two")},
			}},
		},
		{
			name: "nested headings",
			src:  "* Parent\n** Child\n*** Grandchild\n** Second child\n",
			want: &Document{Children: []Node{
				&Heading{Level: 1, Title: txt("Parent"), Children: []Node{
					&Heading{Level: 2, Title: txt("Child"), Children: []Node{
						&Heading{Level: 3, Title: txt("Grandchild")},
					}},
					&Heading{Level: 2, Title: txt("Second child")},
				}},
			}},
		},
		{
			name: "a skipped level still nests",
			src:  "* A\n*** Deep\n* B\n",
			want: &Document{Children: []Node{
				&Heading{Level: 1, Title: txt("A"), Children: []Node{
					&Heading{Level: 3, Title: txt("Deep")},
				}},
				&Heading{Level: 1, Title: txt("B")},
			}},
		},
		{
			name: "a heading owns the body that follows it",
			src:  "Preamble.\n\n* Heading\nBody text.\n\n** Sub\nMore.\n",
			want: &Document{Children: []Node{
				&Paragraph{Lines: textLines("Preamble.")},
				&Heading{Level: 1, Title: txt("Heading"), Children: []Node{
					&Paragraph{Lines: textLines("Body text.")},
					&Heading{Level: 2, Title: txt("Sub"), Children: []Node{
						&Paragraph{Lines: textLines("More.")},
					}},
				}},
			}},
		},
		{
			// Still a paragraph, not a heading: the heading pattern needs
			// whitespace after the stars. The inline phase then reads the
			// emphasis, which is why this is no longer plain text.
			name: "emphasis at the start of a line is not a heading",
			src:  "*bold* opens this line.\n",
			want: &Document{Children: []Node{
				&Paragraph{Lines: [][]Inline{{
					&Bold{Children: txt("bold")},
					&Text{Value: " opens this line."},
				}}},
			}},
		},
		{
			name: "heading with no title",
			src:  "**\n",
			want: &Document{Children: []Node{
				&Heading{Level: 2},
			}},
		},
		{
			name: "heading title is trimmed",
			src:  "*    Spaced out   \n",
			want: &Document{Children: []Node{
				&Heading{Level: 1, Title: txt("Spaced out")},
			}},
		},
		{
			name: "input without a trailing newline",
			src:  "* Heading\nBody",
			want: &Document{Children: []Node{
				&Heading{Level: 1, Title: txt("Heading"), Children: []Node{
					&Paragraph{Lines: textLines("Body")},
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
