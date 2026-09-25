package store

import (
	"fmt"
	"slices"
	"strconv"
	"testing"
	"time"
)

// entry records one thing for a Member.
func entry(t *testing.T, s Store, memberID int64, uid string, kind ActivityKind, at time.Time) Activity {
	t.Helper()
	activity, err := s.CreateActivity(t.Context(), CreateActivityParams{
		UID: uid, MemberID: memberID, Kind: kind,
		Text: "Jonas asked to join", TargetUID: "req_1", At: at,
	})
	if err != nil {
		t.Fatalf("CreateActivity: %v", err)
	}
	return activity
}

func TestActivityStartsUnread(t *testing.T) {
	for _, d := range drivers() {
		t.Run(d.name, func(t *testing.T) {
			s := d.open(t)
			member := newMember(t, s)

			created := entry(t, s, member.ID, "act_1", ActivityJoinRequest, createdAt)
			if !created.Unread() {
				t.Error("Unread() = false for an entry nobody has seen")
			}
			if created.Kind != ActivityJoinRequest {
				t.Errorf("Kind = %q, want %q", created.Kind, ActivityJoinRequest)
			}
		})
	}
}

// The panel is what is waiting now, so the newest entry is the one at the top.
func TestActivityIsNewestFirst(t *testing.T) {
	for _, d := range drivers() {
		t.Run(d.name, func(t *testing.T) {
			s := d.open(t)
			member := newMember(t, s)

			entry(t, s, member.ID, "act_old", ActivityJoinRequest, createdAt)
			entry(t, s, member.ID, "act_new", ActivityListShared, createdAt.Add(time.Hour))

			entries, err := s.ActivityFor(t.Context(), member.ID)
			if err != nil {
				t.Fatalf("ActivityFor: %v", err)
			}
			if len(entries) != 2 {
				t.Fatalf("got %d entries, want 2", len(entries))
			}
			if entries[0].UID != "act_new" {
				t.Errorf("first UID = %q, want the newest entry", entries[0].UID)
			}
		})
	}
}

// Activity is personal: it is the only place anything surfaces, so it must not surface
// to the wrong person.
func TestActivityIsPerMember(t *testing.T) {
	for _, d := range drivers() {
		t.Run(d.name, func(t *testing.T) {
			s := d.open(t)
			anna := newMember(t, s)
			jonas := addMember(t, s, "mem_jonas", "Jonas", "jonas@brunnen.lan")

			entry(t, s, anna.ID, "act_1", ActivityJoinRequest, createdAt)

			entries, err := s.ActivityFor(t.Context(), jonas.ID)
			if err != nil {
				t.Fatalf("ActivityFor: %v", err)
			}
			if len(entries) != 0 {
				t.Errorf("got %d entries for the wrong Member, want 0", len(entries))
			}
		})
	}
}

func TestMarkActivityRead(t *testing.T) {
	for _, d := range drivers() {
		t.Run(d.name, func(t *testing.T) {
			s := d.open(t)
			member := newMember(t, s)
			entry(t, s, member.ID, "act_1", ActivityJoinRequest, createdAt)

			seenAt := createdAt.Add(time.Minute)
			if err := s.MarkActivityRead(t.Context(), member.ID, seenAt); err != nil {
				t.Fatalf("MarkActivityRead: %v", err)
			}

			entries, err := s.ActivityFor(t.Context(), member.ID)
			if err != nil {
				t.Fatalf("ActivityFor: %v", err)
			}
			if entries[0].Unread() {
				t.Error("Unread() = true after the Member has seen it")
			}
			if !entries[0].ReadAt.Equal(seenAt) {
				t.Errorf("ReadAt = %v, want %v", entries[0].ReadAt, seenAt)
			}
		})
	}
}

// Nothing older than the panel holds is worth reading.
func TestActivityIsCapped(t *testing.T) {
	for _, d := range drivers() {
		t.Run(d.name, func(t *testing.T) {
			s := d.open(t)
			member := newMember(t, s)
			for i := range ActivityLimit + 10 {
				entry(t, s, member.ID, "act_"+strconv.Itoa(i), ActivityJoinRequest, createdAt.Add(time.Duration(i)*time.Minute))
			}

			entries, err := s.ActivityFor(t.Context(), member.ID)
			if err != nil {
				t.Fatalf("ActivityFor: %v", err)
			}
			if len(entries) != ActivityLimit {
				t.Errorf("got %d entries, want the newest %d", len(entries), ActivityLimit)
			}
		})
	}
}

