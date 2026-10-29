package md

import (
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/go-spass/org2md/internal/org"
	"github.com/google/go-cmp/cmp"
)

var update = flag.Bool("update", false, "rewrite the golden files in testdata")

func TestRender(t *testing.T) {
	tests := []struct {
		name string
		doc  *org.Document
		want string
	}{
		{
			name: "empty document renders nothing",
			doc:  &org.Document{},
			want: "",
		},
		{
			name: "paragraph keeps its line breaks",
			doc: &org.Document{Children: []org.Node{
				&org.Paragraph{Lines: textLines("One line.", "Another line.")},
			}},
			want: "One line.\nAnother line.\n",
		},
		{
			name: "blocks are separated by a blank line",
			doc: &org.Document{Children: []org.Node{
				&org.Paragraph{Lines: textLines("First.")},
				&org.Paragraph{Lines: textLines("Second.")},
			}},
			want: "First.\n\nSecond.\n",
		},
		{
			name: "heading level maps to hash count",
			doc: &org.Document{Children: []org.Node{
				&org.Heading{Level: 1, Title: txt("One"), Children: []org.Node{
					&org.Heading{Level: 2, Title: txt("Two"), Children: []org.Node{
						&org.Heading{Level: 3, Title: txt("Three")},
					}},
				}},
			}},
			want: "# One\n\n## Two\n\n### Three\n",
		},
		{
			name: "heading body follows the heading",
			doc: &org.Document{Children: []org.Node{
				&org.Heading{Level: 1, Title: txt("Heading"), Children: []org.Node{
					&org.Paragraph{Lines: textLines("Body.")},
				}},
			}},
			want: "# Heading\n\nBody.\n",
		},
		{
			name: "levels deeper than six clamp to six",
			doc: &org.Document{Children: []org.Node{
				&org.Heading{Level: 7, Title: txt("Too deep")},
				&org.Heading{Level: 9, Title: txt("Deeper still")},
			}},
			want: "###### Too deep\n\n###### Deeper still\n",
		},
		{
			name: "heading with no title",
			doc: &org.Document{Children: []org.Node{
				&org.Heading{Level: 2},
			}},
			want: "##\n",
		},
		{
			name: "trailing whitespace is trimmed so it cannot become a hard break",
			doc: &org.Document{Children: []org.Node{
				&org.Paragraph{Lines: textLines("Soft wrapped  ", "in org.\t")},
			}},
			want: "Soft wrapped\nin org.\n",
		},
		{
			name: "markup in a heading title and its body",
			doc: &org.Document{Children: []org.Node{
				&org.Heading{
					Level: 1,
					Title: []org.Inline{
						&org.Text{Value: "A "},
						&org.Bold{Children: txt("bold")},
						&org.Text{Value: " title"},
					},
					Children: []org.Node{
						&org.Paragraph{Lines: [][]org.Inline{{
							&org.Text{Value: "See "},
							&org.Link{Target: "file:other.org", Desc: txt("the other note")},
							&org.Text{Value: "."},
						}}},
					},
				},
			}},
			want: "# A **bold** title\n\nSee [[other|the other note]].\n",
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

// TestGolden converts every testdata/*.org file and compares it with the
// matching .md file. Run `make golden` to rewrite them after an intended
// change, then read the diff before committing it.
func TestGolden(t *testing.T) {
	sources, err := filepath.Glob(filepath.Join("testdata", "*.org"))
	if err != nil {
		t.Fatalf("Glob() error = %v", err)
	}
	if len(sources) == 0 {
		t.Fatal("no testdata/*.org fixtures found")
	}

	for _, src := range sources {
		t.Run(filepath.Base(src), func(t *testing.T) {
			f, err := os.Open(src)
			if err != nil {
				t.Fatalf("Open(%q) error = %v", src, err)
			}
			defer f.Close()

			doc, err := org.Parse(f)
			if err != nil {
				t.Fatalf("org.Parse(%q) error = %v", src, err)
			}
			got, _ := Render(doc)

			golden := strings.TrimSuffix(src, ".org") + ".md"
			if *update {
				if err := os.WriteFile(golden, []byte(got), 0o644); err != nil {
					t.Fatalf("WriteFile(%q) error = %v", golden, err)
				}
				return
			}

			want, err := os.ReadFile(golden)
			if err != nil {
				t.Fatalf("ReadFile(%q) error = %v (run `make golden` to create it)", golden, err)
			}
			if diff := cmp.Diff(string(want), got); diff != "" {
				t.Errorf("%s mismatch (-%s +got):\n%s", src, golden, diff)
			}
		})
	}
}
