package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/uptrace/bun"
)

// ErrNotFound reports that the row a caller asked for does not exist. Callers compare
// with errors.Is rather than inspecting driver errors.
var ErrNotFound = errors.New("not found")

// ErrChangedUnderneath reports that a write was refused because what the caller expected
// to find had already been changed by somebody else.
var ErrChangedUnderneath = errors.New("changed underneath")

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
	// RemovedAt is the zero value for a person and set for what is left of one. A
	// removed Member keeps their row so that what they added to other people's Lists
	// is not carried off by a cascade; the row holds no name, no email anybody could
	// use and no password. See Removed.
	RemovedAt time.Time
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

// CreateMember adds a Member and returns it as stored.
func (s *sqlStore) CreateMember(ctx context.Context, params CreateMemberParams) (Member, error) {
	if params.Email == "" {
		return Member{}, errors.New("store: email is required")
	}
	if params.PasswordHash == "" {
		return Member{}, errors.New("store: password hash is required")
	}

	row := &memberModel{
		UID:                params.UID,
		Name:               params.Name,
		Email:              normaliseEmail(params.Email),
		Role:               string(params.Role),
		PasswordHash:       params.PasswordHash,
		MustChangePassword: params.MustChangePassword,
		CreatedAt:          formatTime(params.CreatedAt),
	}

	if _, err := s.db.NewInsert().Model(row).Returning("*").Exec(ctx); err != nil {
		if isUniqueViolationOn(err, "email") {
			return Member{}, ErrEmailTaken
		}
		// Any other unique violation — a UID collision, say — is a bug rather than
		// something the caller did, so it surfaces as itself.
		return Member{}, fmt.Errorf("create member: %w", err)
	}
	return row.toMember()
}

// MemberByEmail finds a Member by the email they sign in with.
func (s *sqlStore) MemberByEmail(ctx context.Context, email string) (Member, error) {
	return s.oneMember(ctx, "email", normaliseEmail(email), false)
}

// SetMemberRole makes somebody an Admin, or stops them being one.
func (s *sqlStore) SetMemberRole(ctx context.Context, id int64, role Role) error {
	result, err := s.db.NewUpdate().
		Model((*memberModel)(nil)).
		Set("role = ?", string(role)).
		Where("id = ?", id).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("set member role: %w", err)
	}
	return requireOneRow(result, "member")
}

// SetMemberProfile changes a Member's own name and email.
//
// The email is normalised the same way it is on the way in, so that a Member who
// retypes theirs with different capitals still signs in with what they had.
func (s *sqlStore) SetMemberProfile(ctx context.Context, id int64, name, email string) error {
	result, err := s.db.NewUpdate().
		Model((*memberModel)(nil)).
		Set("name = ?", name).
		Set("email = ?", normaliseEmail(email)).
		Where("id = ?", id).
		Exec(ctx)
	if err != nil {
		if isUniqueViolationOn(err, "email") {
			return ErrEmailTaken
		}
		return fmt.Errorf("set member profile: %w", err)
	}
	return requireOneRow(result, "member")
}

/*
RemoveMember empties the row rather than deleting it.

Deleting it took more than the person. `item.added_by_id` and `list.owner_id` are both
ON DELETE CASCADE, so removing a housemate who moved out deleted every List they started
and reached into everybody else's Lists to delete every Item they had ever added to
them. The milk they put on the shared list went with them, and nothing recounted, so a
List somebody else owns was left saying "7 things to get" with three on it.

Keeping the row closes both by construction. Nothing cascades, so there is nothing to
null and nothing to recount, and the rule that what somebody added to another person's
List is never deleted holds because there is no longer a mechanism that could.

What is emptied is the person. The email is rewritten under .invalid, which RFC 2606
reserves so it can never belong to anybody, keyed by uid so two removed people cannot
collide on the unique index. The name goes, which is what makes a row they touched say
the time alone. The password hash goes, so the row cannot be signed in to even if
something found it.

Their sessions and their tokens go outright: those are credentials, not history.

What happens to the Lists they own is the caller's decision and not this one's.
*/
func (s *sqlStore) RemoveMember(ctx context.Context, id int64, at time.Time) error {
	member, err := s.MemberByID(ctx, id)
	if err != nil {
		return err
	}
	if member.Removed() {
		return ErrNotFound
	}

	result, err := s.db.NewUpdate().
		Model((*memberModel)(nil)).
		Set("name = ?", "").
		Set("email = ?", removedEmailFor(member.UID)).
		Set("password_hash = ?", "").
		Set("must_change_password = ?", false).
		Set("removed_at = ?", formatTime(at)).
		Where("id = ? AND removed_at = ''", id).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("remove member: %w", err)
	}
	if err := requireOneRow(result, "member"); err != nil {
		return err
	}

	if err := s.DeleteSessionsFor(ctx, id, ""); err != nil {
		return err
	}
	return s.deleteTokensFor(ctx, id)
}

