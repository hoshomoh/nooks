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

// serve builds a handler over a store with one Member of the given role.
func serve(t *testing.T, role store.Role) (*Handler, string) {
	t.Helper()
	s, err := store.OpenSQLite(t.Context(), filepath.Join(t.TempDir(), "nooks.db"))
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
	return NewHandler(s, resolver, func() time.Time { return testClock }), token
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
