package v1

import (
	"context"
	"errors"
	"strings"
	"time"

	"connectrpc.com/connect"

	apiv1 "github.com/hoshomoh/nooks/proto/gen/nooks/api/v1"
	"github.com/hoshomoh/nooks/server/auth"
	"github.com/hoshomoh/nooks/store"
)

// TokenService covers the keys a Member cuts for things that are not browsers.
type TokenService struct {
	store    store.Store
	now      func() time.Time
	newUID   func() (string, error)
	newToken func() (token string, hash string, err error)
}

// NewTokenService builds the service. Everything it depends on is injected, so a test
// gets the same token every time and never waits for a clock.
func NewTokenService(
	s store.Store,
	now func() time.Time,
	newUID func() (string, error),
	newToken func() (string, string, error),
) *TokenService {
	if now == nil {
		now = time.Now
	}
	if newUID == nil {
		newUID = newMemberUID
	}
	if newToken == nil {
		newToken = auth.NewToken
	}
	return &TokenService{store: s, now: now, newUID: newUID, newToken: newToken}
}

// ListAccessTokens returns the signed-in Member's own tokens.
//
// Their own, and only their own. An Admin needing a token makes one; being able to
// administer an Instance is not the same as being able to act as somebody on it.
func (s *TokenService) ListAccessTokens(
	ctx context.Context,
	_ *connect.Request[apiv1.ListAccessTokensRequest],
) (*connect.Response[apiv1.ListAccessTokensResponse], error) {
	member, err := requireMember(ctx)
	if err != nil {
		return nil, err
	}

	tokens, err := s.store.AccessTokensFor(ctx, member.ID)
	if err != nil {
		return nil, internalError("read access tokens", err)
	}

	out := make([]*apiv1.AccessToken, 0, len(tokens))
	for _, token := range tokens {
		described, err := s.describe(ctx, token, member)
		if err != nil {
			return nil, err
		}
		out = append(out, described)
	}
	return connect.NewResponse(&apiv1.ListAccessTokensResponse{Tokens: out}), nil
}

// errNoLists refuses a token that could not reach anything.
var errNoLists = connect.NewError(connect.CodeInvalidArgument,
	errors.New("say which lists the token may reach"))

// CreateAccessToken cuts one.
func (s *TokenService) CreateAccessToken(
	ctx context.Context,
	req *connect.Request[apiv1.CreateAccessTokenRequest],
) (*connect.Response[apiv1.CreateAccessTokenResponse], error) {
	member, err := requireMember(ctx)
	if err != nil {
		return nil, err
	}

	name := strings.TrimSpace(req.Msg.GetName())
	if err := requireText(name, "a name"); err != nil {
		return nil, err
	}
	permission := permissionFromProto(req.Msg.GetPermission())
	if permission == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument,
			errors.New("say what the token may do"))
	}

	// A token that reaches nothing is a key to no door. Refusing it is kinder than
	// handing somebody a secret that will answer nothing.
	if len(req.Msg.GetListUids()) == 0 {
		return nil, errNoLists
	}
	listIDs, err := s.reachableListIDs(ctx, member, req.Msg.GetListUids())
	if err != nil {
		return nil, err
	}

	expiresAt, err := parseExpiry(req.Msg.GetExpiresAt())
	if err != nil {
		return nil, err
	}

	uid, err := s.newUID()
	if err != nil {
		return nil, internalError("make an identifier", err)
	}
	secret, hash, err := s.newToken()
	if err != nil {
		return nil, internalError("make a token", err)
	}

	token, err := s.store.CreateAccessToken(ctx, store.CreateAccessTokenParams{
		UID: uid, MemberID: member.ID, Name: name, TokenHash: hash,
		Permission: permission, ExpiresAt: expiresAt, ListIDs: listIDs, At: s.now(),
	})
	if err != nil {
		return nil, internalError("create access token", err)
	}

	described, err := s.describe(ctx, token, member)
	if err != nil {
		return nil, err
	}
	// The only moment the secret exists outside the caller's hands.
	return connect.NewResponse(&apiv1.CreateAccessTokenResponse{
		Token: described, Secret: secret,
	}), nil
}