/*
removedEmailFor is the address left on a removed Member's row.

Under .invalid, which RFC 2606 sets aside so that no such name can ever be registered,
so this cannot collide with somebody's real address. Keyed by uid because the column is
unique and a second removal would otherwise fail on the first one's row.
*/
func removedEmailFor(uid string) string {
	return "removed-" + uid + "@invalid"
}

// Removed reports whether this is what is left of a Member rather than one.
func (m Member) Removed() bool { return !m.RemovedAt.IsZero() }

// Members returns everyone on the Instance, by name — who a List can be shared with.
func (s *sqlStore) Members(ctx context.Context) ([]Member, error) {
	var rows []memberModel
	err := s.db.NewSelect().Model(&rows).Where("removed_at = ''").OrderExpr(byName).Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("read members: %w", err)
	}

	members := make([]Member, 0, len(rows))
	for _, row := range rows {
		member, err := row.toMember()
		if err != nil {
			return nil, err
		}
		members = append(members, member)
	}
	return members, nil
}

// MemberByUID finds a Member by their public identifier.
func (s *sqlStore) MemberByUID(ctx context.Context, uid string) (Member, error) {
	return s.oneMember(ctx, "uid", uid, false)
}

// MemberByID finds a Member by internal id, including one who has been removed.
//
// The exception, and deliberately: this is how an Item names whoever added it, and what
// a removed person's row carries is an empty name rather than nothing at all.
func (s *sqlStore) MemberByID(ctx context.Context, id int64) (Member, error) {
	return s.oneMember(ctx, "id", id, true)
}

// CountMembers reports how many Members exist. First run is complete once this is
// above zero.
func (s *sqlStore) CountMembers(ctx context.Context) (int, error) {
	count, err := s.db.NewSelect().Model((*memberModel)(nil)).Where("removed_at = ''").Count(ctx)
	if err != nil {
		return 0, fmt.Errorf("count members: %w", err)
	}
	return count, nil
}

// SetMemberPassword replaces a Member's password and clears the must-change-password
// flag, because replacing it is exactly what clears it.
func (s *sqlStore) SetMemberPassword(ctx context.Context, id int64, hash string) error {
	if hash == "" {
		return errors.New("store: password hash is required")
	}

	result, err := s.db.NewUpdate().
		Model((*memberModel)(nil)).
		Set("password_hash = ?", hash).
		Set("must_change_password = ?", false).
		Where("id = ?", id).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("set member password: %w", err)
	}
	return requireOneRow(result, "member")
}

func (s *sqlStore) MarkMemberSignedIn(ctx context.Context, id int64, at time.Time) error {
	result, err := s.db.NewUpdate().
		Model((*memberModel)(nil)).
		Set("last_signed_in_at = ?", formatTime(at)).
		Where("id = ?", id).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("mark member signed in: %w", err)
	}
	return requireOneRow(result, "member")
}

/*
oneMember finds the Member whose column matches value.

The column name is supplied by this package's own methods and never by a caller, so
interpolating it cannot carry outside input.

alsoRemoved is whether what is left of a removed Member counts as a match. MemberByID is
the only read that takes one, and it has to: it is how an Item names whoever added it,
and a row whose adder has left says the time alone because the name it finds is there
and empty. Every other way of reaching a Member is a way of reaching a person, by
signing in, by sharing a List, by being made an Admin, and a tombstone is none of those.
*/
func (s *sqlStore) oneMember(ctx context.Context, column string, value any, alsoRemoved bool) (Member, error) {
	row := new(memberModel)
	query := s.db.NewSelect().Model(row).Where("? = ?", bun.Ident(column), value)
	if !alsoRemoved {
		query = query.Where("removed_at = ''")
	}
	err := query.Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Member{}, ErrNotFound
		}
		return Member{}, fmt.Errorf("read member: %w", err)
	}
	return row.toMember()
}

// normaliseEmail makes sign-in case-insensitive, since nobody thinks of their address
// as case-sensitive even where the standard allows it.
func normaliseEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
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

// isUniqueViolationOn reports whether an error is a unique-constraint failure naming
// the given column.
//
// The column matters: mapping every unique violation onto one error told a caller
// their email was taken when it was really a UID collision. Neither driver exposes a
// portable constraint name through database/sql, so the message is matched — but both
// name the column in it:
//
//	sqlite:   UNIQUE constraint failed: member.email
//	postgres: duplicate key value violates unique constraint "member_email_key"
func isUniqueViolationOn(err error, column string) bool {
	if err == nil {
		return false
	}
	text := strings.ToLower(err.Error())
	isUnique := strings.Contains(text, "unique constraint") || strings.Contains(text, "duplicate key")
	return isUnique && strings.Contains(text, strings.ToLower(column))
}
