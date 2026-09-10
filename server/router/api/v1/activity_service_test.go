package v1

import (
	"strings"
	"testing"
	"time"

	"connectrpc.com/connect"

	apiv1 "github.com/hoshomoh/nooks/proto/gen/nooks/api/v1"
	"github.com/hoshomoh/nooks/store"
)

// activityFor reads one Member's panel.
func (f listFixture) activityFor(t *testing.T, member store.Member) *apiv1.ListActivityResponse {
	t.Helper()
	svc := NewActivityService(f.store, func() time.Time { return testClock })
	res, err := svc.ListActivity(f.as(t, member), connect.NewRequest(&apiv1.ListActivityRequest{}))
	if err != nil {
		t.Fatalf("ListActivity: %v", err)
	}
	return res.Msg
}

// Nothing is emailed, so a share has to surface somewhere.
func TestSharingByNameTellsThePersonItNames(t *testing.T) {
	f := newListFixture(t)
	uid := f.createList(t, f.anna, "Groceries")

	f.shareWith(t, f.anna, uid, []string{f.jonas.UID}, nil)

	panel := f.activityFor(t, f.jonas)
	if panel.GetUnreadCount() != 1 {
		t.Fatalf("unread = %d, want 1", panel.GetUnreadCount())
	}
	entry := panel.GetActivity()[0]
	if entry.GetKind() != apiv1.ActivityKind_ACTIVITY_KIND_LIST_SHARED {
		t.Errorf("kind = %v, want a shared list", entry.GetKind())
	}
	if !strings.Contains(entry.GetText(), "Groceries") {
		t.Errorf("text = %q, want it to name the List", entry.GetText())
	}
	if entry.GetTargetUid() != uid {
		t.Errorf("target = %q, want the List it points at", entry.GetTargetUid())
	}
}

// Sharing again with the same person is not news.
func TestSharingTwiceTellsThemOnce(t *testing.T) {
	f := newListFixture(t)
	uid := f.createList(t, f.anna, "Groceries")

	f.shareWith(t, f.anna, uid, []string{f.jonas.UID}, nil)
	f.shareWith(t, f.anna, uid, []string{f.jonas.UID}, nil)

	if got := f.activityFor(t, f.jonas).GetUnreadCount(); got != 1 {
		t.Errorf("unread = %d, want 1", got)
	}
}

// Sharing something with yourself is not news either.
func TestSharingWithYourselfTellsYouNothing(t *testing.T) {
	f := newListFixture(t)
	uid := f.createList(t, f.anna, "Groceries")

	f.shareWith(t, f.anna, uid, []string{f.anna.UID}, nil)

	if got := f.activityFor(t, f.anna).GetUnreadCount(); got != 0 {
		t.Errorf("unread = %d, want nothing", got)
	}
}

func TestMarkingActivityRead(t *testing.T) {
	f := newListFixture(t)
	uid := f.createList(t, f.anna, "Groceries")
	f.shareWith(t, f.anna, uid, []string{f.jonas.UID}, nil)

	svc := NewActivityService(f.store, func() time.Time { return testClock })
	if _, err := svc.MarkActivityRead(f.as(t, f.jonas), connect.NewRequest(
		&apiv1.MarkActivityReadRequest{},
	)); err != nil {
		t.Fatalf("MarkActivityRead: %v", err)
	}

	panel := f.activityFor(t, f.jonas)
	if panel.GetUnreadCount() != 0 {
		t.Errorf("unread = %d after reading it, want 0", panel.GetUnreadCount())
	}
	if len(panel.GetActivity()) != 1 {
		t.Errorf("got %d entries, want it kept — read is not gone", len(panel.GetActivity()))
	}
}
