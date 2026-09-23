package v1

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"connectrpc.com/connect"

	apiv1 "github.com/hoshomoh/nooks/proto/gen/nooks/api/v1"
	"github.com/hoshomoh/nooks/store"
)

// askToJoin sends one request under a chosen email, and hands back what the Visitor was
// told along with the identifier, so a refusal can be read as well as a success.
func askToJoin(t *testing.T, svc *AuthService, email string) (string, error) {
	t.Helper()
	res, err := svc.RequestJoin(t.Context(), connect.NewRequest(&apiv1.RequestJoinRequest{
		Name: "Til", Email: email, Message: "It's Til, from upstairs",
	}))
	if err != nil {
		return "", err
	}
	return res.Msg.GetRequestUid(), nil
}

// requestJoin asks for an account under one address and returns the request identifier.
// One address may only have one request waiting, so a test wanting two passes two.
func requestJoin(t *testing.T, svc *AuthService, email string) string {
	t.Helper()
	uid, err := askToJoin(t, svc, email)
	if err != nil {
		t.Fatalf("RequestJoin: %v", err)
	}
	return uid
}

func TestJoinRequestWaitsUntilApproved(t *testing.T) {
	svc, s := newAuthService(t)
	completeSetup(t, svc)
	uid := requestJoin(t, svc, "til@example.com")

	before, err := svc.GetJoinRequest(t.Context(), connect.NewRequest(&apiv1.GetJoinRequestRequest{RequestUid: uid}))
	if err != nil {
		t.Fatalf("GetJoinRequest: %v", err)
	}
	if before.Msg.GetStatus() != apiv1.RequestStatus_REQUEST_STATUS_PENDING {
		t.Errorf("status = %v, want pending", before.Msg.GetStatus())
	}

	if err := s.DecideJoinRequest(t.Context(), uid, store.StatusApproved, testClock); err != nil {
		t.Fatalf("DecideJoinRequest: %v", err)
	}

	after, err := svc.GetJoinRequest(t.Context(), connect.NewRequest(&apiv1.GetJoinRequestRequest{RequestUid: uid}))
	if err != nil {
		t.Fatalf("GetJoinRequest: %v", err)
	}
	if after.Msg.GetStatus() != apiv1.RequestStatus_REQUEST_STATUS_APPROVED {
		t.Errorf("status = %v, want approved", after.Msg.GetStatus())
	}
	if after.Msg.GetEmail() != "til@example.com" {
		t.Errorf("email = %q, want it carried back so Til does not retype it", after.Msg.GetEmail())
	}
}

// An identifier that never existed must look exactly like one that is pending, or this
// becomes a way to probe the Instance.
func TestUnknownJoinRequestReadsAsPending(t *testing.T) {
	svc, _ := newAuthService(t)
	res, err := svc.GetJoinRequest(t.Context(),
		connect.NewRequest(&apiv1.GetJoinRequestRequest{RequestUid: "never-existed"}))
	if err != nil {
		t.Fatalf("GetJoinRequest: %v", err)
	}
	if res.Msg.GetStatus() != apiv1.RequestStatus_REQUEST_STATUS_PENDING {
		t.Errorf("status = %v, want pending", res.Msg.GetStatus())
	}
}

// Ignoring is silent: the sender is never told they were turned down.
func TestIgnoredJoinRequestReadsAsPending(t *testing.T) {
	svc, s := newAuthService(t)
	completeSetup(t, svc)
	uid := requestJoin(t, svc, "til@example.com")

	if err := s.DecideJoinRequest(t.Context(), uid, store.StatusIgnored, testClock); err != nil {
		t.Fatalf("DecideJoinRequest: %v", err)
	}

	res, err := svc.GetJoinRequest(t.Context(), connect.NewRequest(&apiv1.GetJoinRequestRequest{RequestUid: uid}))
	if err != nil {
		t.Fatalf("GetJoinRequest: %v", err)
	}
	if res.Msg.GetStatus() != apiv1.RequestStatus_REQUEST_STATUS_PENDING {
		t.Errorf("status = %v, want pending — an ignored request must not be distinguishable", res.Msg.GetStatus())
	}
}

