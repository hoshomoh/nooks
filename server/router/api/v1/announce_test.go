package v1

import (
	"context"
	"path/filepath"
	"slices"
	"testing"
	"time"

	"connectrpc.com/connect"

	apiv1 "github.com/hoshomoh/nooks/proto/gen/nooks/api/v1"
	"github.com/hoshomoh/nooks/server/auth"
	"github.com/hoshomoh/nooks/store"
)

// heard records what a service announced, so a test can ask who was told.
type heard struct {
	listChanged []string
	listsFor    []int64
	activityFor []int64
}

func (h *heard) ListChanged(_ context.Context, list store.List) {
	h.listChanged = append(h.listChanged, list.UID)
}

func (h *heard) ListsChanged(audience []int64) {
	h.listsFor = append(h.listsFor, audience...)
}

func (h *heard) ActivityArrived(memberID int64) {
	h.activityFor = append(h.activityFor, memberID)
}

// listening is a fixture whose service announces into heard, which every other fixture
// leaves nil — so nothing about announcing was covered anywhere until this.
func listening(t *testing.T) (listFixture, *heard) {
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

	told := &heard{}
	issued := 0
	svc := NewListService(s, func() time.Time { return testClock }, func() (string, error) {
		issued++
		return "uid-" + string(rune('a'+issued)), nil
	})
	svc.announce = told

	return listFixture{
		svc:   svc,
		store: s,
		anna:  member("mem_anna", "Anna", "anna@brunnen.lan", store.RoleAdmin),
		jonas: member("mem_jonas", "Jonas", "jonas@brunnen.lan", store.RoleMember),
	}, told
}

/*
Somebody a List is taken away from is told their sidebar changed.

ListChanged reaches the audience as it stands after the change, so the one group it
cannot reach is the people the change removed. Without something else they keep the List
in their sidebar until the query goes stale or they reload, and clicking it gives them an
error about a List that was theirs to read a moment ago.

Publisher had a ListsChanged for exactly this and nothing ever called it — it was not
even on the Announcer interface the services hold, so it could not be called.
*/
func TestUnsharingTellsWhoeverLostTheList(t *testing.T) {
	f, told := listening(t)
	uid := f.createList(t, f.anna, "Groceries")

	as := auth.WithMember(t.Context(), f.anna)
	share := func(sharing apiv1.Sharing, with ...string) {
		t.Helper()
		if _, err := f.svc.SetListSharing(as, connect.NewRequest(&apiv1.SetListSharingRequest{
			ListUid: uid, Sharing: sharing, CanEdit: true, MemberUids: with,
		})); err != nil {
			t.Fatalf("SetListSharing: %v", err)
		}
	}

	share(apiv1.Sharing_SHARING_SPECIFIC, f.jonas.UID)
	told.listsFor = nil

	// Taken back. Jonas can no longer reach it and is the only one who needs telling.
	share(apiv1.Sharing_SHARING_PRIVATE)

	if !slices.Contains(told.listsFor, f.jonas.ID) {
		t.Errorf("told %v that their lists changed, want Jonas (%d) among them",
			told.listsFor, f.jonas.ID)
	}
	if slices.Contains(told.listsFor, f.anna.ID) {
		t.Error("the owner was told they lost a list they still have")
	}
}

// Sharing with somebody takes nothing away, so nobody is told they lost anything.
func TestSharingTellsNobodyTheyLostAList(t *testing.T) {
	f, told := listening(t)
	uid := f.createList(t, f.anna, "Groceries")

	if _, err := f.svc.SetListSharing(auth.WithMember(t.Context(), f.anna),
		connect.NewRequest(&apiv1.SetListSharingRequest{
			ListUid: uid, Sharing: apiv1.Sharing_SHARING_SPECIFIC,
			CanEdit: true, MemberUids: []string{f.jonas.UID},
		})); err != nil {
		t.Fatalf("SetListSharing: %v", err)
	}

	if len(told.listsFor) != 0 {
		t.Errorf("told %v they lost a list when one was only shared", told.listsFor)
	}
}

