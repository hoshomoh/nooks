package store

import (
	"testing"
)

/*
Two things with the same name come back in the order they were made.

A name is not unique and was never meant to be: two people can each start a List called
Shopping, and an Admin can make two Groups called the same thing by accident. Ordering
stops at the name, so the engine decides the rest, and it is not obliged to decide it the
same way twice. The sidebar then swaps two rows between loads with nobody having touched
anything, which is what presence did when it came back in map order.

What is asked here is the rule rather than the absence of the fault: with the tie broken
on id, the older one is first. An engine that happened to return them that way anyway
would pass this with the tie-break taken out, so it is checked by reversing the tie-break
rather than by removing it.
*/
func TestTwoListsNamedTheSameComeBackOldestFirst(t *testing.T) {
	for _, d := range drivers() {
		t.Run(d.name, func(t *testing.T) {
			s := d.open(t)
			anna := newMember(t, s)

			first := makeList(t, s, anna, "list_one", "Shopping", SharingInstance)
			second := makeList(t, s, anna, "list_two", "Shopping", SharingInstance)

			for range 5 {
				lists, err := s.ListsForMember(t.Context(), anna.ID)
				if err != nil {
					t.Fatalf("ListsForMember: %v", err)
				}
				if len(lists) != 2 {
					t.Fatalf("%d Lists, want the two that were made", len(lists))
				}
				if lists[0].ID != first.ID || lists[1].ID != second.ID {
					t.Fatalf("came back as %d then %d, want the older one first",
						lists[0].ID, lists[1].ID)
				}
			}
		})
	}
}

/*
Two requests made in the same second come back in the order they arrived.

Times are stored to the second, so two people asking to join inside one second are the
same instant as far as the column is concerned. The Members screen draws these in the
order they are given, so without a tie the rows move about.
*/
func TestTwoRequestsInOneSecondComeBackOldestFirst(t *testing.T) {
	for _, d := range drivers() {
		t.Run(d.name, func(t *testing.T) {
			s := d.open(t)

			first, err := s.CreateJoinRequest(t.Context(), CreateJoinRequestParams{
				UID: "req_one", Name: "Til", Email: "til@example.com", CreatedAt: createdAt,
			})
			if err != nil {
				t.Fatalf("CreateJoinRequest: %v", err)
			}
			second, err := s.CreateJoinRequest(t.Context(), CreateJoinRequestParams{
				UID: "req_two", Name: "Mira", Email: "mira@example.com", CreatedAt: createdAt,
			})
			if err != nil {
				t.Fatalf("CreateJoinRequest: %v", err)
			}

			for range 5 {
				waiting, err := s.PendingJoinRequests(t.Context())
				if err != nil {
					t.Fatalf("PendingJoinRequests: %v", err)
				}
				if len(waiting) != 2 {
					t.Fatalf("%d waiting, want the two that asked", len(waiting))
				}
				if waiting[0].ID != first.ID || waiting[1].ID != second.ID {
					t.Fatalf("came back as %d then %d, want the one that asked first",
						waiting[0].ID, waiting[1].ID)
				}
			}
		})
	}
}
