package store

import (
	"errors"
	"testing"
	"time"
)

// newList adds a List owned by a fresh Member, and returns both.
func newList(t *testing.T, s Store, name string, sharing Sharing) (List, Member) {
	t.Helper()
	owner := newMember(t, s)
	list, err := s.CreateList(t.Context(), CreateListParams{
		UID: "list_" + name, Name: name, OwnerID: owner.ID,
		Sharing: sharing, CanEdit: true, At: createdAt,
	})
	if err != nil {
		t.Fatalf("CreateList: %v", err)
	}
	return list, owner
}

func TestCreateAndReadList(t *testing.T) {
	for _, d := range drivers() {
		t.Run(d.name, func(t *testing.T) {
			s := d.open(t)
			list, owner := newList(t, s, "Groceries", SharingInstance)

			if list.ID == 0 {
				t.Error("ID = 0, want the database to have assigned one")
			}
			if list.OwnerID != owner.ID {
				t.Errorf("OwnerID = %d, want %d", list.OwnerID, owner.ID)
			}
			if !list.Shared() {
				t.Error("Shared() = false for an instance-wide List")
			}

			found, err := s.ListByUID(t.Context(), "list_Groceries")
			if err != nil {
				t.Fatalf("ListByUID: %v", err)
			}
			if found.Name != "Groceries" {
				t.Errorf("Name = %q, want Groceries", found.Name)
			}
		})
	}
}

func TestAPrivateListIsNotShared(t *testing.T) {
	for _, d := range drivers() {
		t.Run(d.name, func(t *testing.T) {
			list, _ := newList(t, d.open(t), "Bike", SharingPrivate)
			if list.Shared() {
				t.Error("Shared() = true for a private List")
			}
		})
	}
}

// A Member sees their own Lists and anything shared with the whole Instance, and
// nothing else.
func TestListsForMember(t *testing.T) {
	for _, d := range drivers() {
		t.Run(d.name, func(t *testing.T) {
			s := d.open(t)
			owner := newMember(t, s)

			other, err := s.CreateMember(t.Context(), CreateMemberParams{
				UID: "mem_jonas", Name: "Jonas", Email: "jonas@brunnen.lan",
				Role: RoleMember, PasswordHash: "hash", CreatedAt: createdAt,
			})
			if err != nil {
				t.Fatalf("CreateMember: %v", err)
			}

			add := func(uid, name string, ownerID int64, sharing Sharing) {
				t.Helper()
				if _, err := s.CreateList(t.Context(), CreateListParams{
					UID: uid, Name: name, OwnerID: ownerID, Sharing: sharing,
					CanEdit: true, At: createdAt,
				}); err != nil {
					t.Fatalf("CreateList %s: %v", name, err)
				}
			}
			add("l_mine", "Bike", owner.ID, SharingPrivate)
			add("l_theirs_private", "Their diary", other.ID, SharingPrivate)
			add("l_theirs_shared", "Groceries", other.ID, SharingInstance)

			lists, err := s.ListsForMember(t.Context(), owner.ID)
			if err != nil {
				t.Fatalf("ListsForMember: %v", err)
			}

			names := map[string]bool{}
			for _, list := range lists {
				names[list.Name] = true
			}
			if !names["Bike"] {
				t.Error("their own private List is missing")
			}
			if !names["Groceries"] {
				t.Error("an instance-wide List is missing")
			}
			if names["Their diary"] {
				t.Error("somebody else's private List is visible")
			}
		})
	}
}

func TestRenameList(t *testing.T) {
	for _, d := range drivers() {
		t.Run(d.name, func(t *testing.T) {
			s := d.open(t)
			newList(t, s, "Groceries", SharingPrivate)

			later := createdAt.Add(time.Hour)
			if err := s.RenameList(t.Context(), "list_Groceries", "Shopping", later); err != nil {
				t.Fatalf("RenameList: %v", err)
			}

			list, err := s.ListByUID(t.Context(), "list_Groceries")
			if err != nil {
				t.Fatalf("ListByUID: %v", err)
			}
			if list.Name != "Shopping" {
				t.Errorf("Name = %q, want Shopping", list.Name)
			}
			if !list.UpdatedAt.Equal(later) {
				t.Errorf("UpdatedAt = %v, want %v", list.UpdatedAt, later)
			}
		})
	}
}