// RevokeAccessToken stops one working.
func (s *TokenService) RevokeAccessToken(
	ctx context.Context,
	req *connect.Request[apiv1.RevokeAccessTokenRequest],
) (*connect.Response[apiv1.RevokeAccessTokenResponse], error) {
	member, err := requireMember(ctx)
	if err != nil {
		return nil, err
	}

	token, err := s.store.AccessTokenByUID(ctx, req.Msg.GetTokenUid())
	if errors.Is(err, store.ErrNotFound) {
		return nil, errNoSuchToken
	}
	if err != nil {
		return nil, internalError("read access token", err)
	}

	// A Member revokes their own; an Admin revokes anybody's, because a key somebody
	// left in a door is the Instance's problem rather than only its owner's.
	if token.MemberID != member.ID && !member.IsAdmin() {
		return nil, errNoSuchToken
	}

	if err := s.store.DeleteAccessToken(ctx, token.ID); err != nil {
		return nil, internalError("revoke access token", err)
	}
	return connect.NewResponse(&apiv1.RevokeAccessTokenResponse{}), nil
}

// errNoSuchToken is returned both for a token that does not exist and one the caller
// may not touch. Telling them apart would let anyone probe for tokens.
var errNoSuchToken = connect.NewError(connect.CodeNotFound, errors.New("no such token"))

// reachableListIDs resolves the Lists a token is being scoped to.
//
// Only Lists the Member can already reach: a token is their access narrowed, so it
// cannot be pointed at something they could not open themselves.
func (s *TokenService) reachableListIDs(
	ctx context.Context,
	member store.Member,
	uids []string,
) ([]int64, error) {
	reachable, err := s.store.ListsForMember(ctx, member.ID)
	if err != nil {
		return nil, internalError("read lists", err)
	}

	byUID := make(map[string]int64, len(reachable))
	for _, list := range reachable {
		byUID[list.UID] = list.ID
	}

	ids := make([]int64, 0, len(uids))
	for _, uid := range uids {
		id, ok := byUID[uid]
		if !ok {
			return nil, errListNotFound
		}
		ids = append(ids, id)
	}
	return ids, nil
}

// describe converts a token for the wire, naming the Lists it reaches.
func (s *TokenService) describe(
	ctx context.Context,
	token store.AccessToken,
	owner store.Member,
) (*apiv1.AccessToken, error) {
	ids, err := s.store.TokenListIDs(ctx, token.ID)
	if err != nil {
		return nil, internalError("read token scope", err)
	}

	lists, err := s.store.ListsForMember(ctx, token.MemberID)
	if err != nil {
		return nil, internalError("read lists", err)
	}
	names := make([]string, 0, len(ids))
	for _, list := range lists {
		for _, id := range ids {
			if list.ID == id {
				names = append(names, list.Name)
			}
		}
	}

	return &apiv1.AccessToken{
		Uid:        token.UID,
		Name:       token.Name,
		Permission: permissionToProto(token.Permission),
		ListNames:  names,
		ExpiresAt:  formatMoment(token.ExpiresAt),
		LastUsedAt: formatMoment(token.LastUsedAt),
		CreatedAt:  formatMoment(token.CreatedAt),
		MemberName: owner.Name,
	}, nil
}

// parseExpiry reads an expiry, treating empty as "does not expire".
func parseExpiry(value string) (time.Time, error) {
	if value == "" {
		return time.Time{}, nil
	}
	at, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return time.Time{}, connect.NewError(connect.CodeInvalidArgument,
			errors.New("an expiry has to be a date and time"))
	}
	return at, nil
}

// formatMoment writes a time for the wire, or nothing for the zero value.
func formatMoment(at time.Time) string {
	if at.IsZero() {
		return ""
	}
	return at.Format(time.RFC3339)
}

func permissionFromProto(permission apiv1.Permission) store.Permission {
	switch permission {
	case apiv1.Permission_PERMISSION_READ:
		return store.PermissionRead
	case apiv1.Permission_PERMISSION_WRITE:
		return store.PermissionWrite
	default:
		return ""
	}
}

func permissionToProto(permission store.Permission) apiv1.Permission {
	switch permission {
	case store.PermissionRead:
		return apiv1.Permission_PERMISSION_READ
	case store.PermissionWrite:
		return apiv1.Permission_PERMISSION_WRITE
	default:
		return apiv1.Permission_PERMISSION_UNSPECIFIED
	}
}
