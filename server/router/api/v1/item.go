package v1

import (
	"context"
	"errors"
	"time"

	"connectrpc.com/connect"

	"github.com/hoshomoh/nooks/internal/note"
	apiv1 "github.com/hoshomoh/nooks/proto/gen/nooks/api/v1"
	"github.com/hoshomoh/nooks/store"
)

// CreateItem appends an Item to a List.
func (s *ListService) CreateItem(
	ctx context.Context,
	req *connect.Request[apiv1.CreateItemRequest],
) (*connect.Response[apiv1.CreateItemResponse], error) {
	grant, err := requireGrant(ctx)
	if err != nil {
		return nil, err
	}
	list, err := s.listWithAccess(ctx, req.Msg.GetListUid(), grant, AccessWrite)
	if err != nil {
		return nil, err
	}
	if err := requireText(req.Msg.GetLabel(), "something to add"); err != nil {
		return nil, err
	}
	if err := validDueOn(req.Msg.GetDueOn()); err != nil {
		return nil, err
	}

	uid, err := s.newUID()
	if err != nil {
		return nil, internalError("make item uid", err)
	}
	item, err := s.store.CreateItem(ctx, store.CreateItemParams{
		UID: uid, ListID: list.ID, Label: req.Msg.GetLabel(),
		Quantity: req.Msg.GetQuantity(), DueOn: req.Msg.GetDueOn(),
		AddedByID: grant.Member.ID, At: s.now(),
	})
	if err != nil {
		return nil, internalError("create item", err)
	}
	s.announceListChanged(ctx, list)

	return connect.NewResponse(&apiv1.CreateItemResponse{
		Item: itemToProto(item, map[int64]memberLabel{grant.Member.ID: {Name: grant.Member.Name, UID: grant.Member.UID}}),
	}), nil
}

// UpdateItem changes an Item's label, quantity or due date.
func (s *ListService) UpdateItem(
	ctx context.Context,
	req *connect.Request[apiv1.UpdateItemRequest],
) (*connect.Response[apiv1.UpdateItemResponse], error) {
	grant, err := requireGrant(ctx)
	if err != nil {
		return nil, err
	}
	item, list, err := s.itemWithAccess(ctx, req.Msg.GetItemUid(), grant, AccessWrite)
	if err != nil {
		return nil, err
	}
	if req.Msg.Label != nil {
		if err := requireText(req.Msg.GetLabel(), "something to call it"); err != nil {
			return nil, err
		}
	}
	if req.Msg.DueOn != nil {
		if err := validDueOn(req.Msg.GetDueOn()); err != nil {
			return nil, err
		}
	}

	params := store.UpdateItemParams{
		Label:    req.Msg.Label,
		Quantity: req.Msg.Quantity,
		DueOn:    req.Msg.DueOn,
		Note:     req.Msg.Note,
	}
	if err := s.store.UpdateItem(ctx, item.UID, params, s.now()); err != nil {
		return nil, internalError("update item", err)
	}
	s.announceListChanged(ctx, list)

	updated, err := s.readItem(ctx, item.UID)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&apiv1.UpdateItemResponse{Item: updated}), nil
}

// SetItemDone ticks or unticks an Item.
//
// A tick is a tick whoever made it: this is last-write-wins and never asks a question.
func (s *ListService) SetItemDone(
	ctx context.Context,
	req *connect.Request[apiv1.SetItemDoneRequest],
) (*connect.Response[apiv1.SetItemDoneResponse], error) {
	grant, err := requireGrant(ctx)
	if err != nil {
		return nil, err
	}
	item, list, err := s.itemWithAccess(ctx, req.Msg.GetItemUid(), grant, AccessWrite)
	if err != nil {
		return nil, err
	}

	if req.Msg.GetDone() {
		err = s.store.SetItemDone(ctx, item.UID, grant.Member.ID, s.now())
	} else {
		err = s.store.SetItemNotDone(ctx, item.UID, s.now())
	}
	if err != nil {
		return nil, internalError("tick item", err)
	}
	s.announceListChanged(ctx, list)

	updated, err := s.readItem(ctx, item.UID)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&apiv1.SetItemDoneResponse{Item: updated}), nil
}

