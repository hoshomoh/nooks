package backup

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/hoshomoh/nooks/server/auth"
	"github.com/hoshomoh/nooks/store"
)

var testClock = time.Date(2026, 8, 25, 9, 0, 0, 0, time.UTC)

// served is a handler and the ways of reaching it: a browser's session, and an Access
// token its Member cut for a script.
type served struct {
	handler *Handler
	session string
	token   string
}

// serve builds a handler over a store with one Member of the given role.
func serve(t *testing.T, role store.Role) (*Handler, string) {
	t.Helper()
	data := t.TempDir()
	s, err := store.OpenSQLite(t.Context(), filepath.Join(data, "nooks.db"))
	if err != nil {
		t.Fatalf("OpenSQLite: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })

	member, err := s.CreateMember(t.Context(), store.CreateMemberParams{
		UID: "mem_anna", Name: "Anna", Email: "anna@brunnen.lan", Role: role,
		PasswordHash: "hash", CreatedAt: testClock,
	})
	if err != nil {
		t.Fatalf("CreateMember: %v", err)
	}

	token, hash, err := auth.NewToken()
	if err != nil {
		t.Fatalf("NewToken: %v", err)
	}
	if err := s.CreateSession(t.Context(), store.Session{
		TokenHash: hash, MemberID: member.ID,
		CreatedAt: testClock, ExpiresAt: testClock.Add(time.Hour),
	}); err != nil {
		t.Fatalf("CreateSession: %v", err)
	}

	resolver := auth.NewResolver(s, func() time.Time { return testClock })
	// The snapshot is written beside the database, so the test gives it the same
	// directory the store was opened in rather than letting it fall back to the
	// system's.
	return NewHandler(s, resolver, data, func() time.Time { return testClock }), token
}

// serveWithToken is serve, plus an Access token the Member cut for themselves.
func serveWithToken(t *testing.T, role store.Role, abilities store.TokenAbilities) served {
	t.Helper()
	handler, session := serve(t, role)

	secret, hash, err := auth.NewToken()
	if err != nil {
		t.Fatalf("NewToken: %v", err)
	}
	member, err := handler.store.MemberByUID(t.Context(), "mem_anna")
	if err != nil {
		t.Fatalf("MemberByUID: %v", err)
	}
	if _, err := handler.store.CreateAccessToken(t.Context(), store.CreateAccessTokenParams{
		UID: "tok_1", MemberID: member.ID, Name: "Kitchen tablet", TokenHash: hash,
		Abilities: abilities, AllLists: true, At: testClock,
	}); err != nil {
		t.Fatalf("CreateAccessToken: %v", err)
	}
	return served{handler: handler, session: session, token: secret}
}

// askWithToken makes the request a script would, carrying an Access token.
func askWithToken(h *Handler, token string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/backup", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	res := httptest.NewRecorder()
	h.ServeHTTP(res, req)
	return res
}

// ask makes the request a browser would, with or without a session.
func ask(h *Handler, session string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/backup", nil)
	if session != "" {
		req.AddCookie(&http.Cookie{Name: auth.CookieName, Value: session})
	}
	res := httptest.NewRecorder()
	h.ServeHTTP(res, req)
	return res
}

func TestAnAdminDownloadsTheDatabase(t *testing.T) {
	h, session := serve(t, store.RoleAdmin)

	res := ask(h, session)

	if res.Code != http.StatusOK {
		t.Fatalf("code = %d, want 200", res.Code)
	}
	if got := res.Header().Get("Content-Disposition"); got != `attachment; filename="nooks-2026-08-25.db"` {
		t.Errorf("Content-Disposition = %q", got)
	}
	// A real SQLite file, which is the whole point of exporting it this way.
	if got := res.Body.Bytes(); len(got) < 16 || string(got[:15]) != "SQLite format 3" {
		t.Error("what came back is not a SQLite database")
	}
	// Never a cached copy of somebody's whole household.
	if got := res.Header().Get("Cache-Control"); got != "no-store" {
		t.Errorf("Cache-Control = %q, want no-store", got)
	}
}

// It holds every password hash and every List in the household.
func TestAMemberCannotDownloadTheDatabase(t *testing.T) {
	h, session := serve(t, store.RoleMember)

	if code := ask(h, session).Code; code != http.StatusNotFound {
		t.Errorf("code = %d, want 404 for a Member", code)
	}
}

// Whether this Instance has a backup worth finding is not an anonymous request's to
// learn, so it is answered the same way as a route that does not exist.
func TestAnonymousGetsNothing(t *testing.T) {
	h, _ := serve(t, store.RoleAdmin)

	if code := ask(h, "").Code; code != http.StatusNotFound {
		t.Errorf("code = %d, want 404", code)
	}
}

/*
An Access token cut by an Admin does not download the database.

A token is that Member's access deliberately narrowed, and this file holds every List in
the household and every Member's password hash. It is the one thing no narrowing
survives, so the door asks for a browser rather than only for an Admin — even a token
that may reach every List and do everything to them.

The route is not an RPC, so neither the Connect guards nor the parity test cover it. It
answered a token for as long as it existed.
*/
func TestAnAdminsAccessTokenIsRefused(t *testing.T) {
	it := serveWithToken(t, store.RoleAdmin, store.TokenAbilities{Read: true, Write: true, Delete: true})

	res := askWithToken(it.handler, it.token)

	if res.Code != http.StatusNotFound {
		t.Errorf("code = %d, want 404 for an Access token", res.Code)
	}
	if res.Body.Len() > 64 {
		t.Errorf("answered %d bytes, want nothing that looks like a database", res.Body.Len())
	}

	// The same Admin, in the browser they signed into, still gets it.
	if allowed := ask(it.handler, it.session); allowed.Code != http.StatusOK {
		t.Errorf("code = %d for the browser, want 200", allowed.Code)
	}
}

/*
The snapshot is written beside the database, not in the system's temp directory.

VACUUM INTO writes a complete second copy, so exporting a household of any size needs
that much space somewhere. `os.MkdirTemp("")` follows TMPDIR, which on several
distributions is tmpfs and therefore RAM, and in the shipped image is the container's
own layer rather than the volume somebody mounted. The data directory is sized for the
database by definition, because it is already holding it.

Asserted by pointing the handler at a directory that does not exist. If it honours that,
making the snapshot directory fails and the export answers 500; if it ignores it and
reaches for the system's, the export succeeds and this is how we find out. Checking the
data directory afterwards would prove nothing, because the handler cleans up on the way
out and an empty directory looks the same either way.

My first version of this test did exactly that and passed over nothing.
*/
func TestTheSnapshotIsWrittenBesideTheDatabase(t *testing.T) {
	data := t.TempDir()
	s, err := store.OpenSQLite(t.Context(), filepath.Join(data, "nooks.db"))
	if err != nil {
		t.Fatalf("OpenSQLite: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })

	member, err := s.CreateMember(t.Context(), store.CreateMemberParams{
		UID: "mem_anna", Name: "Anna", Email: "anna@brunnen.lan",
		Role: store.RoleAdmin, PasswordHash: "hash", CreatedAt: testClock,
	})
	if err != nil {
		t.Fatalf("CreateMember: %v", err)
	}
	token, hash, err := auth.NewToken()
	if err != nil {
		t.Fatalf("NewToken: %v", err)
	}
	if err := s.CreateSession(t.Context(), store.Session{
		TokenHash: hash, MemberID: member.ID,
		CreatedAt: testClock, ExpiresAt: testClock.Add(time.Hour),
	}); err != nil {
		t.Fatalf("CreateSession: %v", err)
	}

	resolver := auth.NewResolver(s, func() time.Time { return testClock })
	nowhere := filepath.Join(data, "not-a-directory")
	handler := NewHandler(s, resolver, nowhere, func() time.Time { return testClock })

	request := httptest.NewRequest(http.MethodGet, "/api/v1/backup", nil)
	request.AddCookie(&http.Cookie{Name: auth.CookieName, Value: token})
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)

	if recorder.Code == http.StatusOK {
		t.Error("the export succeeded against a data directory that does not exist, " +
			"so the snapshot went somewhere else")
	}
}
