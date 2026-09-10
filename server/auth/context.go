package auth

import (
	"context"

	"github.com/hoshomoh/nooks/store"
)

// contextKey is unexported so nothing outside this package can write the Member onto a
// context, which is what makes "the context says who this is" trustworthy.
type contextKey struct{}

// WithMember returns a context carrying the signed-in Member.
func WithMember(ctx context.Context, member store.Member) context.Context {
	return context.WithValue(ctx, contextKey{}, member)
}

// MemberFrom returns the signed-in Member, and whether there was one. A false result
// means the request is anonymous — a Visitor, or someone whose session has expired.
func MemberFrom(ctx context.Context) (store.Member, bool) {
	member, ok := ctx.Value(contextKey{}).(store.Member)
	return member, ok
}