// MoveItem places an Item after another one, or at the top.
//
// The new position is the midpoint between its neighbours, so nothing else has to move.
func (s *ListService) MoveItem(
	ctx context.Context,
	req *connect.Request[apiv1.MoveItemRequest],
) (*connect.Response[apiv1.MoveItemResponse], error) {
	grant, err := requireGrant(ctx)
	if err != nil {
		return nil, err
	}
	item, list, err := s.itemWithAccess(ctx, req.Msg.GetItemUid(), grant, AccessWrite)
	if err != nil {
		return nil, err
	}

	siblings, err := s.store.ItemsOnList(ctx, list.ID)
	if err != nil {
		return nil, internalError("read items", err)
	}
	position, err := positionAfter(siblings, item.UID, req.Msg.GetAfterItemUid())
	if err != nil {
		return nil, err
	}
	if err := s.store.MoveItem(ctx, item.UID, position, s.now()); err != nil {
		return nil, internalError("move item", err)
	}
	s.announceListChanged(ctx, list)
	return connect.NewResponse(&apiv1.MoveItemResponse{}), nil
}

// DeleteItem removes an Item.
func (s *ListService) DeleteItem(
	ctx context.Context,
	req *connect.Request[apiv1.DeleteItemRequest],
) (*connect.Response[apiv1.DeleteItemResponse], error) {
	grant, err := requireGrant(ctx)
	if err != nil {
		return nil, err
	}
	item, list, err := s.itemWithAccess(ctx, req.Msg.GetItemUid(), grant, AccessWrite)
	if err != nil {
		return nil, err
	}
	if err := s.store.DeleteItem(ctx, item.UID, s.now()); err != nil {
		return nil, internalError("delete item", err)
	}
	s.announceListChanged(ctx, list)

	return connect.NewResponse(&apiv1.DeleteItemResponse{}), nil
}

// readItem re-reads an Item for a response, so a caller always gets the Item as stored
// rather than as the request imagined it.
func (s *ListService) readItem(ctx context.Context, uid string) (*apiv1.Item, error) {
	item, err := s.store.ItemByUID(ctx, uid)
	if err != nil {
		return nil, internalError("read item", err)
	}
	labels, err := s.memberLabels(ctx, []store.Item{item})
	if err != nil {
		return nil, err
	}
	return itemToProto(item, labels), nil
}

// positionAfter works out where an Item should sit, given the List's current order and
// the Item it is being placed after. An empty after means the top.
func positionAfter(items []store.Item, moving, after string) (float64, error) {
	// Everything except the Item being moved, in order.
	rest := make([]store.Item, 0, len(items))
	for _, item := range items {
		if item.UID != moving {
			rest = append(rest, item)
		}
	}

	if after == "" {
		if len(rest) == 0 {
			return positionGapDefault, nil
		}
		return rest[0].Position / 2, nil
	}

	for i, item := range rest {
		if item.UID != after {
			continue
		}
		if i == len(rest)-1 {
			return item.Position + positionGapDefault, nil
		}
		return (item.Position + rest[i+1].Position) / 2, nil
	}
	return 0, connect.NewError(connect.CodeInvalidArgument,
		errors.New("that item is not on this list"))
}

// positionGapDefault mirrors the store's appending gap.
const positionGapDefault = 1024.0

// memberLabel is what a row needs to know about the Member it names.
//
// The identifier as well as the name, because a reader compares it with their own: a
// tick of your own is not somebody else's tick.
type memberLabel struct {
	Name string
	UID  string
}

// memberLabels looks up what an Item's rows need, once per Member rather than once per
// Item.
func (s *ListService) memberLabels(
	ctx context.Context,
	items []store.Item,
) (map[int64]memberLabel, error) {
	labels := map[int64]memberLabel{}
	for _, item := range items {
		for _, id := range []int64{item.AddedByID, item.DoneByID} {
			if id == 0 {
				continue
			}
			if _, known := labels[id]; known {
				continue
			}
			member, err := s.store.MemberByID(ctx, id)
			if err != nil {
				if errors.Is(err, store.ErrNotFound) {
					continue
				}
				return nil, internalError("read member", err)
			}
			labels[id] = memberLabel{Name: member.Name, UID: member.UID}
		}
	}
	return labels, nil
}

// itemToProto converts an Item for the wire.
func itemToProto(item store.Item, labels map[int64]memberLabel) *apiv1.Item {
	preview := note.PreviewOf(item.Note)
	out := &apiv1.Item{
		Uid:                item.UID,
		Label:              item.Label,
		Quantity:           item.Quantity,
		DueOn:              item.DueOn,
		Done:               item.Done(),
		AddedByName:        labels[item.AddedByID].Name,
		DoneByName:         labels[item.DoneByID].Name,
		DoneByUid:          labels[item.DoneByID].UID,
		Note:               item.Note,
		NoteFirstLine:      preview.FirstLine,
		NoteRemainingLines: int32(preview.RemainingLines),
	}
	if !item.DoneAt.IsZero() {
		out.DoneAt = item.DoneAt.Format(time.RFC3339)
	}
	return out
}
