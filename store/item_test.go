package store

import (
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"
)

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

/*
Copying a List writes once, not once per Item.

It used to be a CreateItem per row, and each of those read the next position, inserted,
and committed on its own — a hundred Items meant a hundred commits, which on the kind of
machine Nooks is meant to run on is felt rather than measured.
*/
func TestCreateItemsKeepsTheOrderItIsGiven(t *testing.T) {
	for _, d := range drivers() {
		t.Run(d.name, func(t *testing.T) {
			s := d.open(t)
			list, owner := newList(t, s, "Groceries", SharingPrivate)

			made, err := s.CreateItems(t.Context(), []CreateItemParams{
				{UID: "item_rye", ListID: list.ID, Label: "Rye flour", AddedByID: owner.ID, At: createdAt},
				{UID: "item_oat", ListID: list.ID, Label: "Oat milk", AddedByID: owner.ID, At: createdAt},
				{UID: "item_tom", ListID: list.ID, Label: "Tomatoes", AddedByID: owner.ID, At: createdAt},
			})
			if err != nil {
				t.Fatalf("CreateItems: %v", err)
			}
			if len(made) != 3 {
				t.Fatalf("created %d Items, want 3", len(made))
			}

			onList, err := s.ItemsOnList(t.Context(), list.ID)
			if err != nil {
				t.Fatalf("ItemsOnList: %v", err)
			}
			for i, want := range []string{"Rye flour", "Oat milk", "Tomatoes"} {
				if onList[i].Label != want {
					t.Errorf("item %d = %q, want %q", i, onList[i].Label, want)
				}
			}
		})
	}
}

func TestCreateItemsGoesAfterWhatIsAlreadyThere(t *testing.T) {
	for _, d := range drivers() {
		t.Run(d.name, func(t *testing.T) {
			s := d.open(t)
			list, owner := newList(t, s, "Groceries", SharingPrivate)
			addItem(t, s, list, owner, "item_milk", "Milk")

			if _, err := s.CreateItems(t.Context(), []CreateItemParams{
				{UID: "item_bread", ListID: list.ID, Label: "Bread", AddedByID: owner.ID, At: createdAt},
			}); err != nil {
				t.Fatalf("CreateItems: %v", err)
			}

			onList, err := s.ItemsOnList(t.Context(), list.ID)
			if err != nil {
				t.Fatalf("ItemsOnList: %v", err)
			}
			if onList[0].Label != "Milk" || onList[1].Label != "Bread" {
				t.Errorf("order = %q then %q, want Milk then Bread", onList[0].Label, onList[1].Label)
			}
		})
	}
}

// One transaction: a batch with a bad row in it leaves the List as it was, rather than
// half copied.
func TestCreateItemsWritesNothingWhenOneIsUnusable(t *testing.T) {
	for _, d := range drivers() {
		t.Run(d.name, func(t *testing.T) {
			s := d.open(t)
			list, owner := newList(t, s, "Groceries", SharingPrivate)

			_, err := s.CreateItems(t.Context(), []CreateItemParams{
				{UID: "item_rye", ListID: list.ID, Label: "Rye flour", AddedByID: owner.ID, At: createdAt},
				{UID: "item_bad", ListID: list.ID, Label: "", AddedByID: owner.ID, At: createdAt},
			})
			if err == nil {
				t.Fatal("CreateItems with an empty label = nil, want an error")
			}

			onList, err := s.ItemsOnList(t.Context(), list.ID)
			if err != nil {
				t.Fatalf("ItemsOnList: %v", err)
			}
			if len(onList) != 0 {
				t.Errorf("wrote %d Items, want none", len(onList))
			}
		})
	}
}

func TestCreateItemsOfNothingIsNothing(t *testing.T) {
	for _, d := range drivers() {
		t.Run(d.name, func(t *testing.T) {
			s := d.open(t)
			if made, err := s.CreateItems(t.Context(), nil); err != nil || made != nil {
				t.Errorf("CreateItems(nil) = %v, %v; want nil, nil", made, err)
			}
		})
	}
}

