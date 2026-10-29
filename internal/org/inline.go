package org

import (
	"regexp"
	"strings"
)

// Org only recognizes an emphasis marker at a word boundary: the character
// before the opening marker must be one of these (or the start of the line),
// and the character after the closing marker must be one of postChars (or the
// end of the line). Without that rule some_variable_name, and/or and 2*3*4
// would all be read as markup.
const (
	preChars  = " \t-–—('\"{"
	postChars = " \t-–—.,:!?;'\")}["
)

// timestampRE matches the inside of an active org timestamp: a date, an
// optional day name, and an optional time and/or repeater.
var timestampRE = regexp.MustCompile(`^(\d{4}-\d{2}-\d{2})(?:\s+[A-Za-z]{3,})?(?:\s+(.*))?$`)

// emphasisKinds maps each marker to a constructor. A nil constructor marks a
// literal marker, whose body is taken verbatim instead of being parsed.
var emphasisKinds = map[rune]func([]Inline) Inline{
	'*': func(c []Inline) Inline { return &Bold{Children: c} },
	'/': func(c []Inline) Inline { return &Italic{Children: c} },
	'_': func(c []Inline) Inline { return &Underline{Children: c} },
	'+': func(c []Inline) Inline { return &Strike{Children: c} },
	'=': nil,
	'~': nil,
}

// parseInlines runs the inline phase over one line of org text. Callers must
// not pass content that forbids markup, such as the body of a source block.
func parseInlines(s string) []Inline {
	p := &inlineParser{src: []rune(s)}
	return p.parse()
}

type inlineParser struct {
	src []rune
	buf []rune
	out []Inline
}

func (p *inlineParser) parse() []Inline {
	for i := 0; i < len(p.src); {
		if n, width := p.match(i); n != nil {
			p.flushText()
			p.out = append(p.out, n)
			i += width
			continue
		}
		p.buf = append(p.buf, p.src[i])
		i++
	}
	p.flushText()
	return p.out
}

func (p *inlineParser) flushText() {
	if len(p.buf) == 0 {
		return
	}
	p.out = append(p.out, &Text{Value: string(p.buf)})
	p.buf = nil
}

// match reports the inline element starting at i and how many runes it spans,
// or nil when i does not begin one.
func (p *inlineParser) match(i int) (Inline, int) {
	if n, width := p.footnoteRef(i); n != nil {
		return n, width
	}
	if n, width := p.link(i); n != nil {
		return n, width
	}
	if n, width := p.timestamp(i); n != nil {
		return n, width
	}
	return p.emphasis(i)
}

func (p *inlineParser) footnoteRef(i int) (Inline, int) {
	if p.src[i] != '[' || !strings.HasPrefix(string(p.src[i:]), "[fn:") {
		return nil, 0
	}
	end := p.indexFrom(i+4, "]")
	if end < 0 || end == i+4 {
		return nil, 0
	}

	label := string(p.src[i+4 : end])
	if strings.ContainsAny(label, "[]:") {
		return nil, 0
	}

	return &FootnoteRef{Label: label}, end + 1 - i
}

func (p *inlineParser) link(i int) (Inline, int) {
	if p.src[i] != '[' || i+1 >= len(p.src) || p.src[i+1] != '[' {
		return nil, 0
	}
	end := p.indexFrom(i+2, "]]")
	if end < 0 {
		return nil, 0
	}

	target, desc := string(p.src[i+2:end]), ""
	if sep := strings.Index(target, "]["); sep >= 0 {
		target, desc = target[:sep], target[sep+2:]
	}
	if target == "" {
		return nil, 0
	}

	return &Link{Target: target, Desc: parseInlines(desc)}, end + 2 - i
}

func (p *inlineParser) timestamp(i int) (Inline, int) {
	if p.src[i] != '<' {
		return nil, 0
	}
	end := p.indexFrom(i+1, ">")
	if end < 0 {
		return nil, 0
	}

	m := timestampRE.FindStringSubmatch(string(p.src[i+1 : end]))
	if m == nil {
		return nil, 0
	}

	return &Timestamp{Date: m[1], Rest: strings.TrimSpace(m[2])}, end + 1 - i
}

func (p *inlineParser) emphasis(i int) (Inline, int) {
	marker := p.src[i]
	wrap, ok := emphasisKinds[marker]
	if !ok || !p.boundaryBefore(i) {
		return nil, 0
	}

	close := p.indexRuneFrom(i+1, marker)
	if close < 0 || !p.boundaryAfter(close) {
		return nil, 0
	}

	body := p.src[i+1 : close]
	if len(body) == 0 || isSpace(body[0]) || isSpace(body[len(body)-1]) {
		return nil, 0
	}

	width := close + 1 - i
	if wrap == nil {
		return &Code{Value: string(body)}, width
	}
	return wrap(parseInlines(string(body))), width
}

func (p *inlineParser) boundaryBefore(i int) bool {
	return i == 0 || strings.ContainsRune(preChars, p.src[i-1])
}

func (p *inlineParser) boundaryAfter(i int) bool {
	return i == len(p.src)-1 || strings.ContainsRune(postChars, p.src[i+1])
}

func (p *inlineParser) indexFrom(start int, sub string) int {
	if start > len(p.src) {
		return -1
	}
	if n := strings.Index(string(p.src[start:]), sub); n >= 0 {
		return start + len([]rune(string(p.src[start:])[:n]))
	}
	return -1
}

func (p *inlineParser) indexRuneFrom(start int, r rune) int {
	for i := start; i < len(p.src); i++ {
		if p.src[i] == r {
			return i
		}
	}
	return -1
}

func isSpace(r rune) bool {
	return r == ' ' || r == '\t'
}
