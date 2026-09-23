/*
Package storetest gives a test a database of its own, without paying to migrate one.

One migrated database is built the first time a test binary asks, and every test after
that gets a copy. `dbtemplate` says what that is worth and how it was measured.

Under store/ rather than internal/, because the dependencies point inwards and internal/
knows nothing of store. This knows a great deal of it, which is the point: it is the
store's own scaffolding, the way httptest belongs to http. A guard said so before a
person did.

The store package's own tests cannot use this, because this imports the package they are
inside. They wire the same thing themselves, which is the one place it is written twice.
*/
package storetest

import (
	"testing"

	"github.com/hoshomoh/nooks/internal/dbtemplate"
	"github.com/hoshomoh/nooks/store"
)

var databases = dbtemplate.Copies{
	Build: func(path string) error {
		s, err := store.OpenSQLite(dbtemplate.Forever, path)
		if err != nil {
			return err
		}
		return s.Close()
	},
}

// Fresh opens a database of this test's own, closed when the test ends.
func Fresh(tb testing.TB) store.Store {
	tb.Helper()
	return open(tb, databases.Fresh(tb))
}

// FreshIn is Fresh, in a directory the test has already chosen, for a test that hands
// that directory to something else as well.
func FreshIn(tb testing.TB, dir string) store.Store {
	tb.Helper()
	return open(tb, databases.FreshIn(tb, dir))
}

func open(tb testing.TB, path string) store.Store {
	tb.Helper()
	s, err := store.OpenSQLite(tb.Context(), path)
	if err != nil {
		tb.Fatalf("OpenSQLite: %v", err)
	}
	tb.Cleanup(func() { _ = s.Close() })
	return s
}