/*
Joining or leaving a Group changes which Lists somebody reaches.

A List shared with a Group is not itself changed by somebody entering or leaving it, so
the ListChanged that normally carries the news is never sent. Nothing else was sent
either, which left both the person who gained a List and the person who lost one looking
at a sidebar that was wrong until they reloaded.
*/
func TestGroupMembershipTellsWhoeverItMoved(t *testing.T) {
	s, err := store.OpenSQLite(t.Context(), filepath.Join(t.TempDir(), "nooks.db"))
	if err != nil {
		t.Fatalf("OpenSQLite: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })

	member := func(uid, name string, role store.Role) store.Member {
		t.Helper()
		m, err := s.CreateMember(t.Context(), store.CreateMemberParams{
			UID: uid, Name: name, Email: uid + "@brunnen.lan", Role: role,
			PasswordHash: "hash", CreatedAt: testClock,
		})
		if err != nil {
			t.Fatalf("CreateMember %s: %v", name, err)
		}
		return m
	}
	anna := member("mem_anna", "Anna", store.RoleAdmin)
	jonas := member("mem_jonas", "Jonas", store.RoleMember)
	mira := member("mem_mira", "Mira", store.RoleMember)

	told := &heard{}
	svc := NewMemberService(s, func() time.Time { return testClock },
		func() (string, error) { return "grp_1", nil }).WithAnnouncer(told)

	as := auth.WithMember(t.Context(), anna)
	made, err := svc.CreateGroup(as, connect.NewRequest(&apiv1.CreateGroupRequest{
		Name: "Flat", MemberUids: []string{jonas.UID},
	}))
	if err != nil {
		t.Fatalf("CreateGroup: %v", err)
	}
	told.listsFor = nil

	// Jonas out, Mira in. Anna was never in it and is not moved by this.
	if _, err := svc.SetGroupMembers(as, connect.NewRequest(&apiv1.SetGroupMembersRequest{
		GroupUid: made.Msg.GetGroup().GetUid(), MemberUids: []string{mira.UID},
	})); err != nil {
		t.Fatalf("SetGroupMembers: %v", err)
	}

	for _, moved := range []struct {
		who  string
		id   int64
		want bool
	}{
		{"Jonas, who left", jonas.ID, true},
		{"Mira, who joined", mira.ID, true},
		{"Anna, who was never in it", anna.ID, false},
	} {
		if got := slices.Contains(told.listsFor, moved.id); got != moved.want {
			t.Errorf("%s: told = %v, want %v (told %v)", moved.who, got, moved.want, told.listsFor)
		}
	}
}

/*
Removing somebody changes what everybody else can see.

Their Lists go with them, and those Lists could have been shared with anybody. Nothing
told the people left: the Admin who clicked had their own screen refreshed by the app,
and everyone else kept a sidebar full of Lists that no longer exist until they reloaded.
*/
func TestRemovingAMemberTellsTheOnesWhoStay(t *testing.T) {
	s, err := store.OpenSQLite(t.Context(), filepath.Join(t.TempDir(), "nooks.db"))
	if err != nil {
		t.Fatalf("OpenSQLite: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })

	member := func(uid, name string, role store.Role) store.Member {
		t.Helper()
		m, err := s.CreateMember(t.Context(), store.CreateMemberParams{
			UID: uid, Name: name, Email: uid + "@brunnen.lan", Role: role,
			PasswordHash: "hash", CreatedAt: testClock,
		})
		if err != nil {
			t.Fatalf("CreateMember %s: %v", name, err)
		}
		return m
	}
	anna := member("mem_anna", "Anna", store.RoleAdmin)
	jonas := member("mem_jonas", "Jonas", store.RoleMember)
	mira := member("mem_mira", "Mira", store.RoleMember)

	told := &heard{}
	svc := NewMemberService(s, func() time.Time { return testClock }, nil).WithAnnouncer(told)

	if _, err := svc.RemoveMember(auth.WithMember(t.Context(), anna),
		connect.NewRequest(&apiv1.RemoveMemberRequest{MemberUid: mira.UID})); err != nil {
		t.Fatalf("RemoveMember: %v", err)
	}

	for _, who := range []struct {
		what string
		id   int64
		want bool
	}{
		{"the admin who did it", anna.ID, true},
		{"everybody else still here", jonas.ID, true},
		{"the member who was removed", mira.ID, false},
	} {
		if got := slices.Contains(told.listsFor, who.id); got != who.want {
			t.Errorf("%s: told = %v, want %v (told %v)", who.what, got, who.want, told.listsFor)
		}
	}
}
