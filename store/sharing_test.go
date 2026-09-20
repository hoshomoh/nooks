package store

import (
	"errors"
	"testing"
)

// addMember adds a second Member for sharing tests.
func addMember(t *testing.T, s Store, uid, name, email string) Member {
	t.Helper()
	member, err := s.CreateMember(t.Context(), CreateMemberParams{
		UID: uid, Name: name, Email: email, Role: RoleMember,
		PasswordHash: "hash", CreatedAt: createdAt,
	})
	if err != nil {
		t.Fatalf("CreateMember %s: %v", name, err)
	}
	return member
}

// names is the set of List names a Member can reach.
func names(t *testing.T, s Store, memberID int64) map[string]bool {
	t.Helper()
	lists, err := s.ListsForMember(t.Context(), memberID)
	if err != nil {
		t.Fatalf("ListsForMember: %v", err)
	}
	found := map[string]bool{}
	for _, list := range lists {
		found[list.Name] = true
	}
	return found
}

// Sharing with one person by name reaches them and nobody else.
func TestSharingWithAMemberByName(t *testing.T) {
	for _, d := range drivers() {
		t.Run(d.name, func(t *testing.T) {
			s := d.open(t)
			list, _ := newList(t, s, "Groceries", SharingPrivate)
			jonas := addMember(t, s, "mem_jonas", "Jonas", "jonas@brunnen.lan")
			mira := addMember(t, s, "mem_mira", "Mira", "mira@brunnen.lan")

			if err := s.ReplaceListShares(t.Context(), list.ID, []int64{jonas.ID}, nil); err != nil {
				t.Fatalf("ReplaceListShares: %v", err)
			}

			if !names(t, s, jonas.ID)["Groceries"] {
				t.Error("Jonas was shared the List and cannot see it")
			}
			if names(t, s, mira.ID)["Groceries"] {
				t.Error("Mira was not shared the List and can see it")
			}
		})
	}
}

// Sharing with a Group reaches everyone in it.
func TestSharingWithAGroup(t *testing.T) {
	for _, d := range drivers() {
		t.Run(d.name, func(t *testing.T) {
			s := d.open(t)
			list, _ := newList(t, s, "Groceries", SharingPrivate)
			jonas := addMember(t, s, "mem_jonas", "Jonas", "jonas@brunnen.lan")
			mira := addMember(t, s, "mem_mira", "Mira", "mira@brunnen.lan")

			group, err := s.CreateGroup(t.Context(), "grp_flatmates", "Flatmates", createdAt)
			if err != nil {
				t.Fatalf("CreateGroup: %v", err)
			}
			if err := s.AddToGroup(t.Context(), group.ID, jonas.ID); err != nil {
				t.Fatalf("AddToGroup: %v", err)
			}
			if err := s.ReplaceListShares(t.Context(), list.ID, nil, []int64{group.ID}); err != nil {
				t.Fatalf("ReplaceListShares: %v", err)
			}

			if !names(t, s, jonas.ID)["Groceries"] {
				t.Error("a Group member cannot see the List shared with their Group")
			}
			if names(t, s, mira.ID)["Groceries"] {
				t.Error("somebody outside the Group can see it")
			}
		})
	}
}

// Someone added to a Group later gets the Lists it already reaches.
func TestJoiningAGroupGrantsItsLists(t *testing.T) {
	for _, d := range drivers() {
		t.Run(d.name, func(t *testing.T) {
			s := d.open(t)
			list, _ := newList(t, s, "Groceries", SharingPrivate)
			mira := addMember(t, s, "mem_mira", "Mira", "mira@brunnen.lan")

			group, err := s.CreateGroup(t.Context(), "grp_flatmates", "Flatmates", createdAt)
			if err != nil {
				t.Fatalf("CreateGroup: %v", err)
			}
			if err := s.ReplaceListShares(t.Context(), list.ID, nil, []int64{group.ID}); err != nil {
				t.Fatalf("ReplaceListShares: %v", err)
			}
			if names(t, s, mira.ID)["Groceries"] {
				t.Fatal("Mira can see it before joining the Group")
			}

			if err := s.AddToGroup(t.Context(), group.ID, mira.ID); err != nil {
				t.Fatalf("AddToGroup: %v", err)
			}
			if !names(t, s, mira.ID)["Groceries"] {
				t.Error("joining the Group did not grant its Lists")
			}
		})
	}
}

