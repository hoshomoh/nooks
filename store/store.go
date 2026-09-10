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

	// Close releases the underlying database handle.
	Close() error
}
