package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/uptrace/bun"
)

// Permission is what a token may do.
//
// Two levels, not a matrix. A token is a key somebody cuts for a script; anything finer
// would be a permission system a household has to administer.
type Permission string

const (
	// PermissionRead is see and print.
	PermissionRead Permission = "READ"
	// PermissionWrite adds ticking, adding and editing.
	PermissionWrite Permission = "WRITE"
)

// AccessToken is how anything that is not a browser reaches an Instance.
//
// It belongs to a Member and can never do more than they can: it is their access,
// narrowed to some of their Lists and to one of two things they may do there.
type AccessToken struct {
	ID       int64
	UID      string
	MemberID int64
	// Name is what it is for, in the Member's words: "kitchen tablet".
	Name       string
	Permission Permission
	// ExpiresAt is the zero value for a token that does not expire.
	ExpiresAt time.Time
	// LastUsedAt is the zero value until something has used it.
	LastUsedAt time.Time
	CreatedAt  time.Time
}

// Expired reports whether a token has passed its expiry.
func (t AccessToken) Expired(now time.Time) bool {
	return !t.ExpiresAt.IsZero() && now.After(t.ExpiresAt)
}

// CreateAccessTokenParams is everything needed to cut a token.
type CreateAccessTokenParams struct {
	UID      string
	MemberID int64
	Name     string
	// TokenHash is all that is kept. The token itself is shown once and never again.
	TokenHash  string
	Permission Permission
	ExpiresAt  time.Time
	// ListIDs are the Lists it may reach. A List not named here is invisible to it.
	ListIDs []int64
	At      time.Time
}

// CreateAccessToken cuts a token and records what it may reach.
func (s *sqlStore) CreateAccessToken(
	ctx context.Context,
	params CreateAccessTokenParams,
) (AccessToken, error) {
	if params.Name == "" {
		return AccessToken{}, errors.New("store: token name is required")
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return AccessToken{}, fmt.Errorf("begin: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	row := &accessTokenModel{
		UID:        params.UID,
		MemberID:   params.MemberID,
		Name:       params.Name,
		TokenHash:  params.TokenHash,
		Permission: string(params.Permission),
		ExpiresAt:  formatTime(params.ExpiresAt),
		CreatedAt:  formatTime(params.At),
	}
	if _, err := tx.NewInsert().Model(row).Returning("*").Exec(ctx); err != nil {
		return AccessToken{}, fmt.Errorf("create access token: %w", err)
	}

	scope := make([]accessTokenListModel, 0, len(params.ListIDs))
	for _, id := range params.ListIDs {
		scope = append(scope, accessTokenListModel{TokenID: row.ID, ListID: id})
	}
	if len(scope) > 0 {
		if _, err := tx.NewInsert().Model(&scope).Exec(ctx); err != nil {
			return AccessToken{}, fmt.Errorf("scope access token: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return AccessToken{}, fmt.Errorf("commit access token: %w", err)
	}
	return row.toToken()
}

// AccessTokenByHash finds a token by what was presented.
func (s *sqlStore) AccessTokenByHash(ctx context.Context, hash string) (AccessToken, error) {
	row := new(accessTokenModel)
	if err := s.db.NewSelect().Model(row).Where("token_hash = ?", hash).Scan(ctx); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return AccessToken{}, ErrNotFound
		}
		return AccessToken{}, fmt.Errorf("read access token: %w", err)
	}
	return row.toToken()
}

// AccessTokensFor lists a Member's own tokens, newest first.
func (s *sqlStore) AccessTokensFor(ctx context.Context, memberID int64) ([]AccessToken, error) {
	var rows []accessTokenModel
	err := s.db.NewSelect().
		Model(&rows).
		Where("member_id = ?", memberID).
		Order("created_at DESC", "id DESC").
		Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("read access tokens: %w", err)
	}

	tokens := make([]AccessToken, 0, len(rows))
	for _, row := range rows {
		token, err := row.toToken()
		if err != nil {
			return nil, err
		}
		tokens = append(tokens, token)
	}
	return tokens, nil
}

// AccessTokenByUID finds a token by its public identifier.
func (s *sqlStore) AccessTokenByUID(ctx context.Context, uid string) (AccessToken, error) {
	row := new(accessTokenModel)
	if err := s.db.NewSelect().Model(row).Where("uid = ?", uid).Scan(ctx); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return AccessToken{}, ErrNotFound
		}
		return AccessToken{}, fmt.Errorf("read access token: %w", err)
	}
	return row.toToken()
}

// TokenListIDs is which Lists a token may reach.
func (s *sqlStore) TokenListIDs(ctx context.Context, tokenID int64) ([]int64, error) {
	var ids []int64
	err := s.db.NewSelect().
		Model((*accessTokenListModel)(nil)).
		Column("list_id").
		Where("token_id = ?", tokenID).
		Scan(ctx, &ids)
	if err != nil {
		return nil, fmt.Errorf("read token scope: %w", err)
	}
	return ids, nil
}

// MarkTokenUsed records that something reached the Instance with this token.
func (s *sqlStore) MarkTokenUsed(ctx context.Context, id int64, at time.Time) error {
	_, err := s.db.NewUpdate().
		Model((*accessTokenModel)(nil)).
		Set("last_used_at = ?", formatTime(at)).
		Where("id = ?", id).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("mark token used: %w", err)
	}
	return nil
}

// DeleteAccessToken revokes a token. Its scope goes with it.
func (s *sqlStore) DeleteAccessToken(ctx context.Context, id int64) error {
	result, err := s.db.NewDelete().
		Model((*accessTokenModel)(nil)).
		Where("id = ?", id).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("revoke access token: %w", err)
	}
	return requireOneRow(result, "access token")
}

// accessTokenModel is the stored shape of a token.
type accessTokenModel struct {
	bun.BaseModel `bun:"table:access_token,alias:access_token"`

	ID         int64  `bun:"id,pk,autoincrement"`
	UID        string `bun:"uid,notnull"`
	MemberID   int64  `bun:"member_id,notnull"`
	Name       string `bun:"name,notnull"`
	TokenHash  string `bun:"token_hash,notnull"`
	Permission string `bun:"permission,notnull"`
	ExpiresAt  string `bun:"expires_at,notnull"`
	LastUsedAt string `bun:"last_used_at,notnull"`
	CreatedAt  string `bun:"created_at,notnull"`
}

func (m accessTokenModel) toToken() (AccessToken, error) {
	expiresAt, err := parseTime(m.ExpiresAt)
	if err != nil {
		return AccessToken{}, err
	}
	lastUsedAt, err := parseTime(m.LastUsedAt)
	if err != nil {
		return AccessToken{}, err
	}
	createdAt, err := parseTime(m.CreatedAt)
	if err != nil {
		return AccessToken{}, err
	}
	return AccessToken{
		ID: m.ID, UID: m.UID, MemberID: m.MemberID, Name: m.Name,
		Permission: Permission(m.Permission), ExpiresAt: expiresAt,
		LastUsedAt: lastUsedAt, CreatedAt: createdAt,
	}, nil
}

// accessTokenListModel is one List a token may reach.
type accessTokenListModel struct {
	bun.BaseModel `bun:"table:access_token_list,alias:access_token_list"`

	TokenID int64 `bun:"token_id,pk"`
	ListID  int64 `bun:"list_id,pk"`
}
