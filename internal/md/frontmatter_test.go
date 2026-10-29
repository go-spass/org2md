package md

import (
	"testing"

	"github.com/go-spass/org2md/internal/org"
	"github.com/google/go-cmp/cmp"
)

func TestFrontmatter(t *testing.T) {
	body := []org.Node{&org.Paragraph{Lines: textLines("Body.")}}

	tests := []struct {
		name string
		doc  *org.Document
		want string
	}{
		{
			name: "no metadata means no frontmatter block",
			doc:  &org.Document{Children: body},
			want: "Body.\n",
		},
		{
			name: "title",
			doc:  &org.Document{Meta: map[string]string{"TITLE": "My Note"}, Children: body},
			want: "---\ntitle: My Note\n---\n\nBody.\n",
		},
		{
			name: "frontmatter with no body",
			doc:  &org.Document{Meta: map[string]string{"TITLE": "My Note"}},
			want: "---\ntitle: My Note\n---\n",
		},
		{
			name: "an empty keyword value is not emitted",
			doc:  &org.Document{Meta: map[string]string{"TITLE": ""}, Children: body},
			want: "Body.\n",
		},
		{
			name: "every mapped keyword",
			doc: &org.Document{Meta: map[string]string{
				"TITLE":    "My Note",
				"DATE":     "2026-10-28",
				"AUTHOR":   "Chris",
				"FILETAGS": ":work:idea:",
			}},
			want: "---\ntitle: My Note\ndate: 2026-10-28\nauthor: Chris\ntags:\n  - work\n  - idea\n---\n",
		},
		{
			name: "an org timestamp becomes a plain date",
			doc:  &org.Document{Meta: map[string]string{"DATE": "<2026-10-28 Wed>"}},
			want: "---\ndate: 2026-10-28\n---\n",
		},
		{
			// Quoting would not change the YAML: both forms are the string
			// "sometime in October". The scalar rules decide, as everywhere else.
			name: "a date that is not a date is left as text",
			doc:  &org.Document{Meta: map[string]string{"DATE": "sometime in October"}},
			want: "---\ndate: sometime in October\n---\n",
		},
		{
			name: "space separated filetags",
			doc:  &org.Document{Meta: map[string]string{"FILETAGS": "work idea"}},
			want: "---\ntags:\n  - work\n  - idea\n---\n",
		},
		{
			name: "a title holding a colon is quoted",
			doc:  &org.Document{Meta: map[string]string{"TITLE": "Note: a subtitle"}},
			want: "---\ntitle: \"Note: a subtitle\"\n---\n",
		},
		{
			name: "a year-only date is quoted so it stays a string",
			doc:  &org.Document{Meta: map[string]string{"DATE": "2026"}},
			want: "---\ndate: \"2026\"\n---\n",
		},
		{
			name: "a title that would read as a number is quoted",
			doc:  &org.Document{Meta: map[string]string{"TITLE": "2026"}},
			want: "---\ntitle: \"2026\"\n---\n",
		},
		{
			name: "a title that would read as a boolean is quoted",
			doc:  &org.Document{Meta: map[string]string{"TITLE": "No"}},
			want: "---\ntitle: \"No\"\n---\n",
		},
		{
			name: "a title with a quote in it is escaped",
			doc:  &org.Document{Meta: map[string]string{"TITLE": `The "good" note`}},
			want: "---\ntitle: \"The \\\"good\\\" note\"\n---\n",
		},
		{
			name: "a backslash in a quoted value is escaped",
			doc:  &org.Document{Meta: map[string]string{"TITLE": `C:\notes\org`}},
			want: "---\ntitle: \"C:\\\\notes\\\\org\"\n---\n",
		},
		{
			name: "properties are lowercased and sorted",
			doc: &org.Document{Properties: map[string]string{
				"ID":       "abc-123",
				"Category": "notes",
			}},
			want: "---\ncategory: notes\nid: abc-123\n---\n",
		},
		{
			name: "a property key that is not a plain key is quoted",
			doc:  &org.Document{Properties: map[string]string{"My Key": "v"}},
			want: "---\n\"my key\": v\n---\n",
		},
		{
			name: "keywords come before properties",
			doc: &org.Document{
				Meta:       map[string]string{"TITLE": "My Note"},
				Properties: map[string]string{"ID": "abc"},
			},
			want: "---\ntitle: My Note\nid: abc\n---\n",
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

func TestWarnings(t *testing.T) {
	tests := []struct {
		name string
		doc  *org.Document
		want []string
	}{
		{
			name: "nothing to report",
			doc: &org.Document{
				Meta:     map[string]string{"TITLE": "My Note"},
				Children: []org.Node{&org.Paragraph{Lines: textLines("Body.")}},
			},
			want: nil,
		},
		{
			name: "a keyword with no frontmatter equivalent",
			doc: &org.Document{Meta: map[string]string{
				"TITLE":   "My Note",
				"OPTIONS": "toc:nil",
				"STARTUP": "overview",
			}},
			want: []string{
				"dropped #+OPTIONS: no frontmatter equivalent",
				"dropped #+STARTUP: no frontmatter equivalent",
			},
		},
		{
			name: "a dropped drawer",
			doc:  doc(&org.Drawer{Name: "NOTES", Lines: []string{"private"}}),
			want: []string{"dropped :NOTES: drawer"},
		},
		{
			name: "a logbook is dropped quietly",
			doc:  doc(&org.Drawer{Name: "LOGBOOK", Lines: []string{"CLOCK: x"}}),
			want: nil,
		},
		{
			name: "a property that would duplicate a key",
			doc: &org.Document{
				Meta:       map[string]string{"TITLE": "From the keyword"},
				Properties: map[string]string{"TITLE": "From the drawer"},
			},
			want: []string{"dropped property TITLE: the title key is already set"},
		},
		{
			name: "keyword warnings come before drawer warnings",
			doc: &org.Document{
				Meta:     map[string]string{"OPTIONS": "toc:nil"},
				Children: []org.Node{&org.Drawer{Name: "NOTES"}},
			},
			want: []string{
				"dropped #+OPTIONS: no frontmatter equivalent",
				"dropped :NOTES: drawer",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, got := Render(tt.doc)
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("Render() warnings mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
