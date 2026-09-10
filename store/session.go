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
	query := fmt.Sprintf(
		`INSERT INTO session (token_hash, member_id, created_at, expires_at) VALUES (%s, %s, %s, %s)`,
		s.dialect.placeholder(1), s.dialect.placeholder(2),
		s.dialect.placeholder(3), s.dialect.placeholder(4))

	_, err := s.db.ExecContext(ctx, query, session.TokenHash, session.MemberID,
		formatTime(session.CreatedAt), formatTime(session.ExpiresAt))
	if err != nil {
		return fmt.Errorf("create session: %w", err)
	}
	return nil
}

// SessionByTokenHash finds a Session. An expired Session is still returned; deciding
// what to do about expiry is the caller's, not the store's.
func (s *sqlStore) SessionByTokenHash(ctx context.Context, tokenHash string) (Session, error) {
	query := fmt.Sprintf(
		`SELECT token_hash, member_id, created_at, expires_at FROM session WHERE token_hash = %s`,
		s.dialect.placeholder(1))

	var (
		session   Session
		createdAt string
		expiresAt string
	)
	err := s.db.QueryRowContext(ctx, query, tokenHash).
		Scan(&session.TokenHash, &session.MemberID, &createdAt, &expiresAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Session{}, ErrNotFound
		}
		return Session{}, fmt.Errorf("read session: %w", err)
	}

	if session.CreatedAt, err = parseTime(createdAt); err != nil {
		return Session{}, err
	}
	if session.ExpiresAt, err = parseTime(expiresAt); err != nil {
		return Session{}, err
	}
	return session, nil
}

// DeleteSession signs one browser out. Deleting a Session that is already gone is not
// an error: signing out twice is not a failure.
func (s *sqlStore) DeleteSession(ctx context.Context, tokenHash string) error {
	query := fmt.Sprintf(`DELETE FROM session WHERE token_hash = %s`, s.dialect.placeholder(1))
	if _, err := s.db.ExecContext(ctx, query, tokenHash); err != nil {
		return fmt.Errorf("delete session: %w", err)
	}
	return nil
}

// DeleteExpiredSessions clears out Sessions that have passed their expiry, and reports
// how many went.
func (s *sqlStore) DeleteExpiredSessions(ctx context.Context, now time.Time) (int64, error) {
	query := fmt.Sprintf(`DELETE FROM session WHERE expires_at <= %s`, s.dialect.placeholder(1))
	result, err := s.db.ExecContext(ctx, query, formatTime(now))
	if err != nil {
		return 0, fmt.Errorf("delete expired sessions: %w", err)
	}
	deleted, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("rows affected: %w", err)
	}
	return deleted, nil
}
