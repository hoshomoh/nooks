package store

import (
	"context"
	"fmt"
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
	// ActivityConflict is two people having edited the same text.
	ActivityConflict ActivityKind = "CONFLICT"
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

// Decided reports whether anybody has acted on this yet.
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

// ActivityFor returns what is waiting for one Member, newest first.
func (s *sqlStore) ActivityFor(ctx context.Context, memberID int64) ([]Activity, error) {
	var rows []activityModel
	err := s.db.NewSelect().
		Model(&rows).
		Where("member_id = ?", memberID).
		Order("created_at DESC", "id DESC").
		Limit(ActivityLimit).
		Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("read activity: %w", err)
	}

	entries := make([]Activity, 0, len(rows))
	for _, row := range rows {
		entry, err := row.toActivity()
		if err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}
	return entries, nil
}

// MarkActivityRead marks everything a Member has now seen.
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

// activityModel is the stored shape of an Activity entry.
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