// Deleting is soft, and a deleted List stops being reachable.
func TestDeleteListHidesIt(t *testing.T) {
	for _, d := range drivers() {
		t.Run(d.name, func(t *testing.T) {
			s := d.open(t)
			_, owner := newList(t, s, "Groceries", SharingPrivate)

			if err := s.DeleteList(t.Context(), "list_Groceries", createdAt); err != nil {
				t.Fatalf("DeleteList: %v", err)
			}
			if _, err := s.ListByUID(t.Context(), "list_Groceries"); !errors.Is(err, ErrNotFound) {
				t.Errorf("ListByUID after deletion = %v, want ErrNotFound", err)
			}

			lists, err := s.ListsForMember(t.Context(), owner.ID)
			if err != nil {
				t.Fatalf("ListsForMember: %v", err)
			}
			if len(lists) != 0 {
				t.Errorf("lists = %d, want none after deletion", len(lists))
			}
		})
	}
}

func TestSetListSharing(t *testing.T) {
	for _, d := range drivers() {
		t.Run(d.name, func(t *testing.T) {
			s := d.open(t)
			newList(t, s, "Groceries", SharingPrivate)

			if err := s.SetListSharing(t.Context(), "list_Groceries", SharingInstance, false, createdAt); err != nil {
				t.Fatalf("SetListSharing: %v", err)
			}

			list, err := s.ListByUID(t.Context(), "list_Groceries")
			if err != nil {
				t.Fatalf("ListByUID: %v", err)
			}
			if list.Sharing != SharingInstance {
				t.Errorf("Sharing = %q, want %q", list.Sharing, SharingInstance)
			}
			// Read-only means see and print, not tick or add.
			if list.CanEdit {
				t.Error("CanEdit = true, want false")
			}
		})
	}
}

// Pinning is per-Member and never affects anyone else's sidebar.
func TestPinningIsPerMember(t *testing.T) {
	for _, d := range drivers() {
		t.Run(d.name, func(t *testing.T) {
			s := d.open(t)
			list, owner := newList(t, s, "Groceries", SharingInstance)

			other, err := s.CreateMember(t.Context(), CreateMemberParams{
				UID: "mem_jonas", Name: "Jonas", Email: "jonas@brunnen.lan",
				Role: RoleMember, PasswordHash: "hash", CreatedAt: createdAt,
			})
			if err != nil {
				t.Fatalf("CreateMember: %v", err)
			}

			if err := s.PinList(t.Context(), owner.ID, list.ID); err != nil {
				t.Fatalf("PinList: %v", err)
			}
			// Pinning twice is not an error.
			if err := s.PinList(t.Context(), owner.ID, list.ID); err != nil {
				t.Fatalf("second PinList: %v", err)
			}

			mine, err := s.PinnedListIDs(t.Context(), owner.ID)
			if err != nil {
				t.Fatalf("PinnedListIDs: %v", err)
			}
			if len(mine) != 1 || mine[0] != list.ID {
				t.Errorf("owner's pins = %v, want [%d]", mine, list.ID)
			}

			theirs, err := s.PinnedListIDs(t.Context(), other.ID)
			if err != nil {
				t.Fatalf("PinnedListIDs: %v", err)
			}
			if len(theirs) != 0 {
				t.Errorf("someone else's pins = %v, want none", theirs)
			}
		})
	}
}

func TestUnpinIsForgiving(t *testing.T) {
	for _, d := range drivers() {
		t.Run(d.name, func(t *testing.T) {
			s := d.open(t)
			list, owner := newList(t, s, "Groceries", SharingPrivate)

			// Unpinning something that was never pinned is not an error.
			if err := s.UnpinList(t.Context(), owner.ID, list.ID); err != nil {
				t.Errorf("UnpinList with no pin = %v, want nil", err)
			}
		})
	}
}

/*
A sidebar is read by a person, so the order has to be a person's.

Comparing raw names compares character codes, and every capital sorts below every
lowercase one: the Lists somebody happened to capitalise come out as a block above the
ones they did not. The two drivers disagreed about it as well, so the same instance
looked different depending on what it was stored in.
*/
func TestListsAreOrderedTheWayAPersonReads(t *testing.T) {
	for _, d := range drivers() {
		t.Run(d.name, func(t *testing.T) {
			s := d.open(t)
			owner := newMember(t, s)

			// Named out of order, and mixing case the way a household does.
			for _, name := range []string{"Groceries", "avocados", "Flat jobs", "bike parts"} {
				if _, err := s.CreateList(t.Context(), CreateListParams{
					UID: "list_" + name, Name: name, OwnerID: owner.ID,
					Sharing: SharingPrivate, CanEdit: true, At: createdAt,
				}); err != nil {
					t.Fatalf("CreateList %q: %v", name, err)
				}
			}

			lists, err := s.ListsForMember(t.Context(), owner.ID)
			if err != nil {
				t.Fatalf("ListsForMember: %v", err)
			}

			got := make([]string, 0, len(lists))
			for _, list := range lists {
				got = append(got, list.Name)
			}
			want := []string{"avocados", "bike parts", "Flat jobs", "Groceries"}
			if len(got) != len(want) {
				t.Fatalf("got %v, want %v", got, want)
			}
			for i := range want {
				if got[i] != want[i] {
					t.Fatalf("got %v, want %v", got, want)
				}
			}
		})
	}
}

