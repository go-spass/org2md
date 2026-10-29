package org

import (
	"bufio"
	"io"
	"regexp"
	"strings"
)

// maxLineLen bounds a single source line. The scanner default of 64 KiB is
// enough for prose but not for a long table row or an embedded data blob.
const maxLineLen = 1 << 20

var headingRE = regexp.MustCompile(`^(\*+)(?:[ \t]+(.*))?$`)

// Parse reads an org document and returns its syntax tree.
func Parse(r io.Reader) (*Document, error) {
	p := &parser{doc: &Document{}}

	s := bufio.NewScanner(r)
	s.Buffer(make([]byte, 0, bufio.MaxScanTokenSize), maxLineLen)
	for s.Scan() {
		p.line(s.Text())
	}
	if err := s.Err(); err != nil {
		return nil, err
	}
	if p.block != nil {
		p.endBlock()
	}
	if p.drawer != nil {
		p.endDrawer()
	}
	p.flush()

	return p.doc, nil
}

type parser struct {
	doc    *Document
	open   []*Heading
	lines  [][]Inline
	levels []listLevel
	block  *openBlock
	drawer *openDrawer
	table  *Table

	// todoKeywords is this file's todo keyword set, built on first use so a
	// #+TODO: line anywhere above the first headline is honoured.
	todoKeywords map[string]bool

	// seenHeading marks the end of the file-level region, where a
	// :PROPERTIES: drawer describes the document itself.
	seenHeading bool

	// footnote labels the paragraph being accumulated, turning it into a
	// footnote definition when it is flushed.
	footnote string
}

// listLevel is one indentation level of the list currently being parsed.
type listLevel struct {
	list   *List
	indent int
}

func (p *parser) line(line string) {
	// A block's contents are literal, so this has to come first: a line
	// looking like a heading or a list item inside a source block is neither.
	if p.block != nil {
		p.blockLine(line)
		return
	}
	if p.drawer != nil {
		p.drawerLine(line)
		return
	}
	if p.beginBlock(line) {
		return
	}
	if p.beginDrawer(line) {
		return
	}
	if p.keyword(line) {
		return
	}
	if commentRE.MatchString(line) {
		return
	}
	// A table runs for as long as its rows do.
	if p.table != nil && !isTableLine(line) {
		p.endTable()
	}
	if isTableLine(line) {
		p.tableLine(line)
		return
	}
	if ruleRE.MatchString(line) {
		p.endList()
		p.flush()
		p.add(&Rule{})
		return
	}
	if m := headingRE.FindStringSubmatch(line); m != nil {
		p.endList()
		p.endTable()
		p.flush()
		p.seenHeading = true

		h := &Heading{Level: len(m[1])}
		p.parseHeadline(h, strings.TrimSpace(m[2]))
		p.openHeading(h)
		return
	}
	if p.planningLine(line) {
		return
	}
	if strings.TrimSpace(line) == "" {
		// A blank line does not end a list: org allows one inside, and what
		// follows decides whether the list continues.
		if !p.inList() {
			p.flush()
		}
		return
	}
	if m := footnoteDefRE.FindStringSubmatch(line); m != nil {
		p.endList()
		p.flush()
		p.footnote = m[1]
		if m[2] != "" {
			p.lines = append(p.lines, parseInlines(m[2]))
		}
		return
	}
	if item, indent, ordered, ok := parseListItem(line); ok {
		p.flush()
		p.addListItem(indent, ordered, item)
		return
	}
	if p.addContinuation(line) {
		return
	}
	p.endList()
	p.lines = append(p.lines, parseInlines(line))
}

// flush ends the paragraph being accumulated, if there is one. A labelled
// paragraph is a footnote definition.
func (p *parser) flush() {
	if p.footnote != "" {
		p.add(&FootnoteDef{Label: p.footnote, Lines: p.lines})
		p.footnote = ""
		p.lines = nil
		return
	}
	if len(p.lines) == 0 {
		return
	}
	p.add(&Paragraph{Lines: p.lines})
	p.lines = nil
}

// openHeading closes every heading at or below h's level, then makes h the
// container for the nodes that follow.
func (p *parser) openHeading(h *Heading) {
	for len(p.open) > 0 && p.open[len(p.open)-1].Level >= h.Level {
		p.open = p.open[:len(p.open)-1]
	}
	p.add(h)
	p.open = append(p.open, h)
}

func (p *parser) add(n Node) {
	if len(p.open) == 0 {
		p.doc.Children = append(p.doc.Children, n)
		return
	}
	parent := p.open[len(p.open)-1]
	parent.Children = append(parent.Children, n)
}
