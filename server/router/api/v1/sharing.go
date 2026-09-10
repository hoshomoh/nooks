package v1

import (
	"context"
	"errors"

	"connectrpc.com/connect"

	apiv1 "github.com/hoshomoh/nooks/proto/gen/nooks/api/v1"
	"github.com/hoshomoh/nooks/store"
)

// GetListShares returns who a List reaches by name.
//
// Only the owner may read it: who else a List is shared with is the owner's business,
// and telling everyone it reaches would leak the shape of a household.
func (s *ListService) GetListShares(
	ctx context.Context,
	req *connect.Request[apiv1.GetListSharesRequest],
) (*connect.Response[apiv1.GetListSharesResponse], error) {
	_, list, err := s.ownedList(ctx, req.Msg.GetListUid())
	if err != nil {
		return nil, err
	}

	shares, err := s.store.ListShares(ctx, list.ID)
	if err != nil {
		return nil, internalError("read list shares", err)
	}

	response := &apiv1.GetListSharesResponse{}
	for _, share := range shares {
		if share.MemberID != 0 {
			member, err := s.store.MemberByID(ctx, share.MemberID)
			if err != nil {
				return nil, internalError("read member", err)
			}
			response.MemberUids = append(response.MemberUids, member.UID)
			continue
		}
		group, err := s.groupByID(ctx, share.GroupID)
		if err != nil {
			return nil, err
		}
		response.GroupUids = append(response.GroupUids, group.UID)
	}
	return connect.NewResponse(response), nil
}

// applyNamedShares replaces who a List reaches by name, and tells each of them.
//
// Named sharing is the only kind that reaches a particular person, so it is the only
// kind worth an Activity entry: being added to an Instance-wide List is not news.
func (s *ListService) applyNamedShares(
	ctx context.Context,
	list store.List,
	sharer store.Member,
	memberUIDs, groupUIDs []string,
) error {
	memberIDs, err := s.memberIDsOf(ctx, memberUIDs)
	if err != nil {
		return err
	}
	groupIDs, err := s.groupIDsOf(ctx, groupUIDs)
	if err != nil {
		return err
	}

	before, err := s.store.ListShares(ctx, list.ID)
	if err != nil {
		return internalError("read list shares", err)
	}
	if err := s.store.ReplaceListShares(ctx, list.ID, memberIDs, groupIDs); err != nil {
		return internalError("share list", err)
	}

	return s.announceNewShares(ctx, list, sharer, reachedBy(before), memberIDs, groupIDs)
}

// announceNewShares records an entry for everyone the List now reaches and did not
// before. Sharing twice does not tell anyone twice.
func (s *ListService) announceNewShares(
	ctx context.Context,
	list store.List,
	sharer store.Member,
	before map[int64]bool,
	memberIDs, groupIDs []int64,
) error {
	reached, err := s.membersReachedBy(ctx, memberIDs, groupIDs)
	if err != nil {
		return err
	}

	for _, id := range reached {
		if before[id] || id == sharer.ID {
			continue
		}
		text := sharer.Name + " shared “" + list.Name + "” with you"
		if err := s.activity().record(ctx, id, store.ActivityListShared, text, list.UID); err != nil {
			return err
		}
	}
	return nil
}

// reachedBy is the set of Members a set of shares named directly. A Group's members are
// not in it: joining a Group later is its own news.
func reachedBy(shares []store.Share) map[int64]bool {
	reached := make(map[int64]bool, len(shares))
	for _, share := range shares {
		if share.MemberID != 0 {
			reached[share.MemberID] = true
		}
	}
	return reached
}

// membersReachedBy expands Groups into the Members in them, without repeats.
func (s *ListService) membersReachedBy(
	ctx context.Context,
	memberIDs, groupIDs []int64,
) ([]int64, error) {
	seen := make(map[int64]bool, len(memberIDs))
	reached := make([]int64, 0, len(memberIDs))

	add := func(id int64) {
		if !seen[id] {
			seen[id] = true
			reached = append(reached, id)
		}
	}

	for _, id := range memberIDs {
		add(id)
	}
	for _, groupID := range groupIDs {
		ids, err := s.store.GroupMemberIDs(ctx, groupID)
		if err != nil {
			return nil, internalError("read group members", err)
		}
		for _, id := range ids {
			add(id)
		}
	}
	return reached, nil
}

// errNoSuchMember and errNoSuchGroup keep a bad identifier from reading as a server
// fault. Nothing is leaked: only an owner reaches this, and both names are already
// visible to everyone on the Instance.
var errNoSuchMember = connect.NewError(connect.CodeInvalidArgument, errors.New("no such member"))

var errNoSuchGroup = connect.NewError(connect.CodeInvalidArgument, errors.New("no such group"))

// memberIDsOf resolves public identifiers to internal ones.
func (s *ListService) memberIDsOf(ctx context.Context, uids []string) ([]int64, error) {
	ids := make([]int64, 0, len(uids))
	for _, uid := range uids {
		member, err := s.store.MemberByUID(ctx, uid)
		if errors.Is(err, store.ErrNotFound) {
			return nil, errNoSuchMember
		}
		if err != nil {
			return nil, internalError("read member", err)
		}
		ids = append(ids, member.ID)
	}
	return ids, nil
}

// groupIDsOf resolves public identifiers to internal ones.
func (s *ListService) groupIDsOf(ctx context.Context, uids []string) ([]int64, error) {
	ids := make([]int64, 0, len(uids))
	for _, uid := range uids {
		group, err := s.store.GroupByUID(ctx, uid)
		if errors.Is(err, store.ErrNotFound) {
			return nil, errNoSuchGroup
		}
		if err != nil {
			return nil, internalError("read group", err)
		}
		ids = append(ids, group.ID)
	}
	return ids, nil
}

// groupByID finds a Group by internal identity, which the share rows carry.
func (s *ListService) groupByID(ctx context.Context, id int64) (store.Group, error) {
	groups, err := s.store.Groups(ctx)
	if err != nil {
		return store.Group{}, internalError("read groups", err)
	}
	for _, group := range groups {
		if group.ID == id {
			return group, nil
		}
	}
	return store.Group{}, errNoSuchGroup
}
