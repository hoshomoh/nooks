package v1

import (
	"context"
	"time"

	"connectrpc.com/connect"

	apiv1 "github.com/hoshomoh/nooks/proto/gen/nooks/api/v1"
	"github.com/hoshomoh/nooks/store"
)

// ActivityService is everything waiting for one Member's attention.
//
// Nooks has no mail server, so this panel is the only place any of it surfaces.
type ActivityService struct {
	store store.Store
	now   func() time.Time
}

// NewActivityService builds the service. The clock is injected so a test need not wait.
func NewActivityService(s store.Store, now func() time.Time) *ActivityService {
	if now == nil {
		now = time.Now
	}
	return &ActivityService{store: s, now: now}
}

// ListActivity returns what is waiting for the signed-in Member, newest first.
func (s *ActivityService) ListActivity(
	ctx context.Context,
	_ *connect.Request[apiv1.ListActivityRequest],
) (*connect.Response[apiv1.ListActivityResponse], error) {
	member, err := requireMember(ctx)
	if err != nil {
		return nil, err
	}

	entries, err := s.store.ActivityFor(ctx, member.ID)
	if err != nil {
		return nil, internalError("read activity", err)
	}

	response := &apiv1.ListActivityResponse{Activity: make([]*apiv1.Activity, 0, len(entries))}
	for _, entry := range entries {
		response.Activity = append(response.Activity, activityToProto(entry))
		if entry.Unread() {
			response.UnreadCount++
		}
	}
	return connect.NewResponse(response), nil
}

// MarkActivityRead marks everything the Member has now seen.
func (s *ActivityService) MarkActivityRead(
	ctx context.Context,
	_ *connect.Request[apiv1.MarkActivityReadRequest],
) (*connect.Response[apiv1.MarkActivityReadResponse], error) {
	member, err := requireMember(ctx)
	if err != nil {
		return nil, err
	}
	if err := s.store.MarkActivityRead(ctx, member.ID, s.now()); err != nil {
		return nil, internalError("mark activity read", err)
	}
	return connect.NewResponse(&apiv1.MarkActivityReadResponse{}), nil
}

// activityToProto converts an entry for the wire.
func activityToProto(entry store.Activity) *apiv1.Activity {
	return &apiv1.Activity{
		Uid:       entry.UID,
		Kind:      activityKindToProto(entry.Kind),
		Text:      entry.Text,
		TargetUid: entry.TargetUID,
		CreatedAt: entry.CreatedAt.Format(time.RFC3339),
		Unread:    entry.Unread(),
	}
}

func activityKindToProto(kind store.ActivityKind) apiv1.ActivityKind {
	switch kind {
	case store.ActivityJoinRequest:
		return apiv1.ActivityKind_ACTIVITY_KIND_JOIN_REQUEST
	case store.ActivityResetRequest:
		return apiv1.ActivityKind_ACTIVITY_KIND_RESET_REQUEST
	case store.ActivityListShared:
		return apiv1.ActivityKind_ACTIVITY_KIND_LIST_SHARED
	case store.ActivityConflict:
		return apiv1.ActivityKind_ACTIVITY_KIND_CONFLICT
	default:
		return apiv1.ActivityKind_ACTIVITY_KIND_UNSPECIFIED
	}
}
