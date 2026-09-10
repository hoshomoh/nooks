package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

// ErrNotFound reports that the row a caller asked for does not exist. Callers compare
// with errors.Is rather than inspecting driver errors.
var ErrNotFound = errors.New("not found")

// ErrEmailTaken reports that another Member already uses that email.
var ErrEmailTaken = errors.New("that email already has an account")

// Role is what a Member may do on this Instance.
type Role string

const (
	// RoleAdmin can reach instance settings, approve requests, and manage Members and
	// Groups. The Member who completes first run is the first Admin.
	RoleAdmin Role = "ADMIN"
	// RoleMember is everyone else.
	RoleMember Role = "MEMBER"
)

// Member is a person with an account on this Instance.
type Member struct {
	// ID is the stable internal identity. It never appears in the API.
	ID int64
	// UID is the public identifier used in resource names and URLs.
	UID string
	// Name is how the Member appears on Items they add.
	Name string
	// Email identifies the Member at sign-in. Nooks never sends mail to it.
	Email string
	Role  Role
	// PasswordHash is a bcrypt hash. The password itself is never stored or logged.
	PasswordHash string
	// MustChangePassword is set when an Admin created the account with a temporary
	// password, and cleared once the Member replaces it.
	MustChangePassword bool
	CreatedAt          time.Time
	// LastSignedInAt is the zero value until the Member has signed in at least once.
	// The Members table greys a Member who never has.
	LastSignedInAt time.Time
}

// IsAdmin reports whether the Member holds the Admin role.
func (m Member) IsAdmin() bool { return m.Role == RoleAdmin }

// CreateMemberParams is everything needed to add a Member. The caller supplies the
// hash, the UID and the clock, so this package neither hashes nor invents time.
type CreateMemberParams struct {
	UID                string
	Name               string
	Email              string
	Role               Role
	PasswordHash       string
	MustChangePassword bool
	CreatedAt          time.Time
}

const memberColumns = `id, uid, name, email, role, password_hash, must_change_password,
	created_at, last_signed_in_at`

// CreateMember adds a Member and returns it as stored.
func (s *sqlStore) CreateMember(ctx context.Context, params CreateMemberParams) (Member, error) {
	if params.Email == "" {
		return Member{}, errors.New("store: email is required")
	}
	if params.PasswordHash == "" {
		return Member{}, errors.New("store: password hash is required")
	}

	query := fmt.Sprintf(`INSERT INTO member
		(uid, name, email, role, password_hash, must_change_password, created_at, last_signed_in_at)
		VALUES (%s, %s, %s, %s, %s, %s, %s, '')
		RETURNING %s`,
		s.dialect.placeholder(1), s.dialect.placeholder(2), s.dialect.placeholder(3),
		s.dialect.placeholder(4), s.dialect.placeholder(5), s.dialect.placeholder(6),
		s.dialect.placeholder(7), memberColumns)

	row := s.db.QueryRowContext(ctx, query,
		params.UID, params.Name, normaliseEmail(params.Email), string(params.Role),
		params.PasswordHash, boolToInt(params.MustChangePassword),
		formatTime(params.CreatedAt))

	member, err := scanMember(row)
	if err != nil {
		if isUniqueViolation(err) {
			return Member{}, ErrEmailTaken
		}
		return Member{}, fmt.Errorf("create member: %w", err)
	}
	return member, nil
}

// MemberByEmail finds a Member by the email they sign in with.
func (s *sqlStore) MemberByEmail(ctx context.Context, email string) (Member, error) {
	query := fmt.Sprintf(`SELECT %s FROM member WHERE email = %s`,
		memberColumns, s.dialect.placeholder(1))
	return s.oneMember(ctx, query, normaliseEmail(email))
}

// MemberByUID finds a Member by their public identifier.
func (s *sqlStore) MemberByUID(ctx context.Context, uid string) (Member, error) {
	query := fmt.Sprintf(`SELECT %s FROM member WHERE uid = %s`,
		memberColumns, s.dialect.placeholder(1))
	return s.oneMember(ctx, query, uid)
}

