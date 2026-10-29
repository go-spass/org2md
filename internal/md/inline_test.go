package md

import (
	"testing"

	"github.com/go-spass/org2md/internal/org"
)

func TestInlines(t *testing.T) {
	tests := []struct {
		name string
		in   []org.Inline
		want string
	}{
		{name: "nothing", in: nil, want: ""},
		{name: "plain text", in: txt("hello"), want: "hello"},
		{name: "bold", in: []org.Inline{&org.Bold{Children: txt("x")}}, want: "**x**"},
		{name: "italic", in: []org.Inline{&org.Italic{Children: txt("x")}}, want: "*x*"},
		{name: "strike", in: []org.Inline{&org.Strike{Children: txt("x")}}, want: "~~x~~"},
		{
			name: "underline becomes inline html",
			in:   []org.Inline{&org.Underline{Children: txt("x")}},
			want: "<u>x</u>",
		},
		{
			name: "code is left exactly as written",
			in:   []org.Inline{&org.Code{Value: "*not bold* _nor this_"}},
			want: "`*not bold* _nor this_`",
		},
		{
			name: "nested emphasis",
			in: []org.Inline{&org.Bold{Children: []org.Inline{
				&org.Text{Value: "bold "},
				&org.Italic{Children: txt("italic")},
			}}},
			want: "**bold *italic***",
		},
		{
			name: "link with a description",
			in:   []org.Inline{&org.Link{Target: "https://example.com", Desc: txt("label")}},
			want: "[label](https://example.com)",
		},
		{
			name: "bare link",
			in:   []org.Inline{&org.Link{Target: "https://example.com"}},
			want: "<https://example.com>",
		},
		{
			name: "org file link becomes a wikilink",
			in:   []org.Inline{&org.Link{Target: "file:note.org", Desc: txt("the note")}},
			want: "[[note|the note]]",
		},
		{
			name: "bare org file link becomes a wikilink",
			in:   []org.Inline{&org.Link{Target: "file:subdir/note.org"}},
			want: "[[note]]",
		},
		{
			name: "heading link becomes an anchor wikilink",
			in:   []org.Inline{&org.Link{Target: "*Some Heading"}},
			want: "[[#Some Heading]]",
		},
		{
			name: "heading link with a description",
			in:   []org.Inline{&org.Link{Target: "*Some Heading", Desc: txt("see above")}},
			want: "[[#Some Heading|see above]]",
		},
		{
			name: "timestamp",
			in:   []org.Inline{&org.Timestamp{Date: "2026-10-28"}},
			want: "2026-10-28",
		},
		{
			name: "timestamp with a time and repeater",
			in:   []org.Inline{&org.Timestamp{Date: "2026-10-28", Rest: "14:30 +1w"}},
			want: "2026-10-28 14:30 +1w",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := inlines(tt.in); got != tt.want {
				t.Errorf("inlines() = %q, want %q", got, tt.want)
			}
		})
	}
}
