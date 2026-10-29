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
	p.flush()

	return p.doc, nil
}

type parser struct {
	doc    *Document
	open   []*Heading
	lines  [][]Inline
	levels []listLevel
}

// listLevel is one indentation level of the list currently being parsed.
type listLevel struct {
	list   *List
	indent int
}

func (p *parser) line(line string) {
	if m := headingRE.FindStringSubmatch(line); m != nil {
		p.endList()
		p.flush()
		p.openHeading(&Heading{Level: len(m[1]), Title: parseInlines(strings.TrimSpace(m[2]))})
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

// flush ends the paragraph being accumulated, if there is one.
func (p *parser) flush() {
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
