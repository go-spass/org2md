package org

import (
	"regexp"
	"strings"
)

var (
	tableRowRE = regexp.MustCompile(`^[ \t]*\|(.*)$`)

	// An alignment cookie is <l>, <r> or <c>, optionally with a column width.
	cookieRE = regexp.MustCompile(`^<([lrc])?(\d+)?>$`)

	// A footnote definition opens a line: [fn:label] followed by its text.
	footnoteDefRE = regexp.MustCompile(`^\[fn:([^\]:]+)\][ \t]*(.*)$`)
)

func isTableLine(line string) bool {
	return tableRowRE.MatchString(line)
}

// tableLine adds one row to the open table, starting the table if needed.
func (p *parser) tableLine(line string) {
	if p.table == nil {
		p.endList()
		p.flush()
		p.table = &Table{}
		p.add(p.table)
	}

	body := tableRowRE.FindStringSubmatch(line)[1]
	if isSeparatorRow(body) {
		p.table.Rows = append(p.table.Rows, &TableRow{Separator: true})
		return
	}

	raw := splitCells(body)
	if align, ok := cookieRow(raw); ok {
		p.table.Align = align
		return
	}

	row := &TableRow{Cells: make([][]Inline, len(raw))}
	for i, cell := range raw {
		if cell == "" {
			continue
		}
		row.Cells[i] = parseInlines(cell)
	}
	p.table.Rows = append(p.table.Rows, row)
}

func (p *parser) endTable() {
	p.table = nil
}

// isSeparatorRow reports whether a row is a horizontal rule: dashes, pluses and
// pipes only, with at least one dash.
func isSeparatorRow(body string) bool {
	if !strings.ContainsRune(body, '-') {
		return false
	}
	return strings.IndexFunc(body, func(r rune) bool {
		return r != '-' && r != '+' && r != '|' && r != ' ' && r != '\t'
	}) < 0
}

// splitCells divides a row body into trimmed cells. The body is what follows
// the leading pipe, so a trailing pipe leaves an empty element to discard.
func splitCells(body string) []string {
	parts := strings.Split(body, "|")
	if len(parts) > 1 && strings.TrimSpace(parts[len(parts)-1]) == "" {
		parts = parts[:len(parts)-1]
	}

	cells := make([]string, len(parts))
	for i, part := range parts {
		cells[i] = strings.TrimSpace(part)
	}
	return cells
}

// cookieRow reports the alignment a row of cookies declares. Every non-empty
// cell has to be a cookie and there has to be at least one, so neither a row of
// numbers nor an entirely empty row is mistaken for one.
func cookieRow(cells []string) ([]Alignment, bool) {
	align := make([]Alignment, len(cells))
	found := false

	for i, cell := range cells {
		if cell == "" {
			continue
		}
		m := cookieRE.FindStringSubmatch(cell)
		if m == nil {
			return nil, false
		}
		found = true

		switch m[1] {
		case "l":
			align[i] = AlignLeft
		case "r":
			align[i] = AlignRight
		case "c":
			align[i] = AlignCenter
		}
	}

	return align, found
}
