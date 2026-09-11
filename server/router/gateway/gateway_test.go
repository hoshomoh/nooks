package gateway

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/hoshomoh/nooks/server/auth"
	v1 "github.com/hoshomoh/nooks/server/router/api/v1"
	"github.com/hoshomoh/nooks/store"
)

var testClock = time.Date(2026, 8, 25, 9, 0, 0, 0, time.UTC)

// instance is an Instance with one Admin, one List, and a REST handler over it.
type instance struct {
	t       *testing.T
	store   store.Store
	handler http.Handler
	member  store.Member
	listUID string
	listID  int64
}

func newInstance(t *testing.T) *instance {
	t.Helper()
	s, err := store.OpenSQLite(t.Context(), filepath.Join(t.TempDir(), "nooks.db"))
	if err != nil {
		t.Fatalf("OpenSQLite: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })

	member, err := s.CreateMember(t.Context(), store.CreateMemberParams{
		UID: "mem_anna", Name: "Anna", Email: "anna@brunnen.lan", Role: store.RoleAdmin,
		PasswordHash: "hash", CreatedAt: testClock,
	})
	if err != nil {
		t.Fatalf("CreateMember: %v", err)
	}
	list, err := s.CreateList(t.Context(), store.CreateListParams{
		UID: "list_groceries", Name: "Groceries", OwnerID: member.ID, At: testClock,
	})
	if err != nil {
		t.Fatalf("CreateList: %v", err)
	}

	now := func() time.Time { return testClock }
	handler, err := New(Services{
		Activity: v1.NewActivityService(s, now),
		Auth:     v1.NewAuthService(s, v1.AuthServiceOptions{Now: now}),
		Instance: v1.NewInstanceService(s),
		List:     v1.NewListService(s, now, nil),
		Member:   v1.NewMemberService(s, now, nil),
		Public:   v1.NewPublicService(s),
		Request:  v1.NewRequestService(s, now),
		Token:    v1.NewTokenService(s, now, nil, nil),
	}, auth.NewResolver(s, now))
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	return &instance{t: t, store: s, handler: handler, member: member, listUID: list.UID, listID: list.ID}
}

// tokenFor cuts an Access token with the given abilities and answers the secret.
func (i *instance) tokenFor(abilities store.TokenAbilities) string {
	i.t.Helper()
	secret, hash, err := auth.NewToken()
	if err != nil {
		i.t.Fatalf("NewToken: %v", err)
	}
	if _, err := i.store.CreateAccessToken(i.t.Context(), store.CreateAccessTokenParams{
		UID: "tok_1", MemberID: i.member.ID, Name: "Script", TokenHash: hash,
		Abilities: abilities, ListIDs: []int64{i.listID}, At: testClock,
	}); err != nil {
		i.t.Fatalf("CreateAccessToken: %v", err)
	}
	return secret
}

// call makes a REST request with a bearer token.
func (i *instance) call(method, path, secret, body string) *httptest.ResponseRecorder {
	var reader *strings.Reader
	if body == "" {
		reader = strings.NewReader("")
	} else {
		reader = strings.NewReader(body)
	}
	req := httptest.NewRequest(method, path, reader)
	if secret != "" {
		req.Header.Set("Authorization", "Bearer "+secret)
	}
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()
	i.handler.ServeHTTP(res, req)
	return res
}

func TestATokenReadsAListOverRest(t *testing.T) {
	i := newInstance(t)
	secret := i.tokenFor(store.TokenAbilities{Read: true})

	res := i.call(http.MethodGet, "/api/v1/lists/"+i.listUID, secret, "")

	if res.Code != http.StatusOK {
		t.Fatalf("code = %d, body = %s", res.Code, res.Body.String())
	}
	var answer struct {
		List struct{ Name string } `json:"list"`
	}
	if err := json.Unmarshal(res.Body.Bytes(), &answer); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if answer.List.Name != "Groceries" {
		t.Errorf("list name = %q, want Groceries", answer.List.Name)
	}
}

// The property the whole design rests on: the same rules, whichever door.
func TestAReadTokenIsRefusedAWriteOverRest(t *testing.T) {
	i := newInstance(t)
	secret := i.tokenFor(store.TokenAbilities{Read: true})

	res := i.call(http.MethodPost, "/api/v1/lists/"+i.listUID+"/items", secret, `{"label":"Milk"}`)

	if res.Code != http.StatusForbidden {
		t.Errorf("code = %d, want 403; body = %s", res.Code, res.Body.String())
	}
}

func TestAWriteTokenAddsAnItemOverRest(t *testing.T) {
	i := newInstance(t)
	secret := i.tokenFor(store.TokenAbilities{Read: true, Write: true})

	res := i.call(http.MethodPost, "/api/v1/lists/"+i.listUID+"/items", secret, `{"label":"Milk"}`)

	if res.Code != http.StatusOK {
		t.Fatalf("code = %d, body = %s", res.Code, res.Body.String())
	}
	if !strings.Contains(res.Body.String(), "Milk") {
		t.Errorf("body = %s, want the item that was added", res.Body.String())
	}
}

// A List a token does not name is invisible, not forbidden — over REST too.
func TestATokenCannotSeePastItsListsOverRest(t *testing.T) {
	i := newInstance(t)
	other, err := i.store.CreateList(t.Context(), store.CreateListParams{
		UID: "list_bike", Name: "Bike", OwnerID: i.member.ID, At: testClock,
	})
	if err != nil {
		t.Fatalf("CreateList: %v", err)
	}
	secret := i.tokenFor(store.TokenAbilities{Read: true, Write: true})

	res := i.call(http.MethodGet, "/api/v1/lists/"+other.UID, secret, "")

	if res.Code != http.StatusNotFound {
		t.Errorf("code = %d, want 404", res.Code)
	}
}

func TestNoTokenIsRefused(t *testing.T) {
	i := newInstance(t)

	if code := i.call(http.MethodGet, "/api/v1/lists", "", "").Code; code != http.StatusUnauthorized {
		t.Errorf("code = %d, want 401", code)
	}
}
