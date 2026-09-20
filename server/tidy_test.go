package server

import (
	"errors"
	"io"
	"log/slog"
	"path/filepath"
	"testing"
	"time"

	"github.com/hoshomoh/nooks/internal/profile"
	"github.com/hoshomoh/nooks/server/auth"
	"github.com/hoshomoh/nooks/store"
)

/*
An Instance clears out what has expired.

Expiry is enforced when a session is read, so a row that outlives it lets nobody in. It
just stays, for the life of the Instance, and `DeleteExpiredSessions` sat written and
uncalled for exactly that reason: nothing went wrong, so nothing said so.

tidyUp is tested rather than the loop around it. The loop is a ticker and a select, and
a test of it would be a test of time.
*/
func TestAnInstanceClearsOutExpiredSessions(t *testing.T) {
	dir := t.TempDir()
	s, err := store.OpenSQLite(t.Context(), filepath.Join(dir, "nooks.db"))
	if err != nil {
		t.Fatalf("OpenSQLite: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })

	member, err := s.CreateMember(t.Context(), store.CreateMemberParams{
		UID: "mem_anna", Name: "Anna", Email: "anna@brunnen.lan", Role: store.RoleAdmin,
		PasswordHash: "hash", CreatedAt: time.Now(),
	})
	if err != nil {
		t.Fatalf("CreateMember: %v", err)
	}

	stale, live := sessionFor(t, s, member, -time.Hour), sessionFor(t, s, member, time.Hour)

	server, err := New(profile.Config{
		Addr: ":0", Data: dir, Driver: profile.DriverSQLite, Mode: profile.ModeProd,
	}, s, slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	server.tidyUp(t.Context())

	if _, err := s.SessionByTokenHash(t.Context(), stale); !errors.Is(err, store.ErrNotFound) {
		t.Errorf("an expired session is still stored: %v", err)
	}
	if _, err := s.SessionByTokenHash(t.Context(), live); err != nil {
		t.Errorf("a live session was swept up with them: %v", err)
	}
}

// sessionFor stores one session that expires the given distance from now, and answers
// the hash it is stored under.
func sessionFor(t *testing.T, s store.Store, member store.Member, in time.Duration) string {
	t.Helper()
	_, hash, err := auth.NewToken()
	if err != nil {
		t.Fatalf("NewToken: %v", err)
	}
	if err := s.CreateSession(t.Context(), store.Session{
		TokenHash: hash, MemberID: member.ID,
		CreatedAt: time.Now().Add(-2 * time.Hour), ExpiresAt: time.Now().Add(in),
	}); err != nil {
		t.Fatalf("CreateSession: %v", err)
	}
	return hash
}
