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

// Whether a Member may watch one List. The event stream asks this on every page load.
func BenchmarkCanReachList(b *testing.B) {
	s, anna := seedBench(b, benchLists, benchItems)
	ctx := context.Background()
	uid := fmt.Sprintf("list_%06d", benchLists/2)

	for b.Loop() {
		if _, err := s.CanReachList(ctx, anna.ID, uid); err != nil {
			b.Fatalf("CanReachList: %v", err)
		}
	}
}

/*
What Today, Upcoming and the calendar cost.

All three are views over DatedItemsForMember, which takes a lower and an upper bound and
no limit at all. Today passes no lower bound on purpose, so that everything overdue is
gathered rather than left behind whatever day it was — which means the answer grows with
how long the household has been going, not with what is on today.

These say what that costs, and how much comes back, because a read that is quick and
answers with thirty thousand rows is still a read nobody wants: it goes over the wire and
into somebody's phone.

On a hundred thousand Items, Today was 4.7s for 30,083 rows and a week of Upcoming was
114ms for 667. That is 0.15ms a row either way — the same constant, so the cost is what
comes back rather than the work of finding it. A partial index on the dated, undone
Items was tried and moved nothing, which is the evidence for that rather than an
argument for it. What Today needs is a bound, and what that bound should be is a
question about what somebody with thirty thousand overdue things should see.
*/
func BenchmarkDatedItemsToday(b *testing.B) {
	s, anna := seedBench(b, benchLists, benchItems)
	dateThem(b, s)
	ctx := context.Background()

	rows := 0
	b.ResetTimer()
	for b.Loop() {
		items, err := s.DatedItemsForMember(ctx, anna.ID, "", benchToday)
		if err != nil {
			b.Fatalf("DatedItemsForMember: %v", err)
		}
		rows = len(items)
	}
	b.ReportMetric(float64(rows), "rows")
}

// BenchmarkDatedItemsWeek is Upcoming: a bounded window rather than everything behind.
func BenchmarkDatedItemsWeek(b *testing.B) {
	s, anna := seedBench(b, benchLists, benchItems)
	dateThem(b, s)
	ctx := context.Background()

	rows := 0
	b.ResetTimer()
	for b.Loop() {
		items, err := s.DatedItemsForMember(ctx, anna.ID, benchToday, benchWeekOut)
		if err != nil {
			b.Fatalf("DatedItemsForMember: %v", err)
		}
		rows = len(items)
	}
	b.ReportMetric(float64(rows), "rows")
}

const (
	benchToday   = "2026-09-16"
	benchWeekOut = "2026-09-23"
)

// dateThem spreads a due date over a third of the Items, most of them already past,
// which is the shape a List that has been used for a while has.
func dateThem(b *testing.B, s Store) {
	b.Helper()

	db := s.(*sqlStore).db
	// printf carries the sign, which a bare concatenation does not: "+-360 days" is not
	// a modifier SQLite knows, and it answers NULL rather than complaining.
	const spread = `UPDATE item
		SET due_on = date('2026-09-16', printf('%+d days', (id % 400) - 360))
		WHERE id % 3 = 0`
	if _, err := db.ExecContext(context.Background(), spread); err != nil {
		b.Fatalf("spread due dates: %v", err)
	}
	if err := s.Analyse(context.Background()); err != nil {
		b.Fatalf("Analyse: %v", err)
	}
}

/*
What one tick costs on a List that has a great deal on it.

The counts beside a List are written onto it, and recount rewrites them by counting
rather than by adding and subtracting — a counter kept by arithmetic drifts the first
time a path forgets to adjust it, and the number a Member reads is then wrong with
nothing to notice. The comment on recount says being right this way "costs nothing worth
saving", which until now was an argument rather than a number.

It runs on every add, every tick and every delete, and it counts the whole List twice.
So the thing to know is whether that stays small as a List grows, because ticking is the
thing a household does most.
*/
func BenchmarkTickOnACrowdedList(b *testing.B) {
	for _, items := range []int{10, 100, 1_000, 10_000, 100_000} {
		b.Run(fmt.Sprintf("%d_items", items), func(b *testing.B) {
			s, anna := seedOneList(b, items)
			ctx := context.Background()

			// A different Item each time, so nothing is measuring a no-op.
			at := time.Date(2026, 9, 16, 9, 0, 0, 0, time.UTC)
			turn := 0
			b.ResetTimer()
			for b.Loop() {
				turn++
				uid := fmt.Sprintf("item_%06d", turn%items)
				if err := s.SetItemDone(ctx, uid, anna.ID, at); err != nil {
					b.Fatalf("SetItemDone: %v", err)
				}
			}
		})
	}
}

// seedOneList fills a single List, which is the shape that makes recount work hardest.
func seedOneList(b *testing.B, items int) (Store, Member) {
	b.Helper()
	return seedBench(b, 1, items)
}

/*
What reading one List costs as it fills up.

ItemsOnList takes no bound: GetList answers with every Item there is, and so does the
MCP get_list an assistant calls. The parked question is whether that should page, and
this is the number that question is about.

A List grows in one direction only — ticking collapses an Item to the foot rather than
removing it — so the size here is not a stress test. It is a weekly shop kept for a few
years.
*/
func BenchmarkReadOneList(b *testing.B) {
	for _, items := range []int{100, 1_000, 10_000, 100_000} {
		b.Run(fmt.Sprintf("%d_items", items), func(b *testing.B) {
			s, _ := seedOneList(b, items)
			ctx := context.Background()

			rows := 0
			b.ResetTimer()
			for b.Loop() {
				read, err := s.ItemsOnList(ctx, 1)
				if err != nil {
					b.Fatalf("ItemsOnList: %v", err)
				}
				rows = len(read)
			}
			b.ReportMetric(float64(rows), "rows")
		})
	}
}
