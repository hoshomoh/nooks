package v1

import (
	"context"
	"testing"
	"time"

	"connectrpc.com/connect"

	apiv1 "github.com/hoshomoh/nooks/proto/gen/nooks/api/v1"
	"github.com/hoshomoh/nooks/server/auth"
	"github.com/hoshomoh/nooks/store"
)

// newRequestService builds the Admin-side service over the same store as the auth
// service, with a fixed clock.
func newRequestService(s store.Store) *RequestService {
	return NewRequestService(s, func() time.Time { return testClock })
}

// asAdmin returns a context carrying the Instance's Admin.
func asAdmin(t *testing.T, s store.Store) context.Context {
	t.Helper()
	admin, err := s.MemberByUID(t.Context(), "mem_test")
	if err != nil {
		t.Fatalf("MemberByUID: %v", err)
	}
	return auth.WithMember(t.Context(), admin)
}

func TestOnlyAnAdminSeesPendingRequests(t *testing.T) {
	authSvc, s := newAuthService(t)
	completeSetup(t, authSvc)
	svc := newRequestService(s)

	_, err := svc.ListPendingRequests(t.Context(),
		connect.NewRequest(&apiv1.ListPendingRequestsRequest{}))
	if got := connect.CodeOf(err); got != connect.CodeUnauthenticated {
		t.Errorf("signed out = %v, want unauthenticated", got)
	}

	member := store.Member{ID: 99, Name: "Mira", Role: store.RoleMember}
	_, err = svc.ListPendingRequests(auth.WithMember(t.Context(), member),
		connect.NewRequest(&apiv1.ListPendingRequestsRequest{}))
	if got := connect.CodeOf(err); got != connect.CodePermissionDenied {
		t.Errorf("a Member who is not an Admin = %v, want permission_denied", got)
	}
}

func TestPendingRequestsAreListedForTheAdmin(t *testing.T) {
	authSvc, s := newAuthService(t)
	completeSetup(t, authSvc)
	svc := newRequestService(s)

	if _, err := authSvc.RequestJoin(t.Context(), connect.NewRequest(&apiv1.RequestJoinRequest{
		Name: "Til", Email: "til@example.com", Message: "It's Til, from upstairs",
	})); err != nil {
		t.Fatalf("RequestJoin: %v", err)
	}
	if _, err := authSvc.RequestPasswordReset(t.Context(),
		connect.NewRequest(&apiv1.RequestPasswordResetRequest{EmailOrName: "anna@brunnen.lan"})); err != nil {
		t.Fatalf("RequestPasswordReset: %v", err)
	}

	res, err := svc.ListPendingRequests(asAdmin(t, s),
		connect.NewRequest(&apiv1.ListPendingRequestsRequest{}))
	if err != nil {
		t.Fatalf("ListPendingRequests: %v", err)
	}
	if len(res.Msg.GetJoinRequests()) != 1 {
		t.Fatalf("join requests = %d, want 1", len(res.Msg.GetJoinRequests()))
	}
	if got := res.Msg.GetJoinRequests()[0].GetMessage(); got != "It's Til, from upstairs" {
		t.Errorf("message = %q, want what Til wrote", got)
	}
	if len(res.Msg.GetResetRequests()) != 1 {
		t.Fatalf("reset requests = %d, want 1", len(res.Msg.GetResetRequests()))
	}
	// An Admin has to recognise who is asking.
	if got := res.Msg.GetResetRequests()[0].GetMember().GetName(); got != "Anna" {
		t.Errorf("reset request member = %q, want Anna", got)
	}
}

func TestApprovingAJoinRequestLetsThemIn(t *testing.T) {
	authSvc, s := newAuthService(t)
	completeSetup(t, authSvc)
	svc := newRequestService(s)

	asked, err := authSvc.RequestJoin(t.Context(), connect.NewRequest(&apiv1.RequestJoinRequest{
		Name: "Til", Email: "til@example.com",
	}))
	if err != nil {
		t.Fatalf("RequestJoin: %v", err)
	}
	uid := asked.Msg.GetRequestUid()

	if _, err := svc.DecideJoinRequest(asAdmin(t, s),
		connect.NewRequest(&apiv1.DecideJoinRequestRequest{RequestUid: uid, Approve: true})); err != nil {
		t.Fatalf("DecideJoinRequest: %v", err)
	}

	if _, err := authSvc.CompleteJoin(t.Context(), connect.NewRequest(&apiv1.CompleteJoinRequest{
		RequestUid: uid, Name: "Til", Password: goodPassword,
	})); err != nil {
		t.Errorf("CompleteJoin after approval: %v", err)
	}
}

