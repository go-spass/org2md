package org

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestParseTaskHeadings(t *testing.T) {
	tests := []struct {
		name string
		src  string
		want *Document
	}{
		{
			name: "todo keyword",
			src:  "* TODO Write the parser\n",
			want: &Document{Children: []Node{
				&Heading{Level: 1, Todo: "TODO", Title: txt("Write the parser")},
			}},
		},
		{
			name: "done keyword",
			src:  "** DONE Ship it\n",
			want: &Document{Children: []Node{
				&Heading{Level: 2, Todo: "DONE", Title: txt("Ship it")},
			}},
		},
		{
			name: "a keyword on its own",
			src:  "* TODO\n",
			want: &Document{Children: []Node{
				&Heading{Level: 1, Todo: "TODO"},
			}},
		},
		{
			name: "priority",
			src:  "* TODO [#A] Urgent\n",
			want: &Document{Children: []Node{
				&Heading{Level: 1, Todo: "TODO", Priority: "A", Title: txt("Urgent")},
			}},
		},
		{
			name: "priority without a keyword",
			src:  "* [#B] Important\n",
			want: &Document{Children: []Node{
				&Heading{Level: 1, Priority: "B", Title: txt("Important")},
			}},
		},
		{
			name: "tags",
			src:  "* Some heading :work:idea:\n",
			want: &Document{Children: []Node{
				&Heading{Level: 1, Title: txt("Some heading"), Tags: []string{"work", "idea"}},
			}},
		},
		{
			name: "tags only",
			src:  "* :work:\n",
			want: &Document{Children: []Node{
				&Heading{Level: 1, Tags: []string{"work"}},
			}},
		},
		{
			name: "keyword, priority, title, markup and tags together",
			src:  "* TODO [#A] Fix the *parser* :work:bug:\n",
			want: &Document{Children: []Node{
				&Heading{
					Level:    1,
					Todo:     "TODO",
					Priority: "A",
					Title: []Inline{
						&Text{Value: "Fix the "},
						&Bold{Children: txt("parser")},
					},
					Tags: []string{"work", "bug"},
				},
			}},
		},
		{
			name: "a word that merely starts with a keyword is not one",
			src:  "* TODOS of the week\n",
			want: &Document{Children: []Node{
				&Heading{Level: 1, Title: txt("TODOS of the week")},
			}},
		},
		{
			name: "keywords are case sensitive",
			src:  "* todo lowercase\n",
			want: &Document{Children: []Node{
				&Heading{Level: 1, Title: txt("todo lowercase")},
			}},
		},
		{
			name: "a mid-line colon is not a tag",
			src:  "* See :config: for details\n",
			want: &Document{Children: []Node{
				&Heading{Level: 1, Title: txt("See :config: for details")},
			}},
		},
		{
			name: "a declared keyword set replaces the default",
			src:  "#+TODO: TODO FINISHED\n* FINISHED done differently\n* DONE no longer a keyword\n",
			want: &Document{
				Meta: map[string]string{"TODO": "TODO FINISHED"},
				Children: []Node{
					&Heading{Level: 1, Todo: "FINISHED", Title: txt("done differently")},
					&Heading{Level: 1, Title: txt("DONE no longer a keyword")},
				},
			},
		},
		{
			name: "the separator in a declared keyword set is not a keyword",
			src:  "#+TODO: TODO | DONE\n* TODO still works\n",
			want: &Document{
				Meta: map[string]string{"TODO": "TODO | DONE"},
				Children: []Node{
					&Heading{Level: 1, Todo: "TODO", Title: txt("still works")},
				},
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

func TestParsePlanning(t *testing.T) {
	tests := []struct {
		name string
		src  string
		want *Document
	}{
		{
			name: "scheduled",
			src:  "* TODO Task\nSCHEDULED: <2026-10-28 Wed>\n",
			want: &Document{Children: []Node{
				&Heading{
					Level:     1,
					Todo:      "TODO",
					Title:     txt("Task"),
					Scheduled: &Timestamp{Date: "2026-10-28"},
				},
			}},
		},
		{
			name: "deadline and scheduled on one line",
			src:  "* TODO Task\nDEADLINE: <2026-11-01 Sun> SCHEDULED: <2026-10-28 Wed>\n",
			want: &Document{Children: []Node{
				&Heading{
					Level:     1,
					Todo:      "TODO",
					Title:     txt("Task"),
					Scheduled: &Timestamp{Date: "2026-10-28"},
					Deadline:  &Timestamp{Date: "2026-11-01"},
				},
			}},
		},
		{
			name: "a time and repeater are kept",
			src:  "* TODO Task\nSCHEDULED: <2026-10-28 Wed 14:30 +1w>\n",
			want: &Document{Children: []Node{
				&Heading{
					Level:     1,
					Todo:      "TODO",
					Title:     txt("Task"),
					Scheduled: &Timestamp{Date: "2026-10-28", Rest: "14:30 +1w"},
				},
			}},
		},
		{
			name: "closed uses an inactive timestamp",
			src:  "* DONE Task\nCLOSED: [2026-10-28 Wed 10:00]\n",
			want: &Document{Children: []Node{
				&Heading{
					Level:  1,
					Todo:   "DONE",
					Title:  txt("Task"),
					Closed: &Timestamp{Date: "2026-10-28", Rest: "10:00"},
				},
			}},
		},
		{
			name: "planning does not swallow the body",
			src:  "* TODO Task\nSCHEDULED: <2026-10-28 Wed>\nBody text.\n",
			want: &Document{Children: []Node{
				&Heading{
					Level:     1,
					Todo:      "TODO",
					Title:     txt("Task"),
					Scheduled: &Timestamp{Date: "2026-10-28"},
					Children:  []Node{&Paragraph{Lines: textLines("Body text.")}},
				},
			}},
		},
		{
			name: "planning attaches to the innermost heading",
			src:  "* Outer\n** TODO Inner\nSCHEDULED: <2026-10-28 Wed>\n",
			want: &Document{Children: []Node{
				&Heading{Level: 1, Title: txt("Outer"), Children: []Node{
					&Heading{
						Level:     2,
						Todo:      "TODO",
						Title:     txt("Inner"),
						Scheduled: &Timestamp{Date: "2026-10-28"},
					},
				}},
			}},
		},
		{
			name: "with no heading above it, planning is prose",
			src:  "SCHEDULED: <2026-10-28 Wed>\n",
			want: &Document{Children: []Node{
				&Paragraph{Lines: [][]Inline{{
					&Text{Value: "SCHEDULED: "},
					&Timestamp{Date: "2026-10-28"},
				}}},
			}},
		},
		{
			name: "a line mixing planning with prose is prose",
			src:  "* Task\nSCHEDULED: <2026-10-28 Wed> and some notes\n",
			want: &Document{Children: []Node{
				&Heading{Level: 1, Title: txt("Task"), Children: []Node{
					&Paragraph{Lines: [][]Inline{{
						&Text{Value: "SCHEDULED: "},
						&Timestamp{Date: "2026-10-28"},
						&Text{Value: " and some notes"},
					}}},
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
