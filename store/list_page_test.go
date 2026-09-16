package store

import (
	"fmt"
	"testing"
	"time"
)

// page reads one page the way the API will, and names what came back.
func page(t *testing.T, s Store, q ListQuery) ([]string, int) {
	t.Helper()
	got, err := s.ListsPage(t.Context(), q)
	if err != nil {
		t.Fatalf("ListsPage: %v", err)
	}
	names := make([]string, 0, len(got.Lists))
	for _, l := range got.Lists {
		names = append(names, l.Name)
	}
	return names, got.Total
}

// withItems adds a List holding open and ticked Items, and returns it.
func withItems(t *testing.T, s Store, owner Member, name string, open, done int, at time.Time) List {
	t.Helper()
	list, err := s.CreateList(t.Context(), CreateListParams{
		UID: "list_" + name, Name: name, OwnerID: owner.ID,
		Sharing: SharingPrivate, CanEdit: true, At: at,
	})
	if err != nil {
		t.Fatalf("CreateList %s: %v", name, err)
	}
	for i := 0; i < open+done; i++ {
		item, err := s.CreateItem(t.Context(), CreateItemParams{
			UID: fmt.Sprintf("item_%s_%d", name, i), ListID: list.ID,
			Label: "thing", AddedByID: owner.ID, At: at,
		})
		if err != nil {
			t.Fatalf("CreateItem: %v", err)
		}
		if i >= open {
			if err := s.SetItemDone(t.Context(), item.UID, owner.ID, at); err != nil {
				t.Fatalf("SetItemDone: %v", err)
			}
		}
	}
	return list
}

func TestListsPageCountsWithoutReadingEveryItem(t *testing.T) {
	for _, d := range drivers() {
		t.Run(d.name, func(t *testing.T) {
			s := d.open(t)
			owner := newMember(t, s)
			withItems(t, s, owner, "Groceries", 3, 7, createdAt)

			got, err := s.ListsPage(t.Context(), ListQuery{MemberID: owner.ID})
			if err != nil {
				t.Fatalf("ListsPage: %v", err)
			}
			if len(got.Lists) != 1 {
				t.Fatalf("got %d lists, want 1", len(got.Lists))
			}
			if got.Lists[0].OpenCount != 3 || got.Lists[0].DoneCount != 7 {
				t.Errorf("open=%d done=%d, want 3 and 7", got.Lists[0].OpenCount, got.Lists[0].DoneCount)
			}
		})
	}
}

/*
A finished List and one nobody has filled in are not the same thing.

Both have nothing open. Only the first belongs under Completed, and the difference is
whether anything was ever ticked on it.
*/
func TestListsPageSeparatesFinishedFromEmpty(t *testing.T) {
	for _, d := range drivers() {
		t.Run(d.name, func(t *testing.T) {
			s := d.open(t)
			owner := newMember(t, s)
			withItems(t, s, owner, "Finished", 0, 4, createdAt)
			withItems(t, s, owner, "Empty", 0, 0, createdAt)
			withItems(t, s, owner, "Busy", 2, 1, createdAt)

			active, _ := page(t, s, ListQuery{MemberID: owner.ID, Status: StatusActive, Order: OrderName})
			if len(active) != 2 {
				t.Errorf("active = %v, want the empty one and the busy one", active)
			}

			completed, _ := page(t, s, ListQuery{MemberID: owner.ID, Status: StatusCompleted})
			if len(completed) != 1 || completed[0] != "Finished" {
				t.Errorf("completed = %v, want [Finished]", completed)
			}
		})
	}
}

func TestListsPageOrdersAsAsked(t *testing.T) {
	for _, d := range drivers() {
		t.Run(d.name, func(t *testing.T) {
			s := d.open(t)
			owner := newMember(t, s)
			withItems(t, s, owner, "avocados", 9, 0, createdAt)
			withItems(t, s, owner, "Bread", 1, 0, createdAt.Add(time.Hour))
			withItems(t, s, owner, "cheese", 4, 0, createdAt.Add(2*time.Hour))

			// A person reads names without capitals coming first.
			byName, _ := page(t, s, ListQuery{MemberID: owner.ID, Order: OrderName})
			if fmt.Sprint(byName) != "[avocados Bread cheese]" {
				t.Errorf("by name = %v", byName)
			}

			byOpen, _ := page(t, s, ListQuery{MemberID: owner.ID, Order: OrderOpen})
			if fmt.Sprint(byOpen) != "[avocados cheese Bread]" {
				t.Errorf("by open = %v", byOpen)
			}

			byUpdated, _ := page(t, s, ListQuery{MemberID: owner.ID, Order: OrderUpdated})
			if fmt.Sprint(byUpdated) != "[cheese Bread avocados]" {
				t.Errorf("by updated = %v", byUpdated)
			}
		})
	}
}