func TestCompleteJoinCreatesTheAccount(t *testing.T) {
	svc, s := newAuthService(t)
	completeSetup(t, svc)
	uid := requestJoin(t, svc, "til@example.com")
	if err := s.DecideJoinRequest(t.Context(), uid, store.StatusApproved, testClock); err != nil {
		t.Fatalf("DecideJoinRequest: %v", err)
	}

	res, err := svc.CompleteJoin(t.Context(), connect.NewRequest(&apiv1.CompleteJoinRequest{
		RequestUid: uid, Name: "Til", Password: goodPassword,
	}))
	if err != nil {
		t.Fatalf("CompleteJoin: %v", err)
	}
	if res.Msg.GetMember().GetRole() != apiv1.Role_ROLE_MEMBER {
		t.Errorf("role = %v, want ROLE_MEMBER — joining does not make an Admin", res.Msg.GetMember().GetRole())
	}
	if res.Header().Get("Set-Cookie") == "" {
		t.Error("CompleteJoin did not sign the new Member in")
	}
}

// One approval creates one account.
func TestCompleteJoinCannotBeReplayed(t *testing.T) {
	svc, s := newAuthService(t)
	completeSetup(t, svc)
	uid := requestJoin(t, svc, "til@example.com")
	if err := s.DecideJoinRequest(t.Context(), uid, store.StatusApproved, testClock); err != nil {
		t.Fatalf("DecideJoinRequest: %v", err)
	}
	if _, err := svc.CompleteJoin(t.Context(), connect.NewRequest(&apiv1.CompleteJoinRequest{
		RequestUid: uid, Name: "Til", Password: goodPassword,
	})); err != nil {
		t.Fatalf("first CompleteJoin: %v", err)
	}

	_, err := svc.CompleteJoin(t.Context(), connect.NewRequest(&apiv1.CompleteJoinRequest{
		RequestUid: uid, Name: "Til Again", Password: goodPassword,
	}))
	if got := connect.CodeOf(err); got != connect.CodeFailedPrecondition {
		t.Errorf("second CompleteJoin code = %v, want failed_precondition", got)
	}
}

func TestCompleteJoinNeedsAnApproval(t *testing.T) {
	svc, _ := newAuthService(t)
	completeSetup(t, svc)
	uid := requestJoin(t, svc, "til@example.com")

	_, err := svc.CompleteJoin(t.Context(), connect.NewRequest(&apiv1.CompleteJoinRequest{
		RequestUid: uid, Name: "Til", Password: goodPassword,
	}))
	if got := connect.CodeOf(err); got != connect.CodeFailedPrecondition {
		t.Errorf("code = %v, want failed_precondition while still pending", got)
	}
}

// Asking about an account that does not exist must look the same as asking about one
// that does.
func TestPasswordResetDoesNotRevealWhoHasAnAccount(t *testing.T) {
	svc, _ := newAuthService(t)
	completeSetup(t, svc)

	real, err := svc.RequestPasswordReset(t.Context(),
		connect.NewRequest(&apiv1.RequestPasswordResetRequest{EmailOrName: "anna@brunnen.lan"}))
	if err != nil {
		t.Fatalf("RequestPasswordReset for a Member: %v", err)
	}
	fake, err := svc.RequestPasswordReset(t.Context(),
		connect.NewRequest(&apiv1.RequestPasswordResetRequest{EmailOrName: "nobody@example.com"}))
	if err != nil {
		t.Fatalf("RequestPasswordReset for a stranger: %v", err)
	}

	if real.Msg.GetRequestUid() == "" || fake.Msg.GetRequestUid() == "" {
		t.Fatal("both requests should return an identifier")
	}

	// And checking back on either says the same thing.
	for name, uid := range map[string]string{"real": real.Msg.GetRequestUid(), "fake": fake.Msg.GetRequestUid()} {
		res, err := svc.GetResetRequest(t.Context(),
			connect.NewRequest(&apiv1.GetResetRequestRequest{RequestUid: uid}))
		if err != nil {
			t.Fatalf("GetResetRequest(%s): %v", name, err)
		}
		if res.Msg.GetStatus() != apiv1.RequestStatus_REQUEST_STATUS_PENDING {
			t.Errorf("GetResetRequest(%s) = %v, want pending", name, res.Msg.GetStatus())
		}
	}
}

