package store

import (
	"database/sql"
	"os"
	"testing"
)

/*
Busy recognises what the drivers actually return.

Written against a real lock rather than a constructed error, because a constructed one
cannot be had: `*sqlite.Error` keeps its code unexported and the driver is the only
thing that builds one. That is a better test anyway. The thing worth knowing is not
whether a switch matches a number somebody typed twice, it is whether the number is the
one the driver sends when a write is held up.

Both halves hold a write and then make a second connection try one without waiting, so
the failure arrives immediately rather than after a timeout somebody has to sit through.
*/
func TestBusyKnowsAHeldLockOnBothDrivers(t *testing.T) {
	t.Run("sqlite", func(t *testing.T) {
		path := sqliteTemplate.Fresh(t)

		held, err := sql.Open("sqlite", "file:"+path+"?_pragma=busy_timeout(0)")
		if err != nil {
			t.Fatalf("open the holder: %v", err)
		}
		t.Cleanup(func() { _ = held.Close() })

		tx, err := held.BeginTx(t.Context(), nil)
		if err != nil {
			t.Fatalf("begin: %v", err)
		}
		t.Cleanup(func() { _ = tx.Rollback() })

		// Any write takes the one write lock, and creating a table takes it while
		// knowing nothing about the schema, so this does not break when a column does.
		if _, err := tx.ExecContext(t.Context(), "CREATE TABLE probe_held (x)"); err != nil {
			t.Fatalf("take the write lock: %v", err)
		}

		other, err := sql.Open("sqlite", "file:"+path+"?_pragma=busy_timeout(0)")
		if err != nil {
			t.Fatalf("open the second connection: %v", err)
		}
		t.Cleanup(func() { _ = other.Close() })

		_, err = other.ExecContext(t.Context(), "CREATE TABLE probe_waiting (x)")
		if err == nil {
			t.Fatal("the second write succeeded while the lock was held")
		}
		if !Busy(err) {
			t.Errorf("Busy says no to what sqlite returns for a held lock: %v", err)
		}
	})

	t.Run("postgres", func(t *testing.T) {
		s := openPostgresForTest(t).(*sqlStore)

		holder, err := s.db.BeginTx(t.Context(), nil)
		if err != nil {
			t.Fatalf("begin the holder: %v", err)
		}
		t.Cleanup(func() { _ = holder.Rollback() })
		if _, err := holder.ExecContext(t.Context(),
			"LOCK TABLE member IN ACCESS EXCLUSIVE MODE"); err != nil {
			t.Fatalf("take the table lock: %v", err)
		}

		waiter, err := s.db.BeginTx(t.Context(), nil)
		if err != nil {
			t.Fatalf("begin the waiter: %v", err)
		}
		t.Cleanup(func() { _ = waiter.Rollback() })

		// NOWAIT, because the ordinary form waits for the holder and would leave the
		// test waiting on itself.
		_, err = waiter.ExecContext(t.Context(),
			"LOCK TABLE member IN ACCESS EXCLUSIVE MODE NOWAIT")
		if err == nil {
			t.Fatal("the second lock was taken while the first was held")
		}
		if !Busy(err) {
			t.Errorf("Busy says no to what postgres returns for a held lock: %v", err)
		}
	})
}

// A failure that is not contention is not contention.
func TestBusyLeavesOtherFailuresAlone(t *testing.T) {
	s := openSQLiteForTest(t).(*sqlStore)
	_, err := s.db.ExecContext(t.Context(), "SELECT * FROM no_such_table")
	if err == nil {
		t.Fatal("reading a table that does not exist succeeded")
	}
	if Busy(err) {
		t.Errorf("Busy says a missing table is contention: %v", err)
	}
	if Busy(nil) {
		t.Error("Busy says nil is contention")
	}
	if Busy(os.ErrNotExist) {
		t.Error("Busy says an unrelated error is contention")
	}
}
