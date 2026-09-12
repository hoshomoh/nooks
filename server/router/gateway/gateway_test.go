package gateway

import (
	"encoding/json"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"connectrpc.com/connect"

	apiv1 "github.com/hoshomoh/nooks/proto/gen/nooks/api/v1"
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
	handler, err := New(v1.Services{
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

// signIn goes through Connect the way a browser does, and answers with both credentials.
func (i *instance) signIn(password string) (access string, refresh *http.Cookie) {
	i.t.Helper()
	hash, err := passwordHash(password)
	if err != nil {
		i.t.Fatalf("hash: %v", err)
	}
	if err := i.store.SetMemberPassword(i.t.Context(), i.member.ID, hash); err != nil {
		i.t.Fatalf("SetMemberPassword: %v", err)
	}

	now := func() time.Time { return testClock }
	svc := v1.NewAuthService(i.store, v1.AuthServiceOptions{Now: now})
	res, err := svc.SignIn(i.t.Context(), connect.NewRequest(&apiv1.SignInRequest{
		Email: i.member.Email, Password: password,
	}))
	if err != nil {
		i.t.Fatalf("SignIn: %v", err)
	}

	for _, raw := range res.Header().Values("Set-Cookie") {
		parsed := (&http.Response{Header: http.Header{"Set-Cookie": []string{raw}}}).Cookies()
		if len(parsed) > 0 && parsed[0].Name == auth.CookieName {
			refresh = parsed[0]
		}
	}
	return res.Msg.GetAccessToken(), refresh
}

// The credential a caller can actually lose travels; the one that lasts a month does not.
func TestSigningInHandsOverAShortLivedToken(t *testing.T) {
	i := newInstance(t)

	access, refresh := i.signIn("a-long-enough-password")

	if access == "" {
		t.Fatal("no access token was handed over")
	}
	if refresh == nil || !refresh.HttpOnly {
		t.Fatal("the refresh token is not in an HttpOnly cookie")
	}
	if access == refresh.Value {
		t.Error("the access token is the refresh token; the split buys nothing")
	}

	// It works as a bearer, which is the point of handing it over.
	if code := i.call(http.MethodGet, "/api/v1/lists", access, "").Code; code != http.StatusOK {
		t.Errorf("code = %d, want the access token to work", code)
	}
}

// Otherwise the long-lived credential would be usable exactly where the short-lived one
// was supposed to be, and the split would buy nothing.
func TestTheRefreshTokenIsNotABearer(t *testing.T) {
	i := newInstance(t)

	_, refresh := i.signIn("a-long-enough-password")

	if code := i.call(http.MethodGet, "/api/v1/lists", refresh.Value, "").Code; code != http.StatusUnauthorized {
		t.Errorf("code = %d, want the refresh token refused as a bearer", code)
	}
}

// passwordHash makes a hash the auth service will verify against.
func passwordHash(plain string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	return string(hash), err
}
