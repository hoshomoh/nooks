package live

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// theKinds is where the Instance declares what it can announce, and theWiring is where
// the browser decides what to do about each one.
const (
	theKinds  = "../../events/broker.go"
	theWiring = "../../../apps/web/src/lib/live-wiring.ts"
)

// declared finds the wire value of every event kind: `KindSomething Kind = "some.thing"`.
var declared = regexp.MustCompile(`Kind\w+\s+Kind\s*=\s*"([^"]+)"`)

/*
Every kind the Instance can send has a case in the browser that receives it.

The two ends are written in different languages and nothing ties them together. The
browser's dispatch has a default branch, which is right — an Instance newer than the tab
it is talking to should do something sensible rather than throw — but it means a kind
added on this side and forgotten on that one is not an error. It re-reads the Lists and
moves on, quietly doing the wrong thing.

That has happened at both ends of this already. `lists.changed` existed here, was handled
there, and was published by nothing. `member.changed` needed its own case because the
default re-reads Lists and the one thing it must re-read is the Member.

Read as text rather than imported, because a Go test cannot import a TypeScript file and
the alternative is writing the list down twice.
*/
func TestEveryKindTheInstanceSendsIsHandledByTheBrowser(t *testing.T) {
	kinds := kindsDeclared(t)
	if len(kinds) < 5 {
		t.Fatalf("found %d kinds in %s, too few to be reading it", len(kinds), theKinds)
	}

	wiring, err := os.ReadFile(filepath.FromSlash(theWiring))
	if err != nil {
		t.Fatalf("read the browser's wiring: %v", err)
	}

	for _, kind := range kinds {
		// The kind named in a case of its own, rather than merely mentioned: falling
		// into the default is exactly what this is about.
		if !strings.Contains(string(wiring), `case "`+kind+`":`) {
			t.Errorf("the Instance can send %q and %s has no case for it — it would fall "+
				"into the default and re-read the Lists", kind, theWiring)
		}
	}
}

// kindsDeclared reads the wire values out of the broker.
func kindsDeclared(t *testing.T) []string {
	t.Helper()

	source, err := os.ReadFile(filepath.FromSlash(theKinds))
	if err != nil {
		t.Fatalf("read the kinds: %v", err)
	}

	found := make([]string, 0, 8)
	for _, match := range declared.FindAllSubmatch(source, -1) {
		found = append(found, string(match[1]))
	}
	return found
}
