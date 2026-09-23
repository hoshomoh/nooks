package conventions

import (
	"path/filepath"
	"strings"
	"testing"
)

// testOnly are the packages that exist for the suite and have no business in a running
// Instance. They import `testing`, which registers flags and pulls the framework in, so
// anything shipped importing one of these would carry it into the binary people run.
var testOnly = []string{
	"internal/dbtemplate",
	"store/storetest",
}

/*
Nothing a Member runs imports a package written for the tests.

An import like this compiles, passes every other check, and is only visible in what the
binary carries. The two packages named above were added to make the suite stop migrating
a database several hundred times over, which is worth having and is worth fencing: the
moment one of them is imported by something real, `testing` is in the Instance.
*/
func TestNothingRealImportsATestOnlyPackage(t *testing.T) {
	files, tests := goFilesIn(t)
	if len(files) < 50 {
		t.Fatalf("read %d files, so this is looking at less of the tree than it did", len(files))
	}

	for path, file := range files {
		// One of these may lean on another: storetest is how a test gets a database and
		// dbtemplate is what makes that cheap. Neither is shipped.
		if writtenForTests(path) {
			continue
		}
		for _, imported := range file.Imports {
			for _, only := range testOnly {
				if strings.Contains(imported.Path.Value, only) {
					t.Errorf("%s imports %s, which exists for the tests alone", path, only)
				}
			}
		}
	}

	// A fence around nothing is a fence that has stopped being read: if these packages
	// were renamed away, this would pass while guarding no one.
	using := 0
	for _, file := range tests {
		for _, imported := range file.Imports {
			for _, only := range testOnly {
				if strings.Contains(imported.Path.Value, only) {
					using++
				}
			}
		}
	}
	if using < 10 {
		t.Errorf("%d test files import one of %v, which is too few to be the packages "+
			"every test gets its database from", using, testOnly)
	}
}

// writtenForTests reports whether a file is inside one of the test-only packages.
func writtenForTests(path string) bool {
	for _, only := range testOnly {
		if strings.Contains(filepath.ToSlash(path), only) {
			return true
		}
	}
	return false
}
