package store

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite" // pure-Go driver: no cgo, so Nook cross-compiles freely
)

// OpenSQLite opens the Instance's single file, creating it and its directory when
// they do not exist, and brings the schema up to date.
func OpenSQLite(ctx context.Context, path string) (Store, error) {
	if path == "" {
		return nil, fmt.Errorf("store: sqlite path is required")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("create data directory: %w", err)
	}

	// WAL keeps reads from blocking the one writer, which matters on a home server
	// where a slow request should not stall the shopping list.
	dsn := "file:" + path + "?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)&_pragma=foreign_keys(1)"
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open sqlite %s: %w", path, err)
	}

	store, err := open(ctx, db, sqliteDialect)
	if err != nil {
		_ = db.Close()
		return nil, err
	}
	return store, nil
}