/*
The page is cut after the order, not before.

Sorting a page that was already cut sorts it within itself and is wrong about every
other page, which is what doing this in the browser bought.
*/
func TestListsPageCutsAfterOrdering(t *testing.T) {
	for _, d := range drivers() {
		t.Run(d.name, func(t *testing.T) {
			s := d.open(t)
			owner := newMember(t, s)
			for _, name := range []string{"d", "a", "c", "b"} {
				withItems(t, s, owner, name, 1, 0, createdAt)
			}

			first, total := page(t, s, ListQuery{MemberID: owner.ID, Order: OrderName, Limit: 2})
			if fmt.Sprint(first) != "[a b]" || total != 4 {
				t.Errorf("first page = %v of %d, want [a b] of 4", first, total)
			}

			second, _ := page(t, s, ListQuery{MemberID: owner.ID, Order: OrderName, Limit: 2, Offset: 2})
			if fmt.Sprint(second) != "[c d]" {
				t.Errorf("second page = %v, want [c d]", second)
			}
		})
	}
}

// The total counts what matched, not what fitted on the page.
func TestListsPageTotalIgnoresTheLimit(t *testing.T) {
	for _, d := range drivers() {
		t.Run(d.name, func(t *testing.T) {
			s := d.open(t)
			owner := newMember(t, s)
			for i := 0; i < 7; i++ {
				withItems(t, s, owner, fmt.Sprintf("list %d", i), 1, 0, createdAt)
			}

			names, total := page(t, s, ListQuery{MemberID: owner.ID, Limit: 2})
			if len(names) != 2 || total != 7 {
				t.Errorf("got %d rows of %d, want 2 of 7", len(names), total)
			}
		})
	}
}

// An archived List is out of every other filter, and only under Archived.
func TestListsPageKeepsArchivedApart(t *testing.T) {
	for _, d := range drivers() {
		t.Run(d.name, func(t *testing.T) {
			s := d.open(t)
			owner := newMember(t, s)
			gone := withItems(t, s, owner, "Move", 2, 0, createdAt)
			withItems(t, s, owner, "Groceries", 2, 0, createdAt)
			if err := s.SetListArchived(t.Context(), gone.UID, owner.ID, createdAt); err != nil {
				t.Fatalf("SetListArchived: %v", err)
			}

			any, total := page(t, s, ListQuery{MemberID: owner.ID})
			if fmt.Sprint(any) != "[Groceries]" || total != 1 {
				t.Errorf("unfiltered = %v of %d, want [Groceries] of 1", any, total)
			}

			archived, _ := page(t, s, ListQuery{MemberID: owner.ID, Status: StatusArchived})
			if fmt.Sprint(archived) != "[Move]" {
				t.Errorf("archived = %v, want [Move]", archived)
			}
		})
	}
}

/*
The sidebar is four bounded groups, not everything a Member can reach.

Each one says how many there are so the row beneath it can offer the rest, and a List
appears in exactly one of them: pinned wins over finished, and finished wins over the
group it would otherwise sit in.
*/
func TestSidebarGroupsEachListOnce(t *testing.T) {
	for _, d := range drivers() {
		t.Run(d.name, func(t *testing.T) {
			s := d.open(t)
			anna := newMember(t, s)

			pinned := withItems(t, s, anna, "Flat jobs", 2, 0, createdAt)
			if err := s.PinList(t.Context(), anna.ID, pinned.ID); err != nil {
				t.Fatalf("PinList: %v", err)
			}
			// Finished, but pinned on purpose, so it stays where it was put.
			alsoPinned := withItems(t, s, anna, "Party", 0, 3, createdAt)
			if err := s.PinList(t.Context(), anna.ID, alsoPinned.ID); err != nil {
				t.Fatalf("PinList: %v", err)
			}
			withItems(t, s, anna, "Groceries", 5, 0, createdAt)
			withItems(t, s, anna, "Winter tyres", 0, 2, createdAt)

			side, err := s.SidebarLists(t.Context(), anna.ID, SidebarCaps{Pinned: 20, Mine: 8, Shared: 8, Completed: 5}, TokenReach{})
			if err != nil {
				t.Fatalf("SidebarLists: %v", err)
			}

			for _, want := range []struct {
				name  string
				group SidebarGroup
				total int
			}{
				{"pinned", side.Pinned, 2},
				{"mine", side.Mine, 1},
				{"shared", side.Shared, 0},
				{"completed", side.Completed, 1},
			} {
				if want.group.Total != want.total {
					t.Errorf("%s total = %d, want %d", want.name, want.group.Total, want.total)
				}
			}
			if len(side.Mine.Lists) != 1 || side.Mine.Lists[0].Name != "Groceries" {
				t.Errorf("mine = %v, want only Groceries", drew(side.Mine))
			}
			if len(side.Completed.Lists) != 1 || side.Completed.Lists[0].Name != "Winter tyres" {
				t.Errorf("completed = %v, want only Winter tyres", drew(side.Completed))
			}
		})
	}
}

