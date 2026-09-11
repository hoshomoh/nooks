package v1

import (
	"context"
	"errors"

	"connectrpc.com/connect"

	"github.com/hoshomoh/nooks/server/auth"
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

// namedShares is the set of Lists a named share reaches for one Member, by internal
// identity — shared with them directly, or with a Group they are in.
//
// Whether a share reaches somebody is a fact about the database rather than about the
// List, so it is read once and passed in. That is what keeps accessTo a pure function
// of its arguments, and what lets one read answer a whole page of Lists.
type namedShares map[int64]bool

// accessTo works out what a caller may do with a List.
//
// It is a pure function of its arguments, so every rule below can be read in one place
// and tested without a database. It is also the *only* place access is decided: a
// browser, a script holding an Access token and an MCP client all arrive here, which is
// what stops a second route growing a second set of rules.
//
// The Member's own access is worked out first, then narrowed by whatever the Grant
// limits. A token is never its own identity — it is somebody else's access with edges.
func accessTo(list store.List, grant auth.Grant, shares namedShares) Access {
	// A List a token does not name is invisible, not forbidden: a token must not be a
	// way to learn which Lists its Member has.
	if !grant.Reaches(list.ID) {
		return AccessNone
	}

	have := memberAccessTo(list, grant.Member, shares)
	if grant.ReadOnly() && have > AccessRead {
		return AccessRead
	}
	return have
}

// memberAccessTo is what the Member themselves may do, before any narrowing.
func memberAccessTo(list store.List, member store.Member, shares namedShares) Access {
	if list.OwnerID == member.ID {
		return AccessOwn
	}
	if !reaches(list, shares) {
		return AccessNone
	}
	// Read-only sharing means see and print, not tick or add.
	if list.CanEdit {
		return AccessWrite
	}
	return AccessRead
}

// reaches reports whether a List is shared with the Member at all.
func reaches(list store.List, shares namedShares) bool {
	switch list.Sharing {
	case store.SharingInstance:
		// Everyone on the Instance, including whoever joins later.
		return true
	case store.SharingSpecific:
		return shares[list.ID]
	default:
		return false
	}
}

// sharesFor reads the Member's named shares once.
//
// A free function rather than a method: every way into Nooks has to work out reach the
// same way, and a helper that hangs off one service is a helper the next one copies.
func sharesFor(ctx context.Context, st store.Store, member store.Member) (namedShares, error) {
	ids, err := st.SharedListIDs(ctx, member.ID)
	if err != nil {
		return nil, internalError("read shared lists", err)
	}
	shares := make(namedShares, len(ids))
	for _, id := range ids {
		shares[id] = true
	}
	return shares, nil
}

// listReach is every List the caller may at least read, in the order the store gives
// them — which is the order a Member sees, and so has to be the same on every read.
//
// This is the one answer to "what can this caller see?", for the browser, the REST API
// and the MCP tools alike.
func listReach(ctx context.Context, st store.Store, grant auth.Grant) ([]store.List, error) {
	lists, err := st.ListsForMember(ctx, grant.Member.ID)
	if err != nil {
		return nil, internalError("read lists", err)
	}
	shares, err := sharesFor(ctx, st, grant.Member)
	if err != nil {
		return nil, err
	}

	reach := make([]store.List, 0, len(lists))
	for _, list := range lists {
		if accessTo(list, grant, shares) >= AccessRead {
			reach = append(reach, list)
		}
	}
	return reach, nil
}

// namedSharesFor is sharesFor with the service's own store.
func (s *ListService) namedSharesFor(ctx context.Context, member store.Member) (namedShares, error) {
	return sharesFor(ctx, s.store, member)
}

// sharesReaching is namedSharesFor for a single List, and reads nothing unless the List
// is shared by name — which most are not.
func (s *ListService) sharesReaching(
	ctx context.Context,
	list store.List,
	member store.Member,
) (namedShares, error) {
	if list.Sharing != store.SharingSpecific || list.OwnerID == member.ID {
		return nil, nil
	}
	return s.namedSharesFor(ctx, member)
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
	grant auth.Grant,
	need Access,
) (store.List, error) {
	list, err := s.store.ListByUID(ctx, uid)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return store.List{}, errListNotFound
		}
		return store.List{}, internalError("read list", err)
	}

	shares, err := s.sharesReaching(ctx, list, grant.Member)
	if err != nil {
		return store.List{}, err
	}

	switch have := accessTo(list, grant, shares); {
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
	grant auth.Grant,
	need Access,
) (store.Item, store.List, error) {
	item, err := s.store.ItemByUID(ctx, uid)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return store.Item{}, store.List{}, errItemNotFound
		}
		return store.Item{}, store.List{}, internalError("read item", err)
	}

	list, err := s.listByIDWithAccess(ctx, item.ListID, grant, need)
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
	grant auth.Grant,
	need Access,
) (store.List, error) {
	lists, err := s.store.ListsForMember(ctx, grant.Member.ID)
	if err != nil {
		return store.List{}, internalError("read lists", err)
	}
	shares, err := s.namedSharesFor(ctx, grant.Member)
	if err != nil {
		return store.List{}, err
	}
	for _, list := range lists {
		if list.ID != id {
			continue
		}
		if accessTo(list, grant, shares) >= need {
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
