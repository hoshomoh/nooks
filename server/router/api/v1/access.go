package v1

import (
	"context"
	"errors"

	"connectrpc.com/connect"

	"github.com/hoshomoh/nooks/store"
)

// Access is what a Member may do with one List.
//
// The three levels are deliberately coarse. Nooks has one container and one sharing
// switch; anything finer would be a permission system nobody asked for.
type Access int

const (
	// AccessNone means the List is invisible: it must not be distinguishable from one
	// that does not exist.
	AccessNone Access = iota
	// AccessRead means see and print, but not tick or add.
	AccessRead
	// AccessWrite means tick, add and edit Items.
	AccessWrite
	// AccessOwn means rename, share and delete the List itself.
	AccessOwn
)

// accessTo works out what a Member may do with a List.
//
// It is a pure function of the two values, so every rule below can be read in one place
// and tested without a database.
func accessTo(list store.List, member store.Member) Access {
	if list.OwnerID == member.ID {
		return AccessOwn
	}
	if list.Sharing == store.SharingInstance {
		// Read-only sharing means see and print, not tick or add.
		if list.CanEdit {
			return AccessWrite
		}
		return AccessRead
	}
	// Named sharing arrives with Groups in M6. Until then, anything else is invisible.
	return AccessNone
}

// errListNotFound is returned both for a List that does not exist and for one the
// Member may not see. Telling them apart would let anyone probe for Lists.
var errListNotFound = connect.NewError(connect.CodeNotFound, errors.New("no such list"))

// errReadOnly reports an attempt to change a List that was shared read-only.
var errReadOnly = connect.NewError(connect.CodePermissionDenied,
	errors.New("this list is read-only for you"))

// errNotOwner reports an attempt to rename, share or delete somebody else's List.
var errNotOwner = connect.NewError(connect.CodePermissionDenied,
	errors.New("only the owner of a list can do that"))

// listWithAccess fetches a List and checks the Member may reach it at the given level.
func (s *ListService) listWithAccess(
	ctx context.Context,
	uid string,
	member store.Member,
	need Access,
) (store.List, error) {
	list, err := s.store.ListByUID(ctx, uid)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return store.List{}, errListNotFound
		}
		return store.List{}, internalError("read list", err)
	}

	switch have := accessTo(list, member); {
	case have >= need:
		return list, nil
	case have == AccessNone:
		return store.List{}, errListNotFound
	case need == AccessOwn:
		return store.List{}, errNotOwner
	default:
		return store.List{}, errReadOnly
	}
}

// itemWithAccess fetches an Item and the List it is on, checking access to the List.
// An Item is only ever as reachable as the List it sits on.
func (s *ListService) itemWithAccess(
	ctx context.Context,
	uid string,
	member store.Member,
	need Access,
) (store.Item, store.List, error) {
	item, err := s.store.ItemByUID(ctx, uid)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return store.Item{}, store.List{}, errItemNotFound
		}
		return store.Item{}, store.List{}, internalError("read item", err)
	}

	list, err := s.listByIDWithAccess(ctx, item.ListID, member, need)
	if err != nil {
		return store.Item{}, store.List{}, err
	}
	return item, list, nil
}

// errItemNotFound mirrors errListNotFound for Items.
var errItemNotFound = connect.NewError(connect.CodeNotFound, errors.New("no such item"))

// listByIDWithAccess is listWithAccess for a List already known by internal identity.
func (s *ListService) listByIDWithAccess(
	ctx context.Context,
	id int64,
	member store.Member,
	need Access,
) (store.List, error) {
	lists, err := s.store.ListsForMember(ctx, member.ID)
	if err != nil {
		return store.List{}, internalError("read lists", err)
	}
	for _, list := range lists {
		if list.ID != id {
			continue
		}
		if accessTo(list, member) >= need {
			return list, nil
		}
		if need == AccessOwn {
			return store.List{}, errNotOwner
		}
		return store.List{}, errReadOnly
	}
	// Not among the Lists they can reach: indistinguishable from not existing.
	return store.List{}, errItemNotFound
}
