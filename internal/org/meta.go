package org

import (
	"regexp"
	"strings"
)

var (
	// A keyword is #+NAME: value. Block delimiters carry no colon, so they do
	// not collide.
	keywordRE = regexp.MustCompile(`^[ \t]*#\+([A-Za-z_]+):[ \t]*(.*)$`)

	// A comment is a hash followed by whitespace or nothing. "#hashtag" is not
	// a comment.
	commentRE = regexp.MustCompile(`^[ \t]*#([ \t].*)?$`)

	drawerBeginRE = regexp.MustCompile(`^[ \t]*:([A-Za-z][A-Za-z0-9_-]*):[ \t]*$`)
	drawerEndRE   = regexp.MustCompile(`(?i)^[ \t]*:END:[ \t]*$`)
	propertyRE    = regexp.MustCompile(`^[ \t]*:([^:]+):[ \t]*(.*)$`)
)

// openDrawer is the drawer currently being collected.
type openDrawer struct {
	name  string
	lines []string
}

// keyword records a #+NAME: value line. It deliberately leaves the parser's
// other state alone, so a keyword between two lines of prose does not split
// the paragraph.
func (p *parser) keyword(line string) bool {
	m := keywordRE.FindStringSubmatch(line)
	if m == nil {
		return false
	}

	if p.doc.Meta == nil {
		p.doc.Meta = map[string]string{}
	}
	p.doc.Meta[strings.ToUpper(m[1])] = m[2]

	return true
}

func (p *parser) beginDrawer(line string) bool {
	m := drawerBeginRE.FindStringSubmatch(line)
	if m == nil || strings.EqualFold(m[1], "END") {
		return false
	}

	p.endList()
	p.endTable()
	p.flush()
	p.drawer = &openDrawer{name: strings.ToUpper(m[1])}

	return true
}

// drawerLine consumes a line while a drawer is open. Drawer contents are
// literal, so nothing in them is parsed as org.
func (p *parser) drawerLine(line string) {
	if drawerEndRE.MatchString(line) {
		p.endDrawer()
		return
	}
	p.drawer.lines = append(p.drawer.lines, line)
}

// endDrawer attaches the collected drawer. A properties drawer before the
// first heading describes the file, so it becomes the document's properties;
// every other drawer is kept as a node for the renderer to deal with.
func (p *parser) endDrawer() {
	d := p.drawer
	p.drawer = nil

	if d.name == "PROPERTIES" && !p.seenHeading {
		for _, line := range d.lines {
			m := propertyRE.FindStringSubmatch(line)
			if m == nil {
				continue
			}
			if p.doc.Properties == nil {
				p.doc.Properties = map[string]string{}
			}
			p.doc.Properties[strings.TrimSpace(m[1])] = m[2]
		}
		return
	}

	p.add(&Drawer{Name: d.name, Lines: d.lines})
}
