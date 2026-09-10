package v1

import (
	"context"
	"errors"
	"time"

	"connectrpc.com/connect"

	apiv1 "github.com/hoshomoh/nooks/proto/gen/nooks/api/v1"
	"github.com/hoshomoh/nooks/server/auth"
	"github.com/hoshomoh/nooks/store"
)

// ListService covers Lists and the Items on them.
type ListService struct {
	store  store.Store
	now    func() time.Time
	newUID func() (string, error)
}

// NewListService builds the service. now and newUID may be nil, in which case the real
// clock and a random identifier are used.
func NewListService(s store.Store, now func() time.Time, newUID func() (string, error)) *ListService {
	if now == nil {
		now = time.Now
	}
	if newUID == nil {
		newUID = newMemberUID
	}
	return &ListService{store: s, now: now, newUID: newUID}
}

// activity records entries in the panel, from the same clock and identifiers this
// service already has.
func (s *ListService) activity() activityRecorder {
	return activityRecorder{store: s.store, now: s.now, newUID: s.newUID}
}

// ListLists returns every List the signed-in Member can reach.
func (s *ListService) ListLists(
	ctx context.Context,
	_ *connect.Request[apiv1.ListListsRequest],
) (*connect.Response[apiv1.ListListsResponse], error) {
	member, err := requireMember(ctx)
	if err != nil {
		return nil, err
	}

	lists, err := s.store.ListsForMember(ctx, member.ID)
	if err != nil {
		return nil, internalError("read lists", err)
	}
	pinned, err := s.pinnedSet(ctx, member.ID)
	if err != nil {
		return nil, err
	}

	out := make([]*apiv1.List, 0, len(lists))
	for _, list := range lists {
		open, err := s.openCount(ctx, list.ID)
		if err != nil {
			return nil, err
		}
		out = append(out, listToProto(list, member, pinned[list.ID], open))
	}
	return connect.NewResponse(&apiv1.ListListsResponse{Lists: out}), nil
}

// GetList returns one List and its Items, in their manual order.
func (s *ListService) GetList(
	ctx context.Context,
	req *connect.Request[apiv1.GetListRequest],
) (*connect.Response[apiv1.GetListResponse], error) {
	member, err := requireMember(ctx)
	if err != nil {
		return nil, err
	}
	list, err := s.listWithAccess(ctx, req.Msg.GetListUid(), member, AccessRead)
	if err != nil {
		return nil, err
	}

	items, err := s.store.ItemsOnList(ctx, list.ID)
	if err != nil {
		return nil, internalError("read items", err)
	}
	names, err := s.memberNames(ctx, items)
	if err != nil {
		return nil, err
	}
	pinned, err := s.pinnedSet(ctx, member.ID)
	if err != nil {
		return nil, err
	}

	out := make([]*apiv1.Item, 0, len(items))
	open := 0
	for _, item := range items {
		if !item.Done() {
			open++
		}
		out = append(out, itemToProto(item, names))
	}
	return connect.NewResponse(&apiv1.GetListResponse{
		List:  listToProto(list, member, pinned[list.ID], open),
		Items: out,
	}), nil
}

// CreateList adds a List. It starts private: sharing is a deliberate second step.
func (s *ListService) CreateList(
	ctx context.Context,
	req *connect.Request[apiv1.CreateListRequest],
) (*connect.Response[apiv1.CreateListResponse], error) {
	member, err := requireMember(ctx)
	if err != nil {
		return nil, err
	}
	if err := requireText(req.Msg.GetName(), "a name"); err != nil {
		return nil, err
	}

	uid, err := s.newUID()
	if err != nil {
		return nil, internalError("make list uid", err)
	}
	list, err := s.store.CreateList(ctx, store.CreateListParams{
		UID: uid, Name: req.Msg.GetName(), OwnerID: member.ID,
		Sharing: store.SharingPrivate, CanEdit: true, At: s.now(),
	})
	if err != nil {
		return nil, internalError("create list", err)
	}
	return connect.NewResponse(&apiv1.CreateListResponse{
		List: listToProto(list, member, false, 0),
	}), nil
}

// RenameList changes a List's name.
func (s *ListService) RenameList(
	ctx context.Context,
	req *connect.Request[apiv1.RenameListRequest],
) (*connect.Response[apiv1.RenameListResponse], error) {
	member, list, err := s.ownedList(ctx, req.Msg.GetListUid())
	if err != nil {
		return nil, err
	}
	if err := requireText(req.Msg.GetName(), "a name"); err != nil {
		return nil, err
	}
	if err := s.store.RenameList(ctx, list.UID, req.Msg.GetName(), s.now()); err != nil {
		return nil, internalError("rename list", err)
	}

	list.Name = req.Msg.GetName()
	return connect.NewResponse(&apiv1.RenameListResponse{
		List: listToProto(list, member, false, 0),
	}), nil
}

