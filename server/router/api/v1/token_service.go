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
	// activity is nil when nothing is listening, which is what a test usually wants.
	activity *TokenActivity
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

// WithActivity tells a Member when somebody else stops one of their keys. Optional: a
// service built without it simply records nothing.
func (s *TokenService) WithActivity(activity *TokenActivity) *TokenService {
	s.activity = activity
	return s
}

// ListAccessTokens returns the signed-in Member's own tokens.
//
// Their own, and only their own. An Admin needing a token makes one; being able to
// administer an Instance is not the same as being able to act as somebody on it.
func (s *TokenService) ListAccessTokens(
	ctx context.Context,
	_ *connect.Request[apiv1.ListAccessTokensRequest],
) (*connect.Response[apiv1.ListAccessTokensResponse], error) {
	member, err := requireBrowser(ctx)
	if err != nil {
		return nil, err
	}

	tokens, err := s.visibleTokens(ctx, member)
	if err != nil {
		return nil, err
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

// visibleTokens is the tokens a Member may be told about.
//
// Their own, and — for an Admin — everybody's, because a key somebody left in a door is
// the Instance's problem. What an Admin is told about somebody else's is deliberately
// thin; see describe.
func (s *TokenService) visibleTokens(
	ctx context.Context,
	member store.Member,
) ([]store.AccessToken, error) {
	if member.IsAdmin() {
		tokens, err := s.store.AllAccessTokens(ctx)
		if err != nil {
			return nil, internalError("read access tokens", err)
		}
		return tokens, nil
	}

	tokens, err := s.store.AccessTokensFor(ctx, member.ID)
	if err != nil {
		return nil, internalError("read access tokens", err)
	}
	return tokens, nil
}

// tellOwnerOfRevocation puts an Admin's revocation in front of whoever cut the token.
func (s *TokenService) tellOwnerOfRevocation(
	ctx context.Context,
	token store.AccessToken,
	by store.Member,
) error {
	if s.activity == nil {
		return nil
	}
	owner, err := s.store.MemberByID(ctx, token.MemberID)
	if err != nil {
		return internalError("read token owner", err)
	}
	return s.activity.TokenRevokedByAdmin(ctx, owner, token, by)
}

// errNoAbilities refuses a token that could do nothing wherever it reached.
var errNoAbilities = connect.NewError(connect.CodeInvalidArgument,
	errors.New("say what the token may do"))

// errNoLists refuses a token that could not reach anything.
var errNoLists = connect.NewError(connect.CodeInvalidArgument,
	errors.New("say which lists the token may reach"))

// CreateAccessToken cuts one.
func (s *TokenService) CreateAccessToken(
	ctx context.Context,
	req *connect.Request[apiv1.CreateAccessTokenRequest],
) (*connect.Response[apiv1.CreateAccessTokenResponse], error) {
	member, err := requireBrowser(ctx)
	if err != nil {
		return nil, err
	}

	name := strings.TrimSpace(req.Msg.GetName())
	if err := requireText(name, "a name"); err != nil {
		return nil, err
	}
	abilities := abilitiesFrom(req.Msg.GetAbilities())
	if !abilities.Usable() {
		return nil, errNoAbilities
	}

	listIDs, err := s.scopeOf(ctx, member, req.Msg)
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
		AllLists:  req.Msg.GetAllLists(),
		Abilities: abilities, ExpiresAt: expiresAt, ListIDs: listIDs, At: s.now(),
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
	member, err := requireBrowser(ctx)
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

	// Somebody whose key has been stopped by an Admin has to be told: the alternative is
	// a script that fails silently and a Member who does not know why.
	if token.MemberID != member.ID {
		if err := s.tellOwnerOfRevocation(ctx, token, member); err != nil {
			return nil, err
		}
	}
	return connect.NewResponse(&apiv1.RevokeAccessTokenResponse{}), nil
}

// errNoSuchToken is returned both for a token that does not exist and one the caller
// may not touch. Telling them apart would let anyone probe for tokens.
var errNoSuchToken = connect.NewError(connect.CodeNotFound, errors.New("no such token"))

// scopeOf resolves what a token is being pointed at.
//
// Reaching everything and naming every List are different answers: the first keeps
// working when a List is made next week, the second deliberately does not.
func (s *TokenService) scopeOf(
	ctx context.Context,
	member store.Member,
	msg *apiv1.CreateAccessTokenRequest,
) ([]int64, error) {
	if msg.GetAllLists() {
		return nil, nil
	}
	// A token that reaches nothing is a key to no door. Refusing it is kinder than
	// handing somebody a secret that will answer nothing.
	if len(msg.GetListUids()) == 0 {
		return nil, errNoLists
	}
	return s.reachableListIDs(ctx, member, msg.GetListUids())
}

// reachableListIDs resolves the Lists a token is being scoped to.
//
// Only Lists the Member can already reach: a token is their access narrowed, so it
// cannot be pointed at something they could not open themselves.
func (s *TokenService) reachableListIDs(
	ctx context.Context,
	member store.Member,
	uids []string,
) ([]int64, error) {
	// Through the same reach the rest of the app uses, so a token cannot be scoped to a
	// List its Member could not open.
	reachable, err := listReach(ctx, s.store, auth.Grant{Member: member})
	if err != nil {
		return nil, err
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

// describe converts a token for the wire.
//
// The Lists it reaches are named only on the reader's own tokens. An Admin may see that
// somebody else's key exists — that is what lets them revoke it — but the names of
// another Member's Lists are not theirs to read.
func (s *TokenService) describe(
	ctx context.Context,
	token store.AccessToken,
	reader store.Member,
) (*apiv1.AccessToken, error) {
	owner, err := s.owner(ctx, token, reader)
	if err != nil {
		return nil, err
	}

	described := &apiv1.AccessToken{
		Uid:        token.UID,
		Name:       token.Name,
		Abilities:  abilitiesToProto(token.Abilities),
		ExpiresAt:  formatMoment(token.ExpiresAt),
		LastUsedAt: formatMoment(token.LastUsedAt),
		CreatedAt:  formatMoment(token.CreatedAt),
		AllLists:   token.AllLists,
		MemberName: owner.Name,
	}
	if token.MemberID != reader.ID {
		return described, nil
	}

	if token.AllLists {
		return described, nil
	}
	described.ListNames, err = s.scopeNames(ctx, token)
	if err != nil {
		return nil, err
	}
	return described, nil
}

// owner is whose token this is, read only when it is not the reader's own.
func (s *TokenService) owner(
	ctx context.Context,
	token store.AccessToken,
	reader store.Member,
) (store.Member, error) {
	if token.MemberID == reader.ID {
		return reader, nil
	}
	owner, err := s.store.MemberByID(ctx, token.MemberID)
	if err != nil {
		return store.Member{}, internalError("read token owner", err)
	}
	return owner, nil
}

// scopeNames is what a token's Lists are called.
func (s *TokenService) scopeNames(
	ctx context.Context,
	token store.AccessToken,
) ([]string, error) {
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
	return names, nil
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

/*
abilitiesFrom reads what a token is being cut for.

Nothing is assumed on. A request that asks for nothing is refused rather than quietly
given read, because a token nobody chose the shape of is a token nobody can reason
about later.
*/
func abilitiesFrom(wanted *apiv1.TokenAbilities) store.TokenAbilities {
	return store.TokenAbilities{
		Read:   wanted.GetRead(),
		Write:  wanted.GetWrite(),
		Delete: wanted.GetDelete(),
	}
}

// abilitiesToProto describes what a token may do.
func abilitiesToProto(abilities store.TokenAbilities) *apiv1.TokenAbilities {
	return &apiv1.TokenAbilities{
		Read:   abilities.Read,
		Write:  abilities.Write,
		Delete: abilities.Delete,
	}
}
