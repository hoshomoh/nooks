package store

import (
	"errors"
	"testing"
	"time"
)

// addItem appends an Item to a List.
func addItem(t *testing.T, s Store, list List, owner Member, uid, label string) Item {
	t.Helper()
	item, err := s.CreateItem(t.Context(), CreateItemParams{
		UID: uid, ListID: list.ID, Label: label, AddedByID: owner.ID, At: createdAt,
	})
	if err != nil {
		t.Fatalf("CreateItem %s: %v", label, err)
	}
	return item
}

func TestItemsKeepTheOrderTheyWereAddedIn(t *testing.T) {
	for _, d := range drivers() {
		t.Run(d.name, func(t *testing.T) {
			s := d.open(t)
			list, owner := newList(t, s, "Groceries", SharingPrivate)

			for i, label := range []string{"Washing-up liquid", "Milk", "Oats, coarse"} {
				addItem(t, s, list, owner, "item_"+label[:3], label)
				_ = i
			}

			items, err := s.ItemsOnList(t.Context(), list.ID)
			if err != nil {
				t.Fatalf("ItemsOnList: %v", err)
			}
			got := []string{}
			for _, item := range items {
				got = append(got, item.Label)
			}
			want := []string{"Washing-up liquid", "Milk", "Oats, coarse"}
			for i := range want {
				if i >= len(got) || got[i] != want[i] {
					t.Fatalf("order = %v, want %v", got, want)
				}
			}
		})
	}
}

// An Item can be dropped between two others by averaging their positions, without
// renumbering anything.
func TestMoveItemBetweenTwoOthers(t *testing.T) {
	for _, d := range drivers() {
		t.Run(d.name, func(t *testing.T) {
			s := d.open(t)
			list, owner := newList(t, s, "Groceries", SharingPrivate)

			first := addItem(t, s, list, owner, "item_a", "Milk")
			second := addItem(t, s, list, owner, "item_b", "Oats")
			last := addItem(t, s, list, owner, "item_c", "Baking paper")

			// Put the last Item between the first two.
			midpoint := (first.Position + second.Position) / 2
			if err := s.MoveItem(t.Context(), last.UID, midpoint, createdAt); err != nil {
				t.Fatalf("MoveItem: %v", err)
			}

			items, err := s.ItemsOnList(t.Context(), list.ID)
			if err != nil {
				t.Fatalf("ItemsOnList: %v", err)
			}
			got := []string{items[0].Label, items[1].Label, items[2].Label}
			want := []string{"Milk", "Baking paper", "Oats"}
			for i := range want {
				if got[i] != want[i] {
					t.Fatalf("order = %v, want %v", got, want)
				}
			}
		})
	}
}

// A tick is a tick whoever made it: last write wins, and it never conflicts.
func TestTickAndUntick(t *testing.T) {
	for _, d := range drivers() {
		t.Run(d.name, func(t *testing.T) {
			s := d.open(t)
			list, owner := newList(t, s, "Groceries", SharingPrivate)
			item := addItem(t, s, list, owner, "item_milk", "Milk")

			if item.Done() {
				t.Fatal("a new Item is already done")
			}

			tickedAt := createdAt.Add(time.Hour)
			if err := s.SetItemDone(t.Context(), item.UID, owner.ID, tickedAt); err != nil {
				t.Fatalf("SetItemDone: %v", err)
			}

			ticked, err := s.ItemByUID(t.Context(), item.UID)
			if err != nil {
				t.Fatalf("ItemByUID: %v", err)
			}
			if !ticked.Done() {
				t.Error("Done() = false after ticking")
			}
			if ticked.DoneByID != owner.ID {
				t.Errorf("DoneByID = %d, want %d", ticked.DoneByID, owner.ID)
			}

			if err := s.SetItemNotDone(t.Context(), item.UID, tickedAt); err != nil {
				t.Fatalf("SetItemNotDone: %v", err)
			}
			unticked, err := s.ItemByUID(t.Context(), item.UID)
			if err != nil {
				t.Fatalf("ItemByUID: %v", err)
			}
			if unticked.Done() {
				t.Error("Done() = true after unticking")
			}
			if unticked.DoneByID != 0 {
				t.Errorf("DoneByID = %d after unticking, want 0", unticked.DoneByID)
			}
		})
	}
}

