package store

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib" // database/sql driver for Postgres
)

/*
DefaultMaxConns is how many connections an Instance opens to Postgres at once.

`database/sql` opens as many as it needs and closes none of them under load, so a burst
took whatever it took. Postgres allows a hundred in total out of the box, shared with
anything else on that server, so a busy moment could exhaust them and break both this
and whatever else was using that database.

Ten, because a household is a handful of people and a few tokens, and because leaving
ninety for pgAdmin, a backup job and the next application is the neighbourly default on
a machine somebody else also uses. `--db-max-conns` is there for whoever has a reason to
disagree.
*/
const DefaultMaxConns = 10

// OpenPostgres connects to an existing Postgres server and brings the schema up to
// date. maxConns bounds the pool; zero takes DefaultMaxConns.
func OpenPostgres(ctx context.Context, dsn string, maxConns int) (Store, error) {
	if dsn == "" {
		return nil, fmt.Errorf("store: postgres dsn is required")
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("open postgres: %w", err)
	}
	if maxConns <= 0 {
		maxConns = DefaultMaxConns
	}
	db.SetMaxOpenConns(maxConns)
	// Idle connections are what make the next request cheap, and holding more of them
	// than the pool may open is not a thing that can happen.
	db.SetMaxIdleConns(maxConns)

	store, err := open(ctx, db, postgresDialect(), "postgres")
	if err != nil {
		// As in sqlite.go: the handle is going nowhere and its close error would bury
		// the one worth reading.
		_ = db.Close()
		return nil, err
	}
	return store, nil
}

// maxOpenConns reports the bound actually put on the pool, for the test beside it.
func (s *sqlStore) maxOpenConns() int { return s.db.DB.Stats().MaxOpenConnections }
