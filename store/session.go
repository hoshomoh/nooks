package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// Session is one signed-in browser.
//
// Only the hash of the token is stored, never the token, so a stolen database does not
// hand over live sessions.
type Session struct {
	TokenHash string
	MemberID  int64
	CreatedAt time.Time
	ExpiresAt time.Time
}

// Expired reports whether the Session has passed its expiry at the given moment. The
// clock is a parameter so this can be tested without waiting.
func (s Session) Expired(now time.Time) bool {
	return !now.Before(s.ExpiresAt)
}

// CreateSession records a signed-in browser.
func (s *sqlStore) CreateSession(ctx context.Context, session Session) error {
	if session.TokenHash == "" {
		return errors.New("store: session token hash is required")
	}
	if _, err := s.db.NewInsert().Model(newSessionModel(session)).Exec(ctx); err != nil {
		return fmt.Errorf("create session: %w", err)
	}
	return nil
}

// SessionByTokenHash finds a Session. An expired Session is still returned; deciding
// what to do about expiry is the caller's, not the store's.
func (s *sqlStore) SessionByTokenHash(ctx context.Context, tokenHash string) (Session, error) {
	row := new(sessionModel)
	err := s.db.NewSelect().Model(row).Where("token_hash = ?", tokenHash).Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Session{}, ErrNotFound
		}
		return Session{}, fmt.Errorf("read session: %w", err)
	}
	return row.toSession()
}

// DeleteSession signs one browser out. Deleting a Session that is already gone is not
// an error: signing out twice is not a failure.
func (s *sqlStore) DeleteSession(ctx context.Context, tokenHash string) error {
	_, err := s.db.NewDelete().
		Model((*sessionModel)(nil)).
		Where("token_hash = ?", tokenHash).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("delete session: %w", err)
	}
	return nil
}

// DeleteExpiredSessions clears out Sessions that have passed their expiry, and reports
// how many went.
func (s *sqlStore) DeleteExpiredSessions(ctx context.Context, now time.Time) (int64, error) {
	result, err := s.db.NewDelete().
		Model((*sessionModel)(nil)).
		Where("expires_at <= ?", formatTime(now)).
		Exec(ctx)
	if err != nil {
		return 0, fmt.Errorf("delete expired sessions: %w", err)
	}
	deleted, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("rows affected: %w", err)
	}
	return deleted, nil
}