// Requests go to every Admin, so the sender does not depend on one person being awake.
func TestAdminIDs(t *testing.T) {
	for _, d := range drivers() {
		t.Run(d.name, func(t *testing.T) {
			s := d.open(t)
			anna := newMember(t, s)
			addMember(t, s, "mem_jonas", "Jonas", "jonas@brunnen.lan")

			ids, err := s.AdminIDs(t.Context())
			if err != nil {
				t.Fatalf("AdminIDs: %v", err)
			}
			if len(ids) != 1 || ids[0] != anna.ID {
				t.Errorf("AdminIDs = %v, want just the Admin %d", ids, anna.ID)
			}
		})
	}
}

// Every Admin has their own row for the same request, so deciding it resolves all of
// them: whoever got there first, the others should see what happened rather than a
// button that now does nothing.
func TestDecidingARequestResolvesEveryAdminsEntry(t *testing.T) {
	for _, d := range drivers() {
		t.Run(d.name, func(t *testing.T) {
			s := d.open(t)
			anna := newMember(t, s)
			jonas := addMember(t, s, "mem_jonas", "Jonas", "jonas@brunnen.lan")

			for i, member := range []Member{anna, jonas} {
				if _, err := s.CreateActivity(t.Context(), CreateActivityParams{
					UID: "act_" + strconv.Itoa(i), MemberID: member.ID,
					Kind: ActivityJoinRequest, Text: "Til asked to join",
					TargetUID: "req_til", At: createdAt,
				}); err != nil {
					t.Fatalf("CreateActivity: %v", err)
				}
			}

			if err := s.ResolveActivity(t.Context(), "req_til", OutcomeApproved); err != nil {
				t.Fatalf("ResolveActivity: %v", err)
			}

			for _, member := range []Member{anna, jonas} {
				entries, err := s.ActivityFor(t.Context(), member.ID)
				if err != nil {
					t.Fatalf("ActivityFor: %v", err)
				}
				if !entries[0].Decided() || entries[0].Outcome != OutcomeApproved {
					t.Errorf("%s still sees an undecided request", member.Name)
				}
			}
		})
	}
}

// An entry that is not about that request is left alone.
func TestResolvingLeavesOtherEntriesAlone(t *testing.T) {
	for _, d := range drivers() {
		t.Run(d.name, func(t *testing.T) {
			s := d.open(t)
			member := newMember(t, s)
			entry(t, s, member.ID, "act_share", ActivityListShared, createdAt)

			if err := s.ResolveActivity(t.Context(), "req_til", OutcomeIgnored); err != nil {
				t.Fatalf("ResolveActivity: %v", err)
			}

			entries, err := s.ActivityFor(t.Context(), member.ID)
			if err != nil {
				t.Fatalf("ActivityFor: %v", err)
			}
			if entries[0].Decided() {
				t.Error("an unrelated entry was marked decided")
			}
		})
	}
}

/*
Sweeping Activity leaves every entry anybody can read, and removes the rest.

Nothing had ever deleted one. The panel shows the newest fifty for a Member, so entry
fifty-one has no call that returns it and no screen that shows it, and every change to
every List had been adding rows nobody could reach for the life of the Instance.

Checked by reading the panel before and after: what it shows must not move.
*/
func TestSweepingActivityChangesNothingAnybodyCanSee(t *testing.T) {
	for _, d := range drivers() {
		t.Run(d.name, func(t *testing.T) {
			s := d.open(t)
			anna := newMember(t, s)
			jonas := addMember(t, s, "mem_jonas", "Jonas", "jonas@brunnen.lan")

			// Comfortably past the limit for one of them, and under it for the other.
			for i := range ActivityLimit + 20 {
				addActivity(t, s, anna, fmt.Sprintf("anna %03d", i), createdAt.Add(time.Duration(i)*time.Minute))
			}
			for i := range 3 {
				addActivity(t, s, jonas, fmt.Sprintf("jonas %03d", i), createdAt.Add(time.Duration(i)*time.Minute))
			}

			before := panelFor(t, s, anna)
			theirs := panelFor(t, s, jonas)

			gone, err := s.DeleteUnreadableActivity(t.Context())
			if err != nil {
				t.Fatalf("DeleteUnreadableActivity: %v", err)
			}
			if gone != 20 {
				t.Errorf("removed %d, want the 20 past the limit", gone)
			}

			if after := panelFor(t, s, anna); fmt.Sprint(after) != fmt.Sprint(before) {
				t.Errorf("the panel moved:\n before %v\n after  %v", before, after)
			}
			// Somebody under the limit keeps everything.
			if after := panelFor(t, s, jonas); fmt.Sprint(after) != fmt.Sprint(theirs) {
				t.Errorf("a Member under the limit lost entries: %v, want %v", after, theirs)
			}
		})
	}
}

