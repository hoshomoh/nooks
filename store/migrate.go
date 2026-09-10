package store

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"path"
	"sort"
	"strings"

	"github.com/uptrace/bun"
)

// migrations holds one directory of numbered .sql files per driver, plus a LATEST.sql
// that states the whole schema at head. LATEST.sql is documentation and is never run.
//
//go:embed migration/sqlite/*.sql migration/postgres/*.sql
var migrations embed.FS

// migrate applies every migration the database has not seen, in filename order, each
// in its own transaction. Applying an already-applied migration is a no-op, so it is
// safe to call on every start.
func migrate(ctx context.Context, db *bun.DB, driver string) error {
	if err := ensureMigrationTable(ctx, db); err != nil {
		return err
	}
	applied, err := appliedMigrations(ctx, db)
	if err != nil {
		return err
	}
	pending, err := pendingMigrations(driver, applied)
	if err != nil {
		return err
	}
	for _, name := range pending {
		if err := applyMigration(ctx, db, driver, name); err != nil {
			return err
		}
	}
	return nil
}

// ensureMigrationTable creates the ledger. Its own DDL cannot be a migration, so it is
// written to be valid in both dialects.
func ensureMigrationTable(ctx context.Context, db *bun.DB) error {
	const query = `CREATE TABLE IF NOT EXISTS schema_migration (name TEXT PRIMARY KEY)`
	if _, err := db.ExecContext(ctx, query); err != nil {
		return fmt.Errorf("create schema_migration: %w", err)
	}
	return nil
}

// appliedMigrations reads the names already recorded in the ledger.
func appliedMigrations(ctx context.Context, db *bun.DB) (map[string]bool, error) {
	rows, err := db.QueryContext(ctx, `SELECT name FROM schema_migration`)
	if err != nil {
		return nil, fmt.Errorf("read schema_migration: %w", err)
	}
	defer rows.Close()

	applied := map[string]bool{}
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, fmt.Errorf("scan schema_migration: %w", err)
		}
		applied[name] = true
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read schema_migration: %w", err)
	}
	return applied, nil
}

// pendingMigrations lists this dialect's migrations that are not yet applied, sorted by
// filename. LATEST.sql is excluded: it describes the schema, it does not build it.
func pendingMigrations(driver string, applied map[string]bool) ([]string, error) {
	dir := path.Join("migration", driver)
	entries, err := fs.ReadDir(migrations, dir)
	if err != nil {
		return nil, fmt.Errorf("read migrations for %s: %w", driver, err)
	}

	var pending []string
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || name == "LATEST.sql" || !strings.HasSuffix(name, ".sql") {
			continue
		}
		if applied[name] {
			continue
		}
		pending = append(pending, name)
	}
	sort.Strings(pending)
	return pending, nil
}

// applyMigration runs one migration and records it, both or neither.
func applyMigration(ctx context.Context, db *bun.DB, driver string, name string) error {
	statements, err := migrations.ReadFile(path.Join("migration", driver, name))
	if err != nil {
		return fmt.Errorf("read migration %s: %w", name, err)
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin migration %s: %w", name, err)
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.ExecContext(ctx, string(statements)); err != nil {
		return fmt.Errorf("apply migration %s: %w", name, err)
	}
	if _, err := tx.NewInsert().Model(&migrationRecord{Name: name}).Exec(ctx); err != nil {
		return fmt.Errorf("record migration %s: %w", name, err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit migration %s: %w", name, err)
	}
	return nil
}

// migrationRecord is the ledger row saying a migration has run.
type migrationRecord struct {
	bun.BaseModel `bun:"table:schema_migration,alias:schema_migration"`

	Name string `bun:"name,pk"`
}
