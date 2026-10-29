package org

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestParseMeta(t *testing.T) {
	tests := []struct {
		name string
		src  string
		want *Document
	}{
		{
			name: "keywords are collected and upcased",
			src:  "#+TITLE: My Note\n#+date: 2026-10-28\n",
			want: &Document{Meta: map[string]string{
				"TITLE": "My Note",
				"DATE":  "2026-10-28",
			}},
		},
		{
			name: "a keyword does not interrupt the prose around it",
			src:  "First line.\n#+TITLE: My Note\nSecond line.\n",
			want: &Document{
				Meta:     map[string]string{"TITLE": "My Note"},
				Children: []Node{&Paragraph{Lines: textLines("First line.", "Second line.")}},
			},
		},
		{
			name: "a keyword with no value",
			src:  "#+TITLE:\n",
			want: &Document{Meta: map[string]string{"TITLE": ""}},
		},
		{
			name: "comments are dropped without breaking a paragraph",
			src:  "First line.\n# a comment\nSecond line.\n",
			want: &Document{Children: []Node{
				&Paragraph{Lines: textLines("First line.", "Second line.")},
			}},
		},
		{
			name: "a bare hash is a comment",
			src:  "#\n",
			want: &Document{},
		},
		{
			name: "a hash without a space is not a comment",
			src:  "#hashtag stays\n",
			want: &Document{Children: []Node{
				&Paragraph{Lines: textLines("#hashtag stays")},
			}},
		},
		{
			name: "a file-level properties drawer",
			src:  ":PROPERTIES:\n:ID: abc-123\n:Category: notes\n:END:\n",
			want: &Document{Properties: map[string]string{
				"ID":       "abc-123",
				"Category": "notes",
			}},
		},
		{
			name: "a properties drawer under a heading is kept as a drawer",
			src:  "* Heading\n:PROPERTIES:\n:ID: abc\n:END:\n",
			want: &Document{Children: []Node{
				&Heading{Level: 1, Title: txt("Heading"), Children: []Node{
					&Drawer{Name: "PROPERTIES", Lines: []string{":ID: abc"}},
				}},
			}},
		},
		{
			name: "a logbook drawer",
			src:  "* Heading\n:LOGBOOK:\nCLOCK: [2026-10-28 Wed 10:00]\n:END:\n",
			want: &Document{Children: []Node{
				&Heading{Level: 1, Title: txt("Heading"), Children: []Node{
					&Drawer{Name: "LOGBOOK", Lines: []string{"CLOCK: [2026-10-28 Wed 10:00]"}},
				}},
			}},
		},
		{
			name: "an unterminated drawer closes at end of input",
			src:  ":NOTES:\nsomething\n",
			want: &Document{Children: []Node{
				&Drawer{Name: "NOTES", Lines: []string{"something"}},
			}},
		},
		{
			name: "drawer contents are literal",
			src:  ":NOTES:\n* not a heading\n:END:\n",
			want: &Document{Children: []Node{
				&Drawer{Name: "NOTES", Lines: []string{"* not a heading"}},
			}},
		},
		{
			name: "keywords and a drawer inside a block stay literal",
			src:  "#+BEGIN_EXAMPLE\n#+TITLE: not meta\n:NOTES:\n#+END_EXAMPLE\n",
			want: &Document{Children: []Node{
				&Block{Lines: []string{"#+TITLE: not meta", ":NOTES:"}},
			}},
		},
		{
			name: "a stray end marker is not a drawer",
			src:  ":END:\n",
			want: &Document{Children: []Node{
				&Paragraph{Lines: textLines(":END:")},
			}},
		},
		{
			name: "keywords, properties and content together",
			src:  "#+TITLE: My Note\n:PROPERTIES:\n:ID: abc\n:END:\n\nBody text.\n",
			want: &Document{
				Meta:       map[string]string{"TITLE": "My Note"},
				Properties: map[string]string{"ID": "abc"},
				Children:   []Node{&Paragraph{Lines: textLines("Body text.")}},
			},
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
