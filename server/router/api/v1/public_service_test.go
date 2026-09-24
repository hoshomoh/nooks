package v1

import (
	"context"
	"testing"
	"time"

	"connectrpc.com/connect"

	apiv1 "github.com/hoshomoh/nooks/proto/gen/nooks/api/v1"
	"github.com/hoshomoh/nooks/store"
)

// publish puts a List on the public page with the given settings.
func (f listFixture) publish(t *testing.T, listUID string, public store.PublicList) {
	t.Helper()
	settings, err := f.store.InstanceSettings(t.Context())
	if err != nil {
		t.Fatalf("InstanceSettings: %v", err)
	}
	settings.Name = "Brunnen Street"
	public.ListUID = listUID
	settings.Public = public
	if err := f.store.SaveInstanceSettings(t.Context(), settings); err != nil {
		t.Fatalf("SaveInstanceSettings: %v", err)
	}
}

// readPublic asks for the public page as a Visitor would: with no session at all.
func (f listFixture) readPublic(t *testing.T) *apiv1.GetPublicListResponse {
	t.Helper()
	res, err := NewPublicService(f.store, nil).GetPublicList(
		t.Context(), connect.NewRequest(&apiv1.GetPublicListRequest{}),
	)
	if err != nil {
		t.Fatalf("GetPublicList: %v", err)
	}
	return res.Msg
}

// The page needs no account, which is the whole point of it.
func TestThePublicListNeedsNoSession(t *testing.T) {
	f := newListFixture(t)
	uid := f.createList(t, f.anna, "Groceries")
	f.addItem(t, f.anna, uid, "Milk")
	f.publish(t, uid, store.PublicList{ShowMeta: true})

	page := f.readPublic(t)
	if !page.GetPublished() {
		t.Fatal("Published = false for a published List")
	}
	if page.GetListName() != "Groceries" {
		t.Errorf("list = %q", page.GetListName())
	}
	if len(page.GetItems()) != 1 {
		t.Fatalf("got %d items, want the one on it", len(page.GetItems()))
	}
}

// An Instance with nothing published answers the same way to everybody: there is no
// version of this that reveals which Lists exist.
func TestNothingPublishedSaysSo(t *testing.T) {
	f := newListFixture(t)
	f.createList(t, f.anna, "Groceries")

	page := f.readPublic(t)
	if page.GetPublished() {
		t.Error("Published = true when nothing is")
	}
	if page.GetListName() != "" || len(page.GetItems()) != 0 {
		t.Error("an unpublished Instance named a List anyway")
	}
}

// A public page is about what needs buying, not about who is in the household.
func TestContributorNamesAreOffUnlessAskedFor(t *testing.T) {
	f := newListFixture(t)
	uid := f.createList(t, f.anna, "Groceries")
	f.addItem(t, f.anna, uid, "Milk")
	f.publish(t, uid, store.PublicList{ShowNames: false, ShowMeta: true})

	if got := f.readPublic(t).GetItems()[0].GetAddedByName(); got != "" {
		t.Errorf("AddedByName = %q, want nothing", got)
	}
}

func TestContributorNamesWhenTheInstanceShowsThem(t *testing.T) {
	f := newListFixture(t)
	uid := f.createList(t, f.anna, "Groceries")
	f.addItem(t, f.anna, uid, "Milk")
	f.publish(t, uid, store.PublicList{ShowNames: true})

	if got := f.readPublic(t).GetItems()[0].GetAddedByName(); got != "Anna" {
		t.Errorf("AddedByName = %q, want Anna", got)
	}
}

// A field the settings turn off is absent rather than blank, so there is nothing to
// leak by rendering it.
func TestQuantitiesAndDatesAreOffUnlessAskedFor(t *testing.T) {
	f := newListFixture(t)
	uid := f.createList(t, f.anna, "Groceries")

	if _, err := f.svc.CreateItem(f.as(t, f.anna), connect.NewRequest(&apiv1.CreateItemRequest{
		ListUid: uid, Label: "Tomatoes", Quantity: "1 kg", DueOn: "2026-08-30",
	})); err != nil {
		t.Fatalf("CreateItem: %v", err)
	}
	f.publish(t, uid, store.PublicList{ShowMeta: false})

	item := f.readPublic(t).GetItems()[0]
	if item.GetQuantity() != "" || item.GetDueOn() != "" {
		t.Errorf("quantity = %q and due = %q, want neither", item.GetQuantity(), item.GetDueOn())
	}
}

