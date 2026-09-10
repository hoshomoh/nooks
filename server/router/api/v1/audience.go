package v1

import (
	"context"

	"github.com/hoshomoh/nooks/store"
)

// AudienceOf is every Member who can reach a List.
//
// It answers the same question accessTo does, from the other side: accessTo asks
// whether one Member may see one List, and this asks who may see it at all. Live
// updates need the second form, because an event has to be addressed before it is sent.
//
// Deliberately not on the broker: who can see what is a permission, and permissions
// live here beside the rest of them.
func AudienceOf(ctx context.Context, s store.Store, list store.List) ([]int64, error) {
	switch list.Sharing {
	case store.SharingInstance:
		return everyone(ctx, s)
	case store.SharingSpecific:
		return named(ctx, s, list)
	default:
		return []int64{list.OwnerID}, nil
	}
}

// everyone is every Member on the Instance. An Instance-wide List reaches whoever joins
// later, so the answer is read rather than remembered.
func everyone(ctx context.Context, s store.Store) ([]int64, error) {
	members, err := s.Members(ctx)
	if err != nil {
		return nil, internalError("read members", err)
	}
	ids := make([]int64, 0, len(members))
	for _, member := range members {
		ids = append(ids, member.ID)
	}
	return ids, nil
}

// named is the owner plus everyone a share reaches, directly or through a Group.
func named(ctx context.Context, s store.Store, list store.List) ([]int64, error) {
	shares, err := s.ListShares(ctx, list.ID)
	if err != nil {
		return nil, internalError("read list shares", err)
	}

	seen := map[int64]bool{list.OwnerID: true}
	ids := []int64{list.OwnerID}
	add := func(id int64) {
		if !seen[id] {
			seen[id] = true
			ids = append(ids, id)
		}
	}

	for _, share := range shares {
		if share.MemberID != 0 {
			add(share.MemberID)
			continue
		}
		members, err := s.GroupMemberIDs(ctx, share.GroupID)
		if err != nil {
			return nil, internalError("read group members", err)
		}
		for _, id := range members {
			add(id)
		}
	}
	return ids, nil
}
