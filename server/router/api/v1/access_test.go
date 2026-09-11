package v1

import (
	"context"
	"testing"
	"time"

	"connectrpc.com/connect"

	apiv1 "github.com/hoshomoh/nooks/proto/gen/nooks/api/v1"
	"github.com/hoshomoh/nooks/server/auth"
	"github.com/hoshomoh/nooks/store"
)

// shareByName puts a List on named sharing and points it at one Member.
func (f listFixture) shareByName(t *testing.T, owner store.Member, uid string, with store.Member) {
	t.Helper()
	_, err := f.svc.SetListSharing(f.as(t, owner), connect.NewRequest(&apiv1.SetListSharingRequest{
		ListUid: uid, Sharing: apiv1.Sharing_SHARING_SPECIFIC, CanEdit: true,
	}))
	if err != nil {
		t.Fatalf("SetListSharing: %v", err)
	}

	list, err := f.store.ListByUID(t.Context(), uid)
	if err != nil {
		t.Fatalf("ListByUID: %v", err)
	}
	if err := f.store.ReplaceListShares(t.Context(), list.ID, []int64{with.ID}, nil); err != nil {
		t.Fatalf("ReplaceListShares: %v", err)
	}
}

func TestAListSharedByNameIsReachable(t *testing.T) {
	f := newListFixture(t)
	uid := f.createList(t, f.anna, "Groceries")
	f.shareByName(t, f.anna, uid, f.jonas)

	if _, err := f.svc.GetList(f.as(t, f.jonas), connect.NewRequest(&apiv1.GetListRequest{
		ListUid: uid,
	})); err != nil {
		t.Fatalf("GetList as the Member it was shared with: %v", err)
	}

	res, err := f.svc.ListLists(f.as(t, f.jonas), connect.NewRequest(&apiv1.ListListsRequest{}))
	if err != nil {
		t.Fatalf("ListLists: %v", err)
	}
	if len(res.Msg.GetLists()) != 1 {
		t.Errorf("got %d lists, want the one shared by name", len(res.Msg.GetLists()))
	}
}

// Named sharing must reach exactly who it names. Anyone else sees a List that does not
// exist.
func TestNamedSharingReachesNobodyElse(t *testing.T) {
	f := newListFixture(t)
	mira, err := f.store.CreateMember(t.Context(), store.CreateMemberParams{
		UID: "mem_mira", Name: "Mira", Email: "mira@brunnen.lan", Role: store.RoleMember,
		PasswordHash: "hash", CreatedAt: testClock,
	})
	if err != nil {
		t.Fatalf("CreateMember: %v", err)
	}

	uid := f.createList(t, f.anna, "Groceries")
	f.shareByName(t, f.anna, uid, f.jonas)

	_, err = f.svc.GetList(f.as(t, mira), connect.NewRequest(&apiv1.GetListRequest{ListUid: uid}))
	if got := connect.CodeOf(err); got != connect.CodeNotFound {
		t.Errorf("code = %v, want not_found for somebody it was not shared with", got)
	}
}

// Search must not reach further than the Member already could — the property that has
// to hold when the same service answers an Access token.
func TestSearchFindsListsSharedByName(t *testing.T) {
	f := newListFixture(t)
	uid := f.createList(t, f.anna, "Groceries")

	hits := f.search(t, f.jonas, "Groceries")
	if len(hits) != 0 {
		t.Fatalf("got %d hits before sharing, want none", len(hits))
	}

	f.shareByName(t, f.anna, uid, f.jonas)

	hits = f.search(t, f.jonas, "Groceries")
	if len(hits) != 1 {
		t.Errorf("got %d hits after sharing by name, want 1", len(hits))
	}
}

// search runs a query as one Member.
func (f listFixture) search(t *testing.T, member store.Member, query string) []*apiv1.SearchHit {
	t.Helper()
	res, err := f.svc.Search(f.as(t, member), connect.NewRequest(&apiv1.SearchRequest{Query: query}))
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	return res.Msg.GetHits()
}

