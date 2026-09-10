package v1

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"connectrpc.com/connect"

	"github.com/hoshomoh/nooks/internal/password"
	apiv1 "github.com/hoshomoh/nooks/proto/gen/nooks/api/v1"
	"github.com/hoshomoh/nooks/server/auth"
	"github.com/hoshomoh/nooks/store"
)

const goodPassword = "the good beans from the market"

var testClock = time.Date(2026, time.March, 3, 9, 30, 0, 0, time.UTC)

// newAuthService builds the service over a real SQLite store, with the clock and the
// identifier fixed so that assertions do not depend on chance.
func newAuthService(t *testing.T) (*AuthService, store.Store) {
	t.Helper()

	s, err := store.OpenSQLite(t.Context(), filepath.Join(t.TempDir(), "nooks.db"))
	if err != nil {
		t.Fatalf("OpenSQLite: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })

	// Tokens are numbered rather than fixed: signing in twice must produce two
	// sessions, which a constant token would collide on.
	issued := 0
	svc := NewAuthService(s, AuthServiceOptions{
		Now:    func() time.Time { return testClock },
		NewUID: func() (string, error) { return "mem_test", nil },
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
