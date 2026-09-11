package v1

import (
	"context"

	"connectrpc.com/connect"

	apiv1 "github.com/hoshomoh/nooks/proto/gen/nooks/api/v1"
	"github.com/hoshomoh/nooks/store"
)

// copySuffix marks a duplicate apart from what it was copied from.
//
// Not translated: the name becomes the Member's own the moment they rename it, and a
// name that changed language when somebody else opened it would be worse.
const copySuffix = " (copy)"

// DuplicateList copies a List and the Items still open on it.
//
// Reading it is enough — anyone who may see a List may make their own copy of it. The
// copy is theirs: they own it, it starts private, and it starts with no history. What
// is already ticked is not carried over, because a duplicate is made to do the same
// thing again rather than to remember the last time.
func (s *ListService) DuplicateList(
	ctx context.Context,
	req *connect.Request[apiv1.DuplicateListRequest],
) (*connect.Response[apiv1.DuplicateListResponse], error) {
	grant, err := requireGrant(ctx)
	if err != nil {
		return nil, err
	}
	source, err := s.listWithAccess(ctx, req.Msg.GetListUid(), grant, AccessRead)
	if err != nil {
		return nil, err
	}

	copied, err := s.newList(ctx, grant.Member, source.Name+copySuffix)
	if err != nil {
		return nil, err
	}
	if err := s.copyOpenItems(ctx, source, copied, grant.Member); err != nil {
		return nil, err
	}

	return connect.NewResponse(&apiv1.DuplicateListResponse{
		List: listToProto(copied, grant.Member, false, 0),
	}), nil
}

// newList makes a private List owned by the Member.
func (s *ListService) newList(
	ctx context.Context,
	member store.Member,
	name string,
) (store.List, error) {
	uid, err := s.newUID()
	if err != nil {
		return store.List{}, internalError("make list uid", err)
	}
	list, err := s.store.CreateList(ctx, store.CreateListParams{
		UID: uid, Name: name, OwnerID: member.ID,
		Sharing: store.SharingPrivate, CanEdit: true, At: s.now(),
	})
	if err != nil {
		return store.List{}, internalError("create list", err)
	}
	return list, nil
}

// copyOpenItems puts everything still open on the source onto the copy, in order.
func (s *ListService) copyOpenItems(
	ctx context.Context,
	source, copied store.List,
	member store.Member,
) error {
	items, err := s.store.ItemsOnList(ctx, source.ID)
	if err != nil {
		return internalError("read items", err)
	}

	for _, item := range items {
		if item.Done() {
			continue
		}
		uid, err := s.newUID()
		if err != nil {
			return internalError("make item uid", err)
		}
		// Whoever makes the copy is who added everything on it: the Items are new, and
		// attributing them to somebody who never touched this List would be a lie.
		_, err = s.store.CreateItem(ctx, store.CreateItemParams{
			UID: uid, ListID: copied.ID, Label: item.Label,
			Quantity: item.Quantity, DueOn: item.DueOn, Note: item.Note,
			AddedByID: member.ID, At: s.now(),
		})
		if err != nil {
			return internalError("copy item", err)
		}
	}
	return nil
}
