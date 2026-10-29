package org

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestParseLists(t *testing.T) {
	tests := []struct {
		name string
		src  string
		want *Document
	}{
		{
			name: "dash bullets",
			src:  "- one\n- two\n",
			want: &Document{Children: []Node{
				&List{Items: []*ListItem{
					{Lines: textLines("one")},
					{Lines: textLines("two")},
				}},
			}},
		},
		{
			name: "plus bullets",
			src:  "+ one\n+ two\n",
			want: &Document{Children: []Node{
				&List{Items: []*ListItem{
					{Lines: textLines("one")},
					{Lines: textLines("two")},
				}},
			}},
		},
		{
			name: "an indented asterisk is a bullet, not a heading",
			src:  "  * one\n  * two\n",
			want: &Document{Children: []Node{
				&List{Items: []*ListItem{
					{Lines: textLines("one")},
					{Lines: textLines("two")},
				}},
			}},
		},
		{
			name: "ordered with dots and with parens",
			src:  "1. one\n2. two\n",
			want: &Document{Children: []Node{
				&List{Ordered: true, Items: []*ListItem{
					{Lines: textLines("one")},
					{Lines: textLines("two")},
				}},
			}},
		},
		{
			name: "ordered with parens",
			src:  "1) one\n2) two\n",
			want: &Document{Children: []Node{
				&List{Ordered: true, Items: []*ListItem{
					{Lines: textLines("one")},
					{Lines: textLines("two")},
				}},
			}},
		},
		{
			name: "changing the marker starts a sibling list",
			src:  "- bullet\n1. ordered\n",
			want: &Document{Children: []Node{
				&List{Items: []*ListItem{{Lines: textLines("bullet")}}},
				&List{Ordered: true, Items: []*ListItem{{Lines: textLines("ordered")}}},
			}},
		},
		{
			name: "nesting by indentation",
			src:  "- one\n  - nested\n    - deeper\n- two\n",
			want: &Document{Children: []Node{
				&List{Items: []*ListItem{
					{
						Lines: textLines("one"),
						Children: []Node{&List{Items: []*ListItem{
							{
								Lines: textLines("nested"),
								Children: []Node{&List{Items: []*ListItem{
									{Lines: textLines("deeper")},
								}}},
							},
						}}},
					},
					{Lines: textLines("two")},
				}},
			}},
		},
		{
			name: "mixed nesting",
			src:  "1. ordered\n   - bullet\n",
			want: &Document{Children: []Node{
				&List{Ordered: true, Items: []*ListItem{
					{
						Lines:    textLines("ordered"),
						Children: []Node{&List{Items: []*ListItem{{Lines: textLines("bullet")}}}},
					},
				}},
			}},
		},
		{
			name: "checkboxes",
			src:  "- [ ] todo\n- [X] done\n- [x] also done\n- [-] partial\n- [ ]\n",
			want: &Document{Children: []Node{
				&List{Items: []*ListItem{
					{Checkbox: Unchecked, Lines: textLines("todo")},
					{Checkbox: Checked, Lines: textLines("done")},
					{Checkbox: Checked, Lines: textLines("also done")},
					{Checkbox: PartiallyChecked, Lines: textLines("partial")},
					{Checkbox: Unchecked},
				}},
			}},
		},
		{
			name: "definition list",
			src:  "- term :: the definition\n- bare ::\n",
			want: &Document{Children: []Node{
				&List{Items: []*ListItem{
					{Term: txt("term"), Lines: textLines("the definition")},
					{Term: txt("bare")},
				}},
			}},
		},
		{
			name: "markup inside an item",
			src:  "- an *emphatic* item\n",
			want: &Document{Children: []Node{
				&List{Items: []*ListItem{{Lines: [][]Inline{{
					&Text{Value: "an "},
					&Bold{Children: txt("emphatic")},
					&Text{Value: " item"},
				}}}}},
			}},
		},
		{
			name: "continuation line belongs to the item",
			src:  "- first line\n  second line\n- next item\n",
			want: &Document{Children: []Node{
				&List{Items: []*ListItem{
					{Lines: textLines("first line", "second line")},
					{Lines: textLines("next item")},
				}},
			}},
		},
		{
			name: "a continuation closes the deeper levels it is outdented from",
			src:  "- one\n  - nested\n  continues one\n- two\n",
			want: &Document{Children: []Node{
				&List{Items: []*ListItem{
					{
						Lines:    textLines("one", "continues one"),
						Children: []Node{&List{Items: []*ListItem{{Lines: textLines("nested")}}}},
					},
					{Lines: textLines("two")},
				}},
			}},
		},
		{
			name: "a blank line does not end the list",
			src:  "- one\n\n- two\n",
			want: &Document{Children: []Node{
				&List{Items: []*ListItem{
					{Lines: textLines("one")},
					{Lines: textLines("two")},
				}},
			}},
		},
		{
			name: "a paragraph ends the list",
			src:  "- one\n\nBack to prose.\n",
			want: &Document{Children: []Node{
				&List{Items: []*ListItem{{Lines: textLines("one")}}},
				&Paragraph{Lines: textLines("Back to prose.")},
			}},
		},
		{
			name: "a heading ends the list",
			src:  "- one\n* Heading\n",
			want: &Document{Children: []Node{
				&List{Items: []*ListItem{{Lines: textLines("one")}}},
				&Heading{Level: 1, Title: txt("Heading")},
			}},
		},
		{
			name: "a list under a heading",
			src:  "* Heading\n- one\n",
			want: &Document{Children: []Node{
				&Heading{Level: 1, Title: txt("Heading"), Children: []Node{
					&List{Items: []*ListItem{{Lines: textLines("one")}}},
				}},
			}},
		},
		{
			name: "a dash that is not a list item",
			src:  "-not a bullet\n",
			want: &Document{Children: []Node{
				&Paragraph{Lines: textLines("-not a bullet")},
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
