package store

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/uptrace/bun"
)

// ActivityKind says what an entry is about.
type ActivityKind string

const (
	// ActivityJoinRequest is somebody asking for an account.
	ActivityJoinRequest ActivityKind = "JOIN_REQUEST"
	// ActivityResetRequest is a Member asking to replace a forgotten password.
	ActivityResetRequest ActivityKind = "RESET_REQUEST"
	// ActivityListShared is a List somebody shared with you.
	ActivityListShared ActivityKind = "LIST_SHARED"
	// ActivityTokenUsed is an Access token reaching the Instance for the first time, or
	// one being stopped by somebody other than its owner.
	ActivityTokenUsed ActivityKind = "TOKEN_USED"
)

// Activity is one thing waiting for a Member's attention.
//
// Nooks has no mail server, so this is the only place any of it surfaces — which is why
// an entry carries its own text rather than a code the reader has to interpret.
type Activity struct {
	ID       int64
	UID      string
	MemberID int64
	Kind     ActivityKind
	Text     string
	// TargetUID is what it points at: a request, a List, or nothing.
	TargetUID string
	// ReadAt is the zero value until the Member has seen it.
	ReadAt time.Time
	// Outcome is what became of a request, once an Admin decided. Empty until one has.
	Outcome   Outcome
	CreatedAt time.Time
}

// Outcome is what an Admin decided about a request.
type Outcome string

const (
	// OutcomeApproved is a request an Admin let through.
	OutcomeApproved Outcome = "APPROVED"
	// OutcomeIgnored is a request an Admin dismissed. The sender is never told.
	OutcomeIgnored Outcome = "IGNORED"
)

func (a Activity) Decided() bool { return a.Outcome != "" }

// Unread reports whether the entry still wants attention.
func (a Activity) Unread() bool { return a.ReadAt.IsZero() }

// CreateActivityParams is one entry, for one Member.
type CreateActivityParams struct {
	UID       string
	MemberID  int64
	Kind      ActivityKind
	Text      string
	TargetUID string
	At        time.Time
}

// CreateActivity records something for a Member to see.
func (s *sqlStore) CreateActivity(ctx context.Context, params CreateActivityParams) (Activity, error) {
	row := &activityModel{
		UID:       params.UID,
		MemberID:  params.MemberID,
		Kind:      string(params.Kind),
		Text:      params.Text,
		TargetUID: params.TargetUID,
		CreatedAt: formatTime(params.At),
	}
	if _, err := s.db.NewInsert().Model(row).Returning("*").Exec(ctx); err != nil {
		return Activity{}, fmt.Errorf("create activity: %w", err)
	}
	return row.toActivity()
}

// ActivityLimit is how much of the panel is worth reading. Older entries are history,
// not attention.
const ActivityLimit = 50

/*
DeleteUnreadableActivity removes entries nothing can reach, and says how many went.

ActivityFor hands back the newest ActivityLimit for a Member and nothing else. There is
no call that returns an older one and no screen that shows it, so an entry past the
fiftieth is already gone as far as anybody is concerned — it is only still occupying a
row. This is not a decision about how long to keep history; it is deleting what is
already unreadable.

One statement, numbering each Member's entries newest first and removing what falls past
the end. The window function is the same on both drivers.
*/
func (s *sqlStore) DeleteUnreadableActivity(ctx context.Context) (int64, error) {
	/*
	 * Everything past the end, except a request nobody has answered.
	 *
	 * An undecided request is not history. It is somebody locked out or waiting for an
	 * account, and Activity is the only place it appears — Nooks sends no mail. Sweeping
	 * one away leaves the row pending in its own table for ever with nothing anywhere
	 * that shows it, and the person waiting is never told either way.
	 */
	const beyondTheLimit = `
		SELECT id FROM (
			SELECT id, ROW_NUMBER() OVER (
				PARTITION BY member_id ORDER BY created_at DESC, id DESC
			) AS place
			FROM activity
			WHERE NOT (outcome = '' AND kind IN ('JOIN_REQUEST', 'RESET_REQUEST'))
		) AS ranked WHERE ranked.place > ?`

	result, err := s.db.NewDelete().
		Model((*activityModel)(nil)).
		Where("id IN (?)", bun.SafeQuery(beyondTheLimit, ActivityLimit)).
		Exec(ctx)
	if err != nil {
		return 0, fmt.Errorf("delete unreadable activity: %w", err)
	}

	gone, err := result.RowsAffected()
	if err != nil {
		return 0, nil
	}
	return gone, nil
}

/*
ActivityFor returns what is waiting for one Member, newest first.

The newest ActivityLimit, plus every request nobody has answered however old it is. A
decision somebody is waiting on does not stop mattering because fifty other things
happened after it, and Activity is the only place it appears — so falling off the end
would leave the request pending in its own table and invisible everywhere.

Two reads rather than one clever one: the window function that would express it in a
single statement is harder to read than the thing it does.
*/
func (s *sqlStore) ActivityFor(ctx context.Context, memberID int64) ([]Activity, error) {
	var recent []activityModel
	err := s.db.NewSelect().
		Model(&recent).
		Where("member_id = ?", memberID).
		Order("created_at DESC", "id DESC").
		Limit(ActivityLimit).
		Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("read activity: %w", err)
	}

	var waiting []activityModel
	err = s.db.NewSelect().
		Model(&waiting).
		Where("member_id = ? AND outcome = '' AND kind IN (?)", memberID,
			bun.In([]ActivityKind{ActivityJoinRequest, ActivityResetRequest})).
		Order("created_at DESC", "id DESC").
		Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("read waiting activity: %w", err)
	}

	return merged(waiting, recent)
}

