package store

import (
	"errors"
	"testing"
	"time"
)

func joinParams() CreateJoinRequestParams {
	return CreateJoinRequestParams{
		UID:       "req_til",
		Name:      "Til",
		Email:     "til@example.com",
		Message:   "It's Til, from upstairs",
		CreatedAt: createdAt,
	}
}

func TestJoinRequestWaitsForAnAdmin(t *testing.T) {
	for _, d := range drivers() {
		t.Run(d.name, func(t *testing.T) {
			s := d.open(t)

			created, err := s.CreateJoinRequest(t.Context(), joinParams())
			if err != nil {
				t.Fatalf("CreateJoinRequest: %v", err)
			}
			if created.Status != StatusPending {
				t.Errorf("Status = %q, want %q", created.Status, StatusPending)
			}
			if !created.DecidedAt.IsZero() {
				t.Error("DecidedAt is set on a request nobody has decided")
			}

			pending, err := s.PendingJoinRequests(t.Context())
			if err != nil {
				t.Fatalf("PendingJoinRequests: %v", err)
			}
			if len(pending) != 1 || pending[0].Name != "Til" {
				t.Errorf("pending = %+v, want one request from Til", pending)
			}
		})
	}
}

func TestDecidingAJoinRequestTakesItOutOfPending(t *testing.T) {
	for _, d := range drivers() {
		t.Run(d.name, func(t *testing.T) {
			s := d.open(t)
			if _, err := s.CreateJoinRequest(t.Context(), joinParams()); err != nil {
				t.Fatalf("CreateJoinRequest: %v", err)
			}

			decidedAt := createdAt.Add(time.Hour)
			if err := s.DecideJoinRequest(t.Context(), "req_til", StatusApproved, decidedAt); err != nil {
				t.Fatalf("DecideJoinRequest: %v", err)
			}

			pending, err := s.PendingJoinRequests(t.Context())
			if err != nil {
				t.Fatalf("PendingJoinRequests: %v", err)
			}
			if len(pending) != 0 {
				t.Errorf("pending = %+v, want none after the decision", pending)
			}

			decided, err := s.JoinRequestByUID(t.Context(), "req_til")
			if err != nil {
				t.Fatalf("JoinRequestByUID: %v", err)
			}
			if decided.Status != StatusApproved {
				t.Errorf("Status = %q, want %q", decided.Status, StatusApproved)
			}
			if !decided.DecidedAt.Equal(decidedAt) {
				t.Errorf("DecidedAt = %v, want %v", decided.DecidedAt, decidedAt)
			}
		})
	}
}

// Deciding twice must not work, or an approval could be replayed after an Admin
// changed their mind.
func TestAJoinRequestIsDecidedOnce(t *testing.T) {
	for _, d := range drivers() {
		t.Run(d.name, func(t *testing.T) {
			s := d.open(t)
			if _, err := s.CreateJoinRequest(t.Context(), joinParams()); err != nil {
				t.Fatalf("CreateJoinRequest: %v", err)
			}
			if err := s.DecideJoinRequest(t.Context(), "req_til", StatusIgnored, createdAt); err != nil {
				t.Fatalf("first DecideJoinRequest: %v", err)
			}

			err := s.DecideJoinRequest(t.Context(), "req_til", StatusApproved, createdAt)
			if !errors.Is(err, ErrNotFound) {
				t.Errorf("second decision = %v, want ErrNotFound", err)
			}
		})
	}
}

func TestApprovingAResetRequestGivesItAnHour(t *testing.T) {
	for _, d := range drivers() {
		t.Run(d.name, func(t *testing.T) {
			s := d.open(t)
			member := newMember(t, s)

			if _, err := s.CreateResetRequest(t.Context(), "req_reset", member.ID, createdAt); err != nil {
				t.Fatalf("CreateResetRequest: %v", err)
			}
			if err := s.DecideResetRequest(t.Context(), "req_reset", StatusApproved, createdAt); err != nil {
				t.Fatalf("DecideResetRequest: %v", err)
			}

			request, err := s.ResetRequestByUID(t.Context(), "req_reset")
			if err != nil {
				t.Fatalf("ResetRequestByUID: %v", err)
			}
			if !request.Usable(createdAt.Add(30 * time.Minute)) {
				t.Error("Usable() = false within the hour, want true")
			}
			if request.Usable(createdAt.Add(2 * time.Hour)) {
				t.Error("Usable() = true after the hour, want false")
			}
		})
	}
}

func TestAPendingResetRequestIsNotUsable(t *testing.T) {
	for _, d := range drivers() {
		t.Run(d.name, func(t *testing.T) {
			s := d.open(t)
			member := newMember(t, s)

			request, err := s.CreateResetRequest(t.Context(), "req_reset", member.ID, createdAt)
			if err != nil {
				t.Fatalf("CreateResetRequest: %v", err)
			}
			if request.Usable(createdAt) {
				t.Error("Usable() = true on a request no Admin has approved")
			}
		})
	}
}