// SetListSharing changes who can reach a List.
func (s *ListService) SetListSharing(
	ctx context.Context,
	req *connect.Request[apiv1.SetListSharingRequest],
) (*connect.Response[apiv1.SetListSharingResponse], error) {
	member, list, err := s.ownedList(ctx, req.Msg.GetListUid())
	if err != nil {
		return nil, err
	}

	sharing := sharingFromProto(req.Msg.GetSharing())
	if sharing == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument,
			errors.New("say who the list is shared with"))
	}
	if err := s.store.SetListSharing(ctx, list.UID, sharing, req.Msg.GetCanEdit(), s.now()); err != nil {
		return nil, internalError("share list", err)
	}

	// Named shares are replaced in the same call: the dialog is one decision, and a
	// List that is no longer shared by name should not keep the names it had.
	memberUIDs, groupUIDs := req.Msg.GetMemberUids(), req.Msg.GetGroupUids()
	if sharing != store.SharingSpecific {
		memberUIDs, groupUIDs = nil, nil
	}
	if err := s.applyNamedShares(ctx, list, member, memberUIDs, groupUIDs); err != nil {
		return nil, err
	}

	list.Sharing = sharing
	list.CanEdit = req.Msg.GetCanEdit()
	return connect.NewResponse(&apiv1.SetListSharingResponse{
		List: listToProto(list, member, false, 0),
	}), nil
}

// DeleteList removes a List and the Items on it.
func (s *ListService) DeleteList(
	ctx context.Context,
	req *connect.Request[apiv1.DeleteListRequest],
) (*connect.Response[apiv1.DeleteListResponse], error) {
	_, list, err := s.ownedList(ctx, req.Msg.GetListUid())
	if err != nil {
		return nil, err
	}
	if err := s.store.DeleteList(ctx, list.UID, s.now()); err != nil {
		return nil, internalError("delete list", err)
	}
	return connect.NewResponse(&apiv1.DeleteListResponse{}), nil
}

// SetListPinned pins or unpins a List in the Member's own sidebar.
//
// Reading the List needs only read access: pinning is a note to yourself about a List
// you can already see, not a change to the List.
func (s *ListService) SetListPinned(
	ctx context.Context,
	req *connect.Request[apiv1.SetListPinnedRequest],
) (*connect.Response[apiv1.SetListPinnedResponse], error) {
	member, err := requireMember(ctx)
	if err != nil {
		return nil, err
	}
	list, err := s.listWithAccess(ctx, req.Msg.GetListUid(), member, AccessRead)
	if err != nil {
		return nil, err
	}

	if req.Msg.GetPinned() {
		err = s.store.PinList(ctx, member.ID, list.ID)
	} else {
		err = s.store.UnpinList(ctx, member.ID, list.ID)
	}
	if err != nil {
		return nil, internalError("pin list", err)
	}
	return connect.NewResponse(&apiv1.SetListPinnedResponse{}), nil
}

// ownedList fetches a List the signed-in Member owns.
func (s *ListService) ownedList(ctx context.Context, uid string) (store.Member, store.List, error) {
	member, err := requireMember(ctx)
	if err != nil {
		return store.Member{}, store.List{}, err
	}
	list, err := s.listWithAccess(ctx, uid, member, AccessOwn)
	if err != nil {
		return store.Member{}, store.List{}, err
	}
	return member, list, nil
}

// pinnedSet is the Member's pins, as a set for cheap lookup.
func (s *ListService) pinnedSet(ctx context.Context, memberID int64) (map[int64]bool, error) {
	ids, err := s.store.PinnedListIDs(ctx, memberID)
	if err != nil {
		return nil, internalError("read pins", err)
	}
	pinned := make(map[int64]bool, len(ids))
	for _, id := range ids {
		pinned[id] = true
	}
	return pinned, nil
}

// openCount is how many Items on a List are not yet ticked — the number the sidebar
// shows beside its name.
func (s *ListService) openCount(ctx context.Context, listID int64) (int, error) {
	items, err := s.store.ItemsOnList(ctx, listID)
	if err != nil {
		return 0, internalError("read items", err)
	}
	open := 0
	for _, item := range items {
		if !item.Done() {
			open++
		}
	}
	return open, nil
}

// requireMember rejects a request with no session.
func requireMember(ctx context.Context) (store.Member, error) {
	member, ok := auth.MemberFrom(ctx)
	if !ok {
		return store.Member{}, connect.NewError(connect.CodeUnauthenticated,
			errors.New("not signed in"))
	}
	return member, nil
}

// listToProto converts a List for the wire, from one Member's point of view.
func listToProto(list store.List, member store.Member, pinned bool, open int) *apiv1.List {
	return &apiv1.List{
		Uid:       list.UID,
		Name:      list.Name,
		Sharing:   sharingToProto(list.Sharing),
		CanEdit:   list.CanEdit,
		IsOwner:   list.OwnerID == member.ID,
		IsPinned:  pinned,
		OpenCount: int32(open),
	}
}

func sharingToProto(sharing store.Sharing) apiv1.Sharing {
	switch sharing {
	case store.SharingPrivate:
		return apiv1.Sharing_SHARING_PRIVATE
	case store.SharingInstance:
		return apiv1.Sharing_SHARING_INSTANCE
	case store.SharingSpecific:
		return apiv1.Sharing_SHARING_SPECIFIC
	default:
		return apiv1.Sharing_SHARING_UNSPECIFIED
	}
}

// sharingFromProto returns an empty Sharing for an unspecified value, which callers
// treat as a missing argument.
func sharingFromProto(sharing apiv1.Sharing) store.Sharing {
	switch sharing {
	case apiv1.Sharing_SHARING_PRIVATE:
		return store.SharingPrivate
	case apiv1.Sharing_SHARING_INSTANCE:
		return store.SharingInstance
	case apiv1.Sharing_SHARING_SPECIFIC:
		return store.SharingSpecific
	default:
		return ""
	}
}
