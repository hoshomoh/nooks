package v1

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"connectrpc.com/connect"

	"github.com/hoshomoh/nooks/internal/password"
	apiv1 "github.com/hoshomoh/nooks/proto/gen/nooks/api/v1"
	"github.com/hoshomoh/nooks/server/auth"
	"github.com/hoshomoh/nooks/store"
	"github.com/hoshomoh/nooks/store/storetest"
)

const goodPassword = "the good beans from the market"

var testClock = time.Date(2026, time.March, 3, 9, 30, 0, 0, time.UTC)

// newAuthService builds the service over a real SQLite store, with the clock and the
// identifier fixed so that assertions do not depend on chance.
func newAuthService(t *testing.T) (*AuthService, store.Store) {
	t.Helper()

	s := storetest.Fresh(t)

	// Tokens are numbered rather than fixed: signing in twice must produce two
	// sessions, which a constant token would collide on.
	issued := 0
	uids := 0
	svc := NewAuthService(s, AuthServiceOptions{
		Now: func() time.Time { return testClock },
		NewUID: func() (string, error) {
			uids++
			if uids == 1 {
				// The first Member is the one tests look up by name.
				return "mem_test", nil
			}
			return fmt.Sprintf("uid-%d", uids), nil
		},
		NewToken: func() (string, string, error) {
			issued++
			token := fmt.Sprintf("token-%d", issued)
			return token, auth.HashToken(token), nil
		},
	})
	return svc, s
}

// completeSetup runs first run and returns the response.
func completeSetup(t *testing.T, svc *AuthService) *connect.Response[apiv1.CompleteSetupResponse] {
	t.Helper()
	res, err := svc.CompleteSetup(t.Context(), connect.NewRequest(&apiv1.CompleteSetupRequest{
		Name:         "Anna",
		Email:        "anna@brunnen.lan",
		Password:     goodPassword,
		InstanceName: "Brunnen Street",
	}))
	if err != nil {
		t.Fatalf("CompleteSetup: %v", err)
	}
	return res
}

func TestCompleteSetupCreatesTheFirstAdmin(t *testing.T) {
	svc, s := newAuthService(t)

	res := completeSetup(t, svc)

	if got := res.Msg.GetMember().GetRole(); got != apiv1.Role_ROLE_ADMIN {
		t.Errorf("role = %v, want ROLE_ADMIN — whoever sets the Instance up is its Admin", got)
	}
	if res.Msg.GetMember().GetMustChangePassword() {
		t.Error("MustChangePassword = true, want false: they chose this password themselves")
	}

	settings, err := s.InstanceSettings(t.Context())
	if err != nil {
		t.Fatalf("InstanceSettings: %v", err)
	}
	if settings.Name != "Brunnen Street" {
		t.Errorf("instance name = %q, want %q", settings.Name, "Brunnen Street")
	}
	if settings.NeedsSetup() {
		t.Error("NeedsSetup() = true after first run")
	}
}

func TestCompleteSetupIssuesASessionCookie(t *testing.T) {
	svc, _ := newAuthService(t)

	res := completeSetup(t, svc)

	cookie := res.Header().Get("Set-Cookie")
	if !strings.Contains(cookie, auth.CookieName+"=token-1") {
		t.Fatalf("Set-Cookie = %q, want it to carry the session token", cookie)
	}
	if !strings.Contains(cookie, "HttpOnly") {
		t.Error("session cookie is not HttpOnly, so a script bug becomes an account takeover")
	}
	if !strings.Contains(cookie, "SameSite=Lax") {
		t.Error("session cookie is not SameSite=Lax")
	}
}

// Once an Instance has an Admin this must close permanently, or whoever reaches it
// next could seize it.
func TestCompleteSetupCannotHappenTwice(t *testing.T) {
	svc, _ := newAuthService(t)
	completeSetup(t, svc)

	_, err := svc.CompleteSetup(t.Context(), connect.NewRequest(&apiv1.CompleteSetupRequest{
		Name: "Mallory", Email: "m@example.com", Password: goodPassword, InstanceName: "Mine Now",
	}))
	if got := connect.CodeOf(err); got != connect.CodeFailedPrecondition {
		t.Errorf("second CompleteSetup code = %v, want failed_precondition", got)
	}
}

