package store

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// postgresDSNEnv names the environment variable that points the suite at a Postgres
// server. When it is unset the Postgres cases skip, so `go test ./...` works on a
// laptop with nothing installed while CI still covers both drivers.
const postgresDSNEnv = "NOOKS_TEST_POSTGRES_DSN"

// driverCase is one driver the shared suite runs against.
type driverCase struct {
	name string
	// open returns a store, or skips the test when the driver is unavailable here.
	open func(t *testing.T) Store
}

// drivers lists every driver the suite covers.
func drivers() []driverCase {
	return []driverCase{
		{name: "sqlite", open: openSQLiteForTest},
		{name: "postgres", open: openPostgresForTest},
	}
}

func openSQLiteForTest(t *testing.T) Store {
	t.Helper()
	s, err := OpenSQLite(t.Context(), filepath.Join(t.TempDir(), "nooks.db"))
	if err != nil {
		t.Fatalf("OpenSQLite: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })
	return s
}

// openPostgresForTest connects to the server named by the environment, and drops the
// schema first so each test starts from nothing.
func openPostgresForTest(t *testing.T) Store {
	t.Helper()
	dsn := os.Getenv(postgresDSNEnv)
	if dsn == "" {
		t.Skipf("%s is not set", postgresDSNEnv)
	}

	s, err := OpenPostgres(t.Context(), dsn)
	if err != nil {
		t.Fatalf("OpenPostgres: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })

	// Postgres is shared between runs, unlike the per-test SQLite file, so each test
	// starts by emptying it. Every table is truncated rather than a named list, so
	// adding a table cannot quietly leave rows behind for the next test.
	concrete, ok := s.(*sqlStore)
	if !ok {
		t.Fatalf("OpenPostgres returned %T, want *sqlStore", s)
	}
	truncateAll(t, concrete)

	if err := s.SaveInstanceSettings(t.Context(), InstanceSettings{}); err != nil {
		t.Fatalf("reset instance settings: %v", err)
	}
	return s
}

// truncateAll empties every table except the migration ledger, which records work that
// has genuinely been done.
func truncateAll(t *testing.T, s *sqlStore) {
	t.Helper()

	var tables []string
	err := s.db.NewSelect().
		ColumnExpr("tablename").
		TableExpr("pg_tables").
		Where("schemaname = current_schema() AND tablename <> ?", "schema_migration").
		Scan(t.Context(), &tables)
	if err != nil {
		t.Fatalf("list tables: %v", err)
	}
	if len(tables) == 0 {
		return
	}

	query := "TRUNCATE " + strings.Join(tables, ", ") + " RESTART IDENTITY CASCADE"
	if _, err := s.db.ExecContext(t.Context(), query); err != nil {
		t.Fatalf("truncate: %v", err)
	}
}

func TestInstanceSettingsBeforeFirstRun(t *testing.T) {
	for _, d := range drivers() {
		t.Run(d.name, func(t *testing.T) {
			settings, err := d.open(t).InstanceSettings(t.Context())
			if err != nil {
				t.Fatalf("InstanceSettings: %v", err)
			}
			if !settings.NeedsSetup() {
				t.Error("NeedsSetup() = false on a fresh Instance, want true")
			}
			if settings.Name != "" {
				t.Errorf("Name = %q on a fresh Instance, want empty", settings.Name)
			}
		})
	}
}

func TestSaveAndReadInstanceSettings(t *testing.T) {
	completedAt := time.Date(2026, time.March, 3, 9, 30, 0, 0, time.UTC)
	want := InstanceSettings{Name: "Brunnen Street", PublicSignup: true, SetupCompletedAt: completedAt}

	for _, d := range drivers() {
		t.Run(d.name, func(t *testing.T) {
			s := d.open(t)
			if err := s.SaveInstanceSettings(t.Context(), want); err != nil {
				t.Fatalf("SaveInstanceSettings: %v", err)
			}

			got, err := s.InstanceSettings(t.Context())
			if err != nil {
				t.Fatalf("InstanceSettings: %v", err)
			}
			if got.Name != want.Name {
				t.Errorf("Name = %q, want %q", got.Name, want.Name)
			}
			if got.PublicSignup != want.PublicSignup {
				t.Errorf("PublicSignup = %v, want %v", got.PublicSignup, want.PublicSignup)
			}
			if !got.SetupCompletedAt.Equal(completedAt) {
				t.Errorf("SetupCompletedAt = %v, want %v", got.SetupCompletedAt, completedAt)
			}
			if got.NeedsSetup() {
				t.Error("NeedsSetup() = true after first run, want false")
			}
		})
	}
}

func TestSaveInstanceSettingsReplacesEarlierValues(t *testing.T) {
	for _, d := range drivers() {
		t.Run(d.name, func(t *testing.T) {
			s := d.open(t)
			ctx := t.Context()

			if err := s.SaveInstanceSettings(ctx, InstanceSettings{Name: "Brunnen Street", PublicSignup: true}); err != nil {
				t.Fatalf("first SaveInstanceSettings: %v", err)
			}
			if err := s.SaveInstanceSettings(ctx, InstanceSettings{Name: "Kastanienallee"}); err != nil {
				t.Fatalf("second SaveInstanceSettings: %v", err)
			}

			got, err := s.InstanceSettings(ctx)
			if err != nil {
				t.Fatalf("InstanceSettings: %v", err)
			}
			if got.Name != "Kastanienallee" {
				t.Errorf("Name = %q, want the second save to have replaced it", got.Name)
			}
			if got.PublicSignup {
				t.Error("PublicSignup = true, want the second save to have turned it off")
			}
		})
	}
}

// TestMigrateIsIdempotent proves reopening an existing file does not re-run migrations,
// which is what makes it safe to migrate on every start.
func TestMigrateIsIdempotent(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "nooks.db")

	first, err := OpenSQLite(ctx, path)
	if err != nil {
		t.Fatalf("first OpenSQLite: %v", err)
	}
	if err := first.SaveInstanceSettings(ctx, InstanceSettings{Name: "Brunnen Street"}); err != nil {
		t.Fatalf("SaveInstanceSettings: %v", err)
	}
	if err := first.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	second, err := OpenSQLite(ctx, path)
	if err != nil {
		t.Fatalf("second OpenSQLite: %v", err)
	}
	defer func() { _ = second.Close() }()

	got, err := second.InstanceSettings(ctx)
	if err != nil {
		t.Fatalf("InstanceSettings: %v", err)
	}
	if got.Name != "Brunnen Street" {
		t.Errorf("Name = %q after reopening, want the saved value", got.Name)
	}
}

func TestOpenRejectsMissingTarget(t *testing.T) {
	if _, err := OpenSQLite(t.Context(), ""); err == nil {
		t.Error("OpenSQLite with an empty path succeeded, want an error")
	}
	if _, err := OpenPostgres(t.Context(), ""); err == nil {
		t.Error("OpenPostgres with an empty dsn succeeded, want an error")
	}
}
