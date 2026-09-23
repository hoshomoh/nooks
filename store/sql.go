package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/dialect/sqlitedialect"
	"github.com/uptrace/bun/schema"
)

// Setting keys. Instance configuration is a handful of rows rather than a table of one
// row, so adding a setting never needs a migration.
const (
	settingInstanceName     = "instance.name"
	settingPublicSignup     = "instance.public_signup"
	settingDefaultLocale    = "instance.default_locale"
	settingSetupCompletedAt = "instance.setup_completed_at"
	settingPublicListUID    = "instance.public_list_uid"
	settingPublicShowNames  = "instance.public_show_names"
	settingPublicShowMeta   = "instance.public_show_meta"
	settingPublicAllowJoin  = "instance.public_allow_join"
	settingSetupClaimedAt   = "instance.setup_claimed_at"
)

/*
byName orders the way a reader would, rather than the way bytes do.

Plain "name ASC" compares character codes, and every capital letter sorts below every
lowercase one: "Groceries" and "Flat jobs" come out above "avocados" and "bike parts"
in a block, which is not a household's idea of alphabetical. Comparing the lowercased
name is what puts them in one sequence.

Accented names are still ordered by code, so "Éclairs" lands after "Zebra" rather than
between "bike parts" and "Flat jobs". Putting that right means language-aware collation
and a decision about whose language, which this is not.
*/
// OrderExpr, not Order: Bun quotes Order's argument as a column name, which would make
// this the identifier "lower(name)" rather than a call.
// byName is how anything a Member picks from is ordered: by name, and then by which
// came first.
//
// The second half is not decoration. Two Lists can be called the same thing, and two
// Groups can too, and an order that stops at the name leaves whichever the engine
// happens to return first, which is not the same answer twice. A sidebar whose rows
// swap places between loads is the same fault the broker had when presence came back in
// map order.
const byName = "lower(name) ASC, id ASC"

// sqlStore implements Store over Bun.
//
// Bun carries the dialect, so a query is written once and runs on both drivers. The
// only per-driver knowledge left in this package is which dialect to construct and
// which migration directory to read.
type sqlStore struct {
	db *bun.DB
	// name identifies the driver in errors and picks its migration directory.
	name string
}

func (s *sqlStore) InstanceSettings(ctx context.Context) (InstanceSettings, error) {
	var rows []settingModel
	if err := s.db.NewSelect().Model(&rows).Scan(ctx); err != nil {
		return InstanceSettings{}, fmt.Errorf("read instance settings: %w", err)
	}

	values := make(map[string]string, len(rows))
	for _, row := range rows {
		values[row.Key] = row.Value
	}
	return settingsFromValues(values)
}

// SaveInstanceSettings writes the Instance's own configuration in full.
func (s *sqlStore) SaveInstanceSettings(ctx context.Context, settings InstanceSettings) error {
	values := valuesFromSettings(settings)
	rows := make([]settingModel, 0, len(values))
	for key, value := range values {
		rows = append(rows, settingModel{Key: key, Value: value})
	}

	_, err := s.db.NewInsert().
		Model(&rows).
		On("CONFLICT (key) DO UPDATE").
		Set("value = EXCLUDED.value").
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("write instance settings: %w", err)
	}
	return nil
}

func (s *sqlStore) Close() error {
	if err := s.db.Close(); err != nil {
		return fmt.Errorf("close %s store: %w", s.name, err)
	}
	return nil
}

/*
ClaimFirstRun takes first run, and reports whether this caller is the one who got it.

CompleteSetup used to count Members, find none, and only then hash a password. Hashing
is bcrypt and deliberately slow, so a couple of hundred milliseconds sat between the
check and the insert, and two requests arriving inside it both passed. With different
emails both became Admins, and losing was silent: an owner who is told the Instance is
already set up redeploys, where an owner who completes setup has no reason to look at
the members list.

It matters at the one moment an Instance is defenceless. The setup page answers anybody
until it is used, so something scanning the internet that reaches a fresh deployment in
that window becomes an Admin of it.

An insert that cannot collide, because `key` is the settings table's primary key: the
second one affects no rows and is told so, on SQLite and Postgres alike. Re-counting
inside a transaction would have been enough on SQLite, which serialises writers, and not
on Postgres under READ COMMITTED, where neither transaction sees the other's row.

Claimed before the hash rather than after, so the slow part happens once the race is
already over.
*/
func (s *sqlStore) ClaimFirstRun(ctx context.Context, at time.Time) (bool, error) {
	result, err := s.db.NewInsert().
		Model(&settingModel{Key: settingSetupClaimedAt, Value: formatTime(at)}).
		On("CONFLICT (key) DO NOTHING").
		Exec(ctx)
	if err != nil {
		return false, fmt.Errorf("claim first run: %w", err)
	}

	claimed, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("read whether first run was claimed: %w", err)
	}
	return claimed == 1, nil
}

