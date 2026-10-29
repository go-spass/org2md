package org

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestParseInlines(t *testing.T) {
	tests := []struct {
		name string
		src  string
		want []Inline
	}{
		{name: "empty string", src: "", want: nil},
		{name: "plain text", src: "hello there", want: txt("hello there")},
		{
			name: "bold",
			src:  "*bold*",
			want: []Inline{&Bold{Children: txt("bold")}},
		},
		{
			name: "emphasis inside a sentence",
			src:  "a *bold* word",
			want: []Inline{&Text{Value: "a "}, &Bold{Children: txt("bold")}, &Text{Value: " word"}},
		},
		{
			name: "italic, underline and strike",
			src:  "/it/ _un_ +gone+",
			want: []Inline{
				&Italic{Children: txt("it")},
				&Text{Value: " "},
				&Underline{Children: txt("un")},
				&Text{Value: " "},
				&Strike{Children: txt("gone")},
			},
		},
		{
			name: "nested emphasis",
			src:  "*bold with /italic/*",
			want: []Inline{&Bold{Children: []Inline{
				&Text{Value: "bold with "},
				&Italic{Children: txt("italic")},
			}}},
		},
		{
			name: "verbatim is literal",
			src:  "=*not bold*=",
			want: []Inline{&Code{Value: "*not bold*"}},
		},
		{
			name: "tilde code is literal",
			src:  "~x := *y*~",
			want: []Inline{&Code{Value: "x := *y*"}},
		},

		// Org requires a word boundary before an emphasis marker. Without that
		// rule these four would all be mangled.
		{name: "snake_case is not underline", src: "some_variable_name", want: txt("some_variable_name")},
		{name: "and/or is not italic", src: "and/or", want: txt("and/or")},
		{name: "arithmetic is not bold", src: "2*3*4", want: txt("2*3*4")},
		{
			name: "a url keeps its slashes",
			src:  "see https://example.com/a/b for more",
			want: txt("see https://example.com/a/b for more"),
		},

		{name: "unclosed marker", src: "*unclosed", want: txt("*unclosed")},
		{name: "body may not start with whitespace", src: "* not bold*", want: txt("* not bold*")},
		{name: "body may not be empty", src: "**", want: txt("**")},
		{
			name: "link with a description",
			src:  "[[https://example.com][the label]]",
			want: []Inline{&Link{Target: "https://example.com", Desc: txt("the label")}},
		},
		{
			name: "bare link",
			src:  "[[https://example.com]]",
			want: []Inline{&Link{Target: "https://example.com"}},
		},
		{
			name: "a link description may hold markup",
			src:  "[[https://example.com][*bold* label]]",
			want: []Inline{&Link{Target: "https://example.com", Desc: []Inline{
				&Bold{Children: txt("bold")},
				&Text{Value: " label"},
			}}},
		},
		{
			name: "unclosed link",
			src:  "[[https://example.com",
			want: txt("[[https://example.com"),
		},
		{
			name: "timestamp drops the day name",
			src:  "<2026-10-28 Wed>",
			want: []Inline{&Timestamp{Date: "2026-10-28"}},
		},
		{
			name: "timestamp keeps the time and repeater",
			src:  "<2026-10-28 Wed 14:30 +1w>",
			want: []Inline{&Timestamp{Date: "2026-10-28", Rest: "14:30 +1w"}},
		},
		{
			name: "timestamp without a day name",
			src:  "<2026-10-28 14:30>",
			want: []Inline{&Timestamp{Date: "2026-10-28", Rest: "14:30"}},
		},
		{
			name: "angle brackets that are not a timestamp",
			src:  "a < b and <not a date>",
			want: txt("a < b and <not a date>"),
		},
		{
			name: "markup inside a heading title",
			src:  "A *bold* heading with =code=",
			want: []Inline{
				&Text{Value: "A "},
				&Bold{Children: txt("bold")},
				&Text{Value: " heading with "},
				&Code{Value: "code"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseInlines(tt.src)
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("parseInlines(%q) mismatch (-want +got):\n%s", tt.src, diff)
			}
		})
	}
}
