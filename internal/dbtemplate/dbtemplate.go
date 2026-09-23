/*
Package dbtemplate hands tests copies of a database that was migrated once.

Opening a fresh SQLite file runs all nineteen migrations, which costs about two hundred
milliseconds; opening one that is already migrated costs thirteen. Nearly every test in
the tree opens one, so the suite was paying the two hundred several hundred times over.
That is most of why the race detector, which CI runs and most machines cannot, took two
packages past the ten minutes `go test` allows.

Measured before it was written, and the first guess was wrong: turning SQLite's
durability off made a fresh open no faster at all, so the cost is the migration SQL
itself rather than waiting for the disk. Running them once is the only thing that helps.

The template is kept as bytes rather than as a file, which is what makes the lifetime
simple: it lives as long as the test binary and needs no cleaning up, while each copy
lands in the test's own directory and goes when the test does.
*/
package dbtemplate

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

// Forever is the context the template is built with: it is built once, before any test
// has started, so there is no test whose ending should cancel it.
var Forever = context.Background()

// Copies hands out copies of one database. The zero value is not usable: Build says how
// to make the first one.
type Copies struct {
	// Build makes a migrated database at a path, and closes whatever it opened. It is
	// supplied by the caller because the packages that need this cannot all import the
	// one that knows how to open a database: the store's own tests are inside it.
	Build func(path string) error

	once     sync.Once
	migrated []byte
	err      error
}

// Fresh writes a copy for one test and returns its path. The copy is that test's own:
// what it writes is seen by nothing else, which is what a fresh database means here.
func (c *Copies) Fresh(tb testing.TB) string {
	tb.Helper()
	return c.FreshIn(tb, tb.TempDir())
}

// FreshIn is Fresh, in a directory the test has already chosen. Some tests hand that
// directory to something else as well, a backup handler being told where the data is.
func (c *Copies) FreshIn(tb testing.TB, dir string) string {
	tb.Helper()

	c.once.Do(c.build)
	if c.err != nil {
		tb.Fatalf("build the database template: %v", c.err)
	}

	path := filepath.Join(dir, "nooks.db")
	if err := os.WriteFile(path, c.migrated, 0o600); err != nil {
		tb.Fatalf("copy the database template: %v", err)
	}
	return path
}

// build makes the template once and keeps its bytes.
//
// The file it was built in goes straight away. What a closed SQLite database leaves
// behind is one file: the write-ahead log is checkpointed into it and removed on the
// last connection closing, so the bytes read here are the whole database.
func (c *Copies) build() {
	dir, err := os.MkdirTemp("", "nooks-template-")
	if err != nil {
		c.err = err
		return
	}
	defer func() { _ = os.RemoveAll(dir) }()

	path := filepath.Join(dir, "nooks.db")
	if err := c.Build(path); err != nil {
		c.err = err
		return
	}
	c.migrated, c.err = os.ReadFile(path)
}