/*
Two Items at one position come back in a settled order.

Appending reads the highest position and then inserts, so two people adding to one List
at the same moment both land on the same one. SQL says nothing about the order of rows
whose sort key ties, so the two can swap places between reads of the List they are both
looking at.

This is a contract test and not a reproduction. Both engines happen to hand these back
in insertion order today even with nothing to break the tie, on a table this size and
through the index they have — so it passes with the tiebreak and without it. What it
holds is the promise: tied or not, a List reads back the same way twice, in the order
things arrived. It would catch a change to what this is ordered by. It would not have
caught the tie.
*/
func TestItemsAtTheSamePositionKeepTheirOrder(t *testing.T) {
	for _, d := range drivers() {
		t.Run(d.name, func(t *testing.T) {
			s := d.open(t)
			owner := newMember(t, s)
			list := withItems(t, s, owner, "Groceries", 0, 0, createdAt)

			for _, label := range []string{"Bread", "Milk", "Tea"} {
				if _, err := s.CreateItem(t.Context(), CreateItemParams{
					UID: "item_" + label, ListID: list.ID, Label: label,
					AddedByID: owner.ID, At: createdAt,
				}); err != nil {
					t.Fatalf("CreateItem %s: %v", label, err)
				}
			}
			// What two simultaneous appends leave behind: one position, three Items.
			if _, err := s.(*sqlStore).db.NewUpdate().
				Model((*itemModel)(nil)).
				Set("position = ?", 1024.0).
				Where("list_id = ?", list.ID).
				Exec(t.Context()); err != nil {
				t.Fatalf("flatten positions: %v", err)
			}
			// Then somebody edits the first one. Postgres rewrites an updated row at the
			// end of the table, so a read with nothing to break the tie hands it back
			// last — which is the swap this is about, reached the way it happens.
			if err := s.UpdateItem(t.Context(), "item_Bread",
				UpdateItemParams{Quantity: ptr("2")}, createdAt); err != nil {
				t.Fatalf("UpdateItem: %v", err)
			}

			var first []string
			for range 5 {
				items, err := s.ItemsOnList(t.Context(), list.ID)
				if err != nil {
					t.Fatalf("ItemsOnList: %v", err)
				}
				got := make([]string, 0, len(items))
				for _, item := range items {
					got = append(got, item.Label)
				}
				if first == nil {
					first = got
					continue
				}
				if fmt.Sprint(got) != fmt.Sprint(first) {
					t.Fatalf("order changed between reads: %v then %v", first, got)
				}
			}

			// The order they arrived in, which is the one a person watched happen.
			if fmt.Sprint(first) != "[Bread Milk Tea]" {
				t.Errorf("order = %v, want them in the order they were added", first)
			}
		})
	}
}

// ptr is a pointer to a value, for the optional fields of UpdateItemParams.
func ptr[T any](v T) *T { return &v }

/*
Two people adding to one List at the same moment land in different places.

Appending used to read the highest position and then insert, which is two statements
with a gap between them, so two callers inside that gap both read the same number and
both wrote it. Reads break the tie on id, so a List still came back in arrival order and
nobody saw anything wrong. But the tie was made, and "put this after that one" has no
answer while two Items share a place.

Raised as optional when it was found, because the order was already settled. Building
drag-to-reorder is what made it matter.

Run concurrently on purpose. In sequence it passes either way, because the second read
sees the first insert.
*/
func TestTwoItemsAddedAtOnceGetDifferentPositions(t *testing.T) {
	at := time.Date(2026, time.April, 1, 9, 0, 0, 0, time.UTC)

	for _, d := range drivers() {
		t.Run(d.name, func(t *testing.T) {
			s := d.open(t)
			anna := newMember(t, s)
			list := makeList(t, s, anna, "list_shop", "Shopping", SharingInstance)

			const adders = 4
			start := make(chan struct{})
			var adding sync.WaitGroup
			for i := range adders {
				adding.Add(1)
				go func() {
					defer adding.Done()
					<-start
					_, _ = s.CreateItem(t.Context(), CreateItemParams{
						UID:       fmt.Sprintf("item_%d", i),
						ListID:    list.ID,
						Label:     fmt.Sprintf("Thing %d", i),
						AddedByID: anna.ID,
						At:        at,
					})
				}()
			}
			close(start)
			adding.Wait()

			items, err := s.ItemsOnList(t.Context(), list.ID)
			if err != nil {
				t.Fatalf("ItemsOnList: %v", err)
			}

			places := make(map[float64]string, len(items))
			for _, item := range items {
				if shared, taken := places[item.Position]; taken {
					t.Errorf("%q and %q are both at position %v", shared, item.Label, item.Position)
				}
				places[item.Position] = item.Label
			}
		})
	}
}