func TestCompletePasswordResetAfterApproval(t *testing.T) {
	svc, s := newAuthService(t)
	completeSetup(t, svc)

	res, err := svc.RequestPasswordReset(t.Context(),
		connect.NewRequest(&apiv1.RequestPasswordResetRequest{EmailOrName: "anna@brunnen.lan"}))
	if err != nil {
		t.Fatalf("RequestPasswordReset: %v", err)
	}
	uid := res.Msg.GetRequestUid()
	if err := s.DecideResetRequest(t.Context(), uid, store.StatusApproved, testClock); err != nil {
		t.Fatalf("DecideResetRequest: %v", err)
	}

	if _, err := svc.CompletePasswordReset(t.Context(),
		connect.NewRequest(&apiv1.CompletePasswordResetRequest{
			RequestUid: uid, NewPassword: "a brand new password",
		})); err != nil {
		t.Fatalf("CompletePasswordReset: %v", err)
	}

	// The new password works and the old one does not.
	if _, err := svc.SignIn(t.Context(), connect.NewRequest(&apiv1.SignInRequest{
		Email: "anna@brunnen.lan", Password: "a brand new password",
	})); err != nil {
		t.Errorf("SignIn with the new password: %v", err)
	}
	if _, err := svc.SignIn(t.Context(), connect.NewRequest(&apiv1.SignInRequest{
		Email: "anna@brunnen.lan", Password: goodPassword,
	})); err == nil {
		t.Error("the old password still works")
	}
}

// An approval nobody used within the hour stops working.
func TestResetApprovalExpires(t *testing.T) {
	svc, s := newAuthService(t)
	completeSetup(t, svc)

	res, err := svc.RequestPasswordReset(t.Context(),
		connect.NewRequest(&apiv1.RequestPasswordResetRequest{EmailOrName: "anna@brunnen.lan"}))
	if err != nil {
		t.Fatalf("RequestPasswordReset: %v", err)
	}
	uid := res.Msg.GetRequestUid()
	// Approved an hour and a half before the service's clock reads.
	if err := s.DecideResetRequest(t.Context(), uid, store.StatusApproved,
		testClock.Add(-90*time.Minute)); err != nil {
		t.Fatalf("DecideResetRequest: %v", err)
	}

	status, err := svc.GetResetRequest(t.Context(),
		connect.NewRequest(&apiv1.GetResetRequestRequest{RequestUid: uid}))
	if err != nil {
		t.Fatalf("GetResetRequest: %v", err)
	}
	if status.Msg.GetStatus() != apiv1.RequestStatus_REQUEST_STATUS_PENDING {
		t.Errorf("status = %v, want pending once the approval has expired", status.Msg.GetStatus())
	}

	_, err = svc.CompletePasswordReset(t.Context(),
		connect.NewRequest(&apiv1.CompletePasswordResetRequest{
			RequestUid: uid, NewPassword: "a brand new password",
		}))
	if got := connect.CodeOf(err); got != connect.CodeFailedPrecondition {
		t.Errorf("code = %v, want failed_precondition on an expired approval", got)
	}
}

/*
Recovering an account signs out whoever was already in it.

Somebody resetting a password has lost their way in, so no browser they hold is worth
keeping — and the reason they are here may be that another one is signed in. Sparing
none is the point of the reset, not a side effect of it.
*/
func TestAPasswordResetEndsEverySession(t *testing.T) {
	svc, s := newAuthService(t)
	completeSetup(t, svc)

	anna, err := s.MemberByEmail(t.Context(), "anna@brunnen.lan")
	if err != nil {
		t.Fatalf("MemberByEmail: %v", err)
	}
	_, theirs := signedInAs(t, s, anna)

	res, err := svc.RequestPasswordReset(t.Context(),
		connect.NewRequest(&apiv1.RequestPasswordResetRequest{EmailOrName: "anna@brunnen.lan"}))
	if err != nil {
		t.Fatalf("RequestPasswordReset: %v", err)
	}
	uid := res.Msg.GetRequestUid()
	if err := s.DecideResetRequest(t.Context(), uid, store.StatusApproved, testClock); err != nil {
		t.Fatalf("DecideResetRequest: %v", err)
	}

	if _, err := svc.CompletePasswordReset(t.Context(),
		connect.NewRequest(&apiv1.CompletePasswordResetRequest{
			RequestUid: uid, NewPassword: "a brand new password",
		})); err != nil {
		t.Fatalf("CompletePasswordReset: %v", err)
	}

	if _, err := s.SessionByTokenHash(t.Context(), theirs); !errors.Is(err, store.ErrNotFound) {
		t.Errorf("a browser signed in before the reset is still signed in: %v", err)
	}
}

