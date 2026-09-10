package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"time"
)

// Setting keys. Instance configuration is a handful of rows rather than a table of
// one row, so adding a setting never needs a migration.
const (
	settingInstanceName     = "instance.name"
	settingPublicSignup     = "instance.public_signup"
	settingSetupCompletedAt = "instance.setup_completed_at"
)

// dialect carries the few things that differ between SQLite and Postgres. It exists
// because two drivers share every query but not their placeholder syntax; it is not
// an abstraction over databases in general.
type dialect struct {
	// name identifies the dialect in errors and in the migration table.
	name string
	// placeholder renders the nth (1-based) bind parameter.
	placeholder func(n int) string
}

var (
	sqliteDialect   = dialect{name: "sqlite", placeholder: func(int) string { return "?" }}
	postgresDialect = dialect{name: "postgres", placeholder: func(n int) string { return "$" + strconv.Itoa(n) }}
)

// sqlStore implements Store over any database/sql handle.
type sqlStore struct {
	db      *sql.DB
	dialect dialect
}

// InstanceSettings reads the Instance's own configuration.
func (s *sqlStore) InstanceSettings(ctx context.Context) (InstanceSettings, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT key, value FROM setting`)
	if err != nil {
		return InstanceSettings{}, fmt.Errorf("read instance settings: %w", err)
	}
	defer rows.Close()

	values := map[string]string{}
	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			return InstanceSettings{}, fmt.Errorf("scan setting: %w", err)
		}
		values[key] = value
	}
	if err := rows.Err(); err != nil {
		return InstanceSettings{}, fmt.Errorf("read instance settings: %w", err)
	}
	return settingsFromValues(values)
}

// SaveInstanceSettings writes the Instance's own configuration in full.
func (s *sqlStore) SaveInstanceSettings(ctx context.Context, settings InstanceSettings) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	for key, value := range valuesFromSettings(settings) {
		if err := s.upsertSetting(ctx, tx, key, value); err != nil {
			return err
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit instance settings: %w", err)
	}
	return nil
}

// upsertSetting writes one key, replacing any existing value.
func (s *sqlStore) upsertSetting(ctx context.Context, tx *sql.Tx, key, value string) error {
	query := fmt.Sprintf(
		`INSERT INTO setting (key, value) VALUES (%s, %s)
		 ON CONFLICT (key) DO UPDATE SET value = excluded.value`,
		s.dialect.placeholder(1), s.dialect.placeholder(2),
	)
	if _, err := tx.ExecContext(ctx, query, key, value); err != nil {
		return fmt.Errorf("write setting %s: %w", key, err)
	}
	return nil
}

// Close releases the underlying database handle.
func (s *sqlStore) Close() error {
	if err := s.db.Close(); err != nil {
		return fmt.Errorf("close %s store: %w", s.dialect.name, err)
	}
	return nil
}

// settingsFromValues turns stored rows into InstanceSettings. A missing key is not an
// error: it means first run has not set it yet.
func settingsFromValues(values map[string]string) (InstanceSettings, error) {
	settings := InstanceSettings{
		Name:         values[settingInstanceName],
		PublicSignup: values[settingPublicSignup] == "true",
	}

	raw, ok := values[settingSetupCompletedAt]
	if !ok || raw == "" {
		return settings, nil
	}
	completedAt, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		return InstanceSettings{}, fmt.Errorf("parse %s %q: %w", settingSetupCompletedAt, raw, err)
	}
	settings.SetupCompletedAt = completedAt
	return settings, nil
}

// valuesFromSettings is the inverse of settingsFromValues.
func valuesFromSettings(settings InstanceSettings) map[string]string {
	completedAt := ""
	if !settings.SetupCompletedAt.IsZero() {
		completedAt = settings.SetupCompletedAt.UTC().Format(time.RFC3339)
	}
	return map[string]string{
		settingInstanceName:     settings.Name,
		settingPublicSignup:     strconv.FormatBool(settings.PublicSignup),
		settingSetupCompletedAt: completedAt,
	}
}

// open verifies the handle and brings the schema up to date.
func open(ctx context.Context, db *sql.DB, d dialect) (Store, error) {
	if db == nil {
		return nil, errors.New("store: database handle is required")
	}
	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("reach %s database: %w", d.name, err)
	}
	if err := migrate(ctx, db, d); err != nil {
		return nil, err
	}
	return &sqlStore{db: db, dialect: d}, nil
}
