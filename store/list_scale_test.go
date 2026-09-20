package store

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"
	"time"
)

/*
What a page costs on an Instance far larger than a household's.

The point of paging is that reading one screenful stops depending on how much there is
to read. These say so with a number rather than by argument: seed a great many Lists and
Items, then time the two reads every signed-in screen makes.

Benchmarks, so nothing here runs in CI. To see them:

	go test ./store/ -run '^$' -bench . -benchtime 10x
*/
const (
	benchLists = 100_000
	benchItems = 100_000
)

// seedBench fills a store with lists Lists and items Items spread across them.
func seedBench(b *testing.B, lists, items int) (Store, Member) {
	b.Helper()

	ctx := context.Background()
	s, err := OpenSQLite(ctx, filepath.Join(b.TempDir(), "nooks.db"))
	if err != nil {
		b.Fatalf("OpenSQLite: %v", err)
	}
	b.Cleanup(func() { _ = s.Close() })

	at := time.Date(2026, 9, 16, 9, 0, 0, 0, time.UTC).Format(time.RFC3339Nano)
	anna, err := s.CreateMember(ctx, CreateMemberParams{
		UID: "mem_anna", Name: "Anna", Email: "anna@brunnen.lan",
		Role: RoleAdmin, PasswordHash: "hash", CreatedAt: time.Now(),
	})
	if err != nil {
		b.Fatalf("CreateMember: %v", err)
	}

	// Straight into the tables rather than through CreateList: a hundred thousand round
	// trips would be measuring the seeding.
	db := s.(*sqlStore).db

	rows := make([]listModel, 0, 1000)
	for i := range lists {
		rows = append(rows, listModel{
			UID: fmt.Sprintf("list_%06d", i), Name: fmt.Sprintf("List %06d", i),
			OwnerID: anna.ID, Sharing: string(SharingPrivate), CanEdit: true,
			CreatedAt: at, UpdatedAt: at,
		})
		if len(rows) == cap(rows) || i == lists-1 {
			if _, err := db.NewInsert().Model(&rows).Exec(ctx); err != nil {
				b.Fatalf("insert lists: %v", err)
			}
			rows = rows[:0]
		}
	}

	things := make([]itemModel, 0, 1000)
	for i := range items {
		// Spread over the Lists, and tick every third, so the counts are real work.
		done := ""
		if i%3 == 0 {
			done = at
		}
		things = append(things, itemModel{
			UID: fmt.Sprintf("item_%06d", i), ListID: int64(i%lists) + 1,
			Label: "thing", Position: float64(i), DoneAt: done,
			AddedByID: anna.ID, CreatedAt: at, UpdatedAt: at,
		})
		if len(things) == cap(things) || i == items-1 {
			if _, err := db.NewInsert().Model(&things).Exec(ctx); err != nil {
				b.Fatalf("insert items: %v", err)
			}
			things = things[:0]
		}
	}

	// The counts the store keeps on every List, which seeding straight into the table
	// went around. Without them every List looks empty and the groups that read them
	// are answered from the wrong end.
	const recount = `UPDATE list SET
		open_count = (SELECT COUNT(*) FROM item WHERE item.list_id = list.id AND item.deleted_at = '' AND item.done_at = ''),
		done_count = (SELECT COUNT(*) FROM item WHERE item.list_id = list.id AND item.deleted_at = '' AND item.done_at <> '')`
	if _, err := db.ExecContext(ctx, recount); err != nil {
		b.Fatalf("recount: %v", err)
	}

	// What a running Instance does on its own timer. Without it SQLite plans these
	// reads for the ten Lists it had when the file was made.
	if err := s.Analyse(ctx); err != nil {
		b.Fatalf("Analyse: %v", err)
	}

	return s, anna
}

func BenchmarkListsPage(b *testing.B) {
	s, anna := seedBench(b, benchLists, benchItems)
	ctx := context.Background()

	for b.Loop() {
		if _, err := s.ListsPage(ctx, ListQuery{
			MemberID: anna.ID, Order: OrderName, Offset: 0, Limit: 25,
		}); err != nil {
			b.Fatalf("ListsPage: %v", err)
		}
	}
}

// The deepest page, which is where an offset has the most rows to walk past.
func BenchmarkListsPageAtTheEnd(b *testing.B) {
	s, anna := seedBench(b, benchLists, benchItems)
	ctx := context.Background()

	for b.Loop() {
		if _, err := s.ListsPage(ctx, ListQuery{
			MemberID: anna.ID, Order: OrderName, Offset: benchLists - 25, Limit: 25,
		}); err != nil {
			b.Fatalf("ListsPage: %v", err)
		}
	}
}

func BenchmarkSidebar(b *testing.B) {
	s, anna := seedBench(b, benchLists, benchItems)
	ctx := context.Background()

	for b.Loop() {
		if _, err := s.SidebarLists(ctx, anna.ID, SidebarCaps{
			Pinned: 20, Mine: 8, Shared: 8, Completed: 5,
		}, TokenReach{}); err != nil {
			b.Fatalf("SidebarLists: %v", err)
		}
	}
}

/*
What authorising one Item costs when a Member has a great many Lists.

Ticking something off asks one question: may this caller change the List this Item is
on. The answer must not get slower because somebody made more Lists years ago, which is
what reading them all to find one did.
*/
func BenchmarkListByID(b *testing.B) {
	s, _ := seedBench(b, benchLists, benchItems)
	ctx := context.Background()

	for b.Loop() {
		if _, err := s.ListByID(ctx, benchLists/2); err != nil {
			b.Fatalf("ListByID: %v", err)
		}
	}
}

// The read it replaced, kept so the difference is a number rather than an assertion.
func BenchmarkListsForMember(b *testing.B) {
	s, anna := seedBench(b, benchLists, benchItems)
	ctx := context.Background()

	for b.Loop() {
		if _, err := s.ListsForMember(ctx, anna.ID); err != nil {
			b.Fatalf("ListsForMember: %v", err)
		}
	}
}
