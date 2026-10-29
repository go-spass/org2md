package md

import (
	"testing"

	"github.com/go-spass/org2md/internal/org"
	"github.com/google/go-cmp/cmp"
)

func TestRenderTaskHeadings(t *testing.T) {
	tests := []struct {
		name string
		doc  *org.Document
		want string
	}{
		{
			name: "a todo keyword becomes a tag",
			doc:  doc(&org.Heading{Level: 1, Todo: "TODO", Title: txt("Task")}),
			want: "# Task #todo\n",
		},
		{
			name: "a done keyword becomes a tag",
			doc:  doc(&org.Heading{Level: 2, Todo: "DONE", Title: txt("Task")}),
			want: "## Task #done\n",
		},
		{
			name: "a custom keyword becomes a tag too",
			doc:  doc(&org.Heading{Level: 1, Todo: "IN-PROGRESS", Title: txt("Task")}),
			want: "# Task #in-progress\n",
		},
		{
			name: "priority",
			doc:  doc(&org.Heading{Level: 1, Priority: "A", Title: txt("Task")}),
			want: "# Task #priority/a\n",
		},
		{
			name: "heading tags keep the case they were written in",
			doc:  doc(&org.Heading{Level: 1, Title: txt("Task"), Tags: []string{"Work", "idea"}}),
			want: "# Task #Work #idea\n",
		},
		{
			name: "status, then priority, then tags",
			doc: doc(&org.Heading{
				Level:    1,
				Todo:     "TODO",
				Priority: "A",
				Title:    txt("Task"),
				Tags:     []string{"work", "bug"},
			}),
			want: "# Task #todo #priority/a #work #bug\n",
		},
		{
			name: "a headline with no title but a keyword",
			doc:  doc(&org.Heading{Level: 1, Todo: "TODO"}),
			want: "# #todo\n",
		},
		{
			name: "markup in the title survives alongside tags",
			doc: doc(&org.Heading{
				Level: 1,
				Todo:  "TODO",
				Title: []org.Inline{&org.Bold{Children: txt("Fix")}},
				Tags:  []string{"bug"},
			}),
			want: "# **Fix** #todo #bug\n",
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

func TestRenderPlanning(t *testing.T) {
	tests := []struct {
		name string
		doc  *org.Document
		want string
	}{
		{
			name: "scheduled",
			doc: doc(&org.Heading{
				Level:     1,
				Title:     txt("Task"),
				Scheduled: &org.Timestamp{Date: "2026-10-28"},
			}),
			want: "# Task\n\n**Scheduled:** 2026-10-28\n",
		},
		{
			name: "deadline",
			doc: doc(&org.Heading{
				Level:    1,
				Title:    txt("Task"),
				Deadline: &org.Timestamp{Date: "2026-11-01"},
			}),
			want: "# Task\n\n**Deadline:** 2026-11-01\n",
		},
		{
			name: "scheduled comes before deadline",
			doc: doc(&org.Heading{
				Level:     1,
				Title:     txt("Task"),
				Scheduled: &org.Timestamp{Date: "2026-10-28"},
				Deadline:  &org.Timestamp{Date: "2026-11-01"},
			}),
			want: "# Task\n\n**Scheduled:** 2026-10-28\n**Deadline:** 2026-11-01\n",
		},
		{
			name: "a time and repeater are kept",
			doc: doc(&org.Heading{
				Level:     1,
				Title:     txt("Task"),
				Scheduled: &org.Timestamp{Date: "2026-10-28", Rest: "14:30 +1w"},
			}),
			want: "# Task\n\n**Scheduled:** 2026-10-28 14:30 +1w\n",
		},
		{
			name: "closed is dropped",
			doc: doc(&org.Heading{
				Level:  1,
				Todo:   "DONE",
				Title:  txt("Task"),
				Closed: &org.Timestamp{Date: "2026-10-28", Rest: "10:00"},
			}),
			want: "# Task #done\n",
		},
		{
			name: "planning comes before the body",
			doc: doc(&org.Heading{
				Level:     1,
				Todo:      "TODO",
				Title:     txt("Task"),
				Scheduled: &org.Timestamp{Date: "2026-10-28"},
				Children:  []org.Node{&org.Paragraph{Lines: textLines("Body.")}},
			}),
			want: "# Task #todo\n\n**Scheduled:** 2026-10-28\n\nBody.\n",
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
