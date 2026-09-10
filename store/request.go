package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/uptrace/bun"
)

// RequestStatus is where a request has got to.
type RequestStatus string

const (
	// StatusPending is waiting for an Admin.
	StatusPending RequestStatus = "PENDING"
	// StatusApproved means the Admin said yes.
	StatusApproved RequestStatus = "APPROVED"
	// StatusIgnored means the Admin said nothing. It is silent, and the sender is
	// never told — see DESIGN.md §9.
	StatusIgnored RequestStatus = "IGNORED"
)

// ResetApprovalLifetime is how long an approved Reset request stays usable. An hour is
// long enough for the Admin to say "go ahead" out loud, and short enough that an
// approval nobody acted on does not linger.
const ResetApprovalLifetime = time.Hour

// JoinRequest is a Visitor's request for an account.
type JoinRequest struct {
	ID    int64
	UID   string
	Name  string
	Email string
	// Message is anything the Visitor added, e.g. "It's Til, from upstairs".
	Message   string
	Status    RequestStatus
	CreatedAt time.Time
	// DecidedAt is the zero value while the request is pending.
	DecidedAt time.Time
}

// ResetRequest is a Member's request to replace a forgotten password.
type ResetRequest struct {
	ID        int64
	UID       string
	MemberID  int64
	Status    RequestStatus
	CreatedAt time.Time
	DecidedAt time.Time
	// ExpiresAt is when an approval stops being usable. Zero while pending.
	ExpiresAt time.Time
}

// Usable reports whether an approved Reset request may still be acted on.
func (r ResetRequest) Usable(now time.Time) bool {
	return r.Status == StatusApproved && now.Before(r.ExpiresAt)
}

// CreateJoinRequestParams is what a Visitor supplies, plus what the server decides.
type CreateJoinRequestParams struct {
	UID       string
	Name      string
	Email     string
	Message   string
	CreatedAt time.Time
}

// CreateJoinRequest records a Visitor's request for an account.
func (s *sqlStore) CreateJoinRequest(ctx context.Context, params CreateJoinRequestParams) (JoinRequest, error) {
	row := &joinRequestModel{
		UID:       params.UID,
		Name:      params.Name,
		Email:     normaliseEmail(params.Email),
		Message:   params.Message,
		Status:    string(StatusPending),
		CreatedAt: formatTime(params.CreatedAt),
	}
	if _, err := s.db.NewInsert().Model(row).Returning("*").Exec(ctx); err != nil {
		return JoinRequest{}, fmt.Errorf("create join request: %w", err)
	}
	return row.toJoinRequest()
}

// PendingJoinRequests lists the requests waiting for an Admin, oldest first.
func (s *sqlStore) PendingJoinRequests(ctx context.Context) ([]JoinRequest, error) {
	var rows []joinRequestModel
	err := s.db.NewSelect().
		Model(&rows).
		Where("status = ?", string(StatusPending)).
		Order("created_at ASC").
		Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("read join requests: %w", err)
	}

	requests := make([]JoinRequest, 0, len(rows))
	for _, row := range rows {
		request, err := row.toJoinRequest()
		if err != nil {
			return nil, err
		}
		requests = append(requests, request)
	}
	return requests, nil
}

// JoinRequestByUID finds one request.
func (s *sqlStore) JoinRequestByUID(ctx context.Context, uid string) (JoinRequest, error) {
	row := new(joinRequestModel)
	if err := s.db.NewSelect().Model(row).Where("uid = ?", uid).Scan(ctx); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return JoinRequest{}, ErrNotFound
		}
		return JoinRequest{}, fmt.Errorf("read join request: %w", err)
	}
	return row.toJoinRequest()
}

// DecideJoinRequest records an Admin's decision.
func (s *sqlStore) DecideJoinRequest(ctx context.Context, uid string, status RequestStatus, at time.Time) error {
	result, err := s.db.NewUpdate().
		Model((*joinRequestModel)(nil)).
		Set("status = ?", string(status)).
		Set("decided_at = ?", formatTime(at)).
		Where("uid = ? AND status = ?", uid, string(StatusPending)).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("decide join request: %w", err)
	}
	return requireOneRow(result, "pending join request")
}

// UseJoinRequest spends an approved request, so one approval creates one account. It
// reports ErrNotFound if the request was never approved or has already been used.
func (s *sqlStore) UseJoinRequest(ctx context.Context, uid string, at time.Time) error {
	result, err := s.db.NewUpdate().
		Model((*joinRequestModel)(nil)).
		Set("status = ?", string(StatusIgnored)).
		Set("decided_at = ?", formatTime(at)).
		Where("uid = ? AND status = ?", uid, string(StatusApproved)).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("use join request: %w", err)
	}
	return requireOneRow(result, "approved join request")
}

