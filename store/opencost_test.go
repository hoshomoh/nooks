package store

import (
	"path/filepath"
	"testing"
)

// BenchmarkOpeningAFreshDatabase is the cost every test pays: a file that does not
// exist, so every migration runs.
func BenchmarkOpeningAFreshDatabase(b *testing.B) {
	for b.Loop() {
		s, err := OpenSQLite(b.Context(), filepath.Join(b.TempDir(), "nooks.db"))
		if err != nil {
			b.Fatal(err)
		}
		_ = s.Close()
	}
}

// BenchmarkOpeningAMigratedDatabase is what it would cost if the schema were already
// there: the same open, with nothing left to migrate.
func BenchmarkOpeningAMigratedDatabase(b *testing.B) {
	path := filepath.Join(b.TempDir(), "nooks.db")
	first, err := OpenSQLite(b.Context(), path)
	if err != nil {
		b.Fatal(err)
	}
	_ = first.Close()

	b.ResetTimer()
	for b.Loop() {
		s, err := OpenSQLite(b.Context(), path)
		if err != nil {
			b.Fatal(err)
		}
		_ = s.Close()
	}
}
