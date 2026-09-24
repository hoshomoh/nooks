package gateway

import (
	"encoding/json"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"connectrpc.com/connect"

	apiv1 "github.com/hoshomoh/nooks/proto/gen/nooks/api/v1"
	"github.com/hoshomoh/nooks/server/auth"
	v1 "github.com/hoshomoh/nooks/server/router/api/v1"
	"github.com/hoshomoh/nooks/store"
	"github.com/hoshomoh/nooks/store/storetest"
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
	s := storetest.Fresh(t)

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
		Public:   v1.NewPublicService(s, nil),
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

/*
And the other way round: an access token is not a cookie.

The two checks are a pair — a cookie carries the refresh token, a bearer carries the
access token, and each resolver path refuses the other kind. Only the first half had a
test, and the second is one `if` in Member: without it an hour-long credential would
quietly be given a month-long cookie's reach, and the split that exists to keep the
long-lived one out of reach of a script would stop meaning anything.
*/
func TestTheAccessTokenIsNotACookie(t *testing.T) {
	i := newInstance(t)
	access, refresh := i.signIn("a-long-enough-password")

	wearing := &http.Cookie{Name: refresh.Name, Value: access}
	res := i.callWithCookie(http.MethodGet, "/api/v1/lists", wearing, "")
	if res.Code != http.StatusUnauthorized {
		t.Errorf("code = %d, want the access token refused as a cookie", res.Code)
	}

	// The same cookie name carrying the right kind still works, so the refusal above is
	// the kind of token and not the cookie.
	if code := i.callWithCookie(http.MethodGet, "/api/v1/lists", refresh, "").Code; code != http.StatusOK {
		t.Errorf("code = %d, want the refresh cookie to work", code)
	}
}

// passwordHash makes a hash the auth service will verify against.
func passwordHash(plain string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	return string(hash), err
}

/*
The refresh route answers the cookie it documents.

The adapters build the Connect request the services expect, and one that built an empty
one handed every service an empty header. RefreshAccess reads the refresh cookie, so
POST /api/v1/auth/refresh — annotated, adapted, and in the published reference —
answered "not signed in" to a caller holding a good one, always, for everybody.

A REST caller signs in, is handed an access token worth an hour, and needs this to get
the next one. Without it the only way on is to sign in again, which means keeping the
password around: worse than the thing the split exists to avoid.
*/
func TestTheRefreshRouteAnswersTheCookieItDocuments(t *testing.T) {
	i := newInstance(t)
	access, refresh := i.signIn("a-long-enough-password")

	res := i.callWithCookie(http.MethodPost, "/api/v1/auth/refresh", refresh, "{}")
	if res.Code != http.StatusOK {
		t.Fatalf("code = %d, want the documented route to answer: %s",
			res.Code, strings.TrimSpace(res.Body.String()))
	}

	var answered struct {
		AccessToken string `json:"accessToken"`
	}
	if err := json.Unmarshal(res.Body.Bytes(), &answered); err != nil {
		t.Fatalf("read the answer: %v", err)
	}
	if answered.AccessToken == "" {
		t.Fatal("no access token came back")
	}
	if answered.AccessToken == access {
		t.Error("the same access token came back, so refreshing bought nothing")
	}

	// It is a working credential, not just a string.
	if code := i.call(http.MethodGet, "/api/v1/lists", answered.AccessToken, "").Code; code != http.StatusOK {
		t.Errorf("code = %d, want the refreshed token to work", code)
	}
}

// Without the cookie it must still refuse, or the route would be a way in for anybody.
func TestTheRefreshRouteRefusesWithoutTheCookie(t *testing.T) {
	i := newInstance(t)
	i.signIn("a-long-enough-password")

	if code := i.call(http.MethodPost, "/api/v1/auth/refresh", "", "{}").Code; code != http.StatusUnauthorized {
		t.Errorf("code = %d, want unauthorized with no refresh cookie", code)
	}
}

// callWithCookie is call with a cookie instead of a bearer token.
func (i *instance) callWithCookie(method, path string, cookie *http.Cookie, body string) *httptest.ResponseRecorder {
	i.t.Helper()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(cookie)
	res := httptest.NewRecorder()
	i.handler.ServeHTTP(res, req)
	return res
}