// The published List was deleted. A Visitor cannot do anything about it either way, so
// the page says nothing is published rather than showing them an error.
func TestAPublishedListThatIsGone(t *testing.T) {
	f := newListFixture(t)
	f.publish(t, "list_that_never_was", store.PublicList{ShowMeta: true})

	if f.readPublic(t).GetPublished() {
		t.Error("Published = true for a List that is not there")
	}
}

// A page somebody keeps open on the way to the shop has to say how much is left and
// how fresh it is.
func TestThePublicPageSaysWhatIsLeftAndWhenItChanged(t *testing.T) {
	f := newListFixture(t)
	uid := f.createList(t, f.anna, "Groceries")
	f.addItem(t, f.anna, uid, "Milk")
	f.addItem(t, f.anna, uid, "Oats")
	f.publish(t, uid, store.PublicList{})

	res := f.readPublic(t)

	if got := res.GetOpenCount(); got != 2 {
		t.Errorf("openCount = %d, want 2", got)
	}
	if res.GetUpdatedAt() == "" {
		t.Error("updatedAt is empty, want when the list last changed")
	}
}

// countingStore is a Store that says how many times the whole List was read.
type countingStore struct {
	store.Store
	reads int
}

func (c *countingStore) ItemsOnList(ctx context.Context, listID int64) ([]store.Item, error) {
	c.reads++
	return c.Store.ItemsOnList(ctx, listID)
}

/*
The page is built again when it could have changed, and not for every Visitor.

Nothing bounded this endpoint. It needs no session, has no rate limit, and reads every
Item on the List, which is 653ms at ten thousand of them. A caller asking a thousand
times paid that a thousand times, and paging the answer would not have helped: the cost
is the repetition, not the size of one reply.

Three things are checked because the cache has to be wrong in none of them: it stops
reading when nothing changed, it notices a change at once rather than when a timer says
so, and it gives up on its own after a while for the changes it cannot see.
*/
func TestThePublicPageIsNotRebuiltForEveryVisitor(t *testing.T) {
	f := newListFixture(t)
	uid := f.createList(t, f.anna, "Groceries")
	f.addItem(t, f.anna, uid, "Milk")
	f.publish(t, uid, store.PublicList{ShowMeta: true})

	counting := &countingStore{Store: f.store}
	clock := testClock
	service := NewPublicService(counting, func() time.Time { return clock })

	ask := func() *apiv1.GetPublicListResponse {
		t.Helper()
		res, err := service.GetPublicList(t.Context(), connect.NewRequest(&apiv1.GetPublicListRequest{}))
		if err != nil {
			t.Fatalf("GetPublicList: %v", err)
		}
		return res.Msg
	}

	ask()
	ask()
	ask()
	if counting.reads != 1 {
		t.Errorf("three Visitors read the List %d times, want 1", counting.reads)
	}

	// Added through the store with a later clock, because the fixture's own is frozen
	// and a List that changed at the same instant did not change.
	list, err := f.store.ListByUID(t.Context(), uid)
	if err != nil {
		t.Fatalf("ListByUID: %v", err)
	}
	if _, err := f.store.CreateItem(t.Context(), store.CreateItemParams{
		UID: "later-item", ListID: list.ID, Label: "Bread",
		AddedByID: f.anna.ID, At: testClock.Add(time.Hour),
	}); err != nil {
		t.Fatalf("CreateItem: %v", err)
	}

	got := ask()
	if counting.reads != 2 {
		t.Errorf("a changed List was read %d times, want 2", counting.reads)
	}
	if len(got.GetItems()) != 2 {
		t.Fatalf("the page shows %d items after one was added, want 2", len(got.GetItems()))
	}

	// Nothing changed, so nothing is read, until it is old enough to have missed
	// something a List does not record: a Member renaming themselves.
	ask()
	if counting.reads != 2 {
		t.Errorf("an unchanged List was read %d times, want 2", counting.reads)
	}
	clock = clock.Add(pageMaxAge + time.Second)
	ask()
	if counting.reads != 3 {
		t.Errorf("a page older than %v was read %d times, want 3", pageMaxAge, counting.reads)
	}
}
