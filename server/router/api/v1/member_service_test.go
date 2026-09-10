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

// memberFixture is a store with an Admin, two Members, and the service under test.
type memberFixture struct {
	svc   *MemberService
	store store.Store
	anna  store.Member
	jonas store.Member
	mira  store.Member
}

func newMemberFixture(t *testing.T) memberFixture {
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
	svc := NewMemberService(s, func() time.Time { return testClock }, func() (string, error) {
		issued++
		return fmt.Sprintf("grp-%d", issued), nil
	})

	return memberFixture{
		svc:   svc,
		store: s,
		anna:  member("mem_anna", "Anna", "anna@brunnen.lan", store.RoleAdmin),
		jonas: member("mem_jonas", "Jonas", "jonas@brunnen.lan", store.RoleMember),
		mira:  member("mem_mira", "Mira", "mira@brunnen.lan", store.RoleMember),
	}
}

func (f memberFixture) as(t *testing.T, member store.Member) context.Context {
	t.Helper()
	return auth.WithMember(t.Context(), member)
}

// addGroup makes a Group with the given Members in it.
func (f memberFixture) addGroup(t *testing.T, name string, members ...store.Member) *apiv1.Group {
	t.Helper()
	created, err := f.svc.CreateGroup(f.as(t, f.anna), connect.NewRequest(&apiv1.CreateGroupRequest{
		Name: name,
	}))
	if err != nil {
		t.Fatalf("CreateGroup: %v", err)
	}

	uids := make([]string, 0, len(members))
	for _, member := range members {
		uids = append(uids, member.UID)
	}
	set, err := f.svc.SetGroupMembers(f.as(t, f.anna), connect.NewRequest(&apiv1.SetGroupMembersRequest{
		GroupUid: created.Msg.GetGroup().GetUid(), MemberUids: uids,
	}))
	if err != nil {
		t.Fatalf("SetGroupMembers: %v", err)
	}
	return set.Msg.GetGroup()
}

// The share dialog has to offer somebody to share with.
func TestEveryMemberCanSeeWhoIsHere(t *testing.T) {
	f := newMemberFixture(t)

	res, err := f.svc.ListMembers(f.as(t, f.jonas), connect.NewRequest(&apiv1.ListMembersRequest{}))
	if err != nil {
		t.Fatalf("ListMembers: %v", err)
	}
	if len(res.Msg.GetMembers()) != 3 {
		t.Errorf("got %d members, want everyone", len(res.Msg.GetMembers()))
	}
}

func TestAGroupNamesWhoIsInIt(t *testing.T) {
	f := newMemberFixture(t)

	group := f.addGroup(t, "Flatmates", f.anna, f.jonas)

	if group.GetName() != "Flatmates" {
		t.Errorf("name = %q", group.GetName())
	}
	if len(group.GetMembers()) != 2 {
		t.Fatalf("got %d members, want two", len(group.GetMembers()))
	}
}

// Membership is one decision, like sharing.
func TestSettingGroupMembersReplacesThem(t *testing.T) {
	f := newMemberFixture(t)
	group := f.addGroup(t, "Flatmates", f.anna, f.jonas)

	set, err := f.svc.SetGroupMembers(f.as(t, f.anna), connect.NewRequest(&apiv1.SetGroupMembersRequest{
		GroupUid: group.GetUid(), MemberUids: []string{f.mira.UID},
	}))
	if err != nil {
		t.Fatalf("SetGroupMembers: %v", err)
	}
	if len(set.Msg.GetGroup().GetMembers()) != 1 {
		t.Fatalf("got %d members, want just the one named", len(set.Msg.GetGroup().GetMembers()))
	}
	if set.Msg.GetGroup().GetMembers()[0].GetName() != "Mira" {
		t.Errorf("member = %q, want Mira", set.Msg.GetGroup().GetMembers()[0].GetName())
	}
}

func TestOnlyAnAdminMakesAGroup(t *testing.T) {
	f := newMemberFixture(t)

	_, err := f.svc.CreateGroup(f.as(t, f.jonas), connect.NewRequest(&apiv1.CreateGroupRequest{
		Name: "Flatmates",
	}))
	if got := connect.CodeOf(err); got != connect.CodePermissionDenied {
		t.Errorf("code = %v, want permission_denied", got)
	}
}

func TestAGroupNeedsAName(t *testing.T) {
	f := newMemberFixture(t)

	_, err := f.svc.CreateGroup(f.as(t, f.anna), connect.NewRequest(&apiv1.CreateGroupRequest{
		Name: "   ",
	}))
	if got := connect.CodeOf(err); got != connect.CodeInvalidArgument {
		t.Errorf("code = %v, want invalid_argument", got)
	}
}
