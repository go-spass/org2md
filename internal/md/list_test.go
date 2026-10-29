package md

import (
	"testing"

	"github.com/go-spass/org2md/internal/org"
	"github.com/google/go-cmp/cmp"
)

func TestRenderLists(t *testing.T) {
	tests := []struct {
		name string
		doc  *org.Document
		want string
	}{
		{
			name: "bullets",
			doc: doc(&org.List{Items: []*org.ListItem{
				{Lines: textLines("one")},
				{Lines: textLines("two")},
			}}),
			want: "- one\n- two\n",
		},
		{
			name: "an ordered list is renumbered",
			doc: doc(&org.List{Ordered: true, Items: []*org.ListItem{
				{Lines: textLines("one")},
				{Lines: textLines("two")},
				{Lines: textLines("three")},
			}}),
			want: "1. one\n2. two\n3. three\n",
		},
		{
			name: "checkbox states",
			doc: doc(&org.List{Items: []*org.ListItem{
				{Checkbox: org.Unchecked, Lines: textLines("todo")},
				{Checkbox: org.Checked, Lines: textLines("done")},
				{Checkbox: org.PartiallyChecked, Lines: textLines("partial")},
				{Checkbox: org.Unchecked},
			}}),
			want: "- [ ] todo\n- [x] done\n- [/] partial\n- [ ]\n",
		},
		{
			name: "definition items",
			doc: doc(&org.List{Items: []*org.ListItem{
				{Term: txt("term"), Lines: textLines("the definition")},
				{Term: txt("bare")},
			}}),
			want: "- **term**: the definition\n- **bare**:\n",
		},
		{
			name: "continuation lines align with the item text",
			doc: doc(&org.List{Items: []*org.ListItem{
				{Lines: textLines("first line", "second line")},
			}}),
			want: "- first line\n  second line\n",
		},
		{
			name: "ordered continuation lines align with the item text",
			doc: doc(&org.List{Ordered: true, Items: []*org.ListItem{
				{Lines: textLines("first line", "second line")},
			}}),
			want: "1. first line\n   second line\n",
		},
		{
			name: "nested bullets indent by two",
			doc: doc(&org.List{Items: []*org.ListItem{
				{
					Lines: textLines("one"),
					Children: []org.Node{&org.List{Items: []*org.ListItem{
						{Lines: textLines("nested")},
					}}},
				},
				{Lines: textLines("two")},
			}}),
			want: "- one\n  - nested\n- two\n",
		},
		{
			name: "a bullet nested under an ordered item indents by three",
			doc: doc(&org.List{Ordered: true, Items: []*org.ListItem{
				{
					Lines: textLines("ordered"),
					Children: []org.Node{&org.List{Items: []*org.ListItem{
						{Lines: textLines("bullet")},
					}}},
				},
			}}),
			want: "1. ordered\n   - bullet\n",
		},
		{
			name: "markup inside an item",
			doc: doc(&org.List{Items: []*org.ListItem{
				{Lines: [][]org.Inline{{
					&org.Text{Value: "see "},
					&org.Link{Target: "file:other.org", Desc: txt("the note")},
				}}},
			}}),
			want: "- see [[other|the note]]\n",
		},
		{
			name: "a list between paragraphs",
			doc: &org.Document{Children: []org.Node{
				&org.Paragraph{Lines: textLines("Before.")},
				&org.List{Items: []*org.ListItem{{Lines: textLines("item")}}},
				&org.Paragraph{Lines: textLines("After.")},
			}},
			want: "Before.\n\n- item\n\nAfter.\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _ := Render(tt.doc)
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("Render() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
