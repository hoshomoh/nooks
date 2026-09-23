package store

import (
	"fmt"
	"sync"
	"testing"
)

/*
One email waits once, and the ceiling holds, even when the asking is concurrent.

A flood is concurrent by definition, so the only honest test of a rule written for one is
a concurrent test. `INSERT ... WHERE NOT EXISTS` decides against whatever the statement
can see: on SQLite that is everything, because there is one writer, and on Postgres it is
a snapshot taken when the statement began, which two callers can share.

This passed before the queue was serialised, which is worth knowing about it: the window
is narrow enough that running it proves little on its own. It earns its place by failing
when the condition is taken out, which it does on both drivers, and by being here when
somebody changes how the queue is written.
*/
func TestOneEmailWaitsOnceUnderContention(t *testing.T) {
	for _, d := range drivers() {
		t.Run(d.name, func(t *testing.T) {
			s := d.open(t)

			const askers = 24
			// Released together, which is the difference between a test that overlaps
			// and one that merely runs several times.
			start := make(chan struct{})
			var asking sync.WaitGroup
			for i := range askers {
				asking.Add(1)
				go func() {
					defer asking.Done()
					<-start
					_, _ = s.CreateJoinRequest(t.Context(), CreateJoinRequestParams{
						UID:       fmt.Sprintf("req_%d", i),
						Name:      "Til",
						Email:     "til@example.com",
						CreatedAt: createdAt,
					})
				}()
			}
			close(start)
			asking.Wait()

			waiting, err := s.PendingJoinRequests(t.Context())
			if err != nil {
				t.Fatalf("PendingJoinRequests: %v", err)
			}
			if len(waiting) != 1 {
				t.Errorf("%d requests waiting from one email, want one", len(waiting))
			}
		})
	}
}

/*
The ceiling holds under contention too, and for the same reason.

Counting inside the insert has the same snapshot problem as looking for a duplicate: the
twenty-sixth and twenty-seventh callers can both count twenty-five.
*/
func TestTheJoinCeilingHoldsUnderContention(t *testing.T) {
	for _, d := range drivers() {
		t.Run(d.name, func(t *testing.T) {
			s := d.open(t)

			const askers = PendingJoinLimit + 12
			start := make(chan struct{})
			var asking sync.WaitGroup
			for i := range askers {
				asking.Add(1)
				go func() {
					defer asking.Done()
					<-start
					_, _ = s.CreateJoinRequest(t.Context(), CreateJoinRequestParams{
						UID:       fmt.Sprintf("req_%d", i),
						Name:      "Til",
						Email:     fmt.Sprintf("til-%d@example.com", i),
						CreatedAt: createdAt,
					})
				}()
			}
			close(start)
			asking.Wait()

			waiting, err := s.PendingJoinRequests(t.Context())
			if err != nil {
				t.Fatalf("PendingJoinRequests: %v", err)
			}
			if len(waiting) > PendingJoinLimit {
				t.Errorf("%d requests waiting, and the ceiling is %d",
					len(waiting), PendingJoinLimit)
			}
			if len(waiting) == 0 {
				t.Error("nothing got through at all, so this is proving nothing")
			}
		})
	}
}
