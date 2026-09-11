package v1

import (
	"testing"

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
	res, err := NewPublicService(f.store).GetPublicList(
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
