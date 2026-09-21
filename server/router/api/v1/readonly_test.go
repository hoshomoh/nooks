package v1

import (
	"path/filepath"
	"testing"
	"time"

	"connectrpc.com/connect"

	apiv1 "github.com/hoshomoh/nooks/proto/gen/nooks/api/v1"
	"github.com/hoshomoh/nooks/server/auth"
	"github.com/hoshomoh/nooks/store"
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
	s, err := store.OpenSQLite(t.Context(), filepath.Join(t.TempDir(), "nooks.db"))
	if err != nil {
		t.Fatalf("OpenSQLite: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })

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
