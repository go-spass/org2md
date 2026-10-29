package md

import (
	"maps"
	"regexp"
	"slices"
	"strings"

	"github.com/go-spass/org2md/internal/org"
)

// frontmatterKeywords are the org keywords with a frontmatter equivalent. Every
// other keyword is reported and dropped.
var frontmatterKeywords = map[string]bool{
	"TITLE":    true,
	"DATE":     true,
	"AUTHOR":   true,
	"FILETAGS": true,
}

var (
	// A scalar is safe unquoted when it starts with an alphanumeric, holds
	// nothing YAML treats specially, and is not padded with whitespace.
	safeScalarRE = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9 _./()-]*$`)
	safeKeyRE    = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]*$`)
	numberRE     = regexp.MustCompile(`^-?\d+(\.\d+)?$`)
	isoDateRE    = regexp.MustCompile(`\d{4}-\d{2}-\d{2}`)
)

// yamlWords are the scalars YAML would read as something other than a string.
var yamlWords = map[string]bool{
	"true": true, "false": true, "yes": true, "no": true,
	"on": true, "off": true, "null": true, "none": true, "~": true,
}

// frontmatter builds the YAML block from the document's keywords and
// properties. It returns an empty string when nothing maps to frontmatter, so
// a note without metadata gets no empty block.
func (r *renderer) frontmatter(doc *org.Document) string {
	var lines []string
	emitted := map[string]bool{}

	add := func(key, value string) {
		lines = append(lines, key+": "+value)
		emitted[key] = true
	}

	if v := doc.Meta["TITLE"]; v != "" {
		add("title", scalar(v))
	}
	if v := doc.Meta["DATE"]; v != "" {
		add("date", date(v))
	}
	if v := doc.Meta["AUTHOR"]; v != "" {
		add("author", scalar(v))
	}
	if tags := tagList(doc.Meta["FILETAGS"]); len(tags) > 0 {
		lines = append(lines, "tags:")
		for _, tag := range tags {
			lines = append(lines, "  - "+scalar(tag))
		}
		emitted["tags"] = true
	}

	for _, name := range slices.Sorted(maps.Keys(doc.Meta)) {
		if !frontmatterKeywords[name] {
			r.warnf("dropped #+%s: no frontmatter equivalent", name)
		}
	}

	for _, name := range slices.Sorted(maps.Keys(doc.Properties)) {
		if lower := strings.ToLower(name); emitted[lower] {
			r.warnf("dropped property %s: the %s key is already set", name, lower)
			continue
		}
		add(key(name), scalar(doc.Properties[name]))
	}

	if len(lines) == 0 {
		return ""
	}
	return "---\n" + strings.Join(lines, "\n") + "\n---\n"
}

// tagList reads org's #+FILETAGS value, which separates tags with colons and
// sometimes with spaces.
func tagList(value string) []string {
	return strings.FieldsFunc(value, func(r rune) bool {
		return r == ':' || r == ' ' || r == '\t'
	})
}

// date keeps a frontmatter date machine-readable: an ISO date anywhere in the
// value becomes the whole value, so an org timestamp loses its brackets and
// day name. Anything else is left as a string.
func date(value string) string {
	if iso := isoDateRE.FindString(value); iso != "" {
		return iso
	}
	return scalar(value)
}

func scalar(value string) string {
	switch {
	case value == "",
		value != strings.TrimSpace(value),
		!safeScalarRE.MatchString(value),
		numberRE.MatchString(value),
		yamlWords[strings.ToLower(value)]:
		return quoted(value)
	}
	return value
}

func quoted(value string) string {
	var b strings.Builder
	b.WriteByte('"')
	for _, r := range value {
		switch r {
		case '\\':
			b.WriteString(`\\`)
		case '"':
			b.WriteString(`\"`)
		default:
			b.WriteRune(r)
		}
	}
	b.WriteByte('"')
	return b.String()
}

func key(name string) string {
	k := strings.ToLower(name)
	if safeKeyRE.MatchString(k) {
		return k
	}
	return quoted(k)
}
