package store

import (
	"context"
	"errors"
	"fmt"
)

// ErrNoBackup reports a driver that cannot make a copy of itself.
//
// Postgres is the case: its backup tool is pg_dump, which already exists and is better
// than anything Nooks would write. Saying so is more useful than a worse export.
var ErrNoBackup = errors.New("store: this driver exports with its own tools")

/*
BackupTo writes a consistent copy of the database to a path that does not yet exist.

SQLite's own VACUUM INTO, which is the whole implementation: it takes a copy while the
Instance keeps running, and what lands is a real database file rather than a format of
ours. Nothing here describes the schema, so nothing here can fall behind it — which is
the failure a hand-written exporter has, silently, one table at a time.
*/
func (s *sqlStore) BackupTo(ctx context.Context, path string) error {
	if s.name != "sqlite" {
		return ErrNoBackup
	}
	// A bound parameter, so a path with a quote in it cannot end the string. VACUUM INTO
	// refuses a file that already exists, which is the check this would otherwise need.
	if _, err := s.db.ExecContext(ctx, "VACUUM INTO ?", path); err != nil {
		return fmt.Errorf("copy database: %w", err)
	}
	return nil
}

// Driver names which engine is holding the data, for anything that has to say so.
func (s *sqlStore) Driver() string { return s.name }
