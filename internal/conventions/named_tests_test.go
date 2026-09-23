package conventions

import (
	"go/ast"
	"regexp"
	"testing"
)

// named finds a test's name written inside a comment, which is how this repository
// points at the thing that holds a claim.
var named = regexp.MustCompile(`\b(Test|Benchmark)[A-Z][A-Za-z0-9]+`)

/*
A comment that names a test names one that is there.

Pointing at the test that holds a claim is the habit this codebase is written in, and it
is worth more than the sentence it sits in: it says where to look when the claim stops
being true. It is also the first thing to rot, because renaming a test is a rename and
the comment is prose.

It had already happened. `nameMembers` explained why a Member might not be readable by
pointing at the test that proved removing one took their Items with them. Question 2
reversed exactly that: removal tombstones now, nothing cascades, and the test was turned
inside out and renamed to say what it says today. The comment was left describing the old
behaviour and naming a test nobody could open.

Its own name is not written out here, for the reason this test exists.
*/
func TestACommentThatNamesATestNamesOneThatExists(t *testing.T) {
	written, tests := goFilesIn(t)
	if len(written) < 50 || len(tests) < 30 {
		t.Fatalf("read %d files and %d test files, which is less of the tree than it was",
			len(written), len(tests))
	}

	there := make(map[string]bool)
	for _, file := range tests {
		for _, one := range file.Decls {
			if fn, ok := one.(*ast.FuncDecl); ok {
				there[fn.Name.Name] = true
			}
		}
	}
	if len(there) < 200 {
		t.Fatalf("found %d test functions, too few to be reading the suite", len(there))
	}

	for path, file := range all(written, tests) {
		for _, group := range file.Comments {
			for _, cited := range named.FindAllString(group.Text(), -1) {
				if !there[cited] {
					t.Errorf("%s points at %s, and there is no such test", path, cited)
				}
			}
		}
	}
}