/*
Finishing a join must answer the same way whatever the reason it cannot be finished.

GetJoinRequest reports every unfinished state as pending, and that is tested. CompleteJoin
is the other half of the same ceremony and is reachable by the same stranger: if it
refused an ignored request differently from one that never existed, the pair would be an
oracle for which identifiers are real, which is what the reporting-as-pending exists to
prevent.
*/
func TestCompleteJoinRefusesEveryUnapprovedStateAlike(t *testing.T) {
	svc, s := newAuthService(t)
	completeSetup(t, svc)

	pending := requestJoin(t, svc, "til@example.com")

	ignored := requestJoin(t, svc, "jonas@example.com")
	if err := s.DecideJoinRequest(t.Context(), ignored, store.StatusIgnored, testClock); err != nil {
		t.Fatalf("DecideJoinRequest: %v", err)
	}

	refusalFor := func(t *testing.T, uid string) string {
		t.Helper()
		_, err := svc.CompleteJoin(t.Context(), connect.NewRequest(&apiv1.CompleteJoinRequest{
			RequestUid: uid, Name: "Til", Password: goodPassword,
		}))
		if err == nil {
			t.Fatalf("CompleteJoin accepted %q", uid)
		}
		return fmt.Sprintf("%v: %s", connect.CodeOf(err), err)
	}

	unknown := refusalFor(t, "never-existed")
	for _, one := range []struct{ what, uid string }{
		{"a request still waiting for an admin", pending},
		{"a request an admin ignored", ignored},
	} {
		if got := refusalFor(t, one.uid); got != unknown {
			t.Errorf("%s is refused as %q, and an identifier that never existed as %q",
				one.what, got, unknown)
		}
	}
}

/*
The Public signup setting decides whether a stranger may ask for an account.

It was stored, returned and drawn as a toggle, and read by nothing: asking always left a
request for an Admin whichever way it was set, so an Admin who turned it off was told
something happened and nothing did. This is the test that used to say so, inverted.

Off refuses outright rather than accepting and discarding. A closed door says nothing
about who lives here, so there is nothing to conceal by pretending otherwise, and
somebody sent by a housemate should be told to ask them rather than left waiting for an
approval that will never come.
*/
func TestSignupSettingDecidesWhoMayAsk(t *testing.T) {
	for _, open := range []bool{true, false} {
		t.Run(fmt.Sprintf("publicSignup=%v", open), func(t *testing.T) {
			svc, s := newAuthService(t)
			completeSetup(t, svc)

			settings, err := s.InstanceSettings(t.Context())
			if err != nil {
				t.Fatalf("InstanceSettings: %v", err)
			}
			settings.PublicSignup = open
			if err := s.SaveInstanceSettings(t.Context(), settings); err != nil {
				t.Fatalf("SaveInstanceSettings: %v", err)
			}

			res, err := svc.RequestJoin(t.Context(), connect.NewRequest(&apiv1.RequestJoinRequest{
				Name: "Til", Email: "til@example.com",
			}))

			if !open {
				if got := connect.CodeOf(err); got != connect.CodePermissionDenied {
					t.Fatalf("with signup off RequestJoin answered %v, want permission_denied", got)
				}
				return
			}

			if err != nil {
				t.Fatalf("with signup on RequestJoin: %v", err)
			}
			got, err := svc.GetJoinRequest(t.Context(), connect.NewRequest(
				&apiv1.GetJoinRequestRequest{RequestUid: res.Msg.GetRequestUid()},
			))
			if err != nil {
				t.Fatalf("GetJoinRequest: %v", err)
			}
			if got.Msg.GetStatus() != apiv1.RequestStatus_REQUEST_STATUS_PENDING {
				t.Errorf("the request is %v, want pending: the approval is still what "+
					"decides an account, the setting only decides who may ask",
					got.Msg.GetStatus())
			}
		})
	}
}

/*
One address cannot fill the panel.

Asking twice looks exactly like asking once from outside: the same success, an
identifier either way. What must not happen is a second row, and what must especially
not happen is the second asker being handed the first asker's identifier, which is the
thing that completes the account.
*/
func TestOneEmailWaitsOnce(t *testing.T) {
	svc, s := newAuthService(t)
	completeSetup(t, svc)

	first, err := askToJoin(t, svc, "til@example.com")
	if err != nil {
		t.Fatalf("RequestJoin: %v", err)
	}
	second, err := askToJoin(t, svc, "til@example.com")
	if err != nil {
		t.Fatalf("RequestJoin again: %v", err)
	}

	if second == first {
		t.Error("asking again handed back the first identifier, which completes that account")
	}
	if _, err := s.JoinRequestByUID(t.Context(), second); !errors.Is(err, store.ErrNotFound) {
		t.Errorf("JoinRequestByUID(second) = %v, want no request behind it", err)
	}

	waiting, err := s.PendingJoinRequests(t.Context())
	if err != nil {
		t.Fatalf("PendingJoinRequests: %v", err)
	}
	if len(waiting) != 1 {
		t.Errorf("%d requests waiting, want one", len(waiting))
	}
}

