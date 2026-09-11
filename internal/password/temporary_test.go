package password

import (
	"strings"
	"testing"
)

// A temporary password has to satisfy the same rule as any other, or a Member would be
// handed one they are told is not allowed.
func TestATemporaryPasswordIsLongEnough(t *testing.T) {
	for range 200 {
		if err := Validate(NewTemporary()); err != nil {
			t.Fatalf("Validate(%q) = %v", NewTemporary(), err)
		}
	}
}

// It is read out loud, so it has to survive being spoken: words, and nothing that has
// to be described as "capital" or "underscore".
func TestATemporaryPasswordCanBeReadAloud(t *testing.T) {
	for range 200 {
		generated := NewTemporary()
		for _, letter := range generated {
			if letter != '-' && (letter < 'a' || letter > 'z') {
				t.Fatalf("%q contains %q, which has to be spelled out", generated, letter)
			}
		}
		if got := strings.Count(generated, temporarySeparator); got != temporaryWords-1 {
			t.Fatalf("%q has %d separators, want %d", generated, got, temporaryWords-1)
		}
	}
}

func TestTwoTemporaryPasswordsDiffer(t *testing.T) {
	seen := map[string]bool{}
	for range 100 {
		seen[NewTemporary()] = true
	}
	if len(seen) < 90 {
		t.Errorf("got %d distinct passwords out of 100, want them to vary", len(seen))
	}
}

// Every word has to be worth reading out: nothing ambiguous, nothing long.
func TestTheAlphabetIsSpeakable(t *testing.T) {
	seen := map[string]bool{}
	for _, word := range alphabet {
		if seen[word] {
			t.Errorf("%q appears twice, which narrows the alphabet silently", word)
		}
		seen[word] = true
		if len(word) < 5 || len(word) > 8 {
			t.Errorf("%q is %d letters; keep them short enough to say and long enough to hear", word, len(word))
		}
	}
}
