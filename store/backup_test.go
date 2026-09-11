package store

import (
	"path/filepath"
	"testing"
	"time"
)

// What lands is a real database, not a format of ours — so the way to check it is to
// open it and read what was in the original.
func TestBackupWritesAReadableDatabase(t *testing.T) {
	dir := t.TempDir()
	s, err := OpenSQLite(t.Context(), filepath.Join(dir, "nooks.db"))
	if err != nil {
		t.Fatalf("OpenSQLite: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })

	at := time.Date(2026, 8, 25, 9, 0, 0, 0, time.UTC)
	if _, err := s.CreateMember(t.Context(), CreateMemberParams{
		UID: "mem_anna", Name: "Anna", Email: "anna@brunnen.lan", Role: RoleAdmin,
		PasswordHash: "hash", CreatedAt: at,
	}); err != nil {
		t.Fatalf("CreateMember: %v", err)
	}

	backup := filepath.Join(dir, "backup.db")
	if err := s.BackupTo(t.Context(), backup); err != nil {
		t.Fatalf("BackupTo: %v", err)
	}

	restored, err := OpenSQLite(t.Context(), backup)
	if err != nil {
		t.Fatalf("open the backup: %v", err)
	}
	t.Cleanup(func() { _ = restored.Close() })

	member, err := restored.MemberByEmail(t.Context(), "anna@brunnen.lan")
	if err != nil {
		t.Fatalf("MemberByEmail in the backup: %v", err)
	}
	if member.Name != "Anna" {
		t.Errorf("name = %q, want the one that was backed up", member.Name)
	}
}

// VACUUM INTO refuses to overwrite, which is the guard against a backup landing on top
// of the database it was taken from.
func TestBackupWillNotOverwrite(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nooks.db")
	s, err := OpenSQLite(t.Context(), path)
	if err != nil {
		t.Fatalf("OpenSQLite: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })

	if err := s.BackupTo(t.Context(), path); err == nil {
		t.Error("BackupTo onto an existing file succeeded, want a refusal")
	}
}
