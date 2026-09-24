package auth

import (
	"net/http"
	"testing"
	"time"

	"github.com/hoshomoh/nooks/store"
	"github.com/hoshomoh/nooks/store/storetest"
)

var resolvedAt = time.Date(2026, time.April, 1, 9, 0, 0, 0, time.UTC)

// instance is a store with one Member and whatever sessions a test gives them.
func instance(t *testing.T) (store.Store, store.Member) {
	t.Helper()
	s := storetest.Fresh(t)

	member, err := s.CreateMember(t.Context(), store.CreateMemberParams{
		UID: "mem_anna", Name: "Anna", Email: "anna@brunnen.lan",
		Role: store.RoleAdmin, PasswordHash: "hash", CreatedAt: resolvedAt,
	})
	if err != nil {
		t.Fatalf("CreateMember: %v", err)
	}
	return s, member
}

// sessioned stores one session of a kind and returns the token that opens it.
func sessioned(t *testing.T, s store.Store, member store.Member, kind store.SessionKind) string {
	t.Helper()
	token := "token-" + string(kind)
	err := s.CreateSession(t.Context(), store.Session{
		TokenHash: HashToken(token),
		MemberID:  member.ID,
		Kind:      kind,
		ExpiresAt: resolvedAt.Add(time.Hour),
		CreatedAt: resolvedAt,
	})
	if err != nil {
		t.Fatalf("CreateSession %s: %v", kind, err)
	}
	return token
}

/*
A credential is only ever what it was issued as.

Two rules, one each way, both written down in `Member` and `accessMember` and neither
held by anything until now. They are not tidiness: the whole browser-only boundary rests
on the first of them.

`requireBrowserGrant` refuses a request whose grant carries a Token, and a grant made
from a cookie carries none. So an Access token accepted as a cookie would arrive as a
browser: able to create more tokens, change the password, remove Members. The check that
stops it is one comparison of a session's kind.

The other way costs less and is the same mistake: a refresh token is long-lived and sits
in a cookie, and taken as a bearer it would be a durable credential doing an access
token's work.
*/
func TestACredentialIsOnlyWhatItWasIssuedAs(t *testing.T) {
	t.Run("an access token is not a cookie", func(t *testing.T) {
		s, member := instance(t)
		access := sessioned(t, s, member, store.SessionAccess)
		resolver := NewResolver(s, func() time.Time { return resolvedAt })

		header := http.Header{}
		header.Set("Cookie", CookieName+"="+access)

		if grant, ok := resolver.Grant(t.Context(), header); ok {
			t.Errorf("an Access token in a cookie was let in as %q, with no Token on the "+
				"grant, which is what requireBrowser reads to mean a browser",
				grant.Member.Name)
		}
	})

	t.Run("a refresh token is not a bearer", func(t *testing.T) {
		s, member := instance(t)
		refresh := sessioned(t, s, member, store.SessionRefresh)
		resolver := NewResolver(s, func() time.Time { return resolvedAt })

		header := http.Header{}
		header.Set("Authorization", "Bearer "+refresh)

		if _, ok := resolver.Grant(t.Context(), header); ok {
			t.Error("a refresh token was accepted as a bearer, which makes a long-lived " +
				"credential do a short-lived one's work")
		}
	})

	// The ordinary way round still works, so neither case above can pass by the
	// resolver refusing everything.
	t.Run("each is accepted where it belongs", func(t *testing.T) {
		s, member := instance(t)
		refresh := sessioned(t, s, member, store.SessionRefresh)
		access := sessioned(t, s, member, store.SessionAccess)
		resolver := NewResolver(s, func() time.Time { return resolvedAt })

		cookie := http.Header{}
		cookie.Set("Cookie", CookieName+"="+refresh)
		if _, ok := resolver.Grant(t.Context(), cookie); !ok {
			t.Error("a refresh token in a cookie was refused, and that is how a browser signs in")
		}

		bearer := http.Header{}
		bearer.Set("Authorization", "Bearer "+access)
		if _, ok := resolver.Grant(t.Context(), bearer); !ok {
			t.Error("an Access token as a bearer was refused, and that is what it is for")
		}
	})
}
