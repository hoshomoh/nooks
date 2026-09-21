package v1

import (
	"context"

	"connectrpc.com/connect"

	apiv1 "github.com/hoshomoh/nooks/proto/gen/nooks/api/v1"
)

// ListDatedItems returns unticked Items with a due date, across every reachable List.
//
// Today, Upcoming and the calendar are all views over this one answer. Undated Items
// never appear: a calendar is a lens on dates, not a second home for Lists.
func (s *ListService) ListDatedItems(
	ctx context.Context,
	req *connect.Request[apiv1.ListDatedItemsRequest],
) (*connect.Response[apiv1.ListDatedItemsResponse], error) {
	grant, err := requireGrant(ctx)
	if err != nil {
		return nil, err
	}
	if err := requireText(req.Msg.GetTo(), "a last day"); err != nil {
		return nil, err
	}
	if err := validDueOn(req.Msg.GetTo()); err != nil {
		return nil, err
	}
	if err := validDueOn(req.Msg.GetFrom()); err != nil {
		return nil, err
	}

	items, err := s.store.DatedItemsForMember(ctx, grant.Member.ID, req.Msg.GetFrom(), req.Msg.GetTo())
	if err != nil {
		return nil, internalError("read dated items", err)
	}
	if len(items) == 0 {
		return connect.NewResponse(&apiv1.ListDatedItemsResponse{}), nil
	}

	// The store limits these to Lists the Member can reach. This read is what limits
	// them to the Lists a token names as well, so the skip below is the narrowing and
	// not only the naming.
	reachable, err := s.reachableLists(ctx, grant)
	if err != nil {
		return nil, err
	}
	names, err := s.rowNamesFor(ctx, items)
	if err != nil {
		return nil, err
	}

	out := make([]*apiv1.DatedItem, 0, len(items))
	for _, item := range items {
		list, ok := reachable[item.ListID]
		if !ok {
			continue
		}
		out = append(out, &apiv1.DatedItem{
			Item:     itemToProto(item, names),
			ListUid:  list.UID,
			ListName: list.Name,
		})
	}
	return connect.NewResponse(&apiv1.ListDatedItemsResponse{Items: out}), nil
}
