package v1

import (
	"fmt"
	"testing"
	"time"

	"connectrpc.com/connect"

	apiv1 "github.com/hoshomoh/nooks/proto/gen/nooks/api/v1"
	"github.com/hoshomoh/nooks/server/auth"
	"github.com/hoshomoh/nooks/store"
)

// tokens builds the service over the fixture's store, with predictable secrets.
func (f listFixture) tokens(t *testing.T) *TokenService {
	t.Helper()
	cut := 0
	return NewTokenService(
		f.store,
		func() time.Time { return testClock },
		func() (string, error) {
			cut++
			return fmt.Sprintf("tok-%d", cut), nil
		},
		func() (string, string, error) {
			return fmt.Sprintf("secret-%d", cut), fmt.Sprintf("hash-%d", cut), nil
		},
	)
}

// cut makes a token over one List.
func (f listFixture) cut(t *testing.T, listUID string, permission apiv1.Permission) *apiv1.CreateAccessTokenResponse {
	t.Helper()
	res, err := f.tokens(t).CreateAccessToken(f.as(t, f.anna), connect.NewRequest(
		&apiv1.CreateAccessTokenRequest{
			Name: "Kitchen tablet", Permission: permission, ListUids: []string{listUID},
		},
	))
	if err != nil {
		t.Fatalf("CreateAccessToken: %v", err)
	}
	return res.Msg
}

// The only moment the secret exists outside the caller's hands.
func TestCuttingATokenHandsTheSecretOverOnce(t *testing.T) {
	f := newListFixture(t)
	uid := f.createList(t, f.anna, "Groceries")

	made := f.cut(t, uid, apiv1.Permission_PERMISSION_WRITE)
	if made.GetSecret() == "" {
		t.Fatal("no secret was handed over")
	}
	if made.GetToken().GetName() != "Kitchen tablet" {
		t.Errorf("name = %q", made.GetToken().GetName())
	}

	listed, err := f.tokens(t).ListAccessTokens(f.as(t, f.anna), connect.NewRequest(
		&apiv1.ListAccessTokensRequest{},
	))
	if err != nil {
		t.Fatalf("ListAccessTokens: %v", err)
	}
	if len(listed.Msg.GetTokens()) != 1 {
		t.Fatalf("got %d tokens, want the one just cut", len(listed.Msg.GetTokens()))
	}
	// Nothing in the description is the secret.
	for _, field := range []string{
		listed.Msg.GetTokens()[0].GetUid(),
		listed.Msg.GetTokens()[0].GetName(),
	} {
		if field == made.GetSecret() {
			t.Error("the secret comes back when the tokens are listed")
		}
	}
}

// A key to no door. Refusing is kinder than handing somebody a secret that answers
// nothing.
func TestATokenHasToReachSomething(t *testing.T) {
	f := newListFixture(t)

	_, err := f.tokens(t).CreateAccessToken(f.as(t, f.anna), connect.NewRequest(
		&apiv1.CreateAccessTokenRequest{
			Name: "Kitchen tablet", Permission: apiv1.Permission_PERMISSION_READ,
		},
	))
	if got := connect.CodeOf(err); got != connect.CodeInvalidArgument {
		t.Errorf("code = %v, want invalid_argument", got)
	}
}

// A token is its Member's access narrowed, so it cannot be pointed at a List they
// could not open themselves.
func TestATokenCannotReachPastItsMember(t *testing.T) {
	f := newListFixture(t)
	private := f.createList(t, f.jonas, "Bike")

	_, err := f.tokens(t).CreateAccessToken(f.as(t, f.anna), connect.NewRequest(
		&apiv1.CreateAccessTokenRequest{
			Name: "Kitchen tablet", Permission: apiv1.Permission_PERMISSION_READ,
			ListUids: []string{private},
		},
	))
	if got := connect.CodeOf(err); got != connect.CodeNotFound {
		t.Errorf("code = %v, want not_found", got)
	}
}

// Being able to administer an Instance is not being able to act as somebody on it.
func TestAMembersTokensAreTheirOwn(t *testing.T) {
	f := newListFixture(t)
	uid := f.createList(t, f.anna, "Groceries")
	f.cut(t, uid, apiv1.Permission_PERMISSION_READ)

	listed, err := f.tokens(t).ListAccessTokens(f.as(t, f.jonas), connect.NewRequest(
		&apiv1.ListAccessTokensRequest{},
	))
	if err != nil {
		t.Fatalf("ListAccessTokens: %v", err)
	}
	if len(listed.Msg.GetTokens()) != 0 {
		t.Errorf("got %d tokens, want none of somebody else's", len(listed.Msg.GetTokens()))
	}
}

