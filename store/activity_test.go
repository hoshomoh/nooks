package store

import (
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
