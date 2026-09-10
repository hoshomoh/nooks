package store

import (
	"context"
	"path/filepath"
	"testing"
	"time"
)

// newSQLiteStore opens a store in a directory the test framework cleans up.
func newSQLiteStore(t *testing.T) Store {
	t.Helper()
	s, err := OpenSQLite(t.Context(), filepath.Join(t.TempDir(), "nook.db"))
	if err != nil {
		t.Fatalf("OpenSQLite: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })
	return s
}

func TestInstanceSettingsBeforeFirstRun(t *testing.T) {
	s := newSQLiteStore(t)

	settings, err := s.InstanceSettings(t.Context())
	if err != nil {
		t.Fatalf("InstanceSettings: %v", err)
	}
	if !settings.NeedsSetup() {
		t.Error("NeedsSetup() = false on a fresh Instance, want true")
	}
	if settings.Name != "" {
		t.Errorf("Name = %q on a fresh Instance, want empty", settings.Name)
	}
}

func TestSaveAndReadInstanceSettings(t *testing.T) {
	s := newSQLiteStore(t)
	completedAt := time.Date(2026, time.March, 3, 9, 30, 0, 0, time.UTC)

	want := InstanceSettings{Name: "Brunnen Street", PublicSignup: true, SetupCompletedAt: completedAt}
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
}

func TestSaveInstanceSettingsReplacesEarlierValues(t *testing.T) {
	s := newSQLiteStore(t)
	ctx := t.Context()

	first := InstanceSettings{Name: "Brunnen Street", PublicSignup: true}
	if err := s.SaveInstanceSettings(ctx, first); err != nil {
		t.Fatalf("SaveInstanceSettings: %v", err)
	}
	second := InstanceSettings{Name: "Kastanienallee", PublicSignup: false}
	if err := s.SaveInstanceSettings(ctx, second); err != nil {
		t.Fatalf("SaveInstanceSettings: %v", err)
	}

	got, err := s.InstanceSettings(ctx)
	if err != nil {
		t.Fatalf("InstanceSettings: %v", err)
	}
	if got.Name != second.Name {
		t.Errorf("Name = %q, want %q", got.Name, second.Name)
	}
	if got.PublicSignup {
		t.Error("PublicSignup = true, want the second save to have turned it off")
	}
}

// TestMigrateIsIdempotent proves reopening an existing file does not re-run migrations,
// which is what makes it safe to migrate on every start.
func TestMigrateIsIdempotent(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "nook.db")

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

func TestOpenSQLiteRejectsEmptyPath(t *testing.T) {
	if _, err := OpenSQLite(t.Context(), ""); err == nil {
		t.Fatal("OpenSQLite with an empty path succeeded, want an error")
	}
}
