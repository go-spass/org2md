package org

// Node is a node of a parsed org document. The set of implementations is closed
// to this package, so a type switch over them is exhaustive.
type Node interface {
	node()
}

// Document is a parsed org file.
type Document struct {
	Children []Node
}

// Heading is an org headline and everything beneath it, including the headings
// nested under it.
type Heading struct {
	Level    int
	Title    string
	Children []Node
}

// Paragraph is a run of consecutive non-blank lines. Source line breaks are
// kept so the renderer can decide what to do with them.
type Paragraph struct {
	Lines []string
}

func (*Heading) node()   {}
func (*Paragraph) node() {}
