// Package store persists everything an Instance knows.
//
// The package exposes one interface, Store, and two implementations of it that
// differ only in dialect. Nothing above this package knows which driver is in use,
// and nothing in this package knows about HTTP.
package store

import (
	"context"
	"time"
)

// InstanceSettings is an Instance's own configuration: the parts of it that the
// sign-in page and the Public list can see before anyone has signed in.
type InstanceSettings struct {
	// Name is the mutable display label chosen at first run, e.g. "Brunnen Street".
	// Empty until first run completes.
	Name string

	// PublicSignup, when false, makes every account a Join request an Admin approves.
	PublicSignup bool

	// SetupCompletedAt is when first run finished. The zero value means it has not,
	// and the app shows first run rather than sign in.
	SetupCompletedAt time.Time
}

// NeedsSetup reports whether first run still has to happen.
func (s InstanceSettings) NeedsSetup() bool {
	return s.SetupCompletedAt.IsZero()
}

// Store is the persistence boundary for an Instance.
//
// Implementations are obtained from OpenSQLite or OpenPostgres. Callers accept this
// interface rather than a concrete type so that a fake can stand in for tests.
type Store interface {
	// InstanceSettings reads the Instance's own configuration. It returns the zero
	// value, not an error, when first run has not happened.
	InstanceSettings(ctx context.Context) (InstanceSettings, error)

	// SaveInstanceSettings writes the Instance's own configuration in full.
	SaveInstanceSettings(ctx context.Context, settings InstanceSettings) error

	// CreateMember adds a Member, returning ErrEmailTaken if the email is in use.
	CreateMember(ctx context.Context, params CreateMemberParams) (Member, error)

	// MemberByEmail, MemberByUID and MemberByID each return ErrNotFound when there is
	// no such Member.
	MemberByEmail(ctx context.Context, email string) (Member, error)
	MemberByUID(ctx context.Context, uid string) (Member, error)
	MemberByID(ctx context.Context, id int64) (Member, error)

	// CountMembers reports how many Members exist.
	CountMembers(ctx context.Context) (int, error)

	// SetMemberPassword replaces a password and clears the must-change flag.
	SetMemberPassword(ctx context.Context, id int64, hash string) error

	// MarkMemberSignedIn records that a Member has just signed in.
	MarkMemberSignedIn(ctx context.Context, id int64, at time.Time) error

	// CreateSession records a signed-in browser.
	CreateSession(ctx context.Context, session Session) error

	// SessionByTokenHash returns ErrNotFound when there is no such Session. An expired
	// Session is still returned; expiry is the caller's decision.
	SessionByTokenHash(ctx context.Context, tokenHash string) (Session, error)

	// DeleteSession signs one browser out, and is not an error if it was already gone.
	DeleteSession(ctx context.Context, tokenHash string) error

	// DeleteExpiredSessions clears out Sessions past their expiry.
	DeleteExpiredSessions(ctx context.Context, now time.Time) (int64, error)

	// Close releases the underlying database handle.
	Close() error
}
