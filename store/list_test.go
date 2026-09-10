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
