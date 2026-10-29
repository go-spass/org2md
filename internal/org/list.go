package org

import (
	"regexp"
	"strings"
)

// A list item is a bullet (-, + or an indented *) or a number followed by . or
// ), then whitespace. Requiring that whitespace is what keeps "-not a bullet"
// and the ----- rule from being read as items.
var listItemRE = regexp.MustCompile(`^([ \t]*)(?:([-+*])|(\d+)[.)])(?:[ \t]+(.*))?$`)

var checkboxRE = regexp.MustCompile(`^\[([ xX-])\](?:[ \t]+(.*))?$`)

// tabWidth is how wide a tab counts when comparing indentation.
const tabWidth = 8

// parseListItem reads one list-item line. ordered reports a numbered item and
// indent is the width of the leading whitespace, which is what decides nesting.
func parseListItem(line string) (item *ListItem, indent int, ordered bool, ok bool) {
	m := listItemRE.FindStringSubmatch(line)
	if m == nil {
		return nil, 0, false, false
	}
	// At column zero an asterisk is a heading, which the heading pattern has
	// already had its chance at.
	if m[2] == "*" && m[1] == "" {
		return nil, 0, false, false
	}

	item = &ListItem{}
	content := m[4]

	if cb := checkboxRE.FindStringSubmatch(content); cb != nil {
		item.Checkbox = checkboxState(cb[1])
		content = cb[2]
	}
	if term, desc, found := strings.Cut(content, " ::"); found {
		item.Term = parseInlines(strings.TrimSpace(term))
		content = strings.TrimSpace(desc)
	}
	if content != "" {
		item.Lines = [][]Inline{parseInlines(content)}
	}

	return item, indentWidth(m[1]), m[3] != "", true
}

func checkboxState(s string) Checkbox {
	switch s {
	case "x", "X":
		return Checked
	case "-":
		return PartiallyChecked
	default:
		return Unchecked
	}
}

// addListItem places an item in the list stack, opening a nested list when the
// indentation grows and closing levels when it shrinks.
func (p *parser) addListItem(indent int, ordered bool, item *ListItem) {
	for len(p.levels) > 0 && p.levels[len(p.levels)-1].indent > indent {
		p.levels = p.levels[:len(p.levels)-1]
	}

	// A different marker at the same indentation starts a sibling list rather
	// than continuing this one.
	if n := len(p.levels); n > 0 && p.levels[n-1].indent == indent && p.levels[n-1].list.Ordered != ordered {
		p.levels = p.levels[:n-1]
	}

	if n := len(p.levels); n == 0 || p.levels[n-1].indent < indent {
		p.pushList(indent, ordered)
	}

	top := p.levels[len(p.levels)-1].list
	top.Items = append(top.Items, item)
}

func (p *parser) pushList(indent int, ordered bool) {
	list := &List{Ordered: ordered}
	if len(p.levels) == 0 {
		p.add(list)
	} else {
		parent := p.levels[len(p.levels)-1].list
		last := parent.Items[len(parent.Items)-1]
		last.Children = append(last.Children, list)
	}
	p.levels = append(p.levels, listLevel{list: list, indent: indent})
}

// addContinuation appends an indented non-item line to the item that owns its
// indentation, closing any deeper levels on the way, and reports whether the
// line belonged to the list at all.
func (p *parser) addContinuation(line string) bool {
	indent := indentWidth(line)
	for len(p.levels) > 0 && indent <= p.levels[len(p.levels)-1].indent {
		p.levels = p.levels[:len(p.levels)-1]
	}
	if len(p.levels) == 0 {
		return false
	}

	top := p.levels[len(p.levels)-1].list
	item := top.Items[len(top.Items)-1]
	item.Lines = append(item.Lines, parseInlines(strings.TrimLeft(line, " \t")))

	return true
}

func (p *parser) inList() bool {
	return len(p.levels) > 0
}

func (p *parser) endList() {
	p.levels = nil
}

func indentWidth(s string) int {
	w := 0
	for _, r := range s {
		switch r {
		case ' ':
			w++
		case '\t':
			w += tabWidth
		default:
			return w
		}
	}
	return w
}
