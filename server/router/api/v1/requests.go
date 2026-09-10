package v1

import (
	"context"
	"errors"
	"strings"

	"connectrpc.com/connect"

	"github.com/hoshomoh/nooks/internal/password"
	apiv1 "github.com/hoshomoh/nooks/proto/gen/nooks/api/v1"
	"github.com/hoshomoh/nooks/store"
)

// RequestJoin asks an Admin for an account.
//
// It always reports success. Whether that email already has an account is not the
// Visitor's business, and answering differently would turn this into a way to discover
// who lives here.
func (s *AuthService) RequestJoin(
	ctx context.Context,
	req *connect.Request[apiv1.RequestJoinRequest],
) (*connect.Response[apiv1.RequestJoinResponse], error) {
	msg := req.Msg
	if err := requireText(msg.GetName(), "a name"); err != nil {
		return nil, err
	}
	if err := requireText(msg.GetEmail(), "an email"); err != nil {
		return nil, err
	}

	uid, err := s.newUID()
	if err != nil {
		return nil, internalError("make request uid", err)
	}

	request, err := s.store.CreateJoinRequest(ctx, store.CreateJoinRequestParams{
		UID:       uid,
		Name:      msg.GetName(),
		Email:     msg.GetEmail(),
		Message:   msg.GetMessage(),
		CreatedAt: s.now(),
	})
	if err != nil {
		return nil, internalError("create join request", err)
	}
	return connect.NewResponse(&apiv1.RequestJoinResponse{RequestUid: request.UID}), nil
}

// GetJoinRequest reports whether an Admin has decided yet.
//
// An unknown identifier reads as pending rather than missing. A Visitor whose request
// was ignored is told nothing — DESIGN.md §9 — and an identifier that never existed
// must be indistinguishable from one that did, or this becomes an oracle.
func (s *AuthService) GetJoinRequest(
	ctx context.Context,
	req *connect.Request[apiv1.GetJoinRequestRequest],
) (*connect.Response[apiv1.GetJoinRequestResponse], error) {
	request, err := s.store.JoinRequestByUID(ctx, req.Msg.GetRequestUid())
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return connect.NewResponse(&apiv1.GetJoinRequestResponse{
				Status: apiv1.RequestStatus_REQUEST_STATUS_PENDING,
			}), nil
		}
		return nil, internalError("read join request", err)
	}

	// Ignored is reported as pending: the sender is never told they were turned down.
	if request.Status != store.StatusApproved {
		return connect.NewResponse(&apiv1.GetJoinRequestResponse{
			Status: apiv1.RequestStatus_REQUEST_STATUS_PENDING,
		}), nil
	}
	return connect.NewResponse(&apiv1.GetJoinRequestResponse{
		Status: apiv1.RequestStatus_REQUEST_STATUS_APPROVED,
		Name:   request.Name,
		Email:  request.Email,
	}), nil
}

// CompleteJoin turns an approved Join request into an account.
func (s *AuthService) CompleteJoin(
	ctx context.Context,
	req *connect.Request[apiv1.CompleteJoinRequest],
) (*connect.Response[apiv1.CompleteJoinResponse], error) {
	request, err := s.approvedJoinRequest(ctx, req.Msg.GetRequestUid())
	if err != nil {
		return nil, err
	}

	// The Visitor may correct the name they gave; the email is the Admin's decision and
	// is taken from the request.
	name := strings.TrimSpace(req.Msg.GetName())
	if name == "" {
		name = request.Name
	}

	member, err := s.createMember(ctx, createMemberInput{
		name:  name,
		email: request.Email,
		plain: req.Msg.GetPassword(),
		role:  store.RoleMember,
	})
	if err != nil {
		return nil, err
	}

	// Spending the request stops one approval creating two accounts.
	if err := s.store.UseJoinRequest(ctx, request.UID, s.now()); err != nil {
		return nil, internalError("spend join request", err)
	}

	res := connect.NewResponse(&apiv1.CompleteJoinResponse{Member: memberToProto(member)})
	if err := s.startSession(ctx, res.Header(), member); err != nil {
		return nil, err
	}
	return res, nil
}

