package profile

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

// thePage is where an operator is told what they can set.
const thePage = "../../apps/website/content/docs/configure.mdx"

// named finds every environment variable mentioned, in source or in prose.
var named = regexp.MustCompile(`NOOKS_[A-Z_]+`)

/*
What an operator can set and what they are told they can set are the same list.

configure.mdx is the whole contract for running this: somebody reading it is deciding
how their household's Instance behaves, and they have no other way to find out. A
variable added here and not written down is one nobody will use; one written down and
not read is worse, because they will set it and believe it did something.

Only the names. Whether the prose around them is true is not something a test can say —
`deploy/reverse-proxy.mdx` spent however long telling people to pass `X-Forwarded-Proto`
so that cookies would be marked Secure, which this server has never read.
*/
func TestEveryVariableIsDocumented(t *testing.T) {
	read := mentionedIn(t, "profile.go")
	if len(read) < 5 {
		t.Fatalf("found %d variables in profile.go, too few to be reading it", len(read))
	}
	written := mentionedIn(t, filepath.FromSlash(thePage))

	for _, name := range read {
		if !contains(written, name) {
			t.Errorf("%s is read and %s does not mention it", name, thePage)
		}
	}
	for _, name := range written {
		if !contains(read, name) {
			t.Errorf("%s says %s can be set and nothing reads it", thePage, name)
		}
	}
}

// mentionedIn is every variable named in one file, without repeats.
func mentionedIn(t *testing.T, path string) []string {
	t.Helper()

	source, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}

	seen := map[string]bool{}
	found := make([]string, 0, 8)
	for _, name := range named.FindAllString(string(source), -1) {
		// The prefix on its own, where the code builds a name rather than naming one.
		if name == "NOOKS_" || seen[name] {
			continue
		}
		seen[name] = true
		found = append(found, name)
	}
	sort.Strings(found)
	return found
}

func contains(names []string, name string) bool {
	return sort.SearchStrings(names, name) < len(names) &&
		strings.EqualFold(names[sort.SearchStrings(names, name)], name)
}