/*
A deleted List comes back, findable, and only for whoever deleted it.

Deleting has always been soft and nothing ever undid it, so the date a List carried was
a promise the code did not keep. The re-indexing is the half worth asserting: DeleteList
takes the List and every Item and Note on it out of the search index, because search
handing back something a Member can no longer open is worse than not finding it, and a
restore that only cleared the date would give back a List that opens and cannot be
searched for.
*/
func TestADeletedListComesBackFindable(t *testing.T) {
	at := time.Date(2026, time.April, 1, 9, 0, 0, 0, time.UTC)

	for _, d := range drivers() {
		t.Run(d.name, func(t *testing.T) {
			s := d.open(t)
			anna := newMember(t, s)
			jonas := addMember(t, s, "mem_jonas", "Jonas", "jonas@brunnen.lan")

			list := makeList(t, s, anna, "list_camp", "Camping", SharingInstance)
			addItem(t, s, list, anna, "item_tarp", "Tarpaulin")

			if err := s.DeleteList(t.Context(), list.UID, at); err != nil {
				t.Fatalf("DeleteList: %v", err)
			}
			if hits, err := s.Search(t.Context(), "Tarpaulin"); err != nil || len(hits) != 0 {
				t.Fatalf("a deleted List is still searchable: %v hits, %v", len(hits), err)
			}

			// Somebody else's deleted List reads as one that never existed.
			if err := s.RestoreList(t.Context(), list.UID, jonas.ID, at); !errors.Is(err, ErrNotFound) {
				t.Errorf("Jonas restoring Anna's List answered %v, want ErrNotFound", err)
			}

			if err := s.RestoreList(t.Context(), list.UID, anna.ID, at); err != nil {
				t.Fatalf("RestoreList: %v", err)
			}
			if _, err := s.ListByUID(t.Context(), list.UID); err != nil {
				t.Errorf("the restored List is %v", err)
			}

			hits, err := s.Search(t.Context(), "Tarpaulin")
			if err != nil {
				t.Fatalf("Search: %v", err)
			}
			if len(hits) != 1 {
				t.Errorf("found %d after restoring, want the Item findable again", len(hits))
			}
		})
	}
}

// The window ends, and what it was protecting goes with it.
func TestADeletedListIsPurgedAfterItsWindow(t *testing.T) {
	at := time.Date(2026, time.April, 1, 9, 0, 0, 0, time.UTC)

	for _, d := range drivers() {
		t.Run(d.name, func(t *testing.T) {
			s := d.open(t)
			anna := newMember(t, s)
			list := makeList(t, s, anna, "list_camp", "Camping", SharingInstance)

			if err := s.DeleteList(t.Context(), list.UID, at); err != nil {
				t.Fatalf("DeleteList: %v", err)
			}

			// An hour short of the window, from the caller's side.
			early := at.Add(DeletedListLifetime).Add(-time.Hour).Add(-DeletedListLifetime)
			if gone, err := s.PurgeDeletedLists(t.Context(), early); err != nil || gone != 0 {
				t.Fatalf("purged %d an hour early (%v), want none", gone, err)
			}

			late := at.Add(DeletedListLifetime).Add(time.Hour).Add(-DeletedListLifetime)
			gone, err := s.PurgeDeletedLists(t.Context(), late)
			if err != nil {
				t.Fatalf("PurgeDeletedLists: %v", err)
			}
			if gone != 1 {
				t.Errorf("purged %d past the window, want the one List", gone)
			}
			if err := s.RestoreList(t.Context(), list.UID, anna.ID, at); !errors.Is(err, ErrNotFound) {
				t.Errorf("restoring a purged List answered %v, want ErrNotFound", err)
			}
		})
	}
}