// withToken runs a request as the Member, narrowed the way a presented token narrows
// it: to the Lists it names, at the permission it carries.
func (f listFixture) withToken(
	t *testing.T,
	member store.Member,
	permission store.Permission,
	listUIDs ...string,
) context.Context {
	t.Helper()

	ids := make([]int64, 0, len(listUIDs))
	for _, uid := range listUIDs {
		list, err := f.store.ListByUID(t.Context(), uid)
		if err != nil {
			t.Fatalf("ListByUID %s: %v", uid, err)
		}
		ids = append(ids, list.ID)
	}

	token := store.AccessToken{ID: 1, MemberID: member.ID, Permission: permission}
	return auth.WithGrant(t.Context(), auth.NewTokenGrant(member, token, ids))
}

// A List a token does not name must be invisible, not forbidden: a token must not be a
// way to learn which Lists its Member has.
func TestATokenReachesOnlyTheListsItNames(t *testing.T) {
	f := newListFixture(t)
	groceries := f.createList(t, f.anna, "Groceries")
	bike := f.createList(t, f.anna, "Bike")

	ctx := f.withToken(t, f.anna, store.PermissionWrite, groceries)

	if _, err := f.svc.GetList(ctx, connect.NewRequest(&apiv1.GetListRequest{
		ListUid: groceries,
	})); err != nil {
		t.Fatalf("GetList on the List the token names: %v", err)
	}

	_, err := f.svc.GetList(ctx, connect.NewRequest(&apiv1.GetListRequest{ListUid: bike}))
	if got := connect.CodeOf(err); got != connect.CodeNotFound {
		t.Errorf("code = %v, want not_found for a List the token does not name", got)
	}
}

// The sidebar is a way of learning what exists too, so it narrows with everything else.
func TestATokenSeesOnlyItsOwnListsEverywhere(t *testing.T) {
	f := newListFixture(t)
	groceries := f.createList(t, f.anna, "Groceries")
	f.createList(t, f.anna, "Bike")

	ctx := f.withToken(t, f.anna, store.PermissionRead, groceries)

	lists, err := f.svc.ListLists(ctx, connect.NewRequest(&apiv1.ListListsRequest{}))
	if err != nil {
		t.Fatalf("ListLists: %v", err)
	}
	if got := len(lists.Msg.GetLists()); got != 1 {
		t.Errorf("got %d lists, want only the one the token names", got)
	}

	hits, err := f.svc.Search(ctx, connect.NewRequest(&apiv1.SearchRequest{Query: "Bike"}))
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if got := len(hits.Msg.GetHits()); got != 0 {
		t.Errorf("got %d hits, want none for a List the token does not name", got)
	}
}

// A read token is its Member's access with one edge taken off: they own the List and
// may still only look.
func TestAReadTokenCannotWriteToItsOwnersList(t *testing.T) {
	f := newListFixture(t)
	groceries := f.createList(t, f.anna, "Groceries")
	ctx := f.withToken(t, f.anna, store.PermissionRead, groceries)

	if _, err := f.svc.GetList(ctx, connect.NewRequest(&apiv1.GetListRequest{
		ListUid: groceries,
	})); err != nil {
		t.Fatalf("GetList with a read token: %v", err)
	}

	_, err := f.svc.CreateItem(ctx, connect.NewRequest(&apiv1.CreateItemRequest{
		ListUid: groceries, Label: "Milk",
	}))
	if got := connect.CodeOf(err); got != connect.CodePermissionDenied {
		t.Errorf("code = %v, want permission_denied for a read token", got)
	}
}

// A row says who did it and only then what through: a token is somebody's access
// narrowed, never an identity of its own.
func TestAnItemAddedByATokenNamesBoth(t *testing.T) {
	f := newListFixture(t)
	groceries := f.createList(t, f.anna, "Groceries")

	list, err := f.store.ListByUID(t.Context(), groceries)
	if err != nil {
		t.Fatalf("ListByUID: %v", err)
	}
	token, err := f.store.CreateAccessToken(t.Context(), store.CreateAccessTokenParams{
		UID: "tok_1", MemberID: f.anna.ID, Name: "Kitchen tablet", TokenHash: "hash-1",
		Permission: store.PermissionWrite, ListIDs: []int64{list.ID}, At: testClock,
	})
	if err != nil {
		t.Fatalf("CreateAccessToken: %v", err)
	}

	ctx := auth.WithGrant(t.Context(),
		auth.NewTokenGrant(f.anna, token, []int64{list.ID}))
	made, err := f.svc.CreateItem(ctx, connect.NewRequest(&apiv1.CreateItemRequest{
		ListUid: groceries, Label: "Milk",
	}))
	if err != nil {
		t.Fatalf("CreateItem through a token: %v", err)
	}

	if got := made.Msg.GetItem().GetAddedByName(); got != f.anna.Name {
		t.Errorf("addedByName = %q, want the Member who owns the token", got)
	}
	if got := made.Msg.GetItem().GetAddedViaToken(); got != "Kitchen tablet" {
		t.Errorf("addedViaToken = %q, want the token's name", got)
	}
}

