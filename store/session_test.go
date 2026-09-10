package store

import (
	"errors"
	"testing"
	"time"
)

// newMember adds a Member for a Session to belong to.
func newMember(t *testing.T, s Store) Member {
	t.Helper()
	member, err := s.CreateMember(t.Context(), anna())
	if err != nil {
		t.Fatalf("CreateMember: %v", err)
	}
	return member
}

func TestCreateAndReadSession(t *testing.T) {
	for _, d := range drivers() {
		t.Run(d.name, func(t *testing.T) {
			s := d.open(t)
			member := newMember(t, s)
			expires := createdAt.Add(30 * 24 * time.Hour)

			want := Session{
				TokenHash: "hash-of-the-token",
				MemberID:  member.ID,
				CreatedAt: createdAt,
				ExpiresAt: expires,
			}
			if err := s.CreateSession(t.Context(), want); err != nil {
				t.Fatalf("CreateSession: %v", err)
			}

			got, err := s.SessionByTokenHash(t.Context(), want.TokenHash)
			if err != nil {
				t.Fatalf("SessionByTokenHash: %v", err)
			}
			if got.MemberID != member.ID {
				t.Errorf("MemberID = %d, want %d", got.MemberID, member.ID)
			}
			if !got.ExpiresAt.Equal(expires) {
				t.Errorf("ExpiresAt = %v, want %v", got.ExpiresAt, expires)
			}
		})
	}
}

func TestSessionNotFound(t *testing.T) {
	for _, d := range drivers() {
		t.Run(d.name, func(t *testing.T) {
			_, err := d.open(t).SessionByTokenHash(t.Context(), "no-such-hash")
			if !errors.Is(err, ErrNotFound) {
				t.Errorf("SessionByTokenHash for an unknown token = %v, want ErrNotFound", err)
			}
		})
	}
}

// Signing out twice is not a failure.
func TestDeleteSessionIsIdempotent(t *testing.T) {
	for _, d := range drivers() {
		t.Run(d.name, func(t *testing.T) {
			s := d.open(t)
			member := newMember(t, s)
			session := Session{TokenHash: "h", MemberID: member.ID, CreatedAt: createdAt, ExpiresAt: createdAt.Add(time.Hour)}
			if err := s.CreateSession(t.Context(), session); err != nil {
				t.Fatalf("CreateSession: %v", err)
			}

			if err := s.DeleteSession(t.Context(), "h"); err != nil {
				t.Fatalf("first DeleteSession: %v", err)
			}
			if err := s.DeleteSession(t.Context(), "h"); err != nil {
				t.Errorf("second DeleteSession: %v, want nil", err)
			}
			if _, err := s.SessionByTokenHash(t.Context(), "h"); !errors.Is(err, ErrNotFound) {
				t.Errorf("session survived deletion: %v", err)
			}
		})
	}
}

// Expiry is the caller's decision, so the store still returns an expired Session.
func TestExpiredSessionIsStillReturned(t *testing.T) {
	for _, d := range drivers() {
		t.Run(d.name, func(t *testing.T) {
			s := d.open(t)
			member := newMember(t, s)
			expired := Session{
				TokenHash: "stale",
				MemberID:  member.ID,
				CreatedAt: createdAt,
				ExpiresAt: createdAt.Add(time.Hour),
			}
			if err := s.CreateSession(t.Context(), expired); err != nil {
				t.Fatalf("CreateSession: %v", err)
			}

			got, err := s.SessionByTokenHash(t.Context(), "stale")
			if err != nil {
				t.Fatalf("SessionByTokenHash: %v", err)
			}
			if !got.Expired(createdAt.Add(2 * time.Hour)) {
				t.Error("Expired(after expiry) = false, want true")
			}
			if got.Expired(createdAt.Add(30 * time.Minute)) {
				t.Error("Expired(before expiry) = true, want false")
			}
		})
	}
}

func TestDeleteExpiredSessions(t *testing.T) {
	for _, d := range drivers() {
		t.Run(d.name, func(t *testing.T) {
			s := d.open(t)
			member := newMember(t, s)
			live := Session{TokenHash: "live", MemberID: member.ID, CreatedAt: createdAt, ExpiresAt: createdAt.Add(48 * time.Hour)}
			stale := Session{TokenHash: "stale", MemberID: member.ID, CreatedAt: createdAt, ExpiresAt: createdAt.Add(time.Hour)}
			for _, session := range []Session{live, stale} {
				if err := s.CreateSession(t.Context(), session); err != nil {
					t.Fatalf("CreateSession: %v", err)
				}
			}

			deleted, err := s.DeleteExpiredSessions(t.Context(), createdAt.Add(2*time.Hour))
			if err != nil {
				t.Fatalf("DeleteExpiredSessions: %v", err)
			}
			if deleted != 1 {
				t.Errorf("deleted %d sessions, want 1", deleted)
			}
			if _, err := s.SessionByTokenHash(t.Context(), "live"); err != nil {
				t.Errorf("the live session was removed: %v", err)
			}
		})
	}
}
