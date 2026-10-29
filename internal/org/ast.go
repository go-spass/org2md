package org

// Node is a node of a parsed org document. The set of implementations is closed
// to this package, so a type switch over them is exhaustive.
type Node interface {
	node()
}

// Document is a parsed org file. Meta holds its #+KEYWORD: lines, keyed by
// upcased name, and Properties the file-level :PROPERTIES: drawer. Both are nil
// when the file has none.
type Document struct {
	Meta       map[string]string
	Properties map[string]string
	Children   []Node
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

// Block is a literal block: a source block, an example block, or a block whose
// type org2md does not recognize. Its lines are never parsed for markup, which
// is what lets org syntax inside a code sample survive.
type Block struct {
	Lang  string
	Lines []string
}

// Quote is a #+BEGIN_QUOTE block. Unlike Block its contents are org markup, so
// they are parsed.
type Quote struct {
	Children []Node
}

// Rule is a horizontal rule.
type Rule struct{}

// Drawer is a drawer other than the file-level :PROPERTIES: one. Its contents
// are literal: what becomes of them is the renderer's decision.
type Drawer struct {
	Name  string
	Lines []string
}

// Table is an org table. Separator rows are kept in place rather than
// interpreted, because how many of them a header may have is a Markdown
// question rather than an org one. Align is set only when the table declares
// alignment cookies.
type Table struct {
	Align []Alignment
	Rows  []*TableRow
}

// TableRow is one row of a table, or one of its horizontal separators.
type TableRow struct {
	Separator bool
	Cells     [][]Inline
}

// Alignment is a column's declared alignment.
type Alignment int

const (
	AlignDefault Alignment = iota
	AlignLeft
	AlignCenter
	AlignRight
)

// FootnoteDef is a footnote definition: [fn:label] and the text that follows.
type FootnoteDef struct {
	Label string
	Lines [][]Inline
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
func (*Block) node()     {}
func (*Quote) node()     {}
func (*Rule) node()      {}
func (*Table) node()     {}
func (*Drawer) node()    {}

func (*FootnoteDef) node() {}

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

// FootnoteRef is a reference to a footnote, [fn:label].
type FootnoteRef struct {
	Label string
}

// Timestamp is an active org timestamp. Date is the ISO date; Rest holds a time
// of day and any repeater. The day name is dropped as redundant.
type Timestamp struct {
	Date string
	Rest string
}

func (*Text) inline()        {}
func (*Bold) inline()        {}
func (*Italic) inline()      {}
func (*Underline) inline()   {}
func (*Strike) inline()      {}
func (*Code) inline()        {}
func (*Link) inline()        {}
func (*Timestamp) inline()   {}
func (*FootnoteRef) inline() {}