// settingsFromValues turns stored rows into InstanceSettings. A missing key is not an
// error: it means first run has not set it yet.
func settingsFromValues(values map[string]string) (InstanceSettings, error) {
	settings := InstanceSettings{
		Name:          values[settingInstanceName],
		PublicSignup:  values[settingPublicSignup] == "true",
		DefaultLocale: values[settingDefaultLocale],
		Public: PublicList{
			ListUID:   values[settingPublicListUID],
			ShowNames: values[settingPublicShowNames] == "true",
			ShowMeta:  values[settingPublicShowMeta] == "true",
			AllowJoin: values[settingPublicAllowJoin] == "true",
		},
	}

	completedAt, err := parseTime(values[settingSetupCompletedAt])
	if err != nil {
		return InstanceSettings{}, fmt.Errorf("parse %s: %w", settingSetupCompletedAt, err)
	}
	settings.SetupCompletedAt = completedAt
	return settings, nil
}

// valuesFromSettings is the inverse of settingsFromValues.
func valuesFromSettings(settings InstanceSettings) map[string]string {
	return map[string]string{
		settingInstanceName:     settings.Name,
		settingPublicSignup:     strconv.FormatBool(settings.PublicSignup),
		settingDefaultLocale:    settings.DefaultLocale,
		settingSetupCompletedAt: formatTime(settings.SetupCompletedAt),
		settingPublicListUID:    settings.Public.ListUID,
		settingPublicShowNames:  strconv.FormatBool(settings.Public.ShowNames),
		settingPublicShowMeta:   strconv.FormatBool(settings.Public.ShowMeta),
		settingPublicAllowJoin:  strconv.FormatBool(settings.Public.AllowJoin),
	}
}

// open wraps a database handle in Bun, verifies it, and brings the schema up to date.
func open(ctx context.Context, db *sql.DB, dialect schema.Dialect, name string) (Store, error) {
	if db == nil {
		return nil, errors.New("store: database handle is required")
	}
	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("reach %s database: %w", name, err)
	}

	bunDB := bun.NewDB(db, dialect)
	if err := migrate(ctx, bunDB, name); err != nil {
		return nil, err
	}
	store := &sqlStore{db: bunDB, name: name}
	if err := store.Analyse(ctx); err != nil {
		return nil, err
	}
	return store, nil
}

/*
Analyse gives the query planner something to plan with.

Without statistics SQLite guesses which index to use, and on a large Instance it guesses
the one that matches the filter over the one that matches the order — which means finding
every List a Member can reach and sorting the lot to hand back twenty-five. Measured on a
hundred thousand Lists that is 945ms; with statistics it is 2ms.

analysis_limit caps how much of each index is sampled, which is what keeps a full ANALYZE
cheap on a big file. It is a setting of one connection rather than of the database, so
both statements go down the same one: run through the pool they can land on different
connections, and the ANALYZE then reads every index end to end.

Postgres keeps its own statistics, so there is nothing to do there.
*/
func (s *sqlStore) Analyse(ctx context.Context) error {
	if s.name != "sqlite" {
		return nil
	}

	conn, err := s.db.Conn(ctx)
	if err != nil {
		return fmt.Errorf("analyse sqlite: %w", err)
	}
	defer func() { _ = conn.Close() }()

	for _, statement := range []string{"PRAGMA analysis_limit = 400", "ANALYZE"} {
		if _, err := conn.ExecContext(ctx, statement); err != nil {
			return fmt.Errorf("analyse sqlite: %w", err)
		}
	}
	return nil
}

// sqliteDialect and postgresDialect are the two Bun dialects Nooks supports.
func sqliteDialect() schema.Dialect   { return sqlitedialect.New() }
func postgresDialect() schema.Dialect { return pgdialect.New() }

/*
The locks a queue is taken one at a time under. Numbers rather than names because that is
what Postgres advisory locks are, and arbitrary because nothing else uses them: what
matters is that two callers doing the same thing pick the same one.
*/
const (
	joinQueueLock  int64 = 8_101
	resetQueueLock int64 = 8_102
)

/*
oneAtATime runs a write with every other write that shares its lock, on the driver that
needs telling.

A rule written as a condition on a single statement is atomic on SQLite, which has one
writer, and is not on Postgres, which decides the condition against a snapshot taken when
the statement began. Two callers in that gap both find the queue empty and both write,
which is how "one request waiting per email" and "at most twenty-five waiting" stop being
true. The Item positions were the same fault found first.

An advisory lock rather than a row, because these queues have no row to hold: the rule is
about the table as a whole. It is held until the transaction ends, so it goes on commit
or rollback and cannot be left behind.

Coarse on purpose. Asking to join and asking for a password are things a person does by
typing their name, so serialising them costs nothing worth measuring, and the alternative
is a lock per email that would have to be right about what an email is.
*/
func (s *sqlStore) oneAtATime(ctx context.Context, lock int64, write func(bun.IDB) error) error {
	if s.name != postgresDriver {
		return write(s.db)
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin the queued write: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.NewRaw("SELECT pg_advisory_xact_lock(?)", lock).Exec(ctx); err != nil {
		return fmt.Errorf("take the queue lock: %w", err)
	}
	if err := write(tx); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit the queued write: %w", err)
	}
	return nil
}