// A different address is a different person, so asking is not refused by somebody else
// having asked.
func TestAnotherEmailStillWaits(t *testing.T) {
	svc, s := newAuthService(t)
	completeSetup(t, svc)

	if _, err := askToJoin(t, svc, "til@example.com"); err != nil {
		t.Fatalf("RequestJoin: %v", err)
	}
	if _, err := askToJoin(t, svc, "jonas@example.com"); err != nil {
		t.Fatalf("RequestJoin: %v", err)
	}

	waiting, err := s.PendingJoinRequests(t.Context())
	if err != nil {
		t.Fatalf("PendingJoinRequests: %v", err)
	}
	if len(waiting) != 2 {
		t.Errorf("%d requests waiting, want two", len(waiting))
	}
}

/*
The queue has an end, and the panel can still show what is in it.

The ceiling is what bounds somebody inventing addresses, which the per-email rule cannot
touch. It is the one refusal a Visitor is allowed to see: a full queue is a fact about
the Instance, in the same way a closed door is.
*/
func TestTheJoinQueueHasACeiling(t *testing.T) {
	svc, s := newAuthService(t)
	completeSetup(t, svc)

	for i := range store.PendingJoinLimit {
		if _, err := askToJoin(t, svc, fmt.Sprintf("til-%d@example.com", i)); err != nil {
			t.Fatalf("RequestJoin %d: %v", i, err)
		}
	}

	_, err := askToJoin(t, svc, "one-too-many@example.com")
	if connect.CodeOf(err) != connect.CodeResourceExhausted {
		t.Fatalf("RequestJoin past the ceiling = %v, want resource exhausted", err)
	}

	waiting, err := s.PendingJoinRequests(t.Context())
	if err != nil {
		t.Fatalf("PendingJoinRequests: %v", err)
	}
	if len(waiting) != store.PendingJoinLimit {
		t.Errorf("%d requests waiting, want the ceiling of %d", len(waiting), store.PendingJoinLimit)
	}
	if len(waiting) > store.ActivityLimit {
		t.Errorf("%d requests waiting is more than the panel shows, so some are invisible",
			len(waiting))
	}
}

// Deciding one makes room, so a full queue is a thing an Admin can clear rather than a
// door that stays shut.
func TestDecidingAJoinRequestMakesRoom(t *testing.T) {
	svc, s := newAuthService(t)
	completeSetup(t, svc)

	for i := range store.PendingJoinLimit {
		if _, err := askToJoin(t, svc, fmt.Sprintf("til-%d@example.com", i)); err != nil {
			t.Fatalf("RequestJoin %d: %v", i, err)
		}
	}

	waiting, err := s.PendingJoinRequests(t.Context())
	if err != nil {
		t.Fatalf("PendingJoinRequests: %v", err)
	}
	if err := s.DecideJoinRequest(t.Context(), waiting[0].UID, store.StatusIgnored, testClock); err != nil {
		t.Fatalf("DecideJoinRequest: %v", err)
	}

	if _, err := askToJoin(t, svc, "one-more@example.com"); err != nil {
		t.Fatalf("RequestJoin after room was made: %v", err)
	}
}

// A Member who asks twice is in the panel once, and the second identifier is as dead as
// the one somebody with no account gets.
func TestOneMemberResetsOnce(t *testing.T) {
	svc, s := newAuthService(t)
	completeSetup(t, svc)

	ask := func() string {
		t.Helper()
		res, err := svc.RequestPasswordReset(t.Context(),
			connect.NewRequest(&apiv1.RequestPasswordResetRequest{EmailOrName: "anna@brunnen.lan"}))
		if err != nil {
			t.Fatalf("RequestPasswordReset: %v", err)
		}
		return res.Msg.GetRequestUid()
	}

	first, second := ask(), ask()
	if second == first {
		t.Error("asking again handed back the first identifier, which replaces that password")
	}
	if _, err := s.ResetRequestByUID(t.Context(), second); !errors.Is(err, store.ErrNotFound) {
		t.Errorf("ResetRequestByUID(second) = %v, want no request behind it", err)
	}

	waiting, err := s.PendingResetRequests(t.Context())
	if err != nil {
		t.Fatalf("PendingResetRequests: %v", err)
	}
	if len(waiting) != 1 {
		t.Errorf("%d resets waiting, want one", len(waiting))
	}
}
