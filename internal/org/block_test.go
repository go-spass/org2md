package org

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestParseBlocks(t *testing.T) {
	tests := []struct {
		name string
		src  string
		want *Document
	}{
		{
			name: "source block keeps its language",
			src:  "#+BEGIN_SRC go\nfmt.Println(\"hi\")\n#+END_SRC\n",
			want: &Document{Children: []Node{
				&Block{Lang: "go", Lines: []string{`fmt.Println("hi")`}},
			}},
		},
		{
			name: "header arguments are not part of the language",
			src:  "#+BEGIN_SRC go :results output :exports both\nx := 1\n#+END_SRC\n",
			want: &Document{Children: []Node{
				&Block{Lang: "go", Lines: []string{"x := 1"}},
			}},
		},
		{
			name: "delimiters are case insensitive",
			src:  "#+begin_src python\nprint(1)\n#+end_src\n",
			want: &Document{Children: []Node{
				&Block{Lang: "python", Lines: []string{"print(1)"}},
			}},
		},
		{
			name: "source block without a language",
			src:  "#+BEGIN_SRC\nplain\n#+END_SRC\n",
			want: &Document{Children: []Node{
				&Block{Lines: []string{"plain"}},
			}},
		},
		{
			name: "example block",
			src:  "#+BEGIN_EXAMPLE\nliteral text\n#+END_EXAMPLE\n",
			want: &Document{Children: []Node{
				&Block{Lines: []string{"literal text"}},
			}},
		},
		{
			name: "an unrecognized block is kept literally",
			src:  "#+BEGIN_VERSE\nline one\nline two\n#+END_VERSE\n",
			want: &Document{Children: []Node{
				&Block{Lines: []string{"line one", "line two"}},
			}},
		},
		{
			name: "empty block",
			src:  "#+BEGIN_EXAMPLE\n#+END_EXAMPLE\n",
			want: &Document{Children: []Node{
				&Block{},
			}},
		},

		// The whole reason this tool parses instead of substituting: block
		// content is literal, so org syntax inside it must survive untouched.
		{
			name: "org syntax inside a source block stays literal",
			src:  "#+BEGIN_SRC sh\n* not a heading\n- not a list\necho *not bold* =not code=\n#+END_SRC\n",
			want: &Document{Children: []Node{
				&Block{Lang: "sh", Lines: []string{
					"* not a heading",
					"- not a list",
					"echo *not bold* =not code=",
				}},
			}},
		},
		{
			name: "blank lines inside a block are kept",
			src:  "#+BEGIN_SRC go\nfunc a() {}\n\nfunc b() {}\n#+END_SRC\n",
			want: &Document{Children: []Node{
				&Block{Lang: "go", Lines: []string{"func a() {}", "", "func b() {}"}},
			}},
		},
		{
			name: "indentation inside a block is kept",
			src:  "#+BEGIN_SRC go\nif x {\n\treturn nil\n}\n#+END_SRC\n",
			want: &Document{Children: []Node{
				&Block{Lang: "go", Lines: []string{"if x {", "\treturn nil", "}"}},
			}},
		},
		{
			name: "an unterminated block closes at end of input",
			src:  "#+BEGIN_SRC go\nx := 1\n",
			want: &Document{Children: []Node{
				&Block{Lang: "go", Lines: []string{"x := 1"}},
			}},
		},
		{
			name: "quote contents are parsed as org",
			src:  "#+BEGIN_QUOTE\nA *quoted* line.\n\n- and a list\n#+END_QUOTE\n",
			want: &Document{Children: []Node{
				&Quote{Children: []Node{
					&Paragraph{Lines: [][]Inline{{
						&Text{Value: "A "},
						&Bold{Children: txt("quoted")},
						&Text{Value: " line."},
					}}},
					&List{Items: []*ListItem{{Lines: textLines("and a list")}}},
				}},
			}},
		},
		{
			name: "horizontal rule",
			src:  "before\n\n-----\n\nafter\n",
			want: &Document{Children: []Node{
				&Paragraph{Lines: textLines("before")},
				&Rule{},
				&Paragraph{Lines: textLines("after")},
			}},
		},
		{
			name: "more than five dashes is still a rule",
			src:  "--------\n",
			want: &Document{Children: []Node{&Rule{}}},
		},
		{
			name: "fewer than five dashes is not a rule",
			src:  "----\n",
			want: &Document{Children: []Node{
				&Paragraph{Lines: textLines("----")},
			}},
		},
		{
			name: "a block under a heading",
			src:  "* Heading\n#+BEGIN_SRC go\nx := 1\n#+END_SRC\n",
			want: &Document{Children: []Node{
				&Heading{Level: 1, Title: txt("Heading"), Children: []Node{
					&Block{Lang: "go", Lines: []string{"x := 1"}},
				}},
			}},
		},
		{
			name: "a block ends the list before it",
			src:  "- item\n#+BEGIN_EXAMPLE\nx\n#+END_EXAMPLE\n",
			want: &Document{Children: []Node{
				&List{Items: []*ListItem{{Lines: textLines("item")}}},
				&Block{Lines: []string{"x"}},
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
