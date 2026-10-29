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
	Title    []Inline
	Children []Node
}

// Paragraph is a run of consecutive non-blank lines. Source line breaks are
// kept so the renderer can decide what to do with them.
type Paragraph struct {
	Lines [][]Inline
}

// List is a run of list items at one indentation level.
type List struct {
	Ordered bool
	Items   []*ListItem
}

// ListItem is one entry in a list. Term is set only for a definition item
// (- term :: definition). Lines holds the item's own text, including any
// indented continuation lines, and Children holds lists nested under it.
type ListItem struct {
	Checkbox Checkbox
	Term     []Inline
	Lines    [][]Inline
	Children []Node
}

// Checkbox is the state of a list item's checkbox, if it has one.
type Checkbox int

const (
	NoCheckbox Checkbox = iota
	Unchecked
	Checked
	PartiallyChecked
)

func (*Heading) node()   {}
func (*Paragraph) node() {}
func (*List) node()      {}

// Inline is a span of text inside a block. The set of implementations is closed
// to this package, so a type switch over them is exhaustive.
type Inline interface {
	inline()
}

// Text is a run of characters carrying no markup.
type Text struct {
	Value string
}

// Bold, Italic, Underline and Strike are org's nestable emphasis markers:
// *bold*, /italic/, _underline_ and +strike+.
type Bold struct {
	Children []Inline
}

type Italic struct {
	Children []Inline
}

type Underline struct {
	Children []Inline
}

type Strike struct {
	Children []Inline
}

// Code holds the body of =verbatim= or ~code~. Its contents are literal: the
// inline phase does not look inside it, which is what lets a bare * survive.
type Code struct {
	Value string
}

// Link is an org link. Target is kept exactly as written, including any file:
// prefix, because turning it into a Markdown or wikilink target is the
// renderer's decision. Desc is empty for a bare [[target]] link.
type Link struct {
	Target string
	Desc   []Inline
}

// Timestamp is an active org timestamp. Date is the ISO date; Rest holds a time
// of day and any repeater. The day name is dropped as redundant.
type Timestamp struct {
	Date string
	Rest string
}

func (*Text) inline()      {}
func (*Bold) inline()      {}
func (*Italic) inline()    {}
func (*Underline) inline() {}
func (*Strike) inline()    {}
func (*Code) inline()      {}
func (*Link) inline()      {}
func (*Timestamp) inline() {}