// Ignoring is silent, and it closes the request for good.
func TestIgnoringAJoinRequestClosesIt(t *testing.T) {
	authSvc, s := newAuthService(t)
	completeSetup(t, authSvc)
	svc := newRequestService(s)

	asked, err := authSvc.RequestJoin(t.Context(), connect.NewRequest(&apiv1.RequestJoinRequest{
		Name: "tomsmith2891", Email: "tomsmith2891@mail.ru",
	}))
	if err != nil {
		t.Fatalf("RequestJoin: %v", err)
	}
	uid := asked.Msg.GetRequestUid()

	if _, err := svc.DecideJoinRequest(asAdmin(t, s),
		connect.NewRequest(&apiv1.DecideJoinRequestRequest{RequestUid: uid, Approve: false})); err != nil {
		t.Fatalf("DecideJoinRequest: %v", err)
	}

	_, err = authSvc.CompleteJoin(t.Context(), connect.NewRequest(&apiv1.CompleteJoinRequest{
		RequestUid: uid, Name: "tomsmith2891", Password: goodPassword,
	}))
	if got := connect.CodeOf(err); got != connect.CodeFailedPrecondition {
		t.Errorf("CompleteJoin after being ignored = %v, want failed_precondition", got)
	}

	listed, err := svc.ListPendingRequests(asAdmin(t, s),
		connect.NewRequest(&apiv1.ListPendingRequestsRequest{}))
	if err != nil {
		t.Fatalf("ListPendingRequests: %v", err)
	}
	if len(listed.Msg.GetJoinRequests()) != 0 {
		t.Error("an ignored request is still listed as pending")
	}
}

// Two Admins acting at once is not a bug, so the second one gets a precondition
// failure rather than an internal error.
func TestDecidingTwiceIsAPreconditionFailure(t *testing.T) {
	authSvc, s := newAuthService(t)
	completeSetup(t, authSvc)
	svc := newRequestService(s)

	asked, err := authSvc.RequestJoin(t.Context(), connect.NewRequest(&apiv1.RequestJoinRequest{
		Name: "Til", Email: "til@example.com",
	}))
	if err != nil {
		t.Fatalf("RequestJoin: %v", err)
	}
	req := connect.NewRequest(&apiv1.DecideJoinRequestRequest{
		RequestUid: asked.Msg.GetRequestUid(), Approve: true,
	})

	if _, err := svc.DecideJoinRequest(asAdmin(t, s), req); err != nil {
		t.Fatalf("first decision: %v", err)
	}
	_, err = svc.DecideJoinRequest(asAdmin(t, s), req)
	if got := connect.CodeOf(err); got != connect.CodeFailedPrecondition {
		t.Errorf("second decision = %v, want failed_precondition", got)
	}
}

func TestApprovingAResetLetsThemSetAPassword(t *testing.T) {
	authSvc, s := newAuthService(t)
	completeSetup(t, authSvc)
	svc := newRequestService(s)

	asked, err := authSvc.RequestPasswordReset(t.Context(),
		connect.NewRequest(&apiv1.RequestPasswordResetRequest{EmailOrName: "anna@brunnen.lan"}))
	if err != nil {
		t.Fatalf("RequestPasswordReset: %v", err)
	}
	uid := asked.Msg.GetRequestUid()

	if _, err := svc.DecideResetRequest(asAdmin(t, s),
		connect.NewRequest(&apiv1.DecideResetRequestRequest{RequestUid: uid, Approve: true})); err != nil {
		t.Fatalf("DecideResetRequest: %v", err)
	}

	if _, err := authSvc.CompletePasswordReset(t.Context(),
		connect.NewRequest(&apiv1.CompletePasswordResetRequest{
			RequestUid: uid, NewPassword: "a password she picked",
		})); err != nil {
		t.Errorf("CompletePasswordReset after approval: %v", err)
	}
}
