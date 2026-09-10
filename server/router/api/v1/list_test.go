package v1

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"connectrpc.com/connect"

	apiv1 "github.com/hoshomoh/nooks/proto/gen/nooks/api/v1"
	"github.com/hoshomoh/nooks/server/auth"
	"github.com/hoshomoh/nooks/store"
)

// listFixture is a store with two Members, and the service under test.
type listFixture struct {
	svc   *ListService
	store store.Store
	anna  store.Member
	jonas store.Member
}

// newListFixture builds the service over a real SQLite store with identifiers that are
// stable, so assertions do not depend on chance.
func newListFixture(t *testing.T) listFixture {
	t.Helper()

	s, err := store.OpenSQLite(t.Context(), filepath.Join(t.TempDir(), "nooks.db"))
	if err != nil {
		t.Fatalf("OpenSQLite: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })

	member := func(uid, name, email string) store.Member {
		t.Helper()
		m, err := s.CreateMember(t.Context(), store.CreateMemberParams{
			UID: uid, Name: name, Email: email, Role: store.RoleMember,
			PasswordHash: "hash", CreatedAt: testClock,
		})
		if err != nil {
			t.Fatalf("CreateMember %s: %v", name, err)
		}
		return m
	}

	issued := 0
	svc := NewListService(s, func() time.Time { return testClock }, func() (string, error) {
		issued++
		return fmt.Sprintf("uid-%d", issued), nil
	})

	return listFixture{
		svc:   svc,
		store: s,
		anna:  member("mem_anna", "Anna", "anna@brunnen.lan"),
		jonas: member("mem_jonas", "Jonas", "jonas@brunnen.lan"),
	}
}

// as returns a context signed in as the given Member.
func (f listFixture) as(t *testing.T, member store.Member) context.Context {
	t.Helper()
	return auth.WithMember(t.Context(), member)
}

// createList adds a List owned by member and returns its uid.
func (f listFixture) createList(t *testing.T, member store.Member, name string) string {
	t.Helper()
	res, err := f.svc.CreateList(f.as(t, member), connect.NewRequest(&apiv1.CreateListRequest{Name: name}))
	if err != nil {
		t.Fatalf("CreateList %s: %v", name, err)
	}
	return res.Msg.GetList().GetUid()
}

// share makes a List reachable by the whole Instance.
func (f listFixture) share(t *testing.T, member store.Member, uid string, canEdit bool) {
	t.Helper()
	_, err := f.svc.SetListSharing(f.as(t, member), connect.NewRequest(&apiv1.SetListSharingRequest{
		ListUid: uid, Sharing: apiv1.Sharing_SHARING_INSTANCE, CanEdit: canEdit,
	}))
	if err != nil {
		t.Fatalf("SetListSharing: %v", err)
	}
}

func TestCreateListStartsPrivate(t *testing.T) {
	f := newListFixture(t)

	res, err := f.svc.CreateList(f.as(t, f.anna), connect.NewRequest(&apiv1.CreateListRequest{
		Name: "Groceries",
	}))
	if err != nil {
		t.Fatalf("CreateList: %v", err)
	}
	if got := res.Msg.GetList().GetSharing(); got != apiv1.Sharing_SHARING_PRIVATE {
		t.Errorf("sharing = %v, want private — sharing is a deliberate second step", got)
	}
	if !res.Msg.GetList().GetIsOwner() {
		t.Error("IsOwner = false for the Member who made it")
	}
}

// A private List must be invisible, not merely forbidden: telling the two apart lets
// anyone probe for Lists.
func TestAPrivateListIsInvisibleToOthers(t *testing.T) {
	f := newListFixture(t)
	uid := f.createList(t, f.anna, "Bike")

	_, err := f.svc.GetList(f.as(t, f.jonas), connect.NewRequest(&apiv1.GetListRequest{ListUid: uid}))
	if got := connect.CodeOf(err); got != connect.CodeNotFound {
		t.Errorf("code = %v, want not_found so it is indistinguishable from absent", got)
	}

	res, err := f.svc.ListLists(f.as(t, f.jonas), connect.NewRequest(&apiv1.ListListsRequest{}))
	if err != nil {
		t.Fatalf("ListLists: %v", err)
	}
	if len(res.Msg.GetLists()) != 0 {
		t.Errorf("lists = %v, want none of somebody else's private Lists", res.Msg.GetLists())
	}
}

func TestASharedListIsReachable(t *testing.T) {
	f := newListFixture(t)
	uid := f.createList(t, f.anna, "Groceries")
	f.share(t, f.anna, uid, true)

	res, err := f.svc.GetList(f.as(t, f.jonas), connect.NewRequest(&apiv1.GetListRequest{ListUid: uid}))
	if err != nil {
		t.Fatalf("GetList: %v", err)
	}
	if res.Msg.GetList().GetIsOwner() {
		t.Error("IsOwner = true for somebody else's List")
	}
}

// Read-only means see and print, not tick or add.
func TestReadOnlySharingBlocksWriting(t *testing.T) {
	f := newListFixture(t)
	uid := f.createList(t, f.anna, "Groceries")
	f.share(t, f.anna, uid, false)

	if _, err := f.svc.GetList(f.as(t, f.jonas), connect.NewRequest(&apiv1.GetListRequest{ListUid: uid})); err != nil {
		t.Fatalf("GetList on a read-only List: %v", err)
	}

	_, err := f.svc.CreateItem(f.as(t, f.jonas), connect.NewRequest(&apiv1.CreateItemRequest{
		ListUid: uid, Label: "Milk",
	}))
	if got := connect.CodeOf(err); got != connect.CodePermissionDenied {
		t.Errorf("adding to a read-only List = %v, want permission_denied", got)
	}
}

// Only the owner may rename, share or delete.
func TestOnlyTheOwnerMayChangeTheListItself(t *testing.T) {
	f := newListFixture(t)
	uid := f.createList(t, f.anna, "Groceries")
	f.share(t, f.anna, uid, true)

	_, err := f.svc.RenameList(f.as(t, f.jonas), connect.NewRequest(&apiv1.RenameListRequest{
		ListUid: uid, Name: "Jonas's now",
	}))
	if got := connect.CodeOf(err); got != connect.CodePermissionDenied {
		t.Errorf("rename by a non-owner = %v, want permission_denied", got)
	}

	_, err = f.svc.DeleteList(f.as(t, f.jonas), connect.NewRequest(&apiv1.DeleteListRequest{ListUid: uid}))
	if got := connect.CodeOf(err); got != connect.CodePermissionDenied {
		t.Errorf("delete by a non-owner = %v, want permission_denied", got)
	}
}

// Even with write access, a Member can tick and add but not restructure.
func TestAWriterMayAddAndTick(t *testing.T) {
	f := newListFixture(t)
	uid := f.createList(t, f.anna, "Groceries")
	f.share(t, f.anna, uid, true)

	added, err := f.svc.CreateItem(f.as(t, f.jonas), connect.NewRequest(&apiv1.CreateItemRequest{
		ListUid: uid, Label: "Milk", Quantity: "2",
	}))
	if err != nil {
		t.Fatalf("CreateItem: %v", err)
	}
	if got := added.Msg.GetItem().GetAddedByName(); got != "Jonas" {
		t.Errorf("AddedByName = %q, want Jonas", got)
	}

	ticked, err := f.svc.SetItemDone(f.as(t, f.jonas), connect.NewRequest(&apiv1.SetItemDoneRequest{
		ItemUid: added.Msg.GetItem().GetUid(), Done: true,
	}))
	if err != nil {
		t.Fatalf("SetItemDone: %v", err)
	}
	if !ticked.Msg.GetItem().GetDone() {
		t.Error("Done = false after ticking")
	}
	if got := ticked.Msg.GetItem().GetDoneByName(); got != "Jonas" {
		t.Errorf("DoneByName = %q, want Jonas", got)
	}
}

// The sidebar's number is how many Items are not yet ticked.
func TestOpenCountIgnoresTickedItems(t *testing.T) {
	f := newListFixture(t)
	uid := f.createList(t, f.anna, "Groceries")
	ctx := f.as(t, f.anna)

	for _, label := range []string{"Milk", "Oats", "Baking paper"} {
		if _, err := f.svc.CreateItem(ctx, connect.NewRequest(&apiv1.CreateItemRequest{
			ListUid: uid, Label: label,
		})); err != nil {
			t.Fatalf("CreateItem: %v", err)
		}
	}

	listed, err := f.svc.GetList(ctx, connect.NewRequest(&apiv1.GetListRequest{ListUid: uid}))
	if err != nil {
		t.Fatalf("GetList: %v", err)
	}
	if got := listed.Msg.GetList().GetOpenCount(); got != 3 {
		t.Fatalf("open count = %d, want 3", got)
	}

	if _, err := f.svc.SetItemDone(ctx, connect.NewRequest(&apiv1.SetItemDoneRequest{
		ItemUid: listed.Msg.GetItems()[0].GetUid(), Done: true,
	})); err != nil {
		t.Fatalf("SetItemDone: %v", err)
	}

	after, err := f.svc.GetList(ctx, connect.NewRequest(&apiv1.GetListRequest{ListUid: uid}))
	if err != nil {
		t.Fatalf("GetList: %v", err)
	}
	if got := after.Msg.GetList().GetOpenCount(); got != 2 {
		t.Errorf("open count = %d after one tick, want 2", got)
	}
}

// Pinning is per-Member and never affects anyone else's sidebar.
func TestPinningIsPrivateToTheMember(t *testing.T) {
	f := newListFixture(t)
	uid := f.createList(t, f.anna, "Groceries")
	f.share(t, f.anna, uid, true)

	if _, err := f.svc.SetListPinned(f.as(t, f.jonas), connect.NewRequest(&apiv1.SetListPinnedRequest{
		ListUid: uid, Pinned: true,
	})); err != nil {
		t.Fatalf("SetListPinned: %v", err)
	}

	forJonas, err := f.svc.ListLists(f.as(t, f.jonas), connect.NewRequest(&apiv1.ListListsRequest{}))
	if err != nil {
		t.Fatalf("ListLists: %v", err)
	}
	if !forJonas.Msg.GetLists()[0].GetIsPinned() {
		t.Error("Jonas pinned it and does not see it pinned")
	}

	forAnna, err := f.svc.ListLists(f.as(t, f.anna), connect.NewRequest(&apiv1.ListListsRequest{}))
	if err != nil {
		t.Fatalf("ListLists: %v", err)
	}
	if forAnna.Msg.GetLists()[0].GetIsPinned() {
		t.Error("Jonas's pin shows in Anna's sidebar")
	}
}

func TestMoveItemReordersWithoutTouchingTheRest(t *testing.T) {
	f := newListFixture(t)
	uid := f.createList(t, f.anna, "Groceries")
	ctx := f.as(t, f.anna)

	uids := map[string]string{}
	for _, label := range []string{"Milk", "Oats", "Baking paper"} {
		res, err := f.svc.CreateItem(ctx, connect.NewRequest(&apiv1.CreateItemRequest{
			ListUid: uid, Label: label,
		}))
		if err != nil {
			t.Fatalf("CreateItem: %v", err)
		}
		uids[label] = res.Msg.GetItem().GetUid()
	}

	// Put the last Item straight after the first.
	if _, err := f.svc.MoveItem(ctx, connect.NewRequest(&apiv1.MoveItemRequest{
		ItemUid: uids["Baking paper"], AfterItemUid: uids["Milk"],
	})); err != nil {
		t.Fatalf("MoveItem: %v", err)
	}

	res, err := f.svc.GetList(ctx, connect.NewRequest(&apiv1.GetListRequest{ListUid: uid}))
	if err != nil {
		t.Fatalf("GetList: %v", err)
	}
	want := []string{"Milk", "Baking paper", "Oats"}
	for i, label := range want {
		if got := res.Msg.GetItems()[i].GetLabel(); got != label {
			t.Fatalf("order[%d] = %q, want %q", i, got, label)
		}
	}
}

// Moving to the top is what an empty after means.
func TestMoveItemToTheTop(t *testing.T) {
	f := newListFixture(t)
	uid := f.createList(t, f.anna, "Groceries")
	ctx := f.as(t, f.anna)

	var last string
	for _, label := range []string{"Milk", "Oats", "Baking paper"} {
		res, err := f.svc.CreateItem(ctx, connect.NewRequest(&apiv1.CreateItemRequest{
			ListUid: uid, Label: label,
		}))
		if err != nil {
			t.Fatalf("CreateItem: %v", err)
		}
		last = res.Msg.GetItem().GetUid()
	}

	if _, err := f.svc.MoveItem(ctx, connect.NewRequest(&apiv1.MoveItemRequest{
		ItemUid: last, AfterItemUid: "",
	})); err != nil {
		t.Fatalf("MoveItem: %v", err)
	}

	res, err := f.svc.GetList(ctx, connect.NewRequest(&apiv1.GetListRequest{ListUid: uid}))
	if err != nil {
		t.Fatalf("GetList: %v", err)
	}
	if got := res.Msg.GetItems()[0].GetLabel(); got != "Baking paper" {
		t.Errorf("first = %q, want the moved Item", got)
	}
}

func TestADueDateHasToBeADate(t *testing.T) {
	f := newListFixture(t)
	uid := f.createList(t, f.anna, "Groceries")

	_, err := f.svc.CreateItem(f.as(t, f.anna), connect.NewRequest(&apiv1.CreateItemRequest{
		ListUid: uid, Label: "Milk", DueOn: "next friday",
	}))
	if got := connect.CodeOf(err); got != connect.CodeInvalidArgument {
		t.Errorf("code = %v, want invalid_argument", got)
	}
}

// Search must never reach further than the Member already could.
func TestSearchRespectsWhoMaySeeWhat(t *testing.T) {
	f := newListFixture(t)
	ctx := f.as(t, f.anna)

	secret := f.createList(t, f.anna, "Secret plans")
	if _, err := f.svc.CreateItem(ctx, connect.NewRequest(&apiv1.CreateItemRequest{
		ListUid: secret, Label: "Surprise coffee grinder",
	})); err != nil {
		t.Fatalf("CreateItem: %v", err)
	}

	shared := f.createList(t, f.anna, "Groceries")
	f.share(t, f.anna, shared, true)
	if _, err := f.svc.CreateItem(ctx, connect.NewRequest(&apiv1.CreateItemRequest{
		ListUid: shared, Label: "Coffee beans",
	})); err != nil {
		t.Fatalf("CreateItem: %v", err)
	}

	// Anna sees both.
	mine, err := f.svc.Search(ctx, connect.NewRequest(&apiv1.SearchRequest{Query: "coffee"}))
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(mine.Msg.GetHits()) != 2 {
		t.Errorf("Anna's hits = %d, want both of her own", len(mine.Msg.GetHits()))
	}

	// Jonas sees only what is shared with him.
	theirs, err := f.svc.Search(f.as(t, f.jonas), connect.NewRequest(&apiv1.SearchRequest{Query: "coffee"}))
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	for _, hit := range theirs.Msg.GetHits() {
		if hit.GetText() == "Surprise coffee grinder" {
			t.Fatal("search leaked an Item from a private List")
		}
	}
	if len(theirs.Msg.GetHits()) != 1 {
		t.Errorf("Jonas's hits = %d, want only the shared one", len(theirs.Msg.GetHits()))
	}
}

// A result reads in context: the List it belongs to travels with it.
func TestSearchHitsNameTheirList(t *testing.T) {
	f := newListFixture(t)
	ctx := f.as(t, f.anna)
	uid := f.createList(t, f.anna, "Groceries")

	if _, err := f.svc.CreateItem(ctx, connect.NewRequest(&apiv1.CreateItemRequest{
		ListUid: uid, Label: "Coffee beans",
	})); err != nil {
		t.Fatalf("CreateItem: %v", err)
	}

	res, err := f.svc.Search(ctx, connect.NewRequest(&apiv1.SearchRequest{Query: "coffee beans"}))
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(res.Msg.GetHits()) == 0 {
		t.Fatal("no hits")
	}
	hit := res.Msg.GetHits()[0]
	if hit.GetListName() != "Groceries" {
		t.Errorf("ListName = %q, want Groceries", hit.GetListName())
	}
	if hit.GetListUid() != uid {
		t.Errorf("ListUid = %q, want the List to open", hit.GetListUid())
	}
	if hit.GetItemUid() == "" {
		t.Error("ItemUid is empty for an Item hit")
	}
}

func TestEverythingNeedsASession(t *testing.T) {
	f := newListFixture(t)

	_, err := f.svc.ListLists(t.Context(), connect.NewRequest(&apiv1.ListListsRequest{}))
	if got := connect.CodeOf(err); got != connect.CodeUnauthenticated {
		t.Errorf("ListLists signed out = %v, want unauthenticated", got)
	}
	_, err = f.svc.Search(t.Context(), connect.NewRequest(&apiv1.SearchRequest{Query: "coffee"}))
	if got := connect.CodeOf(err); got != connect.CodeUnauthenticated {
		t.Errorf("Search signed out = %v, want unauthenticated", got)
	}
}
