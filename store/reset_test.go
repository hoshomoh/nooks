package store

import (
	"path/filepath"
	"testing"
	"time"
)

// Everything goes, and what is left reads as an Instance nobody has set up yet.
func TestResetInstanceEmptiesEverything(t *testing.T) {
	s, err := OpenSQLite(t.Context(), filepath.Join(t.TempDir(), "nooks.db"))
	if err != nil {
		t.Fatalf("OpenSQLite: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })

	at := time.Date(2026, 8, 25, 9, 0, 0, 0, time.UTC)
	member, err := s.CreateMember(t.Context(), CreateMemberParams{
		UID: "mem_anna", Name: "Anna", Email: "anna@brunnen.lan", Role: RoleAdmin,
		PasswordHash: "hash", CreatedAt: at,
	})
	if err != nil {
		t.Fatalf("CreateMember: %v", err)
	}
	list, err := s.CreateList(t.Context(), CreateListParams{
		UID: "list_groceries", Name: "Groceries", OwnerID: member.ID, At: at,
	})
	if err != nil {
		t.Fatalf("CreateList: %v", err)
	}
	if _, err := s.CreateItem(t.Context(), CreateItemParams{
		UID: "item_milk", ListID: list.ID, Label: "Milk", AddedByID: member.ID, At: at,
	}); err != nil {
		t.Fatalf("CreateItem: %v", err)
	}
	if err := s.SaveInstanceSettings(t.Context(), InstanceSettings{
		Name: "Brunnen Street", SetupCompletedAt: at,
	}); err != nil {
		t.Fatalf("SaveInstanceSettings: %v", err)
	}

	if err := s.ResetInstance(t.Context()); err != nil {
		t.Fatalf("ResetInstance: %v", err)
	}

	stats, err := s.Stats(t.Context())
	if err != nil {
		t.Fatalf("Stats: %v", err)
	}
	if stats.Members != 0 || stats.Lists != 0 || stats.Items != 0 {
		t.Errorf("stats = %+v, want everything at zero", stats)
	}

	settings, err := s.InstanceSettings(t.Context())
	if err != nil {
		t.Fatalf("InstanceSettings: %v", err)
	}
	if !settings.NeedsSetup() {
		t.Error("the instance does not read as needing setup")
	}

	// The index is derived, and would otherwise answer queries about things that are
	// gone.
	hits, err := s.Search(t.Context(), "Milk")
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(hits) != 0 {
		t.Errorf("got %d hits, want nothing left to find", len(hits))
	}
}
