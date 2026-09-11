package auth

import (
	"context"

	"github.com/hoshomoh/nooks/store"
)

// grantKey is unexported so nothing outside this package can write a Grant onto a
// context, which is what makes "the context says what this caller may do" trustworthy.
type grantKey struct{}

/*
Grant is what the caller may do, and how it got narrowed.

A browser's Grant is the Member and nothing else. A token's is the same Member with two
limits on top: the Lists it names, and whether it may write. That is the whole of the
token model — a token is never its own identity with its own permissions, it is somebody
else's access with edges.

Keeping it in one value means the rules are applied in one place. Permissions are
decided by accessTo, and everything that reaches Nooks — browser, script, MCP client —
goes through the same decision.
*/
type Grant struct {
	Member store.Member
	// Token is nil for a browser. When set, the two fields below apply.
	Token *store.AccessToken
	// reach is the Lists the token names, by internal identity.
	reach map[int64]bool
}

// NewTokenGrant builds the Grant a presented token earns.
func NewTokenGrant(member store.Member, token store.AccessToken, listIDs []int64) Grant {
	reach := make(map[int64]bool, len(listIDs))
	for _, id := range listIDs {
		reach[id] = true
	}
	return Grant{Member: member, Token: &token, reach: reach}
}

// Reaches reports whether the caller may see a List at all.
//
// A browser reaches everything its Member does. A token reaches only what it names —
// and a List it does not name is invisible rather than forbidden, so this is asked
// before anything else.
func (g Grant) Reaches(listID int64) bool {
	if g.Token == nil || g.Token.AllLists {
		return true
	}
	return g.reach[listID]
}

// TokenID is the Access token the caller presented, or zero for a browser.
//
// For recording what a change came through. Nothing decides access from it: that is
// Reaches and ReadOnly, which are the same question asked the way the rules ask it.
func (g Grant) TokenID() int64 {
	if g.Token == nil {
		return 0
	}
	return g.Token.ID
}

// MayWrite reports whether the caller may add, tick and edit.
//
// A browser may do whatever its Member may. A token may only what it was cut for.
func (g Grant) MayWrite() bool {
	return g.Token == nil || g.Token.Abilities.Write
}

/*
MayDelete reports whether the caller may delete Items and Lists.

Its own question rather than part of writing, because deleting is the one that cannot
be undone and almost no caller needs it. A shortcut that adds shopping should not be
able to empty a List because somebody sent it the wrong request.

It is checked beside accessTo rather than inside it: accessTo answers how far into a
List a caller may reach, and this answers what they may do once there. Folding them
together would mean a level called "write but not delete", which is the shape the token
model deliberately does not have.
*/
func (g Grant) MayDelete() bool {
	return g.Token == nil || g.Token.Abilities.Delete
}

// WithGrant returns a context carrying what the caller may do.
func WithGrant(ctx context.Context, grant Grant) context.Context {
	return WithMember(context.WithValue(ctx, grantKey{}, grant), grant.Member)
}

// GrantFrom returns the caller's Grant.
//
// A context carrying only a Member — which is what a browser's request has, and what a
// test usually builds — reads as an unnarrowed Grant. That way a service never has to
// ask which kind of caller this is.
func GrantFrom(ctx context.Context) (Grant, bool) {
	if grant, ok := ctx.Value(grantKey{}).(Grant); ok {
		return grant, true
	}
	member, ok := MemberFrom(ctx)
	return Grant{Member: member}, ok
}
