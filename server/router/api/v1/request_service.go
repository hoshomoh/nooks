package v1

import (
	"context"
	"errors"
	"time"

	"connectrpc.com/connect"

	apiv1 "github.com/hoshomoh/nooks/proto/gen/nooks/api/v1"
	"github.com/hoshomoh/nooks/store"
)

// RequestService is the Admin's side of Join and Reset requests.
type RequestService struct {
	store store.Store
	now   func() time.Time
}

// NewRequestService builds the service. now may be nil, in which case time.Now is used.
func NewRequestService(s store.Store, now func() time.Time) *RequestService {
	if now == nil {
		now = time.Now
	}
	return &RequestService{store: s, now: now}
}

// ListPendingRequests returns everything waiting for an Admin, oldest first.
func (s *RequestService) ListPendingRequests(
	ctx context.Context,
	_ *connect.Request[apiv1.ListPendingRequestsRequest],
) (*connect.Response[apiv1.ListPendingRequestsResponse], error) {
	if _, err := requireAdmin(ctx); err != nil {
		return nil, err
	}

	joins, err := s.store.PendingJoinRequests(ctx)
	if err != nil {
		return nil, internalError("read join requests", err)
	}
	resets, err := s.store.PendingResetRequests(ctx)
	if err != nil {
		return nil, internalError("read reset requests", err)
	}

	resetRequests, err := s.resetRequestsToProto(ctx, resets)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&apiv1.ListPendingRequestsResponse{
		JoinRequests:  joinRequestsToProto(joins),
		ResetRequests: resetRequests,
	}), nil
}

// joinRequestsToProto converts stored Join requests for the wire.
func joinRequestsToProto(requests []store.JoinRequest) []*apiv1.PendingJoinRequest {
	out := make([]*apiv1.PendingJoinRequest, 0, len(requests))
	for _, request := range requests {
		out = append(out, &apiv1.PendingJoinRequest{
			RequestUid: request.UID,
			Name:       request.Name,
			Email:      request.Email,
			Message:    request.Message,
			CreatedAt:  formatRFC3339(request.CreatedAt),
		})
	}
	return out
}

// resetRequestsToProto converts stored Reset requests, carrying the Member with each:
// an Admin has to recognise who is asking to know whether it is really them.
func (s *RequestService) resetRequestsToProto(
	ctx context.Context,
	requests []store.ResetRequest,
) ([]*apiv1.PendingResetRequest, error) {
	out := make([]*apiv1.PendingResetRequest, 0, len(requests))
	for _, request := range requests {
		member, err := s.store.MemberByID(ctx, request.MemberID)
		if err != nil {
			return nil, internalError("read member", err)
		}
		out = append(out, &apiv1.PendingResetRequest{
			RequestUid: request.UID,
			Member:     memberToProto(member),
			CreatedAt:  formatRFC3339(request.CreatedAt),
		})
	}
	return out, nil
}

// DecideJoinRequest approves or ignores a request for an account.
func (s *RequestService) DecideJoinRequest(
	ctx context.Context,
	req *connect.Request[apiv1.DecideJoinRequestRequest],
) (*connect.Response[apiv1.DecideJoinRequestResponse], error) {
	if _, err := requireAdmin(ctx); err != nil {
		return nil, err
	}

	err := s.store.DecideJoinRequest(ctx, req.Msg.GetRequestUid(),
		decision(req.Msg.GetApprove()), s.now())
	if err != nil {
		return nil, decideError(err)
	}
	if err := s.resolveActivity(ctx, req.Msg.GetRequestUid(), req.Msg.GetApprove()); err != nil {
		return nil, err
	}
	return connect.NewResponse(&apiv1.DecideJoinRequestResponse{}), nil
}

// DecideResetRequest approves or ignores a request to replace a password.
func (s *RequestService) DecideResetRequest(
	ctx context.Context,
	req *connect.Request[apiv1.DecideResetRequestRequest],
) (*connect.Response[apiv1.DecideResetRequestResponse], error) {
	if _, err := requireAdmin(ctx); err != nil {
		return nil, err
	}

	err := s.store.DecideResetRequest(ctx, req.Msg.GetRequestUid(),
		decision(req.Msg.GetApprove()), s.now())
	if err != nil {
		return nil, decideError(err)
	}
	if err := s.resolveActivity(ctx, req.Msg.GetRequestUid(), req.Msg.GetApprove()); err != nil {
		return nil, err
	}
	return connect.NewResponse(&apiv1.DecideResetRequestResponse{}), nil
}

// decision turns the wire's boolean into a status.
func decision(approve bool) store.RequestStatus {
	if approve {
		return store.StatusApproved
	}
	return store.StatusIgnored
}

// decideError reports a request that is gone or already decided as a precondition
// failure rather than an internal one — two Admins acting at once is not a bug.
func decideError(err error) error {
	if errors.Is(err, store.ErrNotFound) {
		return connect.NewError(connect.CodeFailedPrecondition,
			errors.New("that request is no longer waiting"))
	}
	return internalError("decide request", err)
}

// formatRFC3339 renders a timestamp for the wire, or empty for the zero value.
func formatRFC3339(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.UTC().Format(time.RFC3339)
}

// resolveActivity records what became of a request, in every Admin's panel.
//
// Deciding is the act; this is what makes it visible. Without it an Admin clicks
// Approve and nothing about the entry changes — so they cannot tell whether it worked,
// and the other Admins keep a button that now does nothing.
func (s *RequestService) resolveActivity(
	ctx context.Context,
	requestUID string,
	approved bool,
) error {
	outcome := store.OutcomeIgnored
	if approved {
		outcome = store.OutcomeApproved
	}
	if err := s.store.ResolveActivity(ctx, requestUID, outcome); err != nil {
		return internalError("resolve activity", err)
	}
	return nil
}