// Sweeping an Instance with nothing to sweep removes nothing.
func TestSweepingActivityTwiceRemovesNothingTheSecondTime(t *testing.T) {
	for _, d := range drivers() {
		t.Run(d.name, func(t *testing.T) {
			s := d.open(t)
			anna := newMember(t, s)
			for i := range ActivityLimit + 5 {
				addActivity(t, s, anna, fmt.Sprintf("entry %03d", i), createdAt.Add(time.Duration(i)*time.Minute))
			}

			if _, err := s.DeleteUnreadableActivity(t.Context()); err != nil {
				t.Fatalf("first sweep: %v", err)
			}
			gone, err := s.DeleteUnreadableActivity(t.Context())
			if err != nil {
				t.Fatalf("second sweep: %v", err)
			}
			if gone != 0 {
				t.Errorf("the second sweep removed %d, want nothing left to remove", gone)
			}
		})
	}
}

// addActivity puts one entry in front of one Member.
func addActivity(t *testing.T, s Store, member Member, text string, at time.Time) {
	t.Helper()
	if _, err := s.CreateActivity(t.Context(), CreateActivityParams{
		UID: "act_" + text, MemberID: member.ID, Kind: ActivityListShared,
		Text: text, At: at,
	}); err != nil {
		t.Fatalf("CreateActivity %s: %v", text, err)
	}
}

// panelFor is what a Member would see, as text.
func panelFor(t *testing.T, s Store, member Member) []string {
	t.Helper()
	entries, err := s.ActivityFor(t.Context(), member.ID)
	if err != nil {
		t.Fatalf("ActivityFor: %v", err)
	}
	out := make([]string, 0, len(entries))
	for _, entry := range entries {
		out = append(out, entry.Text)
	}
	return out
}

/*
A request nobody has answered stays where the person waiting can be helped.

nooks sends no mail, so Activity is the only place a join or reset request appears. It
used to be the newest fifty entries and nothing else, swept to the same fifty — so a
request that fifty other things happened after was first invisible and then deleted,
while its row sat pending in its own table for ever. The Member who asked to be let back
in waits, and no Admin has anywhere left to see that they asked.
*/
func TestAnUnansweredRequestOutlivesTheLimit(t *testing.T) {
	for _, d := range drivers() {
		t.Run(d.name, func(t *testing.T) {
			s := d.open(t)
			admin := newMember(t, s)

			// The request, then a great deal of ordinary noise on top of it.
			asked := entry(t, s, admin.ID, "act_asked", ActivityResetRequest, createdAt)
			burySomeone(t, s, admin.ID)

			if _, err := s.DeleteUnreadableActivity(t.Context()); err != nil {
				t.Fatalf("DeleteUnreadableActivity: %v", err)
			}

			if !among(t, s, admin.ID, asked.UID) {
				t.Error("the unanswered request is not where an Admin would look for it")
			}
		})
	}
}

// An answered one is history and goes with the rest, or the panel never empties.
func TestAnAnsweredRequestIsSweptLikeAnythingElse(t *testing.T) {
	for _, d := range drivers() {
		t.Run(d.name, func(t *testing.T) {
			s := d.open(t)
			admin := newMember(t, s)

			decided := entry(t, s, admin.ID, "act_decided", ActivityResetRequest, createdAt)
			if err := s.ResolveActivity(t.Context(), decided.TargetUID, OutcomeApproved); err != nil {
				t.Fatalf("ResolveActivity: %v", err)
			}
			burySomeone(t, s, admin.ID)

			if _, err := s.DeleteUnreadableActivity(t.Context()); err != nil {
				t.Fatalf("DeleteUnreadableActivity: %v", err)
			}

			if among(t, s, admin.ID, decided.UID) {
				t.Error("a decided request is history and was kept anyway")
			}
		})
	}
}

// burySomeone puts more than a panel's worth of ordinary entries on top of whatever is
// already there.
func burySomeone(t *testing.T, s Store, memberID int64) {
	t.Helper()
	for i := range ActivityLimit + 10 {
		entry(t, s, memberID, "act_noise_"+strconv.Itoa(i), ActivityListShared,
			createdAt.Add(time.Duration(i+1)*time.Minute))
	}
}

