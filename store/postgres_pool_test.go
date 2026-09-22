package store

import "testing"

/*
The pool is bounded, and by what was asked for.

`database/sql` opens as many connections as it needs and nothing called
SetMaxOpenConns, so a burst took whatever it took against a Postgres that allows a
hundred in total, shared with anything else on that server.

Asked of the pool itself rather than of the code that sets it: what matters is the
bound the driver is actually carrying. Skips without a Postgres, like every other test
here that needs one.
*/
func TestThePoolIsBounded(t *testing.T) {
	t.Run("the default when nothing is asked for", func(t *testing.T) {
		s := openPostgresWithConns(t, 0)
		if got := s.maxOpenConns(); got != DefaultMaxConns {
			t.Errorf("max open connections = %d, want the default of %d", got, DefaultMaxConns)
		}
	})

	t.Run("the number that was asked for", func(t *testing.T) {
		s := openPostgresWithConns(t, 3)
		if got := s.maxOpenConns(); got != 3 {
			t.Errorf("max open connections = %d, want the 3 that was asked for", got)
		}
	})
}
