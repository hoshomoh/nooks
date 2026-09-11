package v1

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"connectrpc.com/connect"

	"github.com/hoshomoh/nooks/internal/password"
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

// Nooks has no mail server, so the password is handed over in person — and read once.
func TestAddingAMemberReadsThePasswordOut(t *testing.T) {
	f := newMemberFixture(t)

	added, err := f.svc.AddMember(f.as(t, f.anna), connect.NewRequest(&apiv1.AddMemberRequest{
		Name: "Til", Email: "til@example.com",
	}))
	if err != nil {
		t.Fatalf("AddMember: %v", err)
	}
	if err := password.Validate(added.Msg.GetTemporaryPassword()); err != nil {
		t.Errorf("the temporary password would be refused as a password: %v", err)
	}
	if !added.Msg.GetMember().GetMustChangePassword() {
		t.Error("MustChangePassword = false, want the Member sent to replace it")
	}
	if added.Msg.GetMember().GetRole() != apiv1.Role_ROLE_MEMBER {
		t.Error("a new account should not start as an Admin")
	}
}

func TestAddingAMemberWithATakenEmail(t *testing.T) {
	f := newMemberFixture(t)

	_, err := f.svc.AddMember(f.as(t, f.anna), connect.NewRequest(&apiv1.AddMemberRequest{
		Name: "Another Anna", Email: "anna@brunnen.lan",
	}))
	if got := connect.CodeOf(err); got != connect.CodeAlreadyExists {
		t.Errorf("code = %v, want already_exists", got)
	}
}

func TestOnlyAnAdminAddsAMember(t *testing.T) {
	f := newMemberFixture(t)

	_, err := f.svc.AddMember(f.as(t, f.jonas), connect.NewRequest(&apiv1.AddMemberRequest{
		Name: "Til", Email: "til@example.com",
	}))
	if got := connect.CodeOf(err); got != connect.CodePermissionDenied {
		t.Errorf("code = %v, want permission_denied", got)
	}
}

func TestMakingSomebodyAnAdmin(t *testing.T) {
	f := newMemberFixture(t)

	res, err := f.svc.SetMemberRole(f.as(t, f.anna), connect.NewRequest(&apiv1.SetMemberRoleRequest{
		MemberUid: f.jonas.UID, Role: apiv1.Role_ROLE_ADMIN,
	}))
	if err != nil {
		t.Fatalf("SetMemberRole: %v", err)
	}
	if res.Msg.GetMember().GetRole() != apiv1.Role_ROLE_ADMIN {
		t.Error("role did not change")
	}
}

// Standing down as the only Admin would lock the Instance for everyone, including the
// person doing it.
func TestTheOnlyAdminCannotStandDown(t *testing.T) {
	f := newMemberFixture(t)

	_, err := f.svc.SetMemberRole(f.as(t, f.anna), connect.NewRequest(&apiv1.SetMemberRoleRequest{
		MemberUid: f.anna.UID, Role: apiv1.Role_ROLE_MEMBER,
	}))
	if got := connect.CodeOf(err); got != connect.CodeFailedPrecondition {
		t.Errorf("code = %v, want failed_precondition", got)
	}
}

func TestRemovingAMember(t *testing.T) {
	f := newMemberFixture(t)

	if _, err := f.svc.RemoveMember(f.as(t, f.anna), connect.NewRequest(
		&apiv1.RemoveMemberRequest{MemberUid: f.jonas.UID},
	)); err != nil {
		t.Fatalf("RemoveMember: %v", err)
	}

	res, err := f.svc.ListMembers(f.as(t, f.anna), connect.NewRequest(&apiv1.ListMembersRequest{}))
	if err != nil {
		t.Fatalf("ListMembers: %v", err)
	}
	if len(res.Msg.GetMembers()) != 2 {
		t.Errorf("got %d members, want the other two", len(res.Msg.GetMembers()))
	}
}

// The one removal nobody could undo.
func TestAnAdminCannotRemoveThemselves(t *testing.T) {
	f := newMemberFixture(t)

	_, err := f.svc.RemoveMember(f.as(t, f.anna), connect.NewRequest(
		&apiv1.RemoveMemberRequest{MemberUid: f.anna.UID},
	))
	if got := connect.CodeOf(err); got != connect.CodeFailedPrecondition {
		t.Errorf("code = %v, want failed_precondition", got)
	}
}

