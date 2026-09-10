package store

import (
	"testing"
)

// searchLabels runs a query and returns the matched text, for readable assertions.
func searchLabels(t *testing.T, s Store, query string) []string {
	t.Helper()
	hits, err := s.Search(t.Context(), query)
	if err != nil {
		t.Fatalf("Search(%q): %v", query, err)
	}
	texts := make([]string, 0, len(hits))
	for _, hit := range hits {
		texts = append(texts, hit.Text)
	}
	return texts
}

// contains reports whether the results include a text.
func contains(texts []string, want string) bool {
	for _, text := range texts {
		if text == want {
			return true
		}
	}
	return false
}

func TestSearchFindsItemsAndLists(t *testing.T) {
	for _, d := range drivers() {
		t.Run(d.name, func(t *testing.T) {
			s := d.open(t)
			list, owner := newList(t, s, "Groceries", SharingPrivate)
			addItem(t, s, list, owner, "i_coffee", "Coffee for the weekend")
			addItem(t, s, list, owner, "i_paper", "Baking paper")

			got := searchLabels(t, s, "coffee")
			if !contains(got, "Coffee for the weekend") {
				t.Errorf("search for coffee = %v, want the coffee Item", got)
			}
			if contains(got, "Baking paper") {
				t.Errorf("search for coffee returned an unrelated Item: %v", got)
			}

			// A List is findable by its name too.
			if got := searchLabels(t, s, "groceries"); !contains(got, "Groceries") {
				t.Errorf("search for groceries = %v, want the List", got)
			}
		})
	}
}

// Results must narrow while a Member is still typing.
func TestSearchMatchesAPrefix(t *testing.T) {
	for _, d := range drivers() {
		t.Run(d.name, func(t *testing.T) {
			s := d.open(t)
			list, owner := newList(t, s, "Groceries", SharingPrivate)
			addItem(t, s, list, owner, "i_coffee", "Coffee for the weekend")

			for _, prefix := range []string{"cof", "coff", "coffee"} {
				if got := searchLabels(t, s, prefix); !contains(got, "Coffee for the weekend") {
					t.Errorf("search for %q = %v, want the coffee Item", prefix, got)
				}
			}
		})
	}
}

func TestSearchIsCaseInsensitive(t *testing.T) {
	for _, d := range drivers() {
		t.Run(d.name, func(t *testing.T) {
			s := d.open(t)
			list, owner := newList(t, s, "Groceries", SharingPrivate)
			addItem(t, s, list, owner, "i_coffee", "Coffee for the weekend")

			if got := searchLabels(t, s, "COFFEE"); !contains(got, "Coffee for the weekend") {
				t.Errorf("search for COFFEE = %v, want the coffee Item", got)
			}
		})
	}
}

// Quantity is text a Member wrote, so "1 kg" is as searchable as "Tomatoes".
func TestSearchCoversQuantity(t *testing.T) {
	for _, d := range drivers() {
		t.Run(d.name, func(t *testing.T) {
			s := d.open(t)
			list, owner := newList(t, s, "Groceries", SharingPrivate)

			if _, err := s.CreateItem(t.Context(), CreateItemParams{
				UID: "i_tomatoes", ListID: list.ID, Label: "Tomatoes", Quantity: "1 kg",
				AddedByID: owner.ID, At: createdAt,
			}); err != nil {
				t.Fatalf("CreateItem: %v", err)
			}

			if got := searchLabels(t, s, "kg"); len(got) == 0 {
				t.Error("search for kg found nothing, want the Item with that quantity")
			}
		})
	}
}

// Every word has to match, so a second word narrows rather than widens.
func TestSearchRequiresEveryWord(t *testing.T) {
	for _, d := range drivers() {
		t.Run(d.name, func(t *testing.T) {
			s := d.open(t)
			list, owner := newList(t, s, "Groceries", SharingPrivate)
			addItem(t, s, list, owner, "i_coffee", "Coffee for the weekend")
			addItem(t, s, list, owner, "i_beans", "Baking paper")

			if got := searchLabels(t, s, "coffee weekend"); !contains(got, "Coffee for the weekend") {
				t.Errorf("search for two words = %v, want the Item with both", got)
			}
			if got := searchLabels(t, s, "coffee spanner"); len(got) != 0 {
				t.Errorf("search for a word that is not there = %v, want nothing", got)
			}
		})
	}
}

