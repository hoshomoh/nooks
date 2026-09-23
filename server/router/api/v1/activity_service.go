package v1

import (
	"context"
	"time"

	"connectrpc.com/connect"

	apiv1 "github.com/hoshomoh/nooks/proto/gen/nooks/api/v1"
	"github.com/hoshomoh/nooks/server/auth"
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

/*
ListActivity returns what is waiting for the signed-in Member, newest first.

An entry naming a List is left out when the caller is a token that does not reach it.
The token model says "a List it does not name is invisible to it", and without this a
token cut for the shopping list was told "Jonas shared Finances with you" — the name of
a List it cannot open, and of the person who owns it.

The count is worked out after that, not before. A number covering entries the caller
cannot see would say how many there are, which is the thing being kept back.
*/
func (s *ActivityService) ListActivity(
	ctx context.Context,
	_ *connect.Request[apiv1.ListActivityRequest],
) (*connect.Response[apiv1.ListActivityResponse], error) {
	grant, err := requireGrant(ctx)
	if err != nil {
		return nil, err
	}

	entries, err := s.store.ActivityFor(ctx, grant.Member.ID)
	if err != nil {
		return nil, internalError("read activity", err)
	}

	named, err := s.listsNamedBy(ctx, grant)
	if err != nil {
		return nil, err
	}

	response := &apiv1.ListActivityResponse{Activity: make([]*apiv1.Activity, 0, len(entries))}
	for _, entry := range entries {
		if named != nil && entry.Kind == store.ActivityListShared && !named[entry.TargetUID] {
			continue
		}
		response.Activity = append(response.Activity, activityToProto(entry))
		if entry.Unread() {
			response.UnreadCount++
		}
	}
	return connect.NewResponse(response), nil
}

/*
listsNamedBy is the Lists a token names, by uid, or nil when the caller narrows nothing.

Nil rather than an empty set, so a browser and a token cut for every List cost no read
at all and are not accidentally filtered down to nothing.
*/
func (s *ActivityService) listsNamedBy(
	ctx context.Context,
	grant auth.Grant,
) (map[string]bool, error) {
	if _, limited := grant.ReachIDs(); !limited {
		return nil, nil
	}

	lists, err := s.store.ListsForMember(ctx, grant.Member.ID)
	if err != nil {
		return nil, internalError("read lists", err)
	}
	named := make(map[string]bool, len(lists))
	for _, list := range lists {
		if grant.Reaches(list.ID) {
			named[list.UID] = true
		}
	}
	return named, nil
}

// MarkActivityRead marks everything the Member has now seen.
func (s *ActivityService) MarkActivityRead(
	ctx context.Context,
	_ *connect.Request[apiv1.MarkActivityReadRequest],
) (*connect.Response[apiv1.MarkActivityReadResponse], error) {
	member, err := requireWriter(ctx)
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
		Outcome:   outcomeToProto(entry.Outcome),
	}
}

func outcomeToProto(outcome store.Outcome) apiv1.ActivityOutcome {
	switch outcome {
	case store.OutcomeApproved:
		return apiv1.ActivityOutcome_ACTIVITY_OUTCOME_APPROVED
	case store.OutcomeIgnored:
		return apiv1.ActivityOutcome_ACTIVITY_OUTCOME_IGNORED
	default:
		return apiv1.ActivityOutcome_ACTIVITY_OUTCOME_UNSPECIFIED
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
	case store.ActivityTokenUsed:
		return apiv1.ActivityKind_ACTIVITY_KIND_TOKEN_USED
	default:
		return apiv1.ActivityKind_ACTIVITY_KIND_UNSPECIFIED
	}
}