// A Member who has not arrived yet has a real account and no sign-in behind them.
func TestAMemberWhoHasNeverSignedIn(t *testing.T) {
	f := newMemberFixture(t)

	res, err := f.svc.ListMembers(f.as(t, f.anna), connect.NewRequest(&apiv1.ListMembersRequest{}))
	if err != nil {
		t.Fatalf("ListMembers: %v", err)
	}
	for _, member := range res.Msg.GetMembers() {
		if member.GetLastSignedInAt() != "" {
			t.Errorf("%s has signed in already, in a store where nobody has", member.GetName())
		}
		if member.GetCreatedAt() == "" {
			t.Errorf("%s has no created date, which the table shows", member.GetName())
		}
	}
}

// An account is a person. Nooks does not let one person edit another, which is why
// there is no "member_uid" on this request at all.
func TestAMemberChangesTheirOwnNameAndEmail(t *testing.T) {
	f := newListFixture(t)
	svc := NewMemberService(f.store, nil, nil)

	res, err := svc.UpdateOwnProfile(f.as(t, f.jonas), connect.NewRequest(
		&apiv1.UpdateOwnProfileRequest{Name: "Jonas B", Email: "Jonas.B@brunnen.lan"},
	))
	if err != nil {
		t.Fatalf("UpdateOwnProfile: %v", err)
	}
	if got := res.Msg.GetMember().GetName(); got != "Jonas B" {
		t.Errorf("name = %q, want the new one", got)
	}

	// Normalised on the way in, so what they retyped still signs them in.
	found, err := f.store.MemberByEmail(t.Context(), "jonas.b@brunnen.lan")
	if err != nil {
		t.Fatalf("MemberByEmail after the change: %v", err)
	}
	if found.ID != f.jonas.ID {
		t.Errorf("the email now finds member %d, want %d", found.ID, f.jonas.ID)
	}
}

func TestAMemberCannotTakeAnEmailSomebodyElseUses(t *testing.T) {
	f := newListFixture(t)
	svc := NewMemberService(f.store, nil, nil)

	_, err := svc.UpdateOwnProfile(f.as(t, f.jonas), connect.NewRequest(
		&apiv1.UpdateOwnProfileRequest{Name: "Jonas", Email: f.anna.Email},
	))
	if got := connect.CodeOf(err); got != connect.CodeAlreadyExists {
		t.Errorf("code = %v, want already_exists", got)
	}
}

// A key that reaches somebody's Lists must not be a way to change the address their
// account signs in with.
func TestATokenCannotChangeAProfile(t *testing.T) {
	f := newListFixture(t)
	groceries := f.createList(t, f.anna, "Groceries")
	ctx := f.withToken(t, f.anna, store.PermissionWrite, groceries)

	_, err := NewMemberService(f.store, nil, nil).UpdateOwnProfile(ctx, connect.NewRequest(
		&apiv1.UpdateOwnProfileRequest{Name: "Anna", Email: "somewhere@else.lan"},
	))
	if got := connect.CodeOf(err); got != connect.CodePermissionDenied {
		t.Errorf("code = %v, want permission_denied", got)
	}
}

// A Group is made in order to share with the people in it, so naming it and filling it
// is one decision rather than two.
func TestAGroupIsMadeWithItsPeopleInIt(t *testing.T) {
	f := newListFixture(t)
	svc := NewMemberService(f.store, nil, nil)

	res, err := svc.CreateGroup(f.as(t, f.anna), connect.NewRequest(&apiv1.CreateGroupRequest{
		Name: "Flatmates", MemberUids: []string{f.anna.UID, f.jonas.UID},
	}))
	if err != nil {
		t.Fatalf("CreateGroup: %v", err)
	}
	if got := len(res.Msg.GetGroup().GetMembers()); got != 2 {
		t.Errorf("got %d members, want the two it was made with", got)
	}
}

// Resolved before the Group exists, so a bad identifier leaves nothing behind.
func TestAGroupWithAnUnknownMemberIsNotMade(t *testing.T) {
	f := newListFixture(t)
	svc := NewMemberService(f.store, nil, nil)

	_, err := svc.CreateGroup(f.as(t, f.anna), connect.NewRequest(&apiv1.CreateGroupRequest{
		Name: "Flatmates", MemberUids: []string{"mem_nobody"},
	}))
	if err == nil {
		t.Fatal("CreateGroup with an unknown member succeeded")
	}

	listed, err := svc.ListGroups(f.as(t, f.anna), connect.NewRequest(&apiv1.ListGroupsRequest{}))
	if err != nil {
		t.Fatalf("ListGroups: %v", err)
	}
	if got := len(listed.Msg.GetGroups()); got != 0 {
		t.Errorf("got %d groups, want none left behind", got)
	}
}
