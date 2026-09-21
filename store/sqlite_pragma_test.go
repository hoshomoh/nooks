package store

import (
	"path/filepath"
	"sync"
	"testing"
)

/*
The settings the store's behaviour rests on are on every connection, not just the first.

They are set in the DSN, and `database/sql` opens as many connections as it likes and
closes them when it likes. If they applied only to the first one, the rest would behave
differently and nothing would say so — which is the sort of thing that shows up as one
request in twenty doing something odd.

Each of the three is load-bearing:

  - foreign_keys, because the schema leans on ON DELETE CASCADE and SET NULL for what
    happens when a Member is removed. Off, the cascades silently do not happen and rows
    are left pointing at nobody.
  - journal_mode, because WAL is what keeps a read from blocking the one writer. Off,
    reading a List would queue behind somebody ticking.
  - busy_timeout, because it is the whole of how a writer that finds the lock held waits
    rather than failing. Zero, and a second writer loses its write immediately instead of
    after five seconds — see BenchmarkWritersOnOneList for what that costs when the wait
    is exhausted rather than absent.
*/
func TestEveryConnectionCarriesThePragmas(t *testing.T) {
	s, err := OpenSQLite(t.Context(), filepath.Join(t.TempDir(), "nooks.db"))
	if err != nil {
		t.Fatalf("OpenSQLite: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })
	db := s.(*sqlStore).db

	// Asked from several goroutines at once so that more than one connection is in use,
	// since one connection answering correctly says nothing about the others.
	const askers = 8
	answers := make(chan [3]string, askers)

	var wg sync.WaitGroup
	for range askers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			var busy, keys, journal string
			row := db.QueryRowContext(t.Context(), `SELECT
				(SELECT * FROM pragma_busy_timeout()),
				(SELECT * FROM pragma_foreign_keys()),
				(SELECT * FROM pragma_journal_mode())`)
			if err := row.Scan(&busy, &keys, &journal); err != nil {
				answers <- [3]string{"unreadable: " + err.Error(), "", ""}
				return
			}
			answers <- [3]string{busy, keys, journal}
		}()
	}
	wg.Wait()
	close(answers)

	seen := 0
	for got := range answers {
		seen++
		if got[0] != "5000" {
			t.Errorf("busy_timeout = %q, want 5000: a writer that finds the lock held "+
				"must wait rather than lose the write", got[0])
		}
		if got[1] != "1" {
			t.Errorf("foreign_keys = %q, want 1: the cascades the schema relies on do "+
				"not happen without it", got[1])
		}
		if got[2] != "wal" {
			t.Errorf("journal_mode = %q, want wal: reads queue behind the writer "+
				"without it", got[2])
		}
	}
	if seen != askers {
		t.Fatalf("heard from %d connections, want %d", seen, askers)
	}
}
