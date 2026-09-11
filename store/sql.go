package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"

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
)

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

// InstanceSettings reads the Instance's own configuration.
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

// Close releases the underlying database handle.
func (s *sqlStore) Close() error {
	if err := s.db.Close(); err != nil {
		return fmt.Errorf("close %s store: %w", s.name, err)
	}
	return nil
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
	return &sqlStore{db: bunDB, name: name}, nil
}

// sqliteDialect and postgresDialect are the two Bun dialects Nooks supports.
func sqliteDialect() schema.Dialect   { return sqlitedialect.New() }
func postgresDialect() schema.Dialect { return pgdialect.New() }
