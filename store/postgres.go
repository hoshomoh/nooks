package store

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib" // database/sql driver for Postgres
)

// OpenPostgres connects to an existing Postgres server and brings the schema up to
// date.
func OpenPostgres(ctx context.Context, dsn string) (Store, error) {
	if dsn == "" {
		return nil, fmt.Errorf("store: postgres dsn is required")
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("open postgres: %w", err)
	}

	store, err := open(ctx, db, postgresDialect)
	if err != nil {
		_ = db.Close()
		return nil, err
	}
	return store, nil
}
