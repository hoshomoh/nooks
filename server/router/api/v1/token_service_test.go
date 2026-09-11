package v1

import (
	"fmt"
	"strings"
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
func (f listFixture) cut(t *testing.T, listUID string, abilities *apiv1.TokenAbilities) *apiv1.CreateAccessTokenResponse {
	t.Helper()
	res, err := f.tokens(t).CreateAccessToken(f.as(t, f.anna), connect.NewRequest(
		&apiv1.CreateAccessTokenRequest{
			Name: "Kitchen tablet", Abilities: abilities, ListUids: []string{listUID},
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

	made := f.cut(t, uid, &apiv1.TokenAbilities{Read: true, Write: true, Delete: true})
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
			Name: "Kitchen tablet", Abilities: &apiv1.TokenAbilities{Read: true},
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
			Name: "Kitchen tablet", Abilities: &apiv1.TokenAbilities{Read: true},
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
	f.cut(t, uid, &apiv1.TokenAbilities{Read: true})

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
	made := f.cut(t, uid, &apiv1.TokenAbilities{Read: true})

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
			Name: "Shortcut", Abilities: &apiv1.TokenAbilities{Read: true},
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
	made := f.cut(t, uid, &apiv1.TokenAbilities{Read: true})

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
	presented := store.AccessToken{ID: 1, MemberID: f.anna.ID, Abilities: store.TokenAbilities{Read: true, Write: true, Delete: true}}
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

// An Admin can see that somebody else's key exists — that is what lets them revoke it —
// but the names of another Member's Lists are not theirs to read.
func TestAnAdminSeesThatOthersTokensExistWithoutTheirScope(t *testing.T) {
	f := newListFixture(t)
	uid := f.createList(t, f.jonas, "Bike")

	_, err := f.tokens(t).CreateAccessToken(f.as(t, f.jonas), connect.NewRequest(
		&apiv1.CreateAccessTokenRequest{
			Name: "Jonas' shortcut", ListUids: []string{uid},
			Abilities: &apiv1.TokenAbilities{Read: true},
		},
	))
	if err != nil {
		t.Fatalf("CreateAccessToken: %v", err)
	}

	listed, err := f.tokens(t).ListAccessTokens(f.as(t, f.anna), connect.NewRequest(
		&apiv1.ListAccessTokensRequest{},
	))
	if err != nil {
		t.Fatalf("ListAccessTokens as an Admin: %v", err)
	}

	tokens := listed.Msg.GetTokens()
	if len(tokens) != 1 {
		t.Fatalf("got %d tokens, want the one somebody else cut", len(tokens))
	}
	if got := tokens[0].GetMemberName(); got != f.jonas.Name {
		t.Errorf("memberName = %q, want %q", got, f.jonas.Name)
	}
	if got := tokens[0].GetListNames(); len(got) != 0 {
		t.Errorf("listNames = %v, want none for somebody else's token", got)
	}
}

// A Member is told about their own tokens and nobody else's.
func TestAMemberSeesOnlyTheirOwnTokens(t *testing.T) {
	f := newListFixture(t)
	mine := f.createList(t, f.jonas, "Bike")
	theirs := f.createList(t, f.anna, "Groceries")

	// One service across both cuts: its secrets are a counter, and a fresh one would
	// hand out the same hash twice.
	svc := f.tokens(t)
	for _, cut := range []struct {
		member store.Member
		list   string
		name   string
	}{{f.jonas, mine, "Mine"}, {f.anna, theirs, "Theirs"}} {
		_, err := svc.CreateAccessToken(f.as(t, cut.member), connect.NewRequest(
			&apiv1.CreateAccessTokenRequest{
				Name: cut.name, ListUids: []string{cut.list},
				Abilities: &apiv1.TokenAbilities{Read: true},
			},
		))
		if err != nil {
			t.Fatalf("CreateAccessToken %s: %v", cut.name, err)
		}
	}

	listed, err := svc.ListAccessTokens(f.as(t, f.jonas), connect.NewRequest(
		&apiv1.ListAccessTokensRequest{},
	))
	if err != nil {
		t.Fatalf("ListAccessTokens: %v", err)
	}

	tokens := listed.Msg.GetTokens()
	if len(tokens) != 1 || tokens[0].GetName() != "Mine" {
		t.Fatalf("got %d tokens, want only the Member's own", len(tokens))
	}
	if got := tokens[0].GetListNames(); len(got) != 1 || got[0] != "Bike" {
		t.Errorf("listNames = %v, want the List their own token names", got)
	}
}

// Somebody whose key has been stopped by an Admin has to be told: the alternative is a
// script that fails silently and a Member who does not know why.
func TestRevokingSomebodyElsesTokenTellsThem(t *testing.T) {
	f := newListFixture(t)
	uid := f.createList(t, f.jonas, "Bike")

	svc := f.tokens(t).WithActivity(NewTokenActivity(f.store, func() time.Time { return testClock }, nil))
	made, err := svc.CreateAccessToken(f.as(t, f.jonas), connect.NewRequest(
		&apiv1.CreateAccessTokenRequest{
			Name: "Shortcut", ListUids: []string{uid},
			Abilities: &apiv1.TokenAbilities{Read: true},
		},
	))
	if err != nil {
		t.Fatalf("CreateAccessToken: %v", err)
	}

	_, err = svc.RevokeAccessToken(f.as(t, f.anna), connect.NewRequest(
		&apiv1.RevokeAccessTokenRequest{TokenUid: made.Msg.GetToken().GetUid()},
	))
	if err != nil {
		t.Fatalf("RevokeAccessToken as an Admin: %v", err)
	}

	entries, err := f.store.ActivityFor(t.Context(), f.jonas.ID)
	if err != nil {
		t.Fatalf("ActivityFor: %v", err)
	}
	if len(entries) != 1 || entries[0].Kind != store.ActivityTokenUsed {
		t.Fatalf("got %d entries, want one about the revoked token", len(entries))
	}
	if !strings.Contains(entries[0].Text, "Shortcut") {
		t.Errorf("text = %q, want it to name the token", entries[0].Text)
	}
}

// Revoking your own key is not news. Telling you about it would be a panel that fills
// up with things you just did.
func TestRevokingYourOwnTokenTellsNobody(t *testing.T) {
	f := newListFixture(t)
	uid := f.createList(t, f.jonas, "Bike")

	svc := f.tokens(t).WithActivity(NewTokenActivity(f.store, func() time.Time { return testClock }, nil))
	made, err := svc.CreateAccessToken(f.as(t, f.jonas), connect.NewRequest(
		&apiv1.CreateAccessTokenRequest{
			Name: "Shortcut", ListUids: []string{uid},
			Abilities: &apiv1.TokenAbilities{Read: true},
		},
	))
	if err != nil {
		t.Fatalf("CreateAccessToken: %v", err)
	}

	_, err = svc.RevokeAccessToken(f.as(t, f.jonas), connect.NewRequest(
		&apiv1.RevokeAccessTokenRequest{TokenUid: made.Msg.GetToken().GetUid()},
	))
	if err != nil {
		t.Fatalf("RevokeAccessToken: %v", err)
	}

	entries, err := f.store.ActivityFor(t.Context(), f.jonas.ID)
	if err != nil {
		t.Fatalf("ActivityFor: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("got %d entries, want none for revoking your own token", len(entries))
	}
}

// Reaching everything and naming every List are different answers: the first keeps
// working when a List is made next week, the second deliberately does not.
func TestATokenForAllListsReachesOnesMadeLater(t *testing.T) {
	f := newListFixture(t)
	f.createList(t, f.anna, "Groceries")

	svc := NewTokenService(f.store, func() time.Time { return testClock }, nil, nil)
	made, err := svc.CreateAccessToken(f.as(t, f.anna), connect.NewRequest(
		&apiv1.CreateAccessTokenRequest{
			Name: "My client", AllLists: true,
			Abilities: &apiv1.TokenAbilities{Read: true, Write: true, Delete: true},
		},
	))
	if err != nil {
		t.Fatalf("CreateAccessToken for all lists: %v", err)
	}
	if !made.Msg.GetToken().GetAllLists() {
		t.Error("the token does not say it reaches everything")
	}

	// Made after the token was cut, which is the whole point.
	later := f.createList(t, f.anna, "Bike")

	resolver := auth.NewResolver(f.store, func() time.Time { return testClock })
	grant, ok := resolver.Grant(t.Context(), bearer(made.Msg.GetSecret()))
	if !ok {
		t.Fatal("the token was not accepted")
	}

	if _, err := f.svc.GetList(auth.WithGrant(t.Context(), grant), connect.NewRequest(
		&apiv1.GetListRequest{ListUid: later},
	)); err != nil {
		t.Errorf("GetList on a List made after the token: %v", err)
	}
}

// A token that reaches nothing is a key to no door.
func TestATokenMustReachSomething(t *testing.T) {
	f := newListFixture(t)

	_, err := f.tokens(t).CreateAccessToken(f.as(t, f.anna), connect.NewRequest(
		&apiv1.CreateAccessTokenRequest{Name: "Nothing"},
	))
	if got := connect.CodeOf(err); got != connect.CodeInvalidArgument {
		t.Errorf("code = %v, want invalid_argument", got)
	}
}
