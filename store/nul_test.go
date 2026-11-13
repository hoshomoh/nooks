package store

import (
	"errors"
	"testing"
	"time"
)

// A NUL is dropped from text, so a label saves as it reads and an id finds the one without it.
func TestTextWithANulInItSavesAndReadsAsBefore(t *testing.T) {
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

			if _, err := s.CreateItem(ctx, CreateItemParams{
				UID: "i1", ListID: list.ID, Label: "Mi\x00lk", AddedByID: owner.ID, At: at,
			}); err != nil {
				t.Fatalf("add an item with a NUL in its label: %v", err)
			}
			item, err := s.ItemByUID(ctx, "i1")
			if err != nil {
				t.Fatalf("read it back: %v", err)
			}
			if item.Label != "Milk" {
				t.Errorf("label saved as %q, want %q", item.Label, "Milk")
			}

			if _, err := s.ListByUID(ctx, "nope\x00"); !errors.Is(err, ErrNotFound) {
				t.Errorf("an id with a NUL in it: got %v, want ErrNotFound", err)
			}
		})
	}
}