/*
Two people completing first run at once, and only one Instance to own.

checkSetupIsOpen counts Members and finds none, then bcrypt hashes a password, which is
deliberately slow. Two requests arriving inside those couple of hundred milliseconds
both passed the count, and with different emails both became Admins. Losing was silent:
somebody told the Instance is already set up redeploys, where somebody who completes
setup has no reason to look at the members list.

It is the one moment an Instance is defenceless, because the setup page answers anybody
until it is used. Something scanning the internet that reaches a fresh deployment inside
that window used to become an Admin of it.

Run concurrently rather than in sequence, because in sequence it passed before the fix:
the second caller found a Member and was refused for the ordinary reason. The race is
the test.
*/
func TestOnlyOnePersonCanCompleteFirstRun(t *testing.T) {
	svc, s := newAuthService(t)

	const racers = 4
	start := make(chan struct{})
	answers := make(chan error, racers)
	var running sync.WaitGroup

	for i := range racers {
		running.Add(1)
		go func() {
			defer running.Done()
			<-start
			_, err := svc.CompleteSetup(t.Context(), connect.NewRequest(&apiv1.CompleteSetupRequest{
				Name:         fmt.Sprintf("Owner %d", i),
				Email:        fmt.Sprintf("owner%d@brunnen.lan", i),
				Password:     goodPassword,
				InstanceName: fmt.Sprintf("Instance %d", i),
			}))
			answers <- err
		}()
	}

	close(start)
	running.Wait()
	close(answers)

	won := 0
	for err := range answers {
		if err == nil {
			won++
			continue
		}
		if got := connect.CodeOf(err); got != connect.CodeFailedPrecondition {
			t.Errorf("a loser was refused with %v, want failed_precondition", got)
		}
	}
	if won != 1 {
		t.Errorf("%d of %d completed first run, want exactly one", won, racers)
	}

	members, err := s.Members(t.Context())
	if err != nil {
		t.Fatalf("Members: %v", err)
	}
	if len(members) != 1 {
		t.Errorf("the Instance has %d Admins, want the one who got there first", len(members))
	}
}

func TestCompleteSetupValidatesItsInput(t *testing.T) {
	cases := []struct {
		name string
		req  *apiv1.CompleteSetupRequest
	}{
		{"no name", &apiv1.CompleteSetupRequest{Email: "a@b.lan", Password: goodPassword, InstanceName: "X"}},
		{"no email", &apiv1.CompleteSetupRequest{Name: "Anna", Password: goodPassword, InstanceName: "X"}},
		{"no instance name", &apiv1.CompleteSetupRequest{Name: "Anna", Email: "a@b.lan", Password: goodPassword}},
		{"short password", &apiv1.CompleteSetupRequest{Name: "Anna", Email: "a@b.lan", Password: "short", InstanceName: "X"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc, _ := newAuthService(t)
			_, err := svc.CompleteSetup(t.Context(), connect.NewRequest(tc.req))
			if got := connect.CodeOf(err); got != connect.CodeInvalidArgument {
				t.Errorf("code = %v, want invalid_argument", got)
			}
		})
	}
}

func TestSignIn(t *testing.T) {
	svc, _ := newAuthService(t)
	completeSetup(t, svc)

	res, err := svc.SignIn(t.Context(), connect.NewRequest(&apiv1.SignInRequest{
		Email: "anna@brunnen.lan", Password: goodPassword,
	}))
	if err != nil {
		t.Fatalf("SignIn: %v", err)
	}
	if res.Msg.GetMember().GetName() != "Anna" {
		t.Errorf("member = %q, want Anna", res.Msg.GetMember().GetName())
	}
	if !strings.Contains(res.Header().Get("Set-Cookie"), auth.CookieName) {
		t.Error("SignIn did not set a session cookie")
	}
}

// A wrong password and an unknown email must be indistinguishable, or sign-in becomes
// a way to discover who has an account.
func TestSignInFailsIdenticallyForWrongPasswordAndUnknownEmail(t *testing.T) {
	svc, _ := newAuthService(t)
	completeSetup(t, svc)

	_, wrongPassword := svc.SignIn(t.Context(), connect.NewRequest(&apiv1.SignInRequest{
		Email: "anna@brunnen.lan", Password: "not the right password",
	}))
	_, unknownEmail := svc.SignIn(t.Context(), connect.NewRequest(&apiv1.SignInRequest{
		Email: "stranger@example.com", Password: goodPassword,
	}))

	if connect.CodeOf(wrongPassword) != connect.CodeUnauthenticated {
		t.Errorf("wrong password code = %v, want unauthenticated", connect.CodeOf(wrongPassword))
	}
	if wrongPassword.Error() != unknownEmail.Error() {
		t.Errorf("the two failures differ:\n  wrong password: %v\n  unknown email:  %v",
			wrongPassword, unknownEmail)
	}
}

