package store

import (
	"errors"
	"testing"
	"time"
)

// cutToken makes a token for a Member, scoped to some Lists.
func cutToken(t *testing.T, s Store, member Member, listIDs []int64, abilities TokenAbilities) AccessToken {
	t.Helper()
	token, err := s.CreateAccessToken(t.Context(), CreateAccessTokenParams{
		UID: "tok_kitchen", MemberID: member.ID, Name: "Kitchen tablet",
		TokenHash: "hash-of-the-token", Abilities: abilities,
		ListIDs: listIDs, At: createdAt,
	})
	if err != nil {
		t.Fatalf("CreateAccessToken: %v", err)
	}
	return token
}

func TestCuttingAToken(t *testing.T) {
	for _, d := range drivers() {
		t.Run(d.name, func(t *testing.T) {
			s := d.open(t)
			list, owner := newList(t, s, "Groceries", SharingPrivate)

			token := cutToken(t, s, owner, []int64{list.ID}, TokenAbilities{Read: true, Write: true, Delete: true})

			if token.ID == 0 {
				t.Error("ID = 0, want the database to have assigned one")
			}
			if !token.Abilities.Write || !token.Abilities.Delete {
				t.Errorf("abilities = %+v, want the ones it was cut with", token.Abilities)
			}
			if !token.LastUsedAt.IsZero() {
				t.Error("a token that has never been used has a last-used time")
			}
		})
	}
}

// A token is found by what was presented, never by what was stored in clear — there is
// nothing stored in clear.
func TestFindingATokenByItsHash(t *testing.T) {
	for _, d := range drivers() {
		t.Run(d.name, func(t *testing.T) {
			s := d.open(t)
			_, owner := newList(t, s, "Groceries", SharingPrivate)
			cutToken(t, s, owner, nil, TokenAbilities{Read: true})

			found, err := s.AccessTokenByHash(t.Context(), "hash-of-the-token")
			if err != nil {
				t.Fatalf("AccessTokenByHash: %v", err)
			}
			if found.Name != "Kitchen tablet" {
				t.Errorf("name = %q", found.Name)
			}
		})
	}
}

func TestATokenThatIsNotThere(t *testing.T) {
	for _, d := range drivers() {
		t.Run(d.name, func(t *testing.T) {
			_, err := d.open(t).AccessTokenByHash(t.Context(), "nothing-like-it")
			if !errors.Is(err, ErrNotFound) {
				t.Errorf("err = %v, want ErrNotFound", err)
			}
		})
	}
}

// A List that is not named is invisible to the token: not forbidden, invisible.
func TestWhichListsATokenReaches(t *testing.T) {
	for _, d := range drivers() {
		t.Run(d.name, func(t *testing.T) {
			s := d.open(t)
			reachable, owner := newList(t, s, "Groceries", SharingPrivate)
			if _, err := s.CreateList(t.Context(), CreateListParams{
				UID: "l_bike", Name: "Bike", OwnerID: owner.ID,
				Sharing: SharingPrivate, CanEdit: true, At: createdAt,
			}); err != nil {
				t.Fatalf("CreateList: %v", err)
			}

			token := cutToken(t, s, owner, []int64{reachable.ID}, TokenAbilities{Read: true})

			ids, err := s.TokenListIDs(t.Context(), token.ID)
			if err != nil {
				t.Fatalf("TokenListIDs: %v", err)
			}
			if len(ids) != 1 || ids[0] != reachable.ID {
				t.Errorf("reaches %v, want only the List it names", ids)
			}
		})
	}
}

func TestATokenExpires(t *testing.T) {
	never := AccessToken{}
	if never.Expired(createdAt) {
		t.Error("a token with no expiry has expired")
	}

	expiring := AccessToken{ExpiresAt: createdAt.Add(time.Hour)}
	if expiring.Expired(createdAt) {
		t.Error("expired before its expiry")
	}
	if !expiring.Expired(createdAt.Add(2 * time.Hour)) {
		t.Error("did not expire after its expiry")
	}
}

// A Member should be able to tell which of their tokens is doing nothing.
func TestRecordingThatATokenWasUsed(t *testing.T) {
	for _, d := range drivers() {
		t.Run(d.name, func(t *testing.T) {
			s := d.open(t)
			_, owner := newList(t, s, "Groceries", SharingPrivate)
			token := cutToken(t, s, owner, nil, TokenAbilities{Read: true})

			used := createdAt.Add(time.Hour)
			if err := s.MarkTokenUsed(t.Context(), token.ID, used); err != nil {
				t.Fatalf("MarkTokenUsed: %v", err)
			}

			found, err := s.AccessTokenByHash(t.Context(), "hash-of-the-token")
			if err != nil {
				t.Fatalf("AccessTokenByHash: %v", err)
			}
			if !found.LastUsedAt.Equal(used) {
				t.Errorf("last used = %v, want %v", found.LastUsedAt, used)
			}
		})
	}
}

func TestRevokingAToken(t *testing.T) {
	for _, d := range drivers() {
		t.Run(d.name, func(t *testing.T) {
			s := d.open(t)
			_, owner := newList(t, s, "Groceries", SharingPrivate)
			token := cutToken(t, s, owner, nil, TokenAbilities{Read: true})

			if err := s.DeleteAccessToken(t.Context(), token.ID); err != nil {
				t.Fatalf("DeleteAccessToken: %v", err)
			}

			if _, err := s.AccessTokenByHash(t.Context(), "hash-of-the-token"); !errors.Is(err, ErrNotFound) {
				t.Errorf("a revoked token still answers: %v", err)
			}
		})
	}
}

// Somebody's tokens are their own: the page shows what they cut, not what anyone else
// did.
func TestAMembersOwnTokens(t *testing.T) {
	for _, d := range drivers() {
		t.Run(d.name, func(t *testing.T) {
			s := d.open(t)
			_, anna := newList(t, s, "Groceries", SharingPrivate)
			jonas := addMember(t, s, "mem_jonas", "Jonas", "jonas@brunnen.lan")
			cutToken(t, s, anna, nil, TokenAbilities{Read: true})

			mine, err := s.AccessTokensFor(t.Context(), jonas.ID)
			if err != nil {
				t.Fatalf("AccessTokensFor: %v", err)
			}
			if len(mine) != 0 {
				t.Errorf("got %d tokens, want none of somebody else's", len(mine))
			}
		})
	}
}