// CreateResetRequest records a Member's request to replace a forgotten password.
func (s *sqlStore) CreateResetRequest(ctx context.Context, uid string, memberID int64, at time.Time) (ResetRequest, error) {
	row := &resetRequestModel{
		UID:       uid,
		MemberID:  memberID,
		Status:    string(StatusPending),
		CreatedAt: formatTime(at),
	}
	if _, err := s.db.NewInsert().Model(row).Returning("*").Exec(ctx); err != nil {
		return ResetRequest{}, fmt.Errorf("create reset request: %w", err)
	}
	return row.toResetRequest()
}

// PendingResetRequests lists the requests waiting for an Admin, oldest first.
func (s *sqlStore) PendingResetRequests(ctx context.Context) ([]ResetRequest, error) {
	var rows []resetRequestModel
	err := s.db.NewSelect().
		Model(&rows).
		Where("status = ?", string(StatusPending)).
		Order("created_at ASC").
		Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("read reset requests: %w", err)
	}

	requests := make([]ResetRequest, 0, len(rows))
	for _, row := range rows {
		request, err := row.toResetRequest()
		if err != nil {
			return nil, err
		}
		requests = append(requests, request)
	}
	return requests, nil
}

// ResetRequestByUID finds one request.
func (s *sqlStore) ResetRequestByUID(ctx context.Context, uid string) (ResetRequest, error) {
	row := new(resetRequestModel)
	if err := s.db.NewSelect().Model(row).Where("uid = ?", uid).Scan(ctx); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ResetRequest{}, ErrNotFound
		}
		return ResetRequest{}, fmt.Errorf("read reset request: %w", err)
	}
	return row.toResetRequest()
}

// DecideResetRequest records an Admin's decision. An approval carries an expiry.
func (s *sqlStore) DecideResetRequest(ctx context.Context, uid string, status RequestStatus, at time.Time) error {
	expiresAt := time.Time{}
	if status == StatusApproved {
		expiresAt = at.Add(ResetApprovalLifetime)
	}

	result, err := s.db.NewUpdate().
		Model((*resetRequestModel)(nil)).
		Set("status = ?", string(status)).
		Set("decided_at = ?", formatTime(at)).
		Set("expires_at = ?", formatTime(expiresAt)).
		Where("uid = ? AND status = ?", uid, string(StatusPending)).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("decide reset request: %w", err)
	}
	return requireOneRow(result, "pending reset request")
}

// UseResetRequest marks an approved request as spent, so one approval sets one
// password. It reports ErrNotFound if the request has already been used.
func (s *sqlStore) UseResetRequest(ctx context.Context, uid string) error {
	result, err := s.db.NewUpdate().
		Model((*resetRequestModel)(nil)).
		Set("status = ?", string(StatusIgnored)).
		Where("uid = ? AND status = ?", uid, string(StatusApproved)).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("use reset request: %w", err)
	}
	return requireOneRow(result, "approved reset request")
}

// joinRequestModel is the stored shape of a JoinRequest.
type joinRequestModel struct {
	bun.BaseModel `bun:"table:join_request,alias:join_request"`

	ID        int64  `bun:"id,pk,autoincrement"`
	UID       string `bun:"uid,notnull"`
	Name      string `bun:"name,notnull"`
	Email     string `bun:"email,notnull"`
	Message   string `bun:"message,notnull"`
	Status    string `bun:"status,notnull"`
	CreatedAt string `bun:"created_at,notnull"`
	DecidedAt string `bun:"decided_at,notnull"`
}

func (m joinRequestModel) toJoinRequest() (JoinRequest, error) {
	createdAt, err := parseTime(m.CreatedAt)
	if err != nil {
		return JoinRequest{}, err
	}
	decidedAt, err := parseTime(m.DecidedAt)
	if err != nil {
		return JoinRequest{}, err
	}
	return JoinRequest{
		ID: m.ID, UID: m.UID, Name: m.Name, Email: m.Email, Message: m.Message,
		Status: RequestStatus(m.Status), CreatedAt: createdAt, DecidedAt: decidedAt,
	}, nil
}

// resetRequestModel is the stored shape of a ResetRequest.
type resetRequestModel struct {
	bun.BaseModel `bun:"table:reset_request,alias:reset_request"`

	ID        int64  `bun:"id,pk,autoincrement"`
	UID       string `bun:"uid,notnull"`
	MemberID  int64  `bun:"member_id,notnull"`
	Status    string `bun:"status,notnull"`
	CreatedAt string `bun:"created_at,notnull"`
	DecidedAt string `bun:"decided_at,notnull"`
	ExpiresAt string `bun:"expires_at,notnull"`
}

func (m resetRequestModel) toResetRequest() (ResetRequest, error) {
	createdAt, err := parseTime(m.CreatedAt)
	if err != nil {
		return ResetRequest{}, err
	}
	decidedAt, err := parseTime(m.DecidedAt)
	if err != nil {
		return ResetRequest{}, err
	}
	expiresAt, err := parseTime(m.ExpiresAt)
	if err != nil {
		return ResetRequest{}, err
	}
	return ResetRequest{
		ID: m.ID, UID: m.UID, MemberID: m.MemberID, Status: RequestStatus(m.Status),
		CreatedAt: createdAt, DecidedAt: decidedAt, ExpiresAt: expiresAt,
	}, nil
}
