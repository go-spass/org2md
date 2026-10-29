package md

import (
	"strings"

	"github.com/go-spass/org2md/internal/org"
)

// minFence is the shortest code fence Markdown accepts.
const minFence = 3

// block renders a literal block. The fence is made longer than any run of
// backticks in the content, so content that itself shows Markdown fences does
// not close the block early.
func block(b *org.Block) string {
	fence := strings.Repeat("`", max(minFence, longestBacktickRun(b.Lines)+1))

	var sb strings.Builder
	sb.WriteString(fence + b.Lang + "\n")
	for _, line := range b.Lines {
		sb.WriteString(line + "\n")
	}
	sb.WriteString(fence)

	return sb.String()
}

func longestBacktickRun(lines []string) int {
	longest := 0
	for _, line := range lines {
		run := 0
		for _, r := range line {
			if r == '`' {
				run++
				longest = max(longest, run)
				continue
			}
			run = 0
		}
	}
	return longest
}

// quote renders a blockquote. Its children are ordinary blocks, so they are
// rendered first and then prefixed line by line.
func (r *renderer) quote(q *org.Quote) string {
	lines := strings.Split(strings.TrimRight(r.blocks(q.Children), "\n"), "\n")
	for i, line := range lines {
		if line == "" {
			lines[i] = ">"
			continue
		}
		lines[i] = "> " + line
	}
	return strings.Join(lines, "\n")
}