// MemberByID finds a Member by internal identity.
func (s *sqlStore) MemberByID(ctx context.Context, id int64) (Member, error) {
	query := fmt.Sprintf(`SELECT %s FROM member WHERE id = %s`,
		memberColumns, s.dialect.placeholder(1))
	return s.oneMember(ctx, query, id)
}

// CountMembers reports how many Members exist. First run is complete once this is
// above zero.
func (s *sqlStore) CountMembers(ctx context.Context) (int, error) {
	var count int
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM member`).Scan(&count); err != nil {
		return 0, fmt.Errorf("count members: %w", err)
	}
	return count, nil
}

// SetMemberPassword replaces a Member's password and clears the
// must-change-password flag, because replacing it is exactly what clears it.
func (s *sqlStore) SetMemberPassword(ctx context.Context, id int64, hash string) error {
	if hash == "" {
		return errors.New("store: password hash is required")
	}
	query := fmt.Sprintf(
		`UPDATE member SET password_hash = %s, must_change_password = 0 WHERE id = %s`,
		s.dialect.placeholder(1), s.dialect.placeholder(2))

	result, err := s.db.ExecContext(ctx, query, hash, id)
	if err != nil {
		return fmt.Errorf("set member password: %w", err)
	}
	return requireOneRow(result, "member")
}

// MarkMemberSignedIn records that a Member has just signed in.
func (s *sqlStore) MarkMemberSignedIn(ctx context.Context, id int64, at time.Time) error {
	query := fmt.Sprintf(`UPDATE member SET last_signed_in_at = %s WHERE id = %s`,
		s.dialect.placeholder(1), s.dialect.placeholder(2))

	result, err := s.db.ExecContext(ctx, query, formatTime(at), id)
	if err != nil {
		return fmt.Errorf("mark member signed in: %w", err)
	}
	return requireOneRow(result, "member")
}

// oneMember runs a query expected to match at most one Member.
func (s *sqlStore) oneMember(ctx context.Context, query string, arg any) (Member, error) {
	member, err := scanMember(s.db.QueryRowContext(ctx, query, arg))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Member{}, ErrNotFound
		}
		return Member{}, fmt.Errorf("read member: %w", err)
	}
	return member, nil
}

// rowScanner is satisfied by *sql.Row and *sql.Rows alike.
type rowScanner interface {
	Scan(dest ...any) error
}

// scanMember reads one row in memberColumns order.
func scanMember(row rowScanner) (Member, error) {
	var (
		m                  Member
		role               string
		mustChangePassword int64
		createdAt          string
		lastSignedInAt     string
	)
	err := row.Scan(&m.ID, &m.UID, &m.Name, &m.Email, &role, &m.PasswordHash,
		&mustChangePassword, &createdAt, &lastSignedInAt)
	if err != nil {
		return Member{}, err
	}

	m.Role = Role(role)
	m.MustChangePassword = mustChangePassword != 0
	if m.CreatedAt, err = parseTime(createdAt); err != nil {
		return Member{}, err
	}
	if m.LastSignedInAt, err = parseTime(lastSignedInAt); err != nil {
		return Member{}, err
	}
	return m, nil
}

// normaliseEmail makes sign-in case-insensitive, since nobody thinks of their address
// as case-sensitive even where the standard allows it.
func normaliseEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
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
		return time.Time{}, fmt.Errorf("parse timestamp %q: %w", raw, err)
	}
	return parsed, nil
}

// requireOneRow turns "updated nothing" into ErrNotFound.
func requireOneRow(result sql.Result, what string) error {
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}
	if affected == 0 {
		return fmt.Errorf("%s: %w", what, ErrNotFound)
	}
	return nil
}

// isUniqueViolation reports whether an error is a unique-constraint failure. The two
// drivers word it differently and neither exposes a portable code through
// database/sql, so the text is matched.
func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	text := strings.ToLower(err.Error())
	return strings.Contains(text, "unique constraint") ||
		strings.Contains(text, "duplicate key")
}
