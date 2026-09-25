package store

import (
	"time"

	"github.com/uptrace/bun"
)

// The persistence models.
//
// They are separate from the domain types in store.go, member.go and session.go so that
// the ORM's tags — and the storage shapes it needs, like timestamps as text — stay out
// of the types the rest of nooks passes around. The mapping functions below are the one
// place those two shapes meet.

type settingModel struct {
	bun.BaseModel `bun:"table:setting,alias:setting"`

	Key   string `bun:"key,pk"`
	Value string `bun:"value,notnull"`
}

// Timestamps are RFC3339 text rather than a native type: SQLite has no date type, and
// storing the same text in both drivers keeps a database file readable by hand — which
// matters when the whole Instance is one file the owner can copy.
type memberModel struct {
	bun.BaseModel `bun:"table:member,alias:member"`

	ID                 int64  `bun:"id,pk,autoincrement"`
	UID                string `bun:"uid,notnull"`
	Name               string `bun:"name,notnull"`
	Email              string `bun:"email,notnull"`
	Role               string `bun:"role,notnull"`
	PasswordHash       string `bun:"password_hash,notnull"`
	MustChangePassword bool   `bun:"must_change_password,notnull"`
	CreatedAt          string `bun:"created_at,notnull"`
	LastSignedInAt     string `bun:"last_signed_in_at,notnull"`
	RemovedAt          string `bun:"removed_at,notnull"`
}

type sessionModel struct {
	bun.BaseModel `bun:"table:session,alias:session"`

	TokenHash  string `bun:"token_hash,pk"`
	MemberID   int64  `bun:"member_id,notnull"`
	CreatedAt  string `bun:"created_at,notnull"`
	ExpiresAt  string `bun:"expires_at,notnull"`
	Kind       string `bun:"kind,notnull"`
	ParentHash string `bun:"parent_hash,notnull"`
}

func (m memberModel) toMember() (Member, error) {
	createdAt, err := parseTime(m.CreatedAt)
	if err != nil {
		return Member{}, err
	}
	lastSignedInAt, err := parseTime(m.LastSignedInAt)
	if err != nil {
		return Member{}, err
	}
	removedAt, err := parseTime(m.RemovedAt)
	if err != nil {
		return Member{}, err
	}
	return Member{
		ID:                 m.ID,
		UID:                m.UID,
		Name:               m.Name,
		Email:              m.Email,
		Role:               Role(m.Role),
		PasswordHash:       m.PasswordHash,
		MustChangePassword: m.MustChangePassword,
		CreatedAt:          createdAt,
		LastSignedInAt:     lastSignedInAt,
		RemovedAt:          removedAt,
	}, nil
}

func (s sessionModel) toSession() (Session, error) {
	createdAt, err := parseTime(s.CreatedAt)
	if err != nil {
		return Session{}, err
	}
	expiresAt, err := parseTime(s.ExpiresAt)
	if err != nil {
		return Session{}, err
	}
	return Session{
		TokenHash:  s.TokenHash,
		MemberID:   s.MemberID,
		CreatedAt:  createdAt,
		ExpiresAt:  expiresAt,
		Kind:       SessionKind(s.Kind),
		ParentHash: s.ParentHash,
	}, nil
}

func newSessionModel(s Session) *sessionModel {
	kind := s.Kind
	if kind == "" {
		kind = SessionRefresh
	}
	return &sessionModel{
		TokenHash:  s.TokenHash,
		MemberID:   s.MemberID,
		CreatedAt:  formatTime(s.CreatedAt),
		ExpiresAt:  formatTime(s.ExpiresAt),
		Kind:       string(kind),
		ParentHash: s.ParentHash,
	}
}

// formatTime renders a timestamp for storage. The zero value stores as empty text,
// which is how "never" is represented.
func formatTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.UTC().Format(time.RFC3339)
}

// parseTime is the inverse of formatTime.
func parseTime(raw string) (time.Time, error) {
	if raw == "" {
		return time.Time{}, nil
	}
	parsed, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		return time.Time{}, err
	}
	return parsed, nil
}