func TestUpdateItemLeavesUntouchedFieldsAlone(t *testing.T) {
	for _, d := range drivers() {
		t.Run(d.name, func(t *testing.T) {
			s := d.open(t)
			list, owner := newList(t, s, "Groceries", SharingPrivate)

			item, err := s.CreateItem(t.Context(), CreateItemParams{
				UID: "item_milk", ListID: list.ID, Label: "Milk", Quantity: "2",
				DueOn: "2026-08-29", AddedByID: owner.ID, At: createdAt,
			})
			if err != nil {
				t.Fatalf("CreateItem: %v", err)
			}

			label := "Milk, oat"
			if err := s.UpdateItem(t.Context(), item.UID, UpdateItemParams{Label: &label}, createdAt); err != nil {
				t.Fatalf("UpdateItem: %v", err)
			}

			updated, err := s.ItemByUID(t.Context(), item.UID)
			if err != nil {
				t.Fatalf("ItemByUID: %v", err)
			}
			if updated.Label != "Milk, oat" {
				t.Errorf("Label = %q, want the new one", updated.Label)
			}
			if updated.Quantity != "2" {
				t.Errorf("Quantity = %q, want it untouched", updated.Quantity)
			}
			if updated.DueOn != "2026-08-29" {
				t.Errorf("DueOn = %q, want it untouched", updated.DueOn)
			}
		})
	}
}

// Quantity is free text, never a number.
func TestQuantityIsFreeText(t *testing.T) {
	for _, d := range drivers() {
		t.Run(d.name, func(t *testing.T) {
			s := d.open(t)
			list, owner := newList(t, s, "Groceries", SharingPrivate)

			item, err := s.CreateItem(t.Context(), CreateItemParams{
				UID: "item_tomatoes", ListID: list.ID, Label: "Tomatoes",
				Quantity: "1 kg", AddedByID: owner.ID, At: createdAt,
			})
			if err != nil {
				t.Fatalf("CreateItem: %v", err)
			}
			if item.Quantity != "1 kg" {
				t.Errorf("Quantity = %q, want %q", item.Quantity, "1 kg")
			}
		})
	}
}

func TestDeleteItemHidesIt(t *testing.T) {
	for _, d := range drivers() {
		t.Run(d.name, func(t *testing.T) {
			s := d.open(t)
			list, owner := newList(t, s, "Groceries", SharingPrivate)
			item := addItem(t, s, list, owner, "item_milk", "Milk")

			if err := s.DeleteItem(t.Context(), item.UID, createdAt); err != nil {
				t.Fatalf("DeleteItem: %v", err)
			}
			if _, err := s.ItemByUID(t.Context(), item.UID); !errors.Is(err, ErrNotFound) {
				t.Errorf("ItemByUID after deletion = %v, want ErrNotFound", err)
			}

			items, err := s.ItemsOnList(t.Context(), list.ID)
			if err != nil {
				t.Fatalf("ItemsOnList: %v", err)
			}
			if len(items) != 0 {
				t.Errorf("items = %d, want none", len(items))
			}
		})
	}
}

// Today and Upcoming are built from this: dated, unticked Items across every List the
// Member can reach.
func TestDatedItemsForMember(t *testing.T) {
	for _, d := range drivers() {
		t.Run(d.name, func(t *testing.T) {
			s := d.open(t)
			list, owner := newList(t, s, "Groceries", SharingPrivate)

			add := func(uid, label, due string) Item {
				t.Helper()
				item, err := s.CreateItem(t.Context(), CreateItemParams{
					UID: uid, ListID: list.ID, Label: label, DueOn: due,
					AddedByID: owner.ID, At: createdAt,
				})
				if err != nil {
					t.Fatalf("CreateItem: %v", err)
				}
				return item
			}
			add("i_today", "Descale the kettle", "2026-08-25")
			add("i_later", "Ferry tickets", "2026-09-05")
			add("i_undated", "Baking paper", "")
			ticked := add("i_done", "Milk", "2026-08-25")
			if err := s.SetItemDone(t.Context(), ticked.UID, owner.ID, createdAt); err != nil {
				t.Fatalf("SetItemDone: %v", err)
			}

			items, err := s.DatedItemsForMember(t.Context(), owner.ID, "2026-08-01", "2026-08-31")
			if err != nil {
				t.Fatalf("DatedItemsForMember: %v", err)
			}
			if len(items) != 1 {
				t.Fatalf("items = %d, want 1 — undated, ticked and out-of-range are excluded", len(items))
			}
			if items[0].Label != "Descale the kettle" {
				t.Errorf("item = %q, want the dated, unticked one in range", items[0].Label)
			}
		})
	}
}