// Removing someone from a Group takes away the Lists they got through it, and nothing
// else.
func TestLeavingAGroupTakesAwayOnlyItsLists(t *testing.T) {
	for _, d := range drivers() {
		t.Run(d.name, func(t *testing.T) {
			s := d.open(t)
			shared, _ := newList(t, s, "Groceries", SharingPrivate)
			jonas := addMember(t, s, "mem_jonas", "Jonas", "jonas@brunnen.lan")

			// A List of his own, which must survive.
			if _, err := s.CreateList(t.Context(), CreateListParams{
				UID: "l_bike", Name: "Bike", OwnerID: jonas.ID,
				Sharing: SharingPrivate, CanEdit: true, At: createdAt,
			}); err != nil {
				t.Fatalf("CreateList: %v", err)
			}

			group, err := s.CreateGroup(t.Context(), "grp_flatmates", "Flatmates", createdAt)
			if err != nil {
				t.Fatalf("CreateGroup: %v", err)
			}
			if err := s.AddToGroup(t.Context(), group.ID, jonas.ID); err != nil {
				t.Fatalf("AddToGroup: %v", err)
			}
			if err := s.ReplaceListShares(t.Context(), shared.ID, nil, []int64{group.ID}); err != nil {
				t.Fatalf("ReplaceListShares: %v", err)
			}

			if err := s.RemoveFromGroup(t.Context(), group.ID, jonas.ID); err != nil {
				t.Fatalf("RemoveFromGroup: %v", err)
			}

			after := names(t, s, jonas.ID)
			if after["Groceries"] {
				t.Error("the Group's List is still reachable after leaving it")
			}
			if !after["Bike"] {
				t.Error("leaving a Group took away a List of his own")
			}
		})
	}
}

// Sharing is one decision, not a sequence of additions.
func TestReplacingSharesClearsWhatWasThere(t *testing.T) {
	for _, d := range drivers() {
		t.Run(d.name, func(t *testing.T) {
			s := d.open(t)
			list, _ := newList(t, s, "Groceries", SharingPrivate)
			jonas := addMember(t, s, "mem_jonas", "Jonas", "jonas@brunnen.lan")
			mira := addMember(t, s, "mem_mira", "Mira", "mira@brunnen.lan")

			if err := s.ReplaceListShares(t.Context(), list.ID, []int64{jonas.ID}, nil); err != nil {
				t.Fatalf("first ReplaceListShares: %v", err)
			}
			if err := s.ReplaceListShares(t.Context(), list.ID, []int64{mira.ID}, nil); err != nil {
				t.Fatalf("second ReplaceListShares: %v", err)
			}

			if names(t, s, jonas.ID)["Groceries"] {
				t.Error("Jonas still reaches the List after being replaced")
			}
			if !names(t, s, mira.ID)["Groceries"] {
				t.Error("Mira does not reach the List she was shared")
			}
		})
	}
}

func TestGroupNotFound(t *testing.T) {
	for _, d := range drivers() {
		t.Run(d.name, func(t *testing.T) {
			if _, err := d.open(t).GroupByUID(t.Context(), "grp_nope"); !errors.Is(err, ErrNotFound) {
				t.Errorf("GroupByUID = %v, want ErrNotFound", err)
			}
		})
	}
}

// A Group is only ever a shortcut for sharing, so what it reaches is the whole of what
// it does — and the Groups page says so rather than making an Admin work it out.
func TestTheListsAGroupReaches(t *testing.T) {
	for _, d := range drivers() {
		t.Run(d.name, func(t *testing.T) {
			s := d.open(t)
			shared, owner := newList(t, s, "Groceries", SharingSpecific)

			if _, err := s.CreateList(t.Context(), CreateListParams{
				UID: "l_bike", Name: "Bike", OwnerID: owner.ID,
				Sharing: SharingPrivate, CanEdit: true, At: createdAt,
			}); err != nil {
				t.Fatalf("CreateList: %v", err)
			}

			group, err := s.CreateGroup(t.Context(), "grp_flatmates", "Flatmates", createdAt)
			if err != nil {
				t.Fatalf("CreateGroup: %v", err)
			}
			if err := s.ReplaceListShares(t.Context(), shared.ID, nil, []int64{group.ID}); err != nil {
				t.Fatalf("ReplaceListShares: %v", err)
			}

			reached, err := s.ListsSharedWithGroup(t.Context(), group.ID)
			if err != nil {
				t.Fatalf("ListsSharedWithGroup: %v", err)
			}
			if len(reached) != 1 || reached[0].Name != "Groceries" {
				t.Errorf("reaches %v, want only the List shared with it", reached)
			}
		})
	}
}

