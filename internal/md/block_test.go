package md

import (
	"testing"

	"github.com/go-spass/org2md/internal/org"
	"github.com/google/go-cmp/cmp"
)

func TestRenderBlocks(t *testing.T) {
	tests := []struct {
		name string
		doc  *org.Document
		want string
	}{
		{
			name: "source block carries its language as the info string",
			doc:  doc(&org.Block{Lang: "go", Lines: []string{`fmt.Println("hi")`}}),
			want: "```go\nfmt.Println(\"hi\")\n```\n",
		},
		{
			name: "example block has no info string",
			doc:  doc(&org.Block{Lines: []string{"literal"}}),
			want: "```\nliteral\n```\n",
		},
		{
			name: "block content is written out exactly",
			doc: doc(&org.Block{Lang: "sh", Lines: []string{
				"* not a heading",
				"echo *not bold*   ",
				"\tindented",
			}}),
			want: "```sh\n* not a heading\necho *not bold*   \n\tindented\n```\n",
		},
		{
			name: "empty block",
			doc:  doc(&org.Block{}),
			want: "```\n```\n",
		},
		{
			name: "a fence longer than any backtick run in the content",
			doc:  doc(&org.Block{Lines: []string{"a ``` fence", "and ```` more"}}),
			want: "`````\na ``` fence\nand ```` more\n`````\n",
		},
		{
			name: "rule",
			doc:  doc(&org.Rule{}),
			want: "---\n",
		},
		{
			name: "quote keeps its markup and marks blank lines",
			doc: doc(&org.Quote{Children: []org.Node{
				&org.Paragraph{Lines: [][]org.Inline{{
					&org.Text{Value: "A "},
					&org.Bold{Children: txt("quoted")},
					&org.Text{Value: " line."},
				}}},
				&org.List{Items: []*org.ListItem{{Lines: textLines("and a list")}}},
			}}),
			want: "> A **quoted** line.\n>\n> - and a list\n",
		},
		{
			name: "empty quote",
			doc:  doc(&org.Quote{}),
			want: ">\n",
		},
		{
			name: "blocks between paragraphs",
			doc: doc(
				&org.Paragraph{Lines: textLines("Before.")},
				&org.Block{Lang: "go", Lines: []string{"x := 1"}},
				&org.Rule{},
				&org.Paragraph{Lines: textLines("After.")},
			),
			want: "Before.\n\n```go\nx := 1\n```\n\n---\n\nAfter.\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Render(tt.doc)
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("Render() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