// among reports whether an entry is in what the Member would be shown.
func among(t *testing.T, s Store, memberID int64, uid string) bool {
	t.Helper()
	entries, err := s.ActivityFor(t.Context(), memberID)
	if err != nil {
		t.Fatalf("ActivityFor: %v", err)
	}
	for _, one := range entries {
		if one.UID == uid {
			return true
		}
	}
	return false
}

/*
The panel stays a panel when requests are what is filling it.

Preferring waiting requests would be a way to make the panel as long as somebody liked:
the join endpoint answers strangers and nothing rate-limits it, so every unanswered
request being kept for ever and shown for ever is a worse thing than the one it fixes.
The cap holds either way — waiting requests win their place in it, they do not remove it.
*/
func TestTheLimitHoldsEvenWhenEverythingIsWaiting(t *testing.T) {
	for _, d := range drivers() {
		t.Run(d.name, func(t *testing.T) {
			s := d.open(t)
			admin := newMember(t, s)

			for i := range ActivityLimit + 10 {
				entry(t, s, admin.ID, "act_asked_"+strconv.Itoa(i), ActivityJoinRequest,
					createdAt.Add(time.Duration(i)*time.Minute))
			}

			entries, err := s.ActivityFor(t.Context(), admin.ID)
			if err != nil {
				t.Fatalf("ActivityFor: %v", err)
			}
			if len(entries) != ActivityLimit {
				t.Errorf("got %d entries, want no more than the %d a panel holds",
					len(entries), ActivityLimit)
			}
		})
	}
}

/*
A request nobody has answered survives the sweep, however old it gets.

The sweep removes what nothing can reach: `ActivityFor` returns the newest ActivityLimit
and there is no call that returns an older entry, so the fifty-first is already gone as
far as anybody is concerned. An undecided request is the exception, and it is the reason
the exception exists. Activity is the only place a join or a reset appears, so sweeping
one leaves the row pending in its own table with nothing anywhere that shows it, and the
person waiting is never told either way.

`ActivityFor` had a test for its half of this and the sweep had none, which is the half
that deletes.
*/
func TestTheSweepKeepsARequestNobodyHasAnswered(t *testing.T) {
	for _, d := range drivers() {
		t.Run(d.name, func(t *testing.T) {
			s := d.open(t)
			anna := newMember(t, s)

			// The oldest thing she has, and then enough after it to bury it.
			waiting := entry(t, s, anna.ID, "act_waiting", ActivityJoinRequest, createdAt)
			for i := range ActivityLimit + 20 {
				addActivity(t, s, anna, fmt.Sprintf("anna %03d", i),
					createdAt.Add(time.Duration(i+1)*time.Minute))
			}

			if _, err := s.DeleteUnreadableActivity(t.Context()); err != nil {
				t.Fatalf("DeleteUnreadableActivity: %v", err)
			}

			panel, err := s.ActivityFor(t.Context(), anna.ID)
			if err != nil {
				t.Fatalf("ActivityFor: %v", err)
			}
			kept := slices.ContainsFunc(panel, func(one Activity) bool {
				return one.UID == waiting.UID
			})
			if !kept {
				t.Error("the request Anna has not answered was swept away, and Activity is " +
					"the only place it appears")
			}
		})
	}
}

// Once it is answered it is history, and history is what the sweep is for.
func TestTheSweepRemovesADecidedRequestPastTheLimit(t *testing.T) {
	for _, d := range drivers() {
		t.Run(d.name, func(t *testing.T) {
			s := d.open(t)
			anna := newMember(t, s)

			decided := entry(t, s, anna.ID, "act_decided", ActivityJoinRequest, createdAt)
			if err := s.ResolveActivity(t.Context(), decided.TargetUID, OutcomeApproved); err != nil {
				t.Fatalf("ResolveActivity: %v", err)
			}
			for i := range ActivityLimit + 20 {
				addActivity(t, s, anna, fmt.Sprintf("anna %03d", i),
					createdAt.Add(time.Duration(i+1)*time.Minute))
			}

			if _, err := s.DeleteUnreadableActivity(t.Context()); err != nil {
				t.Fatalf("DeleteUnreadableActivity: %v", err)
			}

			panel, err := s.ActivityFor(t.Context(), anna.ID)
			if err != nil {
				t.Fatalf("ActivityFor: %v", err)
			}
			for _, one := range panel {
				if one.UID == decided.UID {
					t.Error("a request that was answered long ago is still in the panel")
				}
			}
		})
	}
}