func TestAGroupThatReachesNothing(t *testing.T) {
	for _, d := range drivers() {
		t.Run(d.name, func(t *testing.T) {
			s := d.open(t)
			group, err := s.CreateGroup(t.Context(), "grp_flatmates", "Flatmates", createdAt)
			if err != nil {
				t.Fatalf("CreateGroup: %v", err)
			}

			reached, err := s.ListsSharedWithGroup(t.Context(), group.ID)
			if err != nil {
				t.Fatalf("ListsSharedWithGroup: %v", err)
			}
			if len(reached) != 0 {
				t.Errorf("reaches %v, want nothing", reached)
			}
		})
	}
}

/*
Asking whether one Member may see one List gives the same answer as listing them all.

The event stream asks this to decide whether somebody may watch a List, and it used to
answer by reading every List the Member could reach and looking for the one. The cheap
question has to agree with the expensive one in every case, because the expensive one is
what the rest of the app trusts — and a yes where there should be a no is a private List
announcing its changes to somebody.
*/
func TestCanReachListAgreesWithListingThemAll(t *testing.T) {
	for _, d := range drivers() {
		t.Run(d.name, func(t *testing.T) {
			s := d.open(t)
			owner := newMember(t, s)
			jonas := addMember(t, s, "mem_jonas", "Jonas", "jonas@brunnen.lan")
			mira := addMember(t, s, "mem_mira", "Mira", "mira@brunnen.lan")

			makeList(t, s, owner, "list_private", "Private", SharingPrivate)
			makeList(t, s, owner, "list_everyone", "Everyone", SharingInstance)
			named := makeList(t, s, owner, "list_named", "Named", SharingSpecific)
			viaGroup := makeList(t, s, owner, "list_group", "Via group", SharingSpecific)

			if err := s.ReplaceListShares(t.Context(), named.ID, []int64{jonas.ID}, nil); err != nil {
				t.Fatalf("ReplaceListShares: %v", err)
			}
			group, err := s.CreateGroup(t.Context(), "grp_flatmates", "Flatmates", createdAt)
			if err != nil {
				t.Fatalf("CreateGroup: %v", err)
			}
			if err := s.AddToGroup(t.Context(), group.ID, mira.ID); err != nil {
				t.Fatalf("AddToGroup: %v", err)
			}
			if err := s.ReplaceListShares(t.Context(), viaGroup.ID, nil, []int64{group.ID}); err != nil {
				t.Fatalf("ReplaceListShares: %v", err)
			}

			for _, who := range []Member{owner, jonas, mira} {
				reachable := map[string]bool{}
				lists, err := s.ListsForMember(t.Context(), who.ID)
				if err != nil {
					t.Fatalf("ListsForMember: %v", err)
				}
				for _, list := range lists {
					reachable[list.UID] = true
				}

				for _, uid := range []string{"list_private", "list_everyone", "list_named", "list_group"} {
					got, err := s.CanReachList(t.Context(), who.ID, uid)
					if err != nil {
						t.Fatalf("CanReachList: %v", err)
					}
					if got != reachable[uid] {
						t.Errorf("%s and %s: CanReachList says %v, listing them all says %v",
							who.Name, uid, got, reachable[uid])
					}
				}
			}
		})
	}
}

// A List nobody can name is not reachable, and neither is one that was deleted.
func TestCanReachListSaysNoToWhatIsNotThere(t *testing.T) {
	for _, d := range drivers() {
		t.Run(d.name, func(t *testing.T) {
			s := d.open(t)
			owner := newMember(t, s)
			list := makeList(t, s, owner, "list_gone", "Gone", SharingInstance)

			if err := s.DeleteList(t.Context(), list.UID, createdAt); err != nil {
				t.Fatalf("DeleteList: %v", err)
			}

			for _, uid := range []string{"list_gone", "list_never_existed"} {
				got, err := s.CanReachList(t.Context(), owner.ID, uid)
				if err != nil {
					t.Fatalf("CanReachList: %v", err)
				}
				if got {
					t.Errorf("CanReachList says yes to %s", uid)
				}
			}
		})
	}
}

// makeList adds one List with the sharing given.
func makeList(t *testing.T, s Store, owner Member, uid, name string, sharing Sharing) List {
	t.Helper()
	list, err := s.CreateList(t.Context(), CreateListParams{
		UID: uid, Name: name, OwnerID: owner.ID, Sharing: sharing, CanEdit: true, At: createdAt,
	})
	if err != nil {
		t.Fatalf("CreateList %s: %v", name, err)
	}
	return list
}