func TestRevokingYourOwnToken(t *testing.T) {
	f := newListFixture(t)
	uid := f.createList(t, f.anna, "Groceries")
	made := f.cut(t, uid, apiv1.Permission_PERMISSION_READ)

	if _, err := f.tokens(t).RevokeAccessToken(f.as(t, f.anna), connect.NewRequest(
		&apiv1.RevokeAccessTokenRequest{TokenUid: made.GetToken().GetUid()},
	)); err != nil {
		t.Fatalf("RevokeAccessToken: %v", err)
	}

	listed, err := f.tokens(t).ListAccessTokens(f.as(t, f.anna), connect.NewRequest(
		&apiv1.ListAccessTokensRequest{},
	))
	if err != nil {
		t.Fatalf("ListAccessTokens: %v", err)
	}
	if len(listed.Msg.GetTokens()) != 0 {
		t.Error("the token still answers after being revoked")
	}
}

// A key somebody left in a door is the Instance's problem, not only its owner's.
func TestAnAdminRevokesAnybodysToken(t *testing.T) {
	f := newListFixture(t)
	uid := f.createList(t, f.jonas, "Bike")

	made, err := f.tokens(t).CreateAccessToken(f.as(t, f.jonas), connect.NewRequest(
		&apiv1.CreateAccessTokenRequest{
			Name: "Shortcut", Permission: apiv1.Permission_PERMISSION_READ,
			ListUids: []string{uid},
		},
	))
	if err != nil {
		t.Fatalf("CreateAccessToken: %v", err)
	}

	// Anna is the Admin here.
	if _, err := f.tokens(t).RevokeAccessToken(f.as(t, f.anna), connect.NewRequest(
		&apiv1.RevokeAccessTokenRequest{TokenUid: made.Msg.GetToken().GetUid()},
	)); err != nil {
		t.Fatalf("RevokeAccessToken as an Admin: %v", err)
	}
}

// Telling a missing token apart from one somebody may not touch would let anyone probe
// for tokens.
func TestRevokingSomebodyElsesTokenReadsAsMissing(t *testing.T) {
	f := newListFixture(t)
	uid := f.createList(t, f.anna, "Groceries")
	made := f.cut(t, uid, apiv1.Permission_PERMISSION_READ)

	_, err := f.tokens(t).RevokeAccessToken(f.as(t, f.jonas), connect.NewRequest(
		&apiv1.RevokeAccessTokenRequest{TokenUid: made.GetToken().GetUid()},
	))
	if got := connect.CodeOf(err); got != connect.CodeNotFound {
		t.Errorf("code = %v, want not_found", got)
	}
}

// A token must not be able to mint or revoke tokens: a leaked one would otherwise be
// able to make itself permanent and to lock its Member out of noticing.
func TestATokenCannotManageTokens(t *testing.T) {
	f := newListFixture(t)
	uid := f.createList(t, f.anna, "Groceries")

	list, err := f.store.ListByUID(t.Context(), uid)
	if err != nil {
		t.Fatalf("ListByUID: %v", err)
	}
	presented := store.AccessToken{ID: 1, MemberID: f.anna.ID, Permission: store.PermissionWrite}
	ctx := auth.WithGrant(t.Context(),
		auth.NewTokenGrant(f.anna, presented, []int64{list.ID}))

	_, err = f.tokens(t).CreateAccessToken(ctx, connect.NewRequest(&apiv1.CreateAccessTokenRequest{
		Name: "Another", ListUids: []string{uid},
	}))
	if got := connect.CodeOf(err); got != connect.CodePermissionDenied {
		t.Errorf("CreateAccessToken code = %v, want permission_denied", got)
	}

	_, err = f.tokens(t).ListAccessTokens(ctx, connect.NewRequest(&apiv1.ListAccessTokensRequest{}))
	if got := connect.CodeOf(err); got != connect.CodePermissionDenied {
		t.Errorf("ListAccessTokens code = %v, want permission_denied", got)
	}
}
