package v1

import (
	"context"
	"fmt"
	"time"

	"github.com/hoshomoh/nooks/store"
)

/*
TokenActivity tells a Member what their Access tokens are doing.

Only the moments that need attention: the first time a key is used, and a key being
stopped by somebody other than its owner. An entry per request would be an audit log,
and an audit log in the Activity panel is a panel nobody reads.

It is a separate value from the services so that the resolver — which is the only thing
that sees a token presented — can be told without depending on any of them.
*/
type TokenActivity struct {
	recorder activityRecorder
}

// NewTokenActivity builds one. now and newUID may be nil for the real ones.
func NewTokenActivity(
	s store.Store,
	now func() time.Time,
	newUID func() (string, error),
) *TokenActivity {
	return &TokenActivity{recorder: newRecorder(s, now, newUID)}
}

// WithAnnouncer makes entries arrive without a refresh.
func (a *TokenActivity) WithAnnouncer(announce Announcer) *TokenActivity {
	a.recorder.announce = announce
	return a
}

// TokenFirstUsed records that a key started being used.
//
// Best effort: a token that works must not stop working because saying so failed. The
// Member would rather their script kept running than have a perfect panel.
func (a *TokenActivity) TokenFirstUsed(
	ctx context.Context,
	member store.Member,
	token store.AccessToken,
) {
	text := fmt.Sprintf("%q was used for the first time", token.Name)
	_ = a.recorder.record(ctx, member.ID, store.ActivityTokenUsed, text, token.UID)
}

// TokenRevokedByAdmin tells a Member that somebody else stopped one of their keys.
func (a *TokenActivity) TokenRevokedByAdmin(
	ctx context.Context,
	owner store.Member,
	token store.AccessToken,
	by store.Member,
) error {
	text := fmt.Sprintf("%s revoked your token %q", by.Name, token.Name)
	return a.recorder.record(ctx, owner.ID, store.ActivityTokenUsed, text, token.UID)
}
