package org

import (
	"regexp"
	"strings"
)

var (
	blockBeginRE = regexp.MustCompile(`(?i)^[ \t]*#\+BEGIN_(\S+)[ \t]*(.*)$`)
	blockEndRE   = regexp.MustCompile(`(?i)^[ \t]*#\+END_(\S+)[ \t]*$`)

	// Org's horizontal rule is five or more dashes on a line of their own.
	ruleRE = regexp.MustCompile(`^[ \t]*-{5,}[ \t]*$`)
)

// openBlock is the block currently being collected.
type openBlock struct {
	kind  string
	lang  string
	lines []string
}

// beginBlock starts collecting a block if the line opens one.
func (p *parser) beginBlock(line string) bool {
	m := blockBeginRE.FindStringSubmatch(line)
	if m == nil {
		return false
	}

	p.endList()
	p.flush()

	// Anything after the language is a babel header argument, which has no
	// Markdown equivalent.
	lang, _, _ := strings.Cut(strings.TrimSpace(m[2]), " ")
	p.block = &openBlock{kind: strings.ToUpper(m[1]), lang: lang}

	return true
}

// blockLine consumes a line while a block is open, closing the block when it
// reaches the matching #+END_ delimiter.
func (p *parser) blockLine(line string) {
	if m := blockEndRE.FindStringSubmatch(line); m != nil && strings.EqualFold(m[1], p.block.kind) {
		p.endBlock()
		return
	}
	p.block.lines = append(p.block.lines, line)
}

// endBlock attaches the collected block. A quote holds org markup, so its
// contents are parsed; everything else is literal and is kept as written.
func (p *parser) endBlock() {
	b := p.block
	p.block = nil

	if b.kind == "QUOTE" {
		inner, err := Parse(strings.NewReader(strings.Join(b.lines, "\n")))
		if err == nil {
			p.add(&Quote{Children: inner.Children})
			return
		}
	}

	block := &Block{Lines: b.lines}
	if b.kind == "SRC" {
		block.Lang = b.lang
	}
	p.add(block)
}
