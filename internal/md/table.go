package md

import (
	"strings"
	"unicode/utf8"

	"github.com/go-spass/org2md/internal/org"
)

// minColumnWidth keeps every separator cell at three dashes or more, the
// familiar form and the shortest a centered column can use.
const minColumnWidth = 3

// table renders a GFM table. Markdown allows exactly one header row and
// requires the alignment row, so a table with no separator uses its first row
// as the header and any separator past the first is dropped without losing the
// rows around it.
func table(t *org.Table) string {
	rows, header := gridOf(t)
	if len(rows) == 0 {
		return ""
	}

	widths := columnWidths(rows)

	lines := make([]string, 0, len(rows)+1)
	lines = append(lines, tableLine(rows[0], widths))
	lines = append(lines, alignLine(t.Align, widths))
	for _, row := range rows[header:] {
		lines = append(lines, tableLine(row, widths))
	}

	return strings.Join(lines, "\n")
}

// gridOf renders every cell and reports how many leading rows are header rows.
func gridOf(t *org.Table) (rows [][]string, header int) {
	seenSeparator := false

	for _, r := range t.Rows {
		if r.Separator {
			if !seenSeparator {
				header = len(rows)
				seenSeparator = true
			}
			continue
		}

		row := make([]string, len(r.Cells))
		for i, cell := range r.Cells {
			row[i] = strings.ReplaceAll(inlines(cell), "|", `\|`)
		}
		rows = append(rows, row)
	}

	// With no separator, or one before any row, the first row is the header.
	if header == 0 && len(rows) > 0 {
		header = 1
	}
	return rows, header
}

func columnWidths(rows [][]string) []int {
	columns := 0
	for _, row := range rows {
		columns = max(columns, len(row))
	}

	widths := make([]int, columns)
	for i := range widths {
		widths[i] = minColumnWidth
	}
	for _, row := range rows {
		for i, cell := range row {
			widths[i] = max(widths[i], utf8.RuneCountInString(cell))
		}
	}
	return widths
}

func tableLine(row []string, widths []int) string {
	var b strings.Builder
	for i, width := range widths {
		cell := ""
		if i < len(row) {
			cell = row[i]
		}
		b.WriteString("| " + cell + strings.Repeat(" ", width-utf8.RuneCountInString(cell)) + " ")
	}
	b.WriteString("|")
	return b.String()
}

func alignLine(align []org.Alignment, widths []int) string {
	var b strings.Builder
	for i, width := range widths {
		a := org.AlignDefault
		if i < len(align) {
			a = align[i]
		}

		var cell string
		switch a {
		case org.AlignLeft:
			cell = ":" + strings.Repeat("-", width-1)
		case org.AlignRight:
			cell = strings.Repeat("-", width-1) + ":"
		case org.AlignCenter:
			cell = ":" + strings.Repeat("-", width-2) + ":"
		default:
			cell = strings.Repeat("-", width)
		}
		b.WriteString("| " + cell + " ")
	}
	b.WriteString("|")
	return b.String()
}

// footnoteDef renders a definition. Continuation lines are indented by four
// spaces, which is what keeps them inside the footnote.
func footnoteDef(f *org.FootnoteDef) string {
	var b strings.Builder
	b.WriteString("[^" + f.Label + "]:")

	for i, line := range f.Lines {
		if i == 0 {
			b.WriteString(" " + inlines(line))
			continue
		}
		b.WriteString("\n    " + inlines(line))
	}

	return b.String()
}