// Revoking a key must not take what it added off the List. The row stops saying what it
// came through, which is the only honest thing left to say.
func TestRevokingATokenLeavesWhatItAdded(t *testing.T) {
	f := newListFixture(t)
	groceries := f.createList(t, f.anna, "Groceries")

	list, err := f.store.ListByUID(t.Context(), groceries)
	if err != nil {
		t.Fatalf("ListByUID: %v", err)
	}
	token, err := f.store.CreateAccessToken(t.Context(), store.CreateAccessTokenParams{
		UID: "tok_1", MemberID: f.anna.ID, Name: "Kitchen tablet", TokenHash: "hash-1",
		Permission: store.PermissionWrite, ListIDs: []int64{list.ID}, At: testClock,
	})
	if err != nil {
		t.Fatalf("CreateAccessToken: %v", err)
	}

	ctx := auth.WithGrant(t.Context(), auth.NewTokenGrant(f.anna, token, []int64{list.ID}))
	if _, err := f.svc.CreateItem(ctx, connect.NewRequest(&apiv1.CreateItemRequest{
		ListUid: groceries, Label: "Milk",
	})); err != nil {
		t.Fatalf("CreateItem: %v", err)
	}

	if err := f.store.DeleteAccessToken(t.Context(), token.ID); err != nil {
		t.Fatalf("DeleteAccessToken: %v", err)
	}

	read, err := f.svc.GetList(f.as(t, f.anna), connect.NewRequest(&apiv1.GetListRequest{
		ListUid: groceries,
	}))
	if err != nil {
		t.Fatalf("GetList: %v", err)
	}

	items := read.Msg.GetItems()
	if len(items) != 1 || items[0].GetLabel() != "Milk" {
		t.Fatalf("got %d items, want the one the token added", len(items))
	}
	if got := items[0].GetAddedByName(); got != f.anna.Name {
		t.Errorf("addedByName = %q, want the Member it belonged to", got)
	}
	if got := items[0].GetAddedViaToken(); got != "" {
		t.Errorf("addedViaToken = %q, want nothing once the token is gone", got)
	}
}

// A Grant puts its Member on the context, so a token looks like its owner to anything
// that only asks who is here. That is what makes it convenient, and exactly why the
// account-level doors have to say no.
func TestATokenCannotChangeTheAccountItBelongsTo(t *testing.T) {
	f := newListFixture(t)
	groceries := f.createList(t, f.anna, "Groceries")
	ctx := f.withToken(t, f.anna, store.PermissionWrite, groceries)

	auths := NewAuthService(f.store, AuthServiceOptions{Now: func() time.Time { return testClock }})
	_, err := auths.ReplacePassword(ctx, connect.NewRequest(&apiv1.ReplacePasswordRequest{
		CurrentPassword: "hunter2-hunter2", NewPassword: "a-much-longer-one",
	}))
	if got := connect.CodeOf(err); got != connect.CodePermissionDenied {
		t.Errorf("ReplacePassword code = %v, want permission_denied", got)
	}
}

// Anna is an Admin. Her token must not be.
func TestATokenCannotActAsAnAdmin(t *testing.T) {
	f := newListFixture(t)
	groceries := f.createList(t, f.anna, "Groceries")
	ctx := f.withToken(t, f.anna, store.PermissionWrite, groceries)

	if _, err := requireAdmin(ctx); connect.CodeOf(err) != connect.CodePermissionDenied {
		t.Errorf("requireAdmin code = %v, want permission_denied", connect.CodeOf(err))
	}

	// The same context is still a perfectly good caller for what the token is for.
	if _, err := requireGrant(ctx); err != nil {
		t.Errorf("requireGrant: %v, want the token to still be a caller", err)
	}
}