func TestSignInRecordsWhenItHappened(t *testing.T) {
	svc, s := newAuthService(t)
	completeSetup(t, svc)

	if _, err := svc.SignIn(t.Context(), connect.NewRequest(&apiv1.SignInRequest{
		Email: "anna@brunnen.lan", Password: goodPassword,
	})); err != nil {
		t.Fatalf("SignIn: %v", err)
	}

	member, err := s.MemberByUID(t.Context(), "mem_test")
	if err != nil {
		t.Fatalf("MemberByUID: %v", err)
	}
	if !member.LastSignedInAt.Equal(testClock) {
		t.Errorf("LastSignedInAt = %v, want %v", member.LastSignedInAt, testClock)
	}
}

func TestSignOutClearsTheCookieAndTheSession(t *testing.T) {
	svc, s := newAuthService(t)
	completeSetup(t, svc)

	req := connect.NewRequest(&apiv1.SignOutRequest{})
	req.Header().Set("Cookie", auth.CookieName+"=token-1")

	res, err := svc.SignOut(t.Context(), req)
	if err != nil {
		t.Fatalf("SignOut: %v", err)
	}
	if !strings.Contains(res.Header().Get("Set-Cookie"), "Max-Age=0") {
		t.Errorf("Set-Cookie = %q, want it to clear the cookie", res.Header().Get("Set-Cookie"))
	}
	if _, err := s.SessionByTokenHash(t.Context(), auth.HashToken("token-1")); err == nil {
		t.Error("the session is still in the store after signing out")
	}
}

// Signing out without a session is not a failure.
func TestSignOutWithoutASession(t *testing.T) {
	svc, _ := newAuthService(t)
	if _, err := svc.SignOut(t.Context(), connect.NewRequest(&apiv1.SignOutRequest{})); err != nil {
		t.Errorf("SignOut with no cookie = %v, want nil", err)
	}
}

func TestGetCurrentMemberRequiresASession(t *testing.T) {
	svc, _ := newAuthService(t)
	_, err := svc.GetCurrentMember(t.Context(), connect.NewRequest(&apiv1.GetCurrentMemberRequest{}))
	if got := connect.CodeOf(err); got != connect.CodeUnauthenticated {
		t.Errorf("code = %v, want unauthenticated", got)
	}
}

func TestReplacePasswordClearsMustChange(t *testing.T) {
	svc, s := newAuthService(t)

	// A Member an Admin created with a temporary password.
	hash, err := password.Hash("a temporary password")
	if err != nil {
		t.Fatalf("Hash: %v", err)
	}
	member, err := s.CreateMember(t.Context(), store.CreateMemberParams{
		UID: "mem_ruth", Name: "Ruth", Email: "ruth@brunnen.lan", Role: store.RoleMember,
		PasswordHash: hash, MustChangePassword: true, CreatedAt: testClock,
	})
	if err != nil {
		t.Fatalf("CreateMember: %v", err)
	}

	ctx := auth.WithMember(t.Context(), member)
	res, err := svc.ReplacePassword(ctx, connect.NewRequest(&apiv1.ReplacePasswordRequest{
		CurrentPassword: "a temporary password",
		NewPassword:     "one she picked herself",
	}))
	if err != nil {
		t.Fatalf("ReplacePassword: %v", err)
	}
	if res.Msg.GetMember().GetMustChangePassword() {
		t.Error("MustChangePassword is still true after the password was replaced")
	}

	updated, err := s.MemberByID(t.Context(), member.ID)
	if err != nil {
		t.Fatalf("MemberByID: %v", err)
	}
	if err := password.Verify(updated.PasswordHash, "one she picked herself"); err != nil {
		t.Errorf("the new password does not verify: %v", err)
	}
}

