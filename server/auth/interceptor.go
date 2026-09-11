package auth

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"connectrpc.com/connect"

	"github.com/hoshomoh/nooks/store"
)

// Resolver turns a session cookie into a Member.
//
// Its dependencies are passed in — the store and the clock — so that its behaviour
// around expiry can be tested without waiting for time to pass.
type Resolver struct {
	store store.Store
	now   func() time.Time
}

// NewResolver builds a Resolver. now may be nil, in which case time.Now is used.
func NewResolver(s store.Store, now func() time.Time) *Resolver {
	if now == nil {
		now = time.Now
	}
	return &Resolver{store: s, now: now}
}

// Interceptor attaches the signed-in Member to the context of every request that
// carries a valid session.
//
// It never rejects a request: an anonymous request is legitimate — first run, sign-in
// and the Public list all happen without a session. Deciding who may do what is each
// service's job, using MemberFrom.
func (r *Resolver) Interceptor() connect.UnaryInterceptorFunc {
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			if grant, ok := r.Grant(ctx, req.Header()); ok {
				return next(WithGrant(ctx, grant), req)
			}
			return next(ctx, req)
		}
	}
}

/*
Grant resolves whoever is asking, however they identified themselves.

A session cookie is a browser and gets its Member's own access. A bearer token is a
script, a shortcut or an MCP client, and gets that access narrowed to what the token
names. Both end up as the same value, so nothing downstream has to know which arrived.

A cookie is tried first: a browser that also happens to carry a token header is still a
browser, and narrowing it would be a surprise.
*/
func (r *Resolver) Grant(ctx context.Context, header http.Header) (Grant, bool) {
	if session := cookieToken(header); session != "" {
		if member, ok := r.Member(ctx, session); ok {
			return Grant{Member: member}, true
		}
	}

	presented := bearerToken(header)
	if presented == "" {
		return Grant{}, false
	}
	return r.tokenGrant(ctx, presented)
}

// tokenGrant resolves a presented access token to what it may do.
//
// An expired token is refused rather than deleted: a Member should be able to see on
// their tokens page that the thing which stopped working is the one that ran out.
func (r *Resolver) tokenGrant(ctx context.Context, presented string) (Grant, bool) {
	token, err := r.store.AccessTokenByHash(ctx, HashToken(presented))
	if err != nil {
		return Grant{}, false
	}
	if token.Expired(r.now()) {
		return Grant{}, false
	}

	member, err := r.store.MemberByID(ctx, token.MemberID)
	if err != nil {
		return Grant{}, false
	}
	listIDs, err := r.store.TokenListIDs(ctx, token.ID)
	if err != nil {
		return Grant{}, false
	}

	// Best effort: a token that worked should not stop working because recording that
	// it worked failed.
	_ = r.store.MarkTokenUsed(ctx, token.ID, r.now())

	return NewTokenGrant(member, token, listIDs), true
}

// Member resolves a session token to a Member, reporting whether it is usable.
//
// An expired session is deleted as it is found: the request that trips over it is the
// cheapest place to clean it up, and it saves a scheduled sweep from being the only
// thing that ever does.
func (r *Resolver) Member(ctx context.Context, token string) (store.Member, bool) {
	session, err := r.store.SessionByTokenHash(ctx, HashToken(token))
	if err != nil {
		return store.Member{}, false
	}
	if session.Expired(r.now()) {
		_ = r.store.DeleteSession(ctx, session.TokenHash)
		return store.Member{}, false
	}

	member, err := r.store.MemberByID(ctx, session.MemberID)
	if err != nil {
		// A session whose Member is gone is unusable; ON DELETE CASCADE should have
		// removed it, so this is belt and braces.
		if errors.Is(err, store.ErrNotFound) {
			_ = r.store.DeleteSession(ctx, session.TokenHash)
		}
		return store.Member{}, false
	}
	return member, true
}

// tokenFromHeader pulls the session token out of request headers. http.Request is
// borrowed purely for its cookie parsing.
func cookieToken(header http.Header) string {
	cookie, err := (&http.Request{Header: header}).Cookie(CookieName)
	if err != nil {
		return ""
	}
	return cookie.Value
}

// bearerToken reads an access token from the Authorization header.
func bearerToken(header http.Header) string {
	value := header.Get("Authorization")
	const prefix = "Bearer "
	if !strings.HasPrefix(value, prefix) {
		return ""
	}
	return strings.TrimSpace(strings.TrimPrefix(value, prefix))
}
