package store

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
	"modernc.org/sqlite"
)

/*
The result codes both drivers use for "somebody else has it, this may work later".

Written out rather than taken from the driver's own constants, which for SQLite means
importing the translated C library for two numbers that have been fixed since SQLite 3.
Postgres publishes its classes in the same way: class 40 is a transaction that has to be
rolled back and tried again, and 55P03 is a lock that could not be taken.
*/
const (
	sqliteBusy   = 5
	sqliteLocked = 6

	pgSerializationFailure = "40001"
	pgDeadlockDetected     = "40P01"
	pgLockNotAvailable     = "55P03"
)

/*
Busy reports whether a write failed because something else held the lock.

This is the one failure in the system that is worth trying again. SQLite takes one
writer at a time, so a second writer that outwaits `busy_timeout` is told the file is
locked, and nothing about that says the write was wrong. Postgres does not serialise the
same way but has the same shape of answer when two transactions collide.

It matters because of what the caller is told. Everything else that reaches
`internalError` means the Instance is broken, and a script or an assistant is right to
give up on it. A lock held for a moment is not that, and a caller told it was will stop
rather than retry, which turns a delay into a lost write.

The code is read rather than the message. `isUniqueViolationOn` matches text because
neither driver exposes a constraint name through database/sql, but both do expose this:
SQLite through *sqlite.Error, Postgres through the SQLSTATE on *pgconn.PgError.
*/
func Busy(err error) bool {
	var lite *sqlite.Error
	if errors.As(err, &lite) {
		// The low byte, because SQLite extends a code in the high bits: a busy snapshot
		// comes back as 517, which is SQLITE_BUSY with a reason attached.
		switch lite.Code() & 0xFF {
		case sqliteBusy, sqliteLocked:
			return true
		}
	}

	var pg *pgconn.PgError
	if errors.As(err, &pg) {
		switch pg.Code {
		case pgSerializationFailure, pgDeadlockDetected, pgLockNotAvailable:
			return true
		}
	}

	return false
}