// RequestPasswordReset asks an Admin to unlock an account.
//
// It reports success whether or not the account exists, and returns an identifier
// either way. An identifier with no request behind it simply stays pending forever,
// which is both true from the asker's point of view and useless for enumeration.
func (s *AuthService) RequestPasswordReset(
	ctx context.Context,
	req *connect.Request[apiv1.RequestPasswordResetRequest],
) (*connect.Response[apiv1.RequestPasswordResetResponse], error) {
	if err := requireText(req.Msg.GetEmailOrName(), "your email or name"); err != nil {
		return nil, err
	}

	uid, err := s.newUID()
	if err != nil {
		return nil, internalError("make request uid", err)
	}

	member, err := s.store.MemberByEmail(ctx, req.Msg.GetEmailOrName())
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			// No account, no request — but the same answer.
			return connect.NewResponse(&apiv1.RequestPasswordResetResponse{RequestUid: uid}), nil
		}
		return nil, internalError("read member", err)
	}

	if _, err := s.store.CreateResetRequest(ctx, uid, member.ID, s.now()); err != nil {
		return nil, internalError("create reset request", err)
	}
	return connect.NewResponse(&apiv1.RequestPasswordResetResponse{RequestUid: uid}), nil
}

// GetResetRequest reports whether an Admin has approved a reset yet.
func (s *AuthService) GetResetRequest(
	ctx context.Context,
	req *connect.Request[apiv1.GetResetRequestRequest],
) (*connect.Response[apiv1.GetResetRequestResponse], error) {
	// Anything not usable — missing, pending, ignored, expired, spent — reads as
	// pending, so none of them can be told apart from outside.
	if _, err := s.usableResetRequest(ctx, req.Msg.GetRequestUid()); err != nil {
		if errors.Is(err, errRequestNotApproved) {
			return connect.NewResponse(&apiv1.GetResetRequestResponse{
				Status: apiv1.RequestStatus_REQUEST_STATUS_PENDING,
			}), nil
		}
		return nil, err
	}
	return connect.NewResponse(&apiv1.GetResetRequestResponse{
		Status: apiv1.RequestStatus_REQUEST_STATUS_APPROVED,
	}), nil
}

// CompletePasswordReset sets a new password against an approved reset.
func (s *AuthService) CompletePasswordReset(
	ctx context.Context,
	req *connect.Request[apiv1.CompletePasswordResetRequest],
) (*connect.Response[apiv1.CompletePasswordResetResponse], error) {
	request, err := s.usableResetRequest(ctx, req.Msg.GetRequestUid())
	if err != nil {
		return nil, err
	}

	hash, err := password.Hash(req.Msg.GetNewPassword())
	if err != nil {
		return nil, passwordError(err)
	}
	if err := s.store.SetMemberPassword(ctx, request.MemberID, hash); err != nil {
		return nil, internalError("set member password", err)
	}
	// One approval sets one password.
	if err := s.store.UseResetRequest(ctx, request.UID); err != nil {
		return nil, internalError("spend reset request", err)
	}

	member, err := s.store.MemberByID(ctx, request.MemberID)
	if err != nil {
		return nil, internalError("read member", err)
	}

	res := connect.NewResponse(&apiv1.CompletePasswordResetResponse{Member: memberToProto(member)})
	if err := s.startSession(ctx, res.Header(), member); err != nil {
		return nil, err
	}
	return res, nil
}

// approvedJoinRequest fetches a Join request that may still be acted on.
func (s *AuthService) approvedJoinRequest(ctx context.Context, uid string) (store.JoinRequest, error) {
	request, err := s.store.JoinRequestByUID(ctx, uid)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return store.JoinRequest{}, errRequestNotApproved
		}
		return store.JoinRequest{}, internalError("read join request", err)
	}
	if request.Status != store.StatusApproved {
		return store.JoinRequest{}, errRequestNotApproved
	}
	return request, nil
}

// usableResetRequest fetches a Reset request that is approved and not yet expired.
func (s *AuthService) usableResetRequest(ctx context.Context, uid string) (store.ResetRequest, error) {
	request, err := s.store.ResetRequestByUID(ctx, uid)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return store.ResetRequest{}, errRequestNotApproved
		}
		return store.ResetRequest{}, internalError("read reset request", err)
	}
	if !request.Usable(s.now()) {
		return store.ResetRequest{}, errRequestNotApproved
	}
	return request, nil
}

// errRequestNotApproved covers a request that is missing, pending, ignored, expired or
// already spent. They are one message on purpose: telling them apart would say more
// about the Instance than the asker is entitled to know.
var errRequestNotApproved = connect.NewError(connect.CodeFailedPrecondition,
	errors.New("that request has not been approved, or is no longer valid"))
