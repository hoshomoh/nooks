package v1

import (
	"testing"

	"connectrpc.com/connect"

	apiv1 "github.com/hoshomoh/nooks/proto/gen/nooks/api/v1"
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
