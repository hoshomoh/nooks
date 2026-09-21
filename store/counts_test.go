package store

import (
	"path/filepath"
	"testing"
	"time"
)

/*
The counts beside a List's name survive a Member being removed.

They are denormalised — counted and written onto the List whenever an Item is added,
ticked or removed — and every Go path that changes one recounts. A Member's Items are
not deleted by a Go path: `item.added_by_id` is ON DELETE CASCADE, so removing somebody
takes their Items out from under the database, and nothing calls recount.

What is left is a List whose name has "7 things to get" beside it and three Items on it,
for good, until somebody happens to tick one.
*/
func TestRemovingAMemberLeavesTheCountsTrue(t *testing.T) {
	at := time.Date(2026, 8, 25, 9, 0, 0, 0, time.UTC)
	s, err := OpenSQLite(t.Context(), filepath.Join(t.TempDir(), "nooks.db"))
	if err != nil {
		t.Fatalf("OpenSQLite: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })

	anna, err := s.CreateMember(t.Context(), CreateMemberParams{
		UID: "mem_anna", Name: "Anna", Email: "anna@brunnen.lan", Role: RoleAdmin,
		PasswordHash: "hash", CreatedAt: at,
	})
	if err != nil {
		t.Fatalf("CreateMember: %v", err)
	}
	jonas, err := s.CreateMember(t.Context(), CreateMemberParams{
		UID: "mem_jonas", Name: "Jonas", Email: "jonas@brunnen.lan", Role: RoleMember,
		PasswordHash: "hash", CreatedAt: at,
	})
	if err != nil {
		t.Fatalf("CreateMember: %v", err)
	}

	// Anna's List. Jonas is in the house and puts things on it too.
	list, err := s.CreateList(t.Context(), CreateListParams{
		UID: "lst_groceries", Name: "Groceries", OwnerID: anna.ID, At: at,
	})
	if err != nil {
		t.Fatalf("CreateList: %v", err)
	}

	add := func(uid, label string, by Member) {
		t.Helper()
		if _, err := s.CreateItem(t.Context(), CreateItemParams{
			UID: uid, ListID: list.ID, Label: label, AddedByID: by.ID, At: at,
		}); err != nil {
			t.Fatalf("CreateItem %s: %v", label, err)
		}
	}
	add("itm_1", "Tomatoes", anna)
	add("itm_2", "Bread", jonas)
	add("itm_3", "Milk", jonas)

	before, err := s.ListByID(t.Context(), list.ID)
	if err != nil {
		t.Fatalf("ListByID: %v", err)
	}
	if before.OpenCount != 3 {
		t.Fatalf("open count = %d before anybody left, want 3", before.OpenCount)
	}

	if err := s.DeleteMember(t.Context(), jonas.ID); err != nil {
		t.Fatalf("DeleteMember: %v", err)
	}

	items, err := s.ItemsOnList(t.Context(), list.ID)
	if err != nil {
		t.Fatalf("ItemsOnList: %v", err)
	}
	after, err := s.ListByID(t.Context(), list.ID)
	if err != nil {
		t.Fatalf("ListByID: %v", err)
	}
	if int(after.OpenCount) != len(items) {
		t.Errorf("open count says %d and the List holds %d", after.OpenCount, len(items))
	}
}

// A batch spanning two Lists would order the Items against the wrong one and leave the
// other's counts behind, neither of them visibly, so it is refused rather than allowed.
func TestItemsAreAddedToOneListAtATime(t *testing.T) {
	at := time.Date(2026, 8, 25, 9, 0, 0, 0, time.UTC)

	s, err := OpenSQLite(t.Context(), filepath.Join(t.TempDir(), "nooks.db"))
	if err != nil {
		t.Fatalf("OpenSQLite: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })

	anna, err := s.CreateMember(t.Context(), CreateMemberParams{
		UID: "mem_anna", Name: "Anna", Email: "anna@brunnen.lan", Role: RoleAdmin,
		PasswordHash: "hash", CreatedAt: at,
	})
	if err != nil {
		t.Fatalf("CreateMember: %v", err)
	}
	first, err := s.CreateList(t.Context(), CreateListParams{
		UID: "lst_one", Name: "Groceries", OwnerID: anna.ID, At: at,
	})
	if err != nil {
		t.Fatalf("CreateList: %v", err)
	}
	second, err := s.CreateList(t.Context(), CreateListParams{
		UID: "lst_two", Name: "Bike", OwnerID: anna.ID, At: at,
	})
	if err != nil {
		t.Fatalf("CreateList: %v", err)
	}

	_, err = s.CreateItems(t.Context(), []CreateItemParams{
		{UID: "itm_1", ListID: first.ID, Label: "Tomatoes", AddedByID: anna.ID, At: at},
		{UID: "itm_2", ListID: second.ID, Label: "Inner tube", AddedByID: anna.ID, At: at},
	})
	if err == nil {
		t.Fatal("a batch spanning two lists was accepted")
	}

	// Refused whole: neither List has anything on it, and neither count moved.
	for _, uid := range []string{"lst_one", "lst_two"} {
		list, err := s.ListByUID(t.Context(), uid)
		if err != nil {
			t.Fatalf("ListByUID %s: %v", uid, err)
		}
		if list.OpenCount != 0 {
			t.Errorf("%s says %d open after a refused batch", uid, list.OpenCount)
		}
	}
}
