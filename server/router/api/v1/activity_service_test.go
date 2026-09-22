package v1

import (
	"strings"
	"testing"
	"time"

	"connectrpc.com/connect"

	apiv1 "github.com/hoshomoh/nooks/proto/gen/nooks/api/v1"
	"github.com/hoshomoh/nooks/server/auth"
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

/*
A token is not told about a List it does not name.

"A List it does not name is invisible to it" is what the token model says, and the
Activity panel was the one read that did not do it: ActivityFor is by Member, so a token
cut for the shopping list was handed "Jonas shared “Finances” with you" — the name of a
List it cannot open, and of whoever owns it. The MCP tool describing this as "what has
happened on the lists the caller can reach" made it something an assistant would repeat.

The unread count is asserted as well as the entries. A count covering what the caller
cannot see would give away how many there are, which is the part being kept back.

What is deliberately still shown is everything that names no List at all: somebody
asking to join, a password reset, a token first used. Those are not covered by the rule
above, and whether a narrow token should see them is in the defense log.
*/
func TestATokenIsNotToldAboutAListItDoesNotName(t *testing.T) {
	f := newListFixture(t)

	reachable := f.createList(t, f.jonas, "Groceries")
	beyond := f.createList(t, f.jonas, "Finances")
	for _, uid := range []string{reachable, beyond} {
		if _, err := f.svc.SetListSharing(f.as(t, f.jonas), connect.NewRequest(
			&apiv1.SetListSharingRequest{
				ListUid: uid, Sharing: apiv1.Sharing_SHARING_SPECIFIC, MemberUids: []string{f.anna.UID},
			},
		)); err != nil {
			t.Fatalf("SetListSharing %s: %v", uid, err)
		}
	}

	list, err := f.store.ListByUID(t.Context(), reachable)
	if err != nil {
		t.Fatalf("ListByUID: %v", err)
	}
	token := store.AccessToken{
		ID: 1, UID: "tok_1", MemberID: f.anna.ID, Name: "kitchen tablet",
		Abilities: store.TokenAbilities{Read: true}, CreatedAt: testClock,
	}
	ctx := auth.WithGrant(t.Context(), auth.NewTokenGrant(f.anna, token, []int64{list.ID}))

	svc := NewActivityService(f.store, func() time.Time { return testClock })
	res, err := svc.ListActivity(ctx, connect.NewRequest(&apiv1.ListActivityRequest{}))
	if err != nil {
		t.Fatalf("ListActivity: %v", err)
	}

	for _, entry := range res.Msg.GetActivity() {
		if entry.GetTargetUid() == beyond || strings.Contains(entry.GetText(), "Finances") {
			t.Errorf("a token cut for one List was told %q", entry.GetText())
		}
	}
	if got := len(res.Msg.GetActivity()); got != 1 {
		t.Errorf("entries = %d, want only the one naming the List the token reaches", got)
	}
	if got := res.Msg.GetUnreadCount(); got != 1 {
		t.Errorf("unread = %d, want 1: the count must not say how many are being kept back", got)
	}

	// A browser of the same Member still sees both, so this narrows the token and not
	// the account behind it.
	if got := len(f.activityFor(t, f.anna).GetActivity()); got != 2 {
		t.Errorf("the Member's own panel has %d entries, want both", got)
	}
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