/*
merged reads the sets as one panel: ActivityLimit entries, earlier sets first.

Still a panel rather than a ledger. Waiting requests are offered first so that noise
cannot push one out, and the cap still holds afterwards — otherwise somebody looping the
unauthenticated join endpoint would make the panel as long as they liked, and every
entry in it permanent.
*/
func merged(sets ...[]activityModel) ([]Activity, error) {
	seen := make(map[int64]bool)
	entries := make([]Activity, 0, ActivityLimit)

	for _, rows := range sets {
		for _, row := range rows {
			if len(entries) == ActivityLimit {
				break
			}
			if seen[row.ID] {
				continue
			}
			seen[row.ID] = true
			entry, err := row.toActivity()
			if err != nil {
				return nil, err
			}
			entries = append(entries, entry)
		}
	}

	sort.Slice(entries, func(a, b int) bool {
		if entries[a].CreatedAt.Equal(entries[b].CreatedAt) {
			return entries[a].ID > entries[b].ID
		}
		return entries[a].CreatedAt.After(entries[b].CreatedAt)
	})
	return entries, nil
}

func (s *sqlStore) MarkActivityRead(ctx context.Context, memberID int64, at time.Time) error {
	_, err := s.db.NewUpdate().
		Model((*activityModel)(nil)).
		Set("read_at = ?", formatTime(at)).
		Where("member_id = ? AND read_at = ''", memberID).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("mark activity read: %w", err)
	}
	return nil
}

// ResolveActivity records what became of everything pointing at one request.
//
// Every Admin has their own row for the same request, so deciding it resolves all of
// them: whoever got there first, the others should see what happened rather than a
// button that now does nothing.
func (s *sqlStore) ResolveActivity(ctx context.Context, targetUID string, outcome Outcome) error {
	_, err := s.db.NewUpdate().
		Model((*activityModel)(nil)).
		Set("outcome = ?", string(outcome)).
		Where("target_uid = ? AND outcome = ''", targetUID).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("resolve activity: %w", err)
	}
	return nil
}

/*
AskedAgain says a waiting request has been asked about a second time.

Every Admin has their own row for the same request, so all of them are rewritten: the
one who looks next should see it whoever that is. The row is marked unread again, which
is what puts it back in front of somebody.

`created_at` is deliberately left alone. It says when the request was made, and moving
it to the top of a panel that reads newest first would be the panel telling a small lie
about when somebody asked. What changes is the sentence and the unread mark.

Only an entry nobody has decided is touched. A request that was answered is history, and
history does not become unread because a stranger typed an address again.
*/
func (s *sqlStore) AskedAgain(ctx context.Context, targetUID, text string) error {
	_, err := s.db.NewUpdate().
		Model((*activityModel)(nil)).
		Set("text = ?", text).
		Set("read_at = ?", "").
		Where("target_uid = ? AND outcome = ''", targetUID).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("record that it was asked again: %w", err)
	}
	return nil
}

// AdminIDs lists the Members who can act on a request. Join and Reset requests go to
// every Admin, so the sender does not depend on one person being awake.
func (s *sqlStore) AdminIDs(ctx context.Context) ([]int64, error) {
	var ids []int64
	err := s.db.NewSelect().
		Model((*memberModel)(nil)).
		Column("id").
		Where("role = ?", string(RoleAdmin)).
		Scan(ctx, &ids)
	if err != nil {
		return nil, fmt.Errorf("read admins: %w", err)
	}
	return ids, nil
}

type activityModel struct {
	bun.BaseModel `bun:"table:activity,alias:activity"`

	ID        int64  `bun:"id,pk,autoincrement"`
	UID       string `bun:"uid,notnull"`
	MemberID  int64  `bun:"member_id,notnull"`
	Kind      string `bun:"kind,notnull"`
	Text      string `bun:"text,notnull"`
	TargetUID string `bun:"target_uid,notnull"`
	ReadAt    string `bun:"read_at,notnull"`
	Outcome   string `bun:"outcome,notnull"`
	CreatedAt string `bun:"created_at,notnull"`
}

func (m activityModel) toActivity() (Activity, error) {
	readAt, err := parseTime(m.ReadAt)
	if err != nil {
		return Activity{}, err
	}
	createdAt, err := parseTime(m.CreatedAt)
	if err != nil {
		return Activity{}, err
	}
	return Activity{
		ID: m.ID, UID: m.UID, MemberID: m.MemberID, Kind: ActivityKind(m.Kind),
		Text: m.Text, TargetUID: m.TargetUID, ReadAt: readAt,
		Outcome: Outcome(m.Outcome), CreatedAt: createdAt,
	}, nil
}
