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
	member, err := requireMember(ctx)
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

	items, err := s.store.DatedItemsForMember(ctx, member.ID, req.Msg.GetFrom(), req.Msg.GetTo())
	if err != nil {
		return nil, internalError("read dated items", err)
	}
	if len(items) == 0 {
		return connect.NewResponse(&apiv1.ListDatedItemsResponse{}), nil
	}

	// The store already limits these to Lists the Member can reach; this read names
	// them, and is the same one accessTo would use.
	reachable, err := s.reachableLists(ctx, member)
	if err != nil {
		return nil, err
	}
	labels, err := s.memberLabels(ctx, items)
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
			Item:     itemToProto(item, labels),
			ListUid:  list.UID,
			ListName: list.Name,
		})
	}
	return connect.NewResponse(&apiv1.ListDatedItemsResponse{Items: out}), nil
}
