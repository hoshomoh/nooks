// Package note handles the document attached to an Item.
//
// A Note is markdown. This package does the two things Nooks needs from it that are not
// rendering: summarising it for a row, and reporting what is in it.
package note

import (
	"strings"
)

// Preview is what a List row shows of a Note.
//
// DESIGN.md §13 is explicit that Nooks never summarises a Member: the row shows the
// Note's own first line and a count of what is left, and nothing is generated on their
// behalf.
type Preview struct {
	// FirstLine is the Note's first non-empty line, with any markdown marker removed.
	FirstLine string
	// RemainingLines is how many further non-empty lines there are.
	RemainingLines int
}

// Empty reports whether there is no Note at all.
func (p Preview) Empty() bool { return p.FirstLine == "" && p.RemainingLines == 0 }

// PreviewOf reads a Note's first line and counts the rest.
func PreviewOf(markdown string) Preview {
	lines := meaningfulLines(markdown)
	if len(lines) == 0 {
		return Preview{}
	}
	return Preview{
		FirstLine:      stripMarker(lines[0]),
		RemainingLines: len(lines) - 1,
	}
}

// markers are the shorthands a line can begin with, longest first so that "- [ ] " is
// matched before "- ".
var markers = []string{
	"- [ ] ", "- [x] ", "- [X] ",
	"###### ", "##### ", "#### ", "### ", "## ", "# ",
	"> ", "- ", "* ",
}

// meaningfulLines is every line with something on it.
func meaningfulLines(markdown string) []string {
	var lines []string
	for _, line := range strings.Split(markdown, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			lines = append(lines, trimmed)
		}
	}
	return lines
}

// stripMarker removes the shorthand from the front of a line.
//
// The marker is how the line was typed, not what it says: a row showing "### Where"
// would be showing the Member their own syntax back, which DESIGN.md §10 rules out.
func stripMarker(line string) string {
	for _, marker := range markers {
		if strings.HasPrefix(line, marker) {
			return strings.TrimSpace(strings.TrimPrefix(line, marker))
		}
	}
	// A fenced code block opens with ``` and says nothing on its own.
	if strings.HasPrefix(line, "```") {
		return strings.TrimSpace(strings.TrimPrefix(line, "```"))
	}
	return line
}
