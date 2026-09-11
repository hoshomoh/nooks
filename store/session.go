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
	// Kind says what the token is for. Empty reads as a refresh token, which is what
	// every session was before there were two kinds.
	Kind SessionKind
	// ParentHash is the refresh token that minted this one, on an access token.
	ParentHash string
}

/*
SessionKind separates the credential that lasts from the credential that travels.

A refresh token lives a month and only ever moves in an HttpOnly cookie, so no script
can read it. An access token lives an hour and is the only one ever put in a response
body — which is the only way a caller that is not a browser can hold a credential.

The rule the split buys: the thing that lasts a month is never the thing in a body.
*/
type SessionKind string

const (
	// SessionRefresh is the long-lived credential, cookie only.
	SessionRefresh SessionKind = "REFRESH"
	// SessionAccess is the short-lived credential, handed over in a body.
	SessionAccess SessionKind = "ACCESS"
)

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

// DeleteSessionTree signs one browser out, taking the access tokens that refresh token
// minted with it. A credential that outlives the session it came from is a credential
// nobody knows they still have.
func (s *sqlStore) DeleteSessionTree(ctx context.Context, refreshHash string) error {
	_, err := s.db.NewDelete().
		Model((*sessionModel)(nil)).
		Where("token_hash = ? OR parent_hash = ?", refreshHash, refreshHash).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("delete session tree: %w", err)
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
