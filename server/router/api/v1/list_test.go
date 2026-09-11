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

	member := func(uid, name, email string, role store.Role) store.Member {
		t.Helper()
		m, err := s.CreateMember(t.Context(), store.CreateMemberParams{
			UID: uid, Name: name, Email: email, Role: role,
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
		// Anna deployed the Instance, so she is its Admin — the first Member always is.
		anna:  member("mem_anna", "Anna", "anna@brunnen.lan", store.RoleAdmin),
		jonas: member("mem_jonas", "Jonas", "jonas@brunnen.lan", store.RoleMember),
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

// addDatedItem puts an Item on a List with a due date.
func (f listFixture) addDatedItem(t *testing.T, listUid, label, dueOn string) {
	t.Helper()
	if _, err := f.svc.CreateItem(f.as(t, f.anna), connect.NewRequest(&apiv1.CreateItemRequest{
		ListUid: listUid, Label: label, DueOn: dueOn,
	})); err != nil {
		t.Fatalf("CreateItem %s: %v", label, err)
	}
}

// Today gathers everything overdue as well as everything due, which is what an empty
// lower bound is for.
func TestDatedItemsWithNoLowerBoundIncludeOverdue(t *testing.T) {
	f := newListFixture(t)
	uid := f.createList(t, f.anna, "Groceries")

	f.addDatedItem(t, uid, "Return the deposit form", "2026-08-22")
	f.addDatedItem(t, uid, "Descale the kettle", "2026-08-25")
	f.addDatedItem(t, uid, "Ferry tickets", "2026-09-05")

	res, err := f.svc.ListDatedItems(f.as(t, f.anna), connect.NewRequest(&apiv1.ListDatedItemsRequest{
		To: "2026-08-25",
	}))
	if err != nil {
		t.Fatalf("ListDatedItems: %v", err)
	}
	if len(res.Msg.GetItems()) != 2 {
		t.Fatalf("items = %d, want the overdue one and today's", len(res.Msg.GetItems()))
	}
	// Earliest first.
	if got := res.Msg.GetItems()[0].GetItem().GetLabel(); got != "Return the deposit form" {
		t.Errorf("first = %q, want the overdue one", got)
	}
}

// Upcoming is the same answer with both bounds.
func TestDatedItemsWithinAWindow(t *testing.T) {
	f := newListFixture(t)
	uid := f.createList(t, f.anna, "Groceries")

	f.addDatedItem(t, uid, "Return the deposit form", "2026-08-22")
	f.addDatedItem(t, uid, "Ferry tickets", "2026-09-05")

	res, err := f.svc.ListDatedItems(f.as(t, f.anna), connect.NewRequest(&apiv1.ListDatedItemsRequest{
		From: "2026-08-26", To: "2026-09-08",
	}))
	if err != nil {
		t.Fatalf("ListDatedItems: %v", err)
	}
	if len(res.Msg.GetItems()) != 1 {
		t.Fatalf("items = %d, want only the one inside the window", len(res.Msg.GetItems()))
	}
	if got := res.Msg.GetItems()[0].GetListName(); got != "Groceries" {
		t.Errorf("ListName = %q, want the List named so the row reads away from it", got)
	}
}

// Undated Items never appear: a calendar is a lens on dates, not a second home.
func TestDatedItemsExcludeUndatedAndTicked(t *testing.T) {
	f := newListFixture(t)
	uid := f.createList(t, f.anna, "Groceries")
	ctx := f.as(t, f.anna)

	f.addDatedItem(t, uid, "Descale the kettle", "2026-08-25")
	if _, err := f.svc.CreateItem(ctx, connect.NewRequest(&apiv1.CreateItemRequest{
		ListUid: uid, Label: "Baking paper",
	})); err != nil {
		t.Fatalf("CreateItem: %v", err)
	}

	ticked, err := f.svc.CreateItem(ctx, connect.NewRequest(&apiv1.CreateItemRequest{
		ListUid: uid, Label: "Milk", DueOn: "2026-08-25",
	}))
	if err != nil {
		t.Fatalf("CreateItem: %v", err)
	}
	if _, err := f.svc.SetItemDone(ctx, connect.NewRequest(&apiv1.SetItemDoneRequest{
		ItemUid: ticked.Msg.GetItem().GetUid(), Done: true,
	})); err != nil {
		t.Fatalf("SetItemDone: %v", err)
	}

	res, err := f.svc.ListDatedItems(ctx, connect.NewRequest(&apiv1.ListDatedItemsRequest{
		To: "2026-08-31",
	}))
	if err != nil {
		t.Fatalf("ListDatedItems: %v", err)
	}
	if len(res.Msg.GetItems()) != 1 {
		t.Fatalf("items = %d, want only the dated, unticked one", len(res.Msg.GetItems()))
	}
}

// Dated views must not reach further than the Member's own Lists.
func TestDatedItemsRespectSharing(t *testing.T) {
	f := newListFixture(t)
	uid := f.createList(t, f.anna, "Secret plans")
	f.addDatedItem(t, uid, "Surprise party", "2026-08-25")

	res, err := f.svc.ListDatedItems(f.as(t, f.jonas), connect.NewRequest(&apiv1.ListDatedItemsRequest{
		To: "2026-08-31",
	}))
	if err != nil {
		t.Fatalf("ListDatedItems: %v", err)
	}
	if len(res.Msg.GetItems()) != 0 {
		t.Errorf("items = %d, want none of somebody else's private List", len(res.Msg.GetItems()))
	}
}

func TestDatedItemsNeedsALastDay(t *testing.T) {
	f := newListFixture(t)
	_, err := f.svc.ListDatedItems(f.as(t, f.anna), connect.NewRequest(&apiv1.ListDatedItemsRequest{}))
	if got := connect.CodeOf(err); got != connect.CodeInvalidArgument {
		t.Errorf("code = %v, want invalid_argument", got)
	}
}

// A Note is markdown, kept as the Member wrote it.
func TestANoteIsStoredAndReturnedAsWritten(t *testing.T) {
	f := newListFixture(t)
	uid := f.createList(t, f.anna, "Groceries")
	ctx := f.as(t, f.anna)

	added, err := f.svc.CreateItem(ctx, connect.NewRequest(&apiv1.CreateItemRequest{
		ListUid: uid, Label: "Coffee",
	}))
	if err != nil {
		t.Fatalf("CreateItem: %v", err)
	}

	markdown := "### Where\nSaturday market, second row.\n\n- [ ] Ethiopian, whole bean\n> They pack up around two."
	updated, err := f.svc.UpdateItem(ctx, connect.NewRequest(&apiv1.UpdateItemRequest{
		ItemUid: added.Msg.GetItem().GetUid(), Note: &markdown,
	}))
	if err != nil {
		t.Fatalf("UpdateItem: %v", err)
	}
	if got := updated.Msg.GetItem().GetNote(); got != markdown {
		t.Errorf("Note = %q, want it byte for byte", got)
	}
}

// The row shows the Note's own first line and a count of the rest — never a summary.
func TestANoteGivesTheRowItsPreview(t *testing.T) {
	f := newListFixture(t)
	uid := f.createList(t, f.anna, "Groceries")
	ctx := f.as(t, f.anna)

	added, err := f.svc.CreateItem(ctx, connect.NewRequest(&apiv1.CreateItemRequest{
		ListUid: uid, Label: "Coffee",
	}))
	if err != nil {
		t.Fatalf("CreateItem: %v", err)
	}

	markdown := "### Where\nSaturday market.\nThey pack up around two."
	updated, err := f.svc.UpdateItem(ctx, connect.NewRequest(&apiv1.UpdateItemRequest{
		ItemUid: added.Msg.GetItem().GetUid(), Note: &markdown,
	}))
	if err != nil {
		t.Fatalf("UpdateItem: %v", err)
	}

	item := updated.Msg.GetItem()
	if item.GetNoteFirstLine() != "Where" {
		t.Errorf("NoteFirstLine = %q, want the heading with its marker stripped", item.GetNoteFirstLine())
	}
	if item.GetNoteRemainingLines() != 2 {
		t.Errorf("NoteRemainingLines = %d, want 2", item.GetNoteRemainingLines())
	}
}

// A Note is searchable, so a Member can find an Item by what they wrote about it.
func TestANoteIsSearchable(t *testing.T) {
	f := newListFixture(t)
	uid := f.createList(t, f.anna, "Groceries")
	ctx := f.as(t, f.anna)

	added, err := f.svc.CreateItem(ctx, connect.NewRequest(&apiv1.CreateItemRequest{
		ListUid: uid, Label: "Coffee",
	}))
	if err != nil {
		t.Fatalf("CreateItem: %v", err)
	}

	markdown := "Saturday market, second row from the Kastanienallee entrance."
	if _, err := f.svc.UpdateItem(ctx, connect.NewRequest(&apiv1.UpdateItemRequest{
		ItemUid: added.Msg.GetItem().GetUid(), Note: &markdown,
	})); err != nil {
		t.Fatalf("UpdateItem: %v", err)
	}

	res, err := f.svc.Search(ctx, connect.NewRequest(&apiv1.SearchRequest{Query: "kastanienallee"}))
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(res.Msg.GetHits()) == 0 {
		t.Fatal("searching a word only in the Note found nothing")
	}
	if got := res.Msg.GetHits()[0].GetKind(); got != apiv1.SearchHitKind_SEARCH_HIT_KIND_NOTE {
		t.Errorf("kind = %v, want a note hit so the result can say where the words were", got)
	}
}

// Emptying a Note takes it out of the index.
func TestClearingANoteRemovesItFromSearch(t *testing.T) {
	f := newListFixture(t)
	uid := f.createList(t, f.anna, "Groceries")
	ctx := f.as(t, f.anna)

	added, err := f.svc.CreateItem(ctx, connect.NewRequest(&apiv1.CreateItemRequest{
		ListUid: uid, Label: "Coffee",
	}))
	if err != nil {
		t.Fatalf("CreateItem: %v", err)
	}

	markdown := "Kastanienallee entrance"
	if _, err := f.svc.UpdateItem(ctx, connect.NewRequest(&apiv1.UpdateItemRequest{
		ItemUid: added.Msg.GetItem().GetUid(), Note: &markdown,
	})); err != nil {
		t.Fatalf("UpdateItem: %v", err)
	}

	empty := ""
	if _, err := f.svc.UpdateItem(ctx, connect.NewRequest(&apiv1.UpdateItemRequest{
		ItemUid: added.Msg.GetItem().GetUid(), Note: &empty,
	})); err != nil {
		t.Fatalf("UpdateItem: %v", err)
	}

	res, err := f.svc.Search(ctx, connect.NewRequest(&apiv1.SearchRequest{Query: "kastanienallee"}))
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(res.Msg.GetHits()) != 0 {
		t.Errorf("hits = %d, want none once the Note is gone", len(res.Msg.GetHits()))
	}
}

// A ticked Item stays on its List. It moves to the done section rather than leaving,
// and untick puts it back — a tick is not a delete.
func TestATickedItemStaysAndCanBeUnticked(t *testing.T) {
	f := newListFixture(t)
	uid := f.createList(t, f.anna, "Groceries")
	ctx := f.as(t, f.anna)

	added, err := f.svc.CreateItem(ctx, connect.NewRequest(&apiv1.CreateItemRequest{
		ListUid: uid, Label: "Milk",
	}))
	if err != nil {
		t.Fatalf("CreateItem: %v", err)
	}
	itemUID := added.Msg.GetItem().GetUid()

	if _, err := f.svc.SetItemDone(ctx, connect.NewRequest(&apiv1.SetItemDoneRequest{
		ItemUid: itemUID, Done: true,
	})); err != nil {
		t.Fatalf("tick: %v", err)
	}

	ticked, err := f.svc.GetList(ctx, connect.NewRequest(&apiv1.GetListRequest{ListUid: uid}))
	if err != nil {
		t.Fatalf("GetList: %v", err)
	}
	if len(ticked.Msg.GetItems()) != 1 {
		t.Fatalf("got %d items after ticking, want the ticked one still there", len(ticked.Msg.GetItems()))
	}
	if !ticked.Msg.GetItems()[0].GetDone() {
		t.Error("Done = false on the Item that was just ticked")
	}

	untickedItem, err := f.svc.SetItemDone(ctx, connect.NewRequest(&apiv1.SetItemDoneRequest{
		ItemUid: itemUID, Done: false,
	}))
	if err != nil {
		t.Fatalf("untick: %v", err)
	}
	if untickedItem.Msg.GetItem().GetDone() {
		t.Error("Done = true after unticking")
	}
	if got := untickedItem.Msg.GetItem().GetDoneByName(); got != "" {
		t.Errorf("DoneByName = %q after unticking, want nobody", got)
	}
}

// An Item's name is changed where it is read, so every view has to be able to send one.
func TestRenamingAnItem(t *testing.T) {
	f := newListFixture(t)
	uid := f.createList(t, f.anna, "Groceries")
	ctx := f.as(t, f.anna)

	added, err := f.svc.CreateItem(ctx, connect.NewRequest(&apiv1.CreateItemRequest{
		ListUid: uid, Label: "Milk",
	}))
	if err != nil {
		t.Fatalf("CreateItem: %v", err)
	}

	label := "Oat milk"
	renamed, err := f.svc.UpdateItem(ctx, connect.NewRequest(&apiv1.UpdateItemRequest{
		ItemUid: added.Msg.GetItem().GetUid(), Label: &label,
	}))
	if err != nil {
		t.Fatalf("UpdateItem: %v", err)
	}
	if got := renamed.Msg.GetItem().GetLabel(); got != "Oat milk" {
		t.Errorf("Label = %q, want the new name", got)
	}
}

// A duplicate is a new List that begins with the same things on it, not a second view
// of the first.
func TestDuplicatingAList(t *testing.T) {
	f := newListFixture(t)
	uid := f.createList(t, f.anna, "Groceries")
	ctx := f.as(t, f.anna)

	for _, label := range []string{"Milk", "Oats"} {
		if _, err := f.svc.CreateItem(ctx, connect.NewRequest(&apiv1.CreateItemRequest{
			ListUid: uid, Label: label,
		})); err != nil {
			t.Fatalf("CreateItem: %v", err)
		}
	}

	copied, err := f.svc.DuplicateList(ctx, connect.NewRequest(&apiv1.DuplicateListRequest{
		ListUid: uid,
	}))
	if err != nil {
		t.Fatalf("DuplicateList: %v", err)
	}
	if got := copied.Msg.GetList().GetName(); got != "Groceries (copy)" {
		t.Errorf("name = %q", got)
	}
	if got := copied.Msg.GetList().GetSharing(); got != apiv1.Sharing_SHARING_PRIVATE {
		t.Errorf("sharing = %v, want private — a copy is yours", got)
	}

	items, err := f.svc.GetList(ctx, connect.NewRequest(&apiv1.GetListRequest{
		ListUid: copied.Msg.GetList().GetUid(),
	}))
	if err != nil {
		t.Fatalf("GetList: %v", err)
	}
	if len(items.Msg.GetItems()) != 2 {
		t.Errorf("got %d items, want both of them", len(items.Msg.GetItems()))
	}
}

// A duplicate is made to do the same thing again, not to remember the last time.
func TestDuplicatingLeavesTheTickedItemsBehind(t *testing.T) {
	f := newListFixture(t)
	uid := f.createList(t, f.anna, "Groceries")
	ctx := f.as(t, f.anna)

	added, err := f.svc.CreateItem(ctx, connect.NewRequest(&apiv1.CreateItemRequest{
		ListUid: uid, Label: "Milk",
	}))
	if err != nil {
		t.Fatalf("CreateItem: %v", err)
	}
	if _, err := f.svc.SetItemDone(ctx, connect.NewRequest(&apiv1.SetItemDoneRequest{
		ItemUid: added.Msg.GetItem().GetUid(), Done: true,
	})); err != nil {
		t.Fatalf("SetItemDone: %v", err)
	}

	copied, err := f.svc.DuplicateList(ctx, connect.NewRequest(&apiv1.DuplicateListRequest{
		ListUid: uid,
	}))
	if err != nil {
		t.Fatalf("DuplicateList: %v", err)
	}

	items, err := f.svc.GetList(ctx, connect.NewRequest(&apiv1.GetListRequest{
		ListUid: copied.Msg.GetList().GetUid(),
	}))
	if err != nil {
		t.Fatalf("GetList: %v", err)
	}
	if len(items.Msg.GetItems()) != 0 {
		t.Errorf("got %d items, want the ticked one left behind", len(items.Msg.GetItems()))
	}
}

// Anyone who may read a List may make their own copy of it.
func TestDuplicatingSomebodyElsesList(t *testing.T) {
	f := newListFixture(t)
	uid := f.createList(t, f.anna, "Groceries")
	f.share(t, f.anna, uid, false)

	copied, err := f.svc.DuplicateList(f.as(t, f.jonas), connect.NewRequest(
		&apiv1.DuplicateListRequest{ListUid: uid},
	))
	if err != nil {
		t.Fatalf("DuplicateList: %v", err)
	}
	if !copied.Msg.GetList().GetIsOwner() {
		t.Error("IsOwner = false, want the copy to belong to whoever made it")
	}
}
