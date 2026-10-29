package md

import (
	"testing"

	"github.com/go-spass/org2md/internal/org"
	"github.com/google/go-cmp/cmp"
)

func row(ss ...string) *org.TableRow {
	r := &org.TableRow{Cells: make([][]org.Inline, len(ss))}
	for i, s := range ss {
		if s == "" {
			continue
		}
		r.Cells[i] = txt(s)
	}
	return r
}

func TestRenderTables(t *testing.T) {
	tests := []struct {
		name string
		doc  *org.Document
		want string
	}{
		{
			name: "header, alignment row and body",
			doc: doc(&org.Table{Rows: []*org.TableRow{
				row("Name", "Qty"),
				{Separator: true},
				row("Apple", "3"),
			}}),
			want: "| Name  | Qty |\n| ----- | --- |\n| Apple | 3   |\n",
		},
		{
			name: "without a separator the first row is the header",
			doc: doc(&org.Table{Rows: []*org.TableRow{
				row("Name", "Qty"),
				row("Apple", "3"),
			}}),
			want: "| Name  | Qty |\n| ----- | --- |\n| Apple | 3   |\n",
		},
		{
			name: "declared alignment",
			doc: doc(&org.Table{
				Align: []org.Alignment{org.AlignLeft, org.AlignRight, org.AlignCenter},
				Rows: []*org.TableRow{
					row("Name", "Qty", "Mid"),
					row("Apple", "3", "x"),
				},
			}),
			want: "| Name  | Qty | Mid |\n| :---- | --: | :-: |\n| Apple | 3   | x   |\n",
		},
		{
			name: "extra separators do not lose rows",
			doc: doc(&org.Table{Rows: []*org.TableRow{
				row("a"),
				{Separator: true},
				row("b"),
				{Separator: true},
				row("c"),
			}}),
			want: "| a   |\n| --- |\n| b   |\n| c   |\n",
		},
		{
			name: "ragged rows are padded to the widest",
			doc: doc(&org.Table{Rows: []*org.TableRow{
				row("a", "b", "c"),
				{Separator: true},
				row("d"),
			}}),
			want: "| a   | b   | c   |\n| --- | --- | --- |\n| d   |     |     |\n",
		},
		{
			name: "a pipe in a cell is escaped",
			doc: doc(&org.Table{Rows: []*org.TableRow{
				row("a|b"),
				{Separator: true},
				row("c"),
			}}),
			want: "| a\\|b |\n| ---- |\n| c    |\n",
		},
		{
			name: "markup inside cells",
			doc: doc(&org.Table{Rows: []*org.TableRow{
				{Cells: [][]org.Inline{
					{&org.Bold{Children: txt("bold")}},
					{&org.Link{Target: "file:n.org", Desc: txt("link")}},
				}},
				{Separator: true},
				row("x", "y"),
			}}),
			want: "| **bold** | [[n\\|link]] |\n| -------- | ----------- |\n| x        | y           |\n",
		},
		{
			name: "a table of separators renders nothing",
			doc:  doc(&org.Table{Rows: []*org.TableRow{{Separator: true}}}),
			want: "",
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

func TestRenderFootnotes(t *testing.T) {
	tests := []struct {
		name string
		doc  *org.Document
		want string
	}{
		{
			name: "reference and definition",
			doc: doc(
				&org.Paragraph{Lines: [][]org.Inline{{
					&org.Text{Value: "A note"},
					&org.FootnoteRef{Label: "1"},
					&org.Text{Value: "."},
				}}},
				&org.FootnoteDef{Label: "1", Lines: textLines("The note.")},
			),
			want: "A note[^1].\n\n[^1]: The note.\n",
		},
		{
			name: "a named label",
			doc:  doc(&org.FootnoteDef{Label: "why", Lines: textLines("Because.")}),
			want: "[^why]: Because.\n",
		},
		{
			name: "continuation lines are indented",
			doc:  doc(&org.FootnoteDef{Label: "1", Lines: textLines("First.", "Second.")}),
			want: "[^1]: First.\n    Second.\n",
		},
		{
			name: "a definition with no text",
			doc:  doc(&org.FootnoteDef{Label: "1"}),
			want: "[^1]:\n",
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
