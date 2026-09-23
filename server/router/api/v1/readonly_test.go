package v1

import (
	"testing"
	"time"

	"connectrpc.com/connect"

	apiv1 "github.com/hoshomoh/nooks/proto/gen/nooks/api/v1"
	"github.com/hoshomoh/nooks/server/auth"
	"github.com/hoshomoh/nooks/store"
	"github.com/hoshomoh/nooks/store/storetest"
)

/*
A token cut to read cannot change anything.

Most changes are to a List, so accessTo asks this on the way in and there is nothing
left to check. These two are the changes that have no List to reach — making one, and
clearing the unread count — so they are the two that can miss it, and both did.

Asserted by the code rather than by the fact of a refusal: an empty request is refused
by nearly everything for some other reason, and a test satisfied by that would pass
just as happily with the check gone.
*/
func TestAReadOnlyTokenCannotChangeAnything(t *testing.T) {
	s := storetest.Fresh(t)

	now := func() time.Time { return testClock }
	anna, err := s.CreateMember(t.Context(), store.CreateMemberParams{
		UID: "mem_anna", Name: "Anna", Email: "anna@brunnen.lan", Role: store.RoleAdmin,
		PasswordHash: "hash", CreatedAt: testClock,
	})
	if err != nil {
		t.Fatalf("CreateMember: %v", err)
	}

	reading := store.AccessToken{
		ID: 1, UID: "tok_read", MemberID: anna.ID, Name: "read only",
		Abilities: store.TokenAbilities{Read: true}, AllLists: true, CreatedAt: testClock,
	}
	writing := reading
	writing.ID, writing.UID = 2, "tok_write"
	writing.Abilities.Write = true

	lists := NewListService(s, now, func() (string, error) { return "lst_1", nil })
	activity := NewActivityService(s, nil)

	changes := map[string]func(grant auth.Grant) error{
		"CreateList": func(grant auth.Grant) error {
			ctx := auth.WithGrant(t.Context(), grant)
			_, err := lists.CreateList(ctx, connect.NewRequest(&apiv1.CreateListRequest{Name: "Bread"}))
			return err
		},
		"MarkActivityRead": func(grant auth.Grant) error {
			ctx := auth.WithGrant(t.Context(), grant)
			_, err := activity.MarkActivityRead(ctx, connect.NewRequest(&apiv1.MarkActivityReadRequest{}))
			return err
		},
	}

	for name, change := range changes {
		t.Run(name, func(t *testing.T) {
			err := change(auth.NewTokenGrant(anna, reading, nil))
			if err == nil {
				t.Fatalf("%s let a read-only token change something", name)
			}
			if got := connect.CodeOf(err); got != connect.CodePermissionDenied {
				t.Errorf("%s refused with %v, want permission_denied", name, got)
			}

			// The same call with a token that may write, so the refusal above is the
			// ability and not the request.
			if err := change(auth.NewTokenGrant(anna, writing, nil)); err != nil {
				t.Errorf("%s refused a token that may write: %v", name, err)
			}
		})
	}
}

/*
The other half of that rule is not held: a token cut with read off reads anyway.

`read` is accepted by CreateAccessToken, written to `can_read`, handed back in the
token's abilities so the settings screen draws it as chosen, and described in the
published API reference as "see lists and items". Nothing consults it. There is
MayWrite and MayDelete on Grant and no MayRead, and Abilities.Read is read nowhere
outside the row it is stored in.

So a Member who cuts a token that may add but not look gets one that looks. The token
still cannot exceed its Member, which is why this is not an escalation — but it exceeds
what the Member asked of it, and the three-way model exists precisely so that asking is
worth something.

This test asserts the hole rather than the rule, because closing it is a decision
about whether a write-only token is a thing Nooks offers at all: Usable() is
`Read || Write`, so today it is offered. Whoever decides comes here, and the guard
worth adding with the fix is that every field of TokenAbilities is consulted by some
Grant method — which is what would have caught this.
*/
func TestTheReadAbilityDecidesNothing(t *testing.T) {
	f := newListFixture(t)
	uid := f.createList(t, f.anna, "Shopping")

	token := store.AccessToken{
		ID: 1, UID: "tok_write_only", MemberID: f.anna.ID, Name: "may add, not look",
		Abilities: store.TokenAbilities{Read: false, Write: true},
		AllLists:  true, CreatedAt: testClock,
	}
	ctx := auth.WithGrant(t.Context(), auth.NewTokenGrant(f.anna, token, nil))

	res, err := f.svc.GetList(ctx, connect.NewRequest(&apiv1.GetListRequest{ListUid: uid}))
	if err != nil {
		t.Fatalf("read is enforced now, so this test has outlived the hole it records — delete it and say what a write-only token may do: %v", err)
	}
	if res.Msg.GetList().GetName() != "Shopping" {
		t.Errorf("read the List as %q, want Shopping", res.Msg.GetList().GetName())
	}
}