// The current password is required even when it was a temporary one, so an unattended
// browser cannot be used to take the account over.
func TestReplacePasswordNeedsTheCurrentOne(t *testing.T) {
	svc, s := newAuthService(t)
	completeSetup(t, svc)

	member, err := s.MemberByUID(t.Context(), "mem_test")
	if err != nil {
		t.Fatalf("MemberByUID: %v", err)
	}

	ctx := auth.WithMember(t.Context(), member)
	_, err = svc.ReplacePassword(ctx, connect.NewRequest(&apiv1.ReplacePasswordRequest{
		CurrentPassword: "not their password",
		NewPassword:     "a perfectly fine new one",
	}))
	if got := connect.CodeOf(err); got != connect.CodeInvalidArgument {
		t.Errorf("code = %v, want invalid_argument", got)
	}
}

func TestReplacePasswordEnforcesTheLengthRule(t *testing.T) {
	svc, s := newAuthService(t)
	completeSetup(t, svc)

	member, err := s.MemberByUID(t.Context(), "mem_test")
	if err != nil {
		t.Fatalf("MemberByUID: %v", err)
	}

	_, err = svc.ReplacePassword(auth.WithMember(t.Context(), member),
		connect.NewRequest(&apiv1.ReplacePasswordRequest{
			CurrentPassword: goodPassword,
			NewPassword:     "short",
		}))
	if got := connect.CodeOf(err); got != connect.CodeInvalidArgument {
		t.Errorf("code = %v, want invalid_argument", got)
	}
}

func TestReplacePasswordRequiresASession(t *testing.T) {
	svc, _ := newAuthService(t)
	_, err := svc.ReplacePassword(context.Background(),
		connect.NewRequest(&apiv1.ReplacePasswordRequest{}))
	if got := connect.CodeOf(err); got != connect.CodeUnauthenticated {
		t.Errorf("code = %v, want unauthenticated", got)
	}
}

// signedInAs puts a Member behind a real session and answers the context that browser
// would arrive with, plus the stored form of its cookie.
func signedInAs(t *testing.T, s store.Store, member store.Member) (context.Context, string) {
	t.Helper()
	token, hash, err := auth.NewToken()
	if err != nil {
		t.Fatalf("NewToken: %v", err)
	}
	if err := s.CreateSession(t.Context(), store.Session{
		TokenHash: hash, MemberID: member.ID,
		CreatedAt: testClock, ExpiresAt: testClock.Add(24 * time.Hour),
	}); err != nil {
		t.Fatalf("CreateSession: %v", err)
	}
	_ = token
	return auth.WithGrant(t.Context(), auth.Grant{Member: member, Session: hash}), hash
}

// withPassword makes a Member who knows their own password.
func withPassword(t *testing.T, s store.Store, plain string) store.Member {
	t.Helper()
	hash, err := password.Hash(plain)
	if err != nil {
		t.Fatalf("Hash: %v", err)
	}
	member, err := s.CreateMember(t.Context(), store.CreateMemberParams{
		UID: "mem_ruth", Name: "Ruth", Email: "ruth@brunnen.lan", Role: store.RoleMember,
		PasswordHash: hash, CreatedAt: testClock,
	})
	if err != nil {
		t.Fatalf("CreateMember: %v", err)
	}
	return member
}

/*
Changing a password signs every other browser out.

It is what somebody does when they think another device has their account, so a password
change that leaves those sessions alive makes the one remedy they reach for do nothing.
The browser doing the changing keeps its session: it has just proved it knows the old
password, and answering a settings change with a sign-in page is not a remedy either.
*/
func TestReplacingAPasswordEndsEveryOtherBrowser(t *testing.T) {
	svc, s := newAuthService(t)
	member := withPassword(t, s, "the one she has now")

	mine, mineHash := signedInAs(t, s, member)
	_, theirsHash := signedInAs(t, s, member)

	if _, err := svc.ReplacePassword(mine, connect.NewRequest(&apiv1.ReplacePasswordRequest{
		CurrentPassword: "the one she has now",
		NewPassword:     "one nobody else knows",
	})); err != nil {
		t.Fatalf("ReplacePassword: %v", err)
	}

	if _, err := s.SessionByTokenHash(t.Context(), theirsHash); !errors.Is(err, store.ErrNotFound) {
		t.Errorf("the other browser is still signed in: %v", err)
	}
	if _, err := s.SessionByTokenHash(t.Context(), mineHash); err != nil {
		t.Errorf("the browser that changed the password was signed out: %v", err)
	}
}
