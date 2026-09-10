package v1

import (
	"testing"

	"connectrpc.com/connect"

	apiv1 "github.com/hoshomoh/nooks/proto/gen/nooks/api/v1"
	"github.com/hoshomoh/nooks/store"
)

// shareWith puts a List on named sharing through the API, as the owner would.
func (f listFixture) shareWith(t *testing.T, owner store.Member, uid string, memberUIDs, groupUIDs []string) {
	t.Helper()
	_, err := f.svc.SetListSharing(f.as(t, owner), connect.NewRequest(&apiv1.SetListSharingRequest{
		ListUid:    uid,
		Sharing:    apiv1.Sharing_SHARING_SPECIFIC,
		CanEdit:    true,
		MemberUids: memberUIDs,
		GroupUids:  groupUIDs,
	}))
	if err != nil {
		t.Fatalf("SetListSharing: %v", err)
	}
}

func TestSharingByNameReachesTheNamedMember(t *testing.T) {
	f := newListFixture(t)
	uid := f.createList(t, f.anna, "Groceries")

	f.shareWith(t, f.anna, uid, []string{f.jonas.UID}, nil)

	if _, err := f.svc.GetList(f.as(t, f.jonas), connect.NewRequest(&apiv1.GetListRequest{
		ListUid: uid,
	})); err != nil {
		t.Fatalf("GetList as the Member it names: %v", err)
	}
}

// The dialog is one decision, so what it sends replaces what was there.
func TestSharingByNameReplacesWhoItReached(t *testing.T) {
	f := newListFixture(t)
	uid := f.createList(t, f.anna, "Groceries")

	f.shareWith(t, f.anna, uid, []string{f.jonas.UID}, nil)
	f.shareWith(t, f.anna, uid, nil, nil)

	_, err := f.svc.GetList(f.as(t, f.jonas), connect.NewRequest(&apiv1.GetListRequest{ListUid: uid}))
	if got := connect.CodeOf(err); got != connect.CodeNotFound {
		t.Errorf("code = %v, want not_found once he is no longer named", got)
	}
}

// Taking a List off named sharing must not leave the names behind, waiting to come
// back the next time somebody picks "specific people".
func TestLeavingNamedSharingClearsTheNames(t *testing.T) {
	f := newListFixture(t)
	uid := f.createList(t, f.anna, "Groceries")
	f.shareWith(t, f.anna, uid, []string{f.jonas.UID}, nil)

	if _, err := f.svc.SetListSharing(f.as(t, f.anna), connect.NewRequest(&apiv1.SetListSharingRequest{
		ListUid: uid, Sharing: apiv1.Sharing_SHARING_PRIVATE,
	})); err != nil {
		t.Fatalf("SetListSharing: %v", err)
	}

	shares, err := f.svc.GetListShares(f.as(t, f.anna), connect.NewRequest(&apiv1.GetListSharesRequest{
		ListUid: uid,
	}))
	if err != nil {
		t.Fatalf("GetListShares: %v", err)
	}
	if len(shares.Msg.GetMemberUids()) != 0 {
		t.Errorf("member uids = %v, want none", shares.Msg.GetMemberUids())
	}
}

// Who else a List reaches is the owner's business.
func TestOnlyTheOwnerReadsWhoAListReaches(t *testing.T) {
	f := newListFixture(t)
	uid := f.createList(t, f.anna, "Groceries")
	f.shareWith(t, f.anna, uid, []string{f.jonas.UID}, nil)

	_, err := f.svc.GetListShares(f.as(t, f.jonas), connect.NewRequest(&apiv1.GetListSharesRequest{
		ListUid: uid,
	}))
	if got := connect.CodeOf(err); got != connect.CodePermissionDenied {
		t.Errorf("code = %v, want permission_denied for somebody who does not own it", got)
	}
}

func TestSharingWithSomebodyWhoIsNotThere(t *testing.T) {
	f := newListFixture(t)
	uid := f.createList(t, f.anna, "Groceries")

	_, err := f.svc.SetListSharing(f.as(t, f.anna), connect.NewRequest(&apiv1.SetListSharingRequest{
		ListUid: uid, Sharing: apiv1.Sharing_SHARING_SPECIFIC, MemberUids: []string{"mem_nobody"},
	}))
	if got := connect.CodeOf(err); got != connect.CodeInvalidArgument {
		t.Errorf("code = %v, want invalid_argument", got)
	}
}

// Sharing with a Group reaches everyone in it, including whoever is added later.
func TestSharingWithAGroupReachesItsMembers(t *testing.T) {
	f := newListFixture(t)
	members := NewMemberService(f.store, f.svc.now, f.svc.newUID)
	ctx := f.as(t, f.anna)

	// Anna is the Admin here: the first Member of an Instance always is.
	created, err := members.CreateGroup(ctx, connect.NewRequest(&apiv1.CreateGroupRequest{
		Name: "Flatmates",
	}))
	if err != nil {
		t.Fatalf("CreateGroup: %v", err)
	}
	groupUID := created.Msg.GetGroup().GetUid()

	if _, err := members.SetGroupMembers(ctx, connect.NewRequest(&apiv1.SetGroupMembersRequest{
		GroupUid: groupUID, MemberUids: []string{f.jonas.UID},
	})); err != nil {
		t.Fatalf("SetGroupMembers: %v", err)
	}

	uid := f.createList(t, f.anna, "Groceries")
	f.shareWith(t, f.anna, uid, nil, []string{groupUID})

	if _, err := f.svc.GetList(f.as(t, f.jonas), connect.NewRequest(&apiv1.GetListRequest{
		ListUid: uid,
	})); err != nil {
		t.Fatalf("GetList as somebody in the Group: %v", err)
	}
}
