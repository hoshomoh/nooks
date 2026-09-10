package note

import "testing"

func TestPreviewOfAnEmptyNote(t *testing.T) {
	for _, markdown := range []string{"", "   ", "\n\n\n"} {
		preview := PreviewOf(markdown)
		if !preview.Empty() {
			t.Errorf("PreviewOf(%q) = %+v, want empty", markdown, preview)
		}
	}
}

func TestPreviewIsTheNotesOwnFirstLine(t *testing.T) {
	markdown := "Not the supermarket ones — the stall at the Saturday market.\n\nSecond line.\nThird."

	preview := PreviewOf(markdown)

	if preview.FirstLine != "Not the supermarket ones — the stall at the Saturday market." {
		t.Errorf("FirstLine = %q, want the Note's own first line", preview.FirstLine)
	}
	if preview.RemainingLines != 2 {
		t.Errorf("RemainingLines = %d, want 2", preview.RemainingLines)
	}
}

// Blank lines are spacing, not content, so they do not count towards "+N lines".
func TestPreviewIgnoresBlankLines(t *testing.T) {
	preview := PreviewOf("Where\n\n\n\nSaturday market")
	if preview.RemainingLines != 1 {
		t.Errorf("RemainingLines = %d, want 1", preview.RemainingLines)
	}
}

// A row must never show a Member their own markdown back.
func TestPreviewStripsTheShorthand(t *testing.T) {
	cases := map[string]string{
		"### Where":                   "Where",
		"# Heading":                   "Heading",
		"- [ ] Ethiopian, whole bean": "Ethiopian, whole bean",
		"- [x] Already got it":        "Already got it",
		"> They pack up around two.":  "They pack up around two.",
		"- A plain bullet":            "A plain bullet",
		"```":                         "",
		"Just a sentence.":            "Just a sentence.",
	}
	for markdown, want := range cases {
		if got := PreviewOf(markdown).FirstLine; got != want {
			t.Errorf("PreviewOf(%q).FirstLine = %q, want %q", markdown, got, want)
		}
	}
}

// A hash inside a sentence is not a heading.
func TestPreviewLeavesTextAlone(t *testing.T) {
	const line = "Ask for #4 at the counter"
	if got := PreviewOf(line).FirstLine; got != line {
		t.Errorf("FirstLine = %q, want it untouched", got)
	}
}

// "- [ ] " has to be matched before "- ", or a checklist line keeps its box.
func TestPreviewPrefersTheLongerMarker(t *testing.T) {
	if got := PreviewOf("- [ ] Whole bean").FirstLine; got != "Whole bean" {
		t.Errorf("FirstLine = %q, want the checklist marker stripped whole", got)
	}
}
