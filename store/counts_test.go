package store

import (
	"errors"
	"fmt"
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

	if err := s.RemoveMember(t.Context(), jonas.ID, createdAt); err != nil {
		t.Fatalf("RemoveMember: %v", err)
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

/*
The counts agree with what is on the List, after everything that can change them.

They are written onto the List rather than added up on every read, which is what made
reading a hundred thousand Lists cost milliseconds instead of seconds. The price is that
they can be wrong, and a wrong one is silent: nothing fails, a Member simply reads a
number that is not true and has no reason to doubt it.

So rather than testing each path on its own, this runs a Member's afternoon through the
store and asks the same question after every step — does the stored count equal what
counting right now would give. A path added later that forgets to recount fails here
whatever it is.
*/
func TestTheCountsAgreeAfterEveryThingThatChangesThem(t *testing.T) {
	at := time.Date(2026, 8, 25, 9, 0, 0, 0, time.UTC)

	s, err := OpenSQLite(t.Context(), filepath.Join(t.TempDir(), "nooks.db"))
	if err != nil {
		t.Fatalf("OpenSQLite: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })

	member := func(uid, name string) Member {
		t.Helper()
		m, err := s.CreateMember(t.Context(), CreateMemberParams{
			UID: uid, Name: name, Email: uid + "@brunnen.lan", Role: RoleMember,
			PasswordHash: "hash", CreatedAt: at,
		})
		if err != nil {
			t.Fatalf("CreateMember %s: %v", name, err)
		}
		return m
	}
	anna, jonas := member("mem_anna", "Anna"), member("mem_jonas", "Jonas")

	list, err := s.CreateList(t.Context(), CreateListParams{
		UID: "lst_groceries", Name: "Groceries", OwnerID: anna.ID, At: at,
	})
	if err != nil {
		t.Fatalf("CreateList: %v", err)
	}

	uids := 0
	add := func(label string, by Member) Item {
		t.Helper()
		uids++
		item, err := s.CreateItem(t.Context(), CreateItemParams{
			UID: fmt.Sprintf("itm_%d", uids), ListID: list.ID, Label: label,
			AddedByID: by.ID, At: at,
		})
		if err != nil {
			t.Fatalf("CreateItem %s: %v", label, err)
		}
		return item
	}

	var bread Item
	for _, step := range []struct {
		what string
		do   func()
	}{
		{"one Item added", func() { add("Tomatoes", anna) }},
		{"another, by somebody else", func() { bread = add("Bread", jonas) }},
		{"a third", func() { add("Milk", jonas) }},
		{"one ticked", func() {
			if err := s.SetItemDone(t.Context(), bread.UID, anna.ID, at); err != nil {
				t.Fatalf("SetItemDone: %v", err)
			}
		}},
		{"the same one unticked", func() {
			if err := s.SetItemNotDone(t.Context(), bread.UID, at); err != nil {
				t.Fatalf("SetItemNotDone: %v", err)
			}
		}},
		{"one removed", func() {
			if err := s.DeleteItem(t.Context(), bread.UID, at); err != nil {
				t.Fatalf("DeleteItem: %v", err)
			}
		}},
		{"a whole List copied", func() {
			copied, err := s.CreateList(t.Context(), CreateListParams{
				UID: "lst_copy", Name: "Groceries (copy)", OwnerID: anna.ID, At: at,
			})
			if err != nil {
				t.Fatalf("CreateList: %v", err)
			}
			if _, err := s.CreateItems(t.Context(), []CreateItemParams{
				{UID: "itm_c1", ListID: copied.ID, Label: "Tomatoes", AddedByID: anna.ID, At: at},
				{UID: "itm_c2", ListID: copied.ID, Label: "Milk", AddedByID: anna.ID, At: at},
			}); err != nil {
				t.Fatalf("CreateItems: %v", err)
			}
		}},
		{"the Member who added things removed", func() {
			if err := s.RemoveMember(t.Context(), jonas.ID, createdAt); err != nil {
				t.Fatalf("RemoveMember: %v", err)
			}
		}},
	} {
		step.do()
		countsAgree(t, s, step.what)
	}
}

// countsAgree checks every live List against what counting its Items now would give.
func countsAgree(t *testing.T, s Store, after string) {
	t.Helper()

	// By uid rather than by a read that narrows to one Member: this is about the
	// numbers on the rows, not about who may see them.
	for _, uid := range []string{"lst_groceries", "lst_copy"} {
		list, err := s.ListByUID(t.Context(), uid)
		if errors.Is(err, ErrNotFound) {
			continue
		}
		if err != nil {
			t.Fatalf("ListByUID %s: %v", uid, err)
		}

		items, err := s.ItemsOnList(t.Context(), list.ID)
		if err != nil {
			t.Fatalf("ItemsOnList %s: %v", uid, err)
		}
		open, done := 0, 0
		for _, item := range items {
			if item.Done() {
				done++
				continue
			}
			open++
		}

		if int(list.OpenCount) != open || int(list.DoneCount) != done {
			t.Errorf("after %s, %s says %d open and %d done, and holds %d open and %d done",
				after, uid, list.OpenCount, list.DoneCount, open, done)
		}
	}
}
