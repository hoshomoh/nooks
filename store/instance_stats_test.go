package store

import (
	"testing"
	"time"
)

/*
About counts what somebody can see, and an Item on a List in the bin is not that.

Deleting a List leaves its Items untouched, which is what lets restoring bring them
back whole. The count only checked each Item's own deletion, so a List deleted with six
Items on it left About reading "0 lists · 6 items": Items with nowhere to be.

Restoring is checked as well, because the fix must not lose them for good: they come
back into the count with their List.
*/
func TestAboutCountsNothingOnADeletedList(t *testing.T) {
	for _, driver := range drivers() {
		t.Run(driver.name, func(t *testing.T) {
			s := driver.open(t)
			ctx := t.Context()
			at := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)

			owner, err := s.CreateMember(ctx, CreateMemberParams{
				UID: "m1", Name: "Ada", Email: "ada@example.invalid",
				PasswordHash: "x", CreatedAt: at,
			})
			if err != nil {
				t.Fatalf("member: %v", err)
			}
			list, err := s.CreateList(ctx, CreateListParams{
				UID: "l1", Name: "Shopping", OwnerID: owner.ID, At: at,
			})
			if err != nil {
				t.Fatalf("list: %v", err)
			}
			for _, label := range []string{"Milk", "Bread", "Eggs"} {
				if _, err := s.CreateItem(ctx, CreateItemParams{
					UID: "i-" + label, ListID: list.ID, Label: label, AddedByID: owner.ID, At: at,
				}); err != nil {
					t.Fatalf("add %s: %v", label, err)
				}
			}

			counted := func(when string, lists, items int) {
				t.Helper()
				stats, err := s.Stats(ctx)
				if err != nil {
					t.Fatalf("Stats %s: %v", when, err)
				}
				if stats.Lists != lists || stats.Items != items {
					t.Errorf("%s: About says %d lists and %d items, want %d and %d",
						when, stats.Lists, stats.Items, lists, items)
				}
			}

			counted("before deleting", 1, 3)

			if err := s.DeleteList(ctx, list.UID, at.Add(time.Hour)); err != nil {
				t.Fatalf("DeleteList: %v", err)
			}
			counted("with the List in the bin", 0, 0)

			if err := s.RestoreList(ctx, list.UID, owner.ID, at.Add(2*time.Hour)); err != nil {
				t.Fatalf("RestoreList: %v", err)
			}
			counted("after restoring", 1, 3)
		})
	}
}