// The cap is what it draws; the total is what there is.
func TestSidebarCapsWhatItDraws(t *testing.T) {
	for _, d := range drivers() {
		t.Run(d.name, func(t *testing.T) {
			s := d.open(t)
			anna := newMember(t, s)
			for i := 0; i < 24; i++ {
				withItems(t, s, anna, fmt.Sprintf("list %02d", i), 1, 0, createdAt)
			}

			side, err := s.SidebarLists(t.Context(), anna.ID, SidebarCaps{Pinned: 20, Mine: 8, Shared: 8, Completed: 5}, TokenReach{})
			if err != nil {
				t.Fatalf("SidebarLists: %v", err)
			}
			if len(side.Mine.Lists) != 8 {
				t.Errorf("drew %d rows, want 8", len(side.Mine.Lists))
			}
			if side.Mine.Total != 24 {
				t.Errorf("total = %d, want 24", side.Mine.Total)
			}
		})
	}
}

// An archived List is out of the sidebar entirely — that is what archiving is.
func TestSidebarLeavesArchivedOut(t *testing.T) {
	for _, d := range drivers() {
		t.Run(d.name, func(t *testing.T) {
			s := d.open(t)
			anna := newMember(t, s)
			gone := withItems(t, s, anna, "Move", 1, 0, createdAt)
			if err := s.SetListArchived(t.Context(), gone.UID, anna.ID, createdAt); err != nil {
				t.Fatalf("SetListArchived: %v", err)
			}

			side, err := s.SidebarLists(t.Context(), anna.ID, SidebarCaps{Pinned: 20, Mine: 8, Shared: 8, Completed: 5}, TokenReach{})
			if err != nil {
				t.Fatalf("SidebarLists: %v", err)
			}
			if side.Mine.Total != 0 || side.Completed.Total != 0 || side.Pinned.Total != 0 {
				t.Error("an archived List is still in the sidebar")
			}
		})
	}
}

// drew is what a group drew, for a failure that reads.
func drew(g SidebarGroup) []string {
	out := make([]string, 0, len(g.Lists))
	for _, l := range g.Lists {
		out = append(out, l.Name)
	}
	return out
}

/*
A token reaches the Lists it names and nothing else, and the query is what enforces it.

Filtering after reading everything is how this worked before, which meant a read of the
whole database to answer about one List. The narrowing has to be part of the query, and
a query that forgot it would hand a token somebody's whole household.
*/
func TestListsPageNarrowsToWhatATokenNames(t *testing.T) {
	for _, d := range drivers() {
		t.Run(d.name, func(t *testing.T) {
			s := d.open(t)
			owner := newMember(t, s)
			groceries := withItems(t, s, owner, "Groceries", 2, 0, createdAt)
			withItems(t, s, owner, "Bike", 1, 0, createdAt)

			named, total := page(t, s, ListQuery{
				MemberID: owner.ID,
				Reach:    TokenReach{Limited: true, ListIDs: []int64{groceries.ID}},
			})
			if fmt.Sprint(named) != "[Groceries]" || total != 1 {
				t.Errorf("a token naming one List saw %v of %d", named, total)
			}

			// A token that names nothing reaches nothing, which is not the same as a
			// caller that is not a token at all.
			none, total := page(t, s, ListQuery{
				MemberID: owner.ID,
				Reach:    TokenReach{Limited: true},
			})
			if len(none) != 0 || total != 0 {
				t.Errorf("a token naming nothing saw %v of %d", none, total)
			}

			all, total := page(t, s, ListQuery{MemberID: owner.ID})
			if len(all) != 2 || total != 2 {
				t.Errorf("a browser saw %v of %d, want both", all, total)
			}
		})
	}
}