// One approval sets one password.
func TestAResetRequestIsSpentWhenUsed(t *testing.T) {
	for _, d := range drivers() {
		t.Run(d.name, func(t *testing.T) {
			s := d.open(t)
			member := newMember(t, s)

			if _, err := s.CreateResetRequest(t.Context(), "req_reset", member.ID, createdAt); err != nil {
				t.Fatalf("CreateResetRequest: %v", err)
			}
			if err := s.DecideResetRequest(t.Context(), "req_reset", StatusApproved, createdAt); err != nil {
				t.Fatalf("DecideResetRequest: %v", err)
			}

			if err := s.UseResetRequest(t.Context(), "req_reset"); err != nil {
				t.Fatalf("first UseResetRequest: %v", err)
			}
			if err := s.UseResetRequest(t.Context(), "req_reset"); !errors.Is(err, ErrNotFound) {
				t.Errorf("second UseResetRequest = %v, want ErrNotFound", err)
			}
		})
	}
}

func TestRequestNotFound(t *testing.T) {
	for _, d := range drivers() {
		t.Run(d.name, func(t *testing.T) {
			s := d.open(t)
			if _, err := s.JoinRequestByUID(t.Context(), "nope"); !errors.Is(err, ErrNotFound) {
				t.Errorf("JoinRequestByUID = %v, want ErrNotFound", err)
			}
			if _, err := s.ResetRequestByUID(t.Context(), "nope"); !errors.Is(err, ErrNotFound) {
				t.Errorf("ResetRequestByUID = %v, want ErrNotFound", err)
			}
		})
	}
}

/*
An answered request is swept eventually, and an unanswered one never is.

The second half is the one that matters and it is why this is a rule rather than a
retention setting. nooks sends no email, so a join or a reset appears in exactly one
place: the Activity panel. A sweep that took an undecided request would leave somebody
locked out waiting on a decision no Admin can see, which is what pass 39 found the last
time a sweep here was written without that rule in it.

What is swept is history the Activity entry beside it already carries: what happened,
and what an Admin decided about it.
*/
func TestOnlyAnAnsweredRequestIsSwept(t *testing.T) {
	// The cutoff, worked out the way tidyUp works it out, so this reads like the caller
	// rather than like the query.
	cutoffAt := func(now time.Time) time.Time { return now.Add(-DecidedRequestLifetime) }
	longAfter := cutoffAt(createdAt.Add(DecidedRequestLifetime).Add(time.Hour))

	for _, d := range drivers() {
		t.Run(d.name, func(t *testing.T) {
			s := d.open(t)
			member := newMember(t, s)

			waiting, err := s.CreateJoinRequest(t.Context(), joinParams())
			if err != nil {
				t.Fatalf("CreateJoinRequest: %v", err)
			}

			answered := joinParams()
			answered.UID = "req_mara"
			answered.Email = "mara@example.com"
			decided, err := s.CreateJoinRequest(t.Context(), answered)
			if err != nil {
				t.Fatalf("CreateJoinRequest: %v", err)
			}
			if err := s.DecideJoinRequest(t.Context(), decided.UID, StatusIgnored, createdAt); err != nil {
				t.Fatalf("DecideJoinRequest: %v", err)
			}

			reset, err := s.CreateResetRequest(t.Context(), "req_reset", member.ID, createdAt)
			if err != nil {
				t.Fatalf("CreateResetRequest: %v", err)
			}
			if err := s.DecideResetRequest(t.Context(), reset.UID, StatusApproved, createdAt); err != nil {
				t.Fatalf("DecideResetRequest: %v", err)
			}

			gone, err := s.DeleteDecidedRequests(t.Context(), longAfter)
			if err != nil {
				t.Fatalf("DeleteDecidedRequests: %v", err)
			}
			if gone != 2 {
				t.Errorf("swept %d, want the answered join and the answered reset", gone)
			}

			// The one nobody answered, however long it has been waiting.
			if _, err := s.JoinRequestByUID(t.Context(), waiting.UID); err != nil {
				t.Errorf("the request nobody answered is %v, and sweeping it is the fault "+
					"this rule exists for", err)
			}
			if _, err := s.JoinRequestByUID(t.Context(), decided.UID); !errors.Is(err, ErrNotFound) {
				t.Errorf("the answered request is %v, want it swept", err)
			}
		})
	}
}

// Nothing is swept before its time, or a mistake stops being recoverable early.
func TestAnAnsweredRequestIsKeptForItsMonth(t *testing.T) {
	for _, d := range drivers() {
		t.Run(d.name, func(t *testing.T) {
			s := d.open(t)

			request, err := s.CreateJoinRequest(t.Context(), joinParams())
			if err != nil {
				t.Fatalf("CreateJoinRequest: %v", err)
			}
			if err := s.DecideJoinRequest(t.Context(), request.UID, StatusIgnored, createdAt); err != nil {
				t.Fatalf("DecideJoinRequest: %v", err)
			}

			// An hour short of the month, from the caller's side.
			anHourEarly := createdAt.Add(DecidedRequestLifetime).Add(-time.Hour).
				Add(-DecidedRequestLifetime)
			gone, err := s.DeleteDecidedRequests(t.Context(), anHourEarly)
			if err != nil {
				t.Fatalf("DeleteDecidedRequests: %v", err)
			}
			if gone != 0 {
				t.Errorf("swept %d an hour before the month was up, want none", gone)
			}
		})
	}
}
