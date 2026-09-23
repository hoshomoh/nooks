package store

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

/*
The counts on a List still match its Items when several people add at once.

The counts are cached on the List and rewritten from the Items after every write, which
is one statement recomputing from truth. On SQLite that statement sees everything
committed before it. On Postgres it sees a snapshot taken when it began, so a recount
that started early can finish last and write a number that was true a moment ago.

It sticks, which is what makes it worth a test: nothing recounts a List until the next
write to it, so a List can show the wrong number until somebody touches it again.
*/
func TestTheCountsSurviveSeveralPeopleAddingAtOnce(t *testing.T) {
	at := time.Date(2026, time.April, 1, 9, 0, 0, 0, time.UTC)
	for _, d := range drivers() {
		t.Run(d.name, func(t *testing.T) {
			s := d.open(t)
			anna := newMember(t, s)
			list := makeList(t, s, anna, "list_shop", "Shopping", SharingInstance)

			const adders = 8
			start := make(chan struct{})
			var adding sync.WaitGroup
			for i := range adders {
				adding.Add(1)
				go func() {
					defer adding.Done()
					<-start
					_, _ = s.CreateItem(t.Context(), CreateItemParams{
						UID: fmt.Sprintf("item_%d", i), ListID: list.ID,
						Label: fmt.Sprintf("Thing %d", i), AddedByID: anna.ID, At: at,
					})
				}()
			}
			close(start)
			adding.Wait()

			items, err := s.ItemsOnList(t.Context(), list.ID)
			if err != nil {
				t.Fatalf("ItemsOnList: %v", err)
			}
			again, err := s.ListByID(t.Context(), list.ID)
			if err != nil {
				t.Fatalf("ListByID: %v", err)
			}
			if again.OpenCount != len(items) {
				t.Errorf("the List says %d open and has %d", again.OpenCount, len(items))
			}
		})
	}
}

/*
And they still match when several people tick at once.

Ticking is the thing two people in a shop do at the same moment, and it moves an Item
from one count to the other, so a stale recount shows both numbers wrong rather than one.
A different path from adding: the write is an update and the List is found from the Item.
*/
func TestTheCountsSurviveSeveralPeopleTickingAtOnce(t *testing.T) {
	at := time.Date(2026, time.April, 1, 9, 0, 0, 0, time.UTC)
	for _, d := range drivers() {
		t.Run(d.name, func(t *testing.T) {
			s := d.open(t)
			anna := newMember(t, s)
			list := makeList(t, s, anna, "list_shop", "Shopping", SharingInstance)

			const shopping = 8
			for i := range shopping {
				_, err := s.CreateItem(t.Context(), CreateItemParams{
					UID: fmt.Sprintf("item_%d", i), ListID: list.ID,
					Label: fmt.Sprintf("Thing %d", i), AddedByID: anna.ID, At: at,
				})
				if err != nil {
					t.Fatalf("CreateItem: %v", err)
				}
			}

			start := make(chan struct{})
			var ticking sync.WaitGroup
			for i := range shopping {
				ticking.Add(1)
				go func() {
					defer ticking.Done()
					<-start
					_ = s.SetItemDone(t.Context(), fmt.Sprintf("item_%d", i), anna.ID, at)
				}()
			}
			close(start)
			ticking.Wait()

			again, err := s.ListByID(t.Context(), list.ID)
			if err != nil {
				t.Fatalf("ListByID: %v", err)
			}
			if again.OpenCount != 0 || again.DoneCount != shopping {
				t.Errorf("the List says %d open and %d done, and everything was ticked",
					again.OpenCount, again.DoneCount)
			}
		})
	}
}
