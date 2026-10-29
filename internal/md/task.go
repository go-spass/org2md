package md

import (
	"strings"

	"github.com/go-spass/org2md/internal/org"
)

// headingTags renders a headline's metadata as Obsidian tags. A status keyword
// is lowercased, since #todo reads better than #TODO and the original casing
// carries no information; the author's own tags are left as written.
func headingTags(h *org.Heading) []string {
	var tags []string

	if h.Todo != "" {
		tags = append(tags, "#"+strings.ToLower(h.Todo))
	}
	if h.Priority != "" {
		tags = append(tags, "#priority/"+strings.ToLower(h.Priority))
	}
	for _, tag := range h.Tags {
		tags = append(tags, "#"+tag)
	}

	return tags
}

// planning renders the scheduled and deadline dates as the first lines of the
// heading's body. A closed date is dropped: it duplicates the done keyword and
// records when rather than what.
func planning(h *org.Heading) string {
	var lines []string

	if h.Scheduled != nil {
		lines = append(lines, "**Scheduled:** "+timestamp(h.Scheduled))
	}
	if h.Deadline != nil {
		lines = append(lines, "**Deadline:** "+timestamp(h.Deadline))
	}

	return strings.Join(lines, "\n")
}
