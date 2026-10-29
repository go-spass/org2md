package org

import (
	"regexp"
	"strings"
)

// defaultTodoKeywords is used when the file declares none. Org itself defaults
// to TODO and DONE and reads the rest from the user's Emacs configuration,
// which is not in the file, so the common workflow words are included too.
var defaultTodoKeywords = []string{
	"TODO", "NEXT", "STARTED", "WAITING", "HOLD", "SOMEDAY",
	"DONE", "CANCELLED", "CANCELED",
}

// todoDeclarations are the keyword lines that replace the default set.
var todoDeclarations = []string{"TODO", "SEQ_TODO", "TYP_TODO"}

var (
	firstWordRE = regexp.MustCompile(`^(\S+)(?:[ \t]+(.*))?$`)
	priorityRE  = regexp.MustCompile(`^\[#([A-Za-z])\](?:[ \t]+(.*))?$`)

	// Org tags are colon-delimited and sit at the very end of the headline.
	tagsRE = regexp.MustCompile(`^(?:(.*?)[ \t]+)?((?::[A-Za-z0-9_@#%]+)+:)[ \t]*$`)

	// A planning entry is a keyword and a timestamp, active or inactive.
	planningRE = regexp.MustCompile(`(SCHEDULED|DEADLINE|CLOSED):[ \t]*([<\[][^>\]]*[>\]])`)
)

// parseHeadline reads everything a headline carries besides its level: the todo
// keyword, the priority, the tags, and the title left over.
func (p *parser) parseHeadline(h *Heading, text string) {
	if m := tagsRE.FindStringSubmatch(text); m != nil {
		h.Tags = strings.Split(strings.Trim(m[2], ":"), ":")
		text = m[1]
	}
	if m := firstWordRE.FindStringSubmatch(text); m != nil && p.isTodoKeyword(m[1]) {
		h.Todo = m[1]
		text = m[2]
	}
	if m := priorityRE.FindStringSubmatch(text); m != nil {
		h.Priority = strings.ToUpper(m[1])
		text = m[2]
	}

	h.Title = parseInlines(strings.TrimSpace(text))
}

// isTodoKeyword reports whether a word is a todo keyword for this file. The
// set comes from a #+TODO: line when the file has one, so a custom workflow is
// recognized and the default words it replaces are not.
func (p *parser) isTodoKeyword(word string) bool {
	if p.todoKeywords == nil {
		p.todoKeywords = map[string]bool{}

		declared := false
		for _, name := range todoDeclarations {
			for _, kw := range strings.Fields(p.doc.Meta[name]) {
				if kw == "|" {
					continue
				}
				p.todoKeywords[kw] = true
				declared = true
			}
		}
		if !declared {
			for _, kw := range defaultTodoKeywords {
				p.todoKeywords[kw] = true
			}
		}
	}

	return p.todoKeywords[word]
}

// planningLine attaches a line of planning entries to the heading it belongs
// to. A line that mixes planning with anything else is not one, and neither is
// a line with no heading above it.
func (p *parser) planningLine(line string) bool {
	if len(p.open) == 0 {
		return false
	}

	matches := planningRE.FindAllStringSubmatch(line, -1)
	if len(matches) == 0 || strings.TrimSpace(planningRE.ReplaceAllString(line, "")) != "" {
		return false
	}

	h := p.open[len(p.open)-1]
	for _, m := range matches {
		ts := parseTimestamp(m[2])
		if ts == nil {
			continue
		}
		switch m[1] {
		case "SCHEDULED":
			h.Scheduled = ts
		case "DEADLINE":
			h.Deadline = ts
		case "CLOSED":
			h.Closed = ts
		}
	}

	return true
}

// parseTimestamp reads a bracketed org timestamp, active or inactive.
func parseTimestamp(s string) *Timestamp {
	inner := strings.TrimRight(strings.TrimLeft(s, "<["), ">]")
	m := timestampRE.FindStringSubmatch(inner)
	if m == nil {
		return nil
	}
	return &Timestamp{Date: m[1], Rest: strings.TrimSpace(m[2])}
}