// Renaming has to update the index, or search hands back the old name.
func TestRenamingAListUpdatesTheIndex(t *testing.T) {
	for _, d := range drivers() {
		t.Run(d.name, func(t *testing.T) {
			s := d.open(t)
			newList(t, s, "Groceries", SharingPrivate)

			if err := s.RenameList(t.Context(), "list_Groceries", "Shopping", createdAt); err != nil {
				t.Fatalf("RenameList: %v", err)
			}

			if got := searchLabels(t, s, "shopping"); !contains(got, "Shopping") {
				t.Errorf("search for the new name = %v, want it found", got)
			}
			if got := searchLabels(t, s, "groceries"); contains(got, "Groceries") {
				t.Errorf("the old name is still findable: %v", got)
			}
		})
	}
}

func TestEditingAnItemUpdatesTheIndex(t *testing.T) {
	for _, d := range drivers() {
		t.Run(d.name, func(t *testing.T) {
			s := d.open(t)
			list, owner := newList(t, s, "Groceries", SharingPrivate)
			item := addItem(t, s, list, owner, "i_milk", "Milk")

			label := "Oat milk"
			if err := s.UpdateItem(t.Context(), item.UID, UpdateItemParams{Label: &label}, createdAt); err != nil {
				t.Fatalf("UpdateItem: %v", err)
			}

			if got := searchLabels(t, s, "oat"); !contains(got, "Oat milk") {
				t.Errorf("search for the new label = %v, want it found", got)
			}
		})
	}
}

// Search must not hand back things a Member can no longer open.
func TestDeletingRemovesFromTheIndex(t *testing.T) {
	for _, d := range drivers() {
		t.Run(d.name, func(t *testing.T) {
			s := d.open(t)
			list, owner := newList(t, s, "Groceries", SharingPrivate)
			item := addItem(t, s, list, owner, "i_coffee", "Coffee for the weekend")

			if err := s.DeleteItem(t.Context(), item.UID, createdAt); err != nil {
				t.Fatalf("DeleteItem: %v", err)
			}
			if got := searchLabels(t, s, "coffee"); contains(got, "Coffee for the weekend") {
				t.Errorf("a deleted Item is still findable: %v", got)
			}
		})
	}
}

// Deleting a List takes its Items out of the index with it.
func TestDeletingAListRemovesItsItemsFromTheIndex(t *testing.T) {
	for _, d := range drivers() {
		t.Run(d.name, func(t *testing.T) {
			s := d.open(t)
			list, owner := newList(t, s, "Groceries", SharingPrivate)
			addItem(t, s, list, owner, "i_coffee", "Coffee for the weekend")

			if err := s.DeleteList(t.Context(), list.UID, createdAt); err != nil {
				t.Fatalf("DeleteList: %v", err)
			}

			if got := searchLabels(t, s, "coffee"); len(got) != 0 {
				t.Errorf("Items of a deleted List are still findable: %v", got)
			}
			if got := searchLabels(t, s, "groceries"); len(got) != 0 {
				t.Errorf("a deleted List is still findable: %v", got)
			}
		})
	}
}

// An empty query is not an error, and matches nothing.
func TestSearchWithNothingToSearchFor(t *testing.T) {
	for _, d := range drivers() {
		t.Run(d.name, func(t *testing.T) {
			s := d.open(t)
			for _, query := range []string{"", "   ", "!!!"} {
				hits, err := s.Search(t.Context(), query)
				if err != nil {
					t.Errorf("Search(%q) = %v, want no error", query, err)
				}
				if len(hits) != 0 {
					t.Errorf("Search(%q) = %v, want nothing", query, hits)
				}
			}
		})
	}
}

// Punctuation is not an operator: both engines have query syntaxes of their own, and a
// Member typing an apostrophe is not asking for them.
func TestSearchDoesNotTreatPunctuationAsSyntax(t *testing.T) {
	for _, d := range drivers() {
		t.Run(d.name, func(t *testing.T) {
			s := d.open(t)
			list, owner := newList(t, s, "Groceries", SharingPrivate)
			addItem(t, s, list, owner, "i_coffee", "Coffee for the weekend")

			for _, query := range []string{`coffee"`, "coffee*", "coffee & weekend", "coffee:*"} {
				if _, err := s.Search(t.Context(), query); err != nil {
					t.Errorf("Search(%q) = %v, want no error", query, err)
				}
			}
		})
	}
}
