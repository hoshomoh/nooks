package v1

import (
	"path/filepath"
	"strings"
	"testing"
	"time"

	"connectrpc.com/connect"

	apiv1 "github.com/hoshomoh/nooks/proto/gen/nooks/api/v1"
	"github.com/hoshomoh/nooks/server/auth"
	"github.com/hoshomoh/nooks/store"
	"google.golang.org/protobuf/encoding/prototext"
)

// theHash is distinctive so that a substring search cannot match anything else.
const theHash = "$2a$10$ThisIsTheStoredBcryptHashAndItMustNotTravel"

// TestListMembersNeverCarriesAStoredHash holds memberToProto to what it claims.
//
// Every Member can call ListMembers — the share dialog needs it — so a Member message
// that grew a hash field would put everybody's credential in front of the whole
// household. The conversion is right today and nothing checks it.
//
// Marshalled and searched rather than compared field by field, because the mistake this
// catches is a field nobody thought about. A test that named the fields it knew would
// pass straight through one.
func TestListMembersNeverCarriesAStoredHash(t *testing.T) {
	s, err := store.OpenSQLite(t.Context(), filepath.Join(t.TempDir(), "nooks.db"))
	if err != nil {
		t.Fatalf("OpenSQLite: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })

	_, err = s.CreateMember(t.Context(), store.CreateMemberParams{
		UID: "mem_anna", Name: "Anna", Email: "anna@brunnen.lan",
		Role: store.RoleAdmin, PasswordHash: theHash, CreatedAt: testClock,
	})
	if err != nil {
		t.Fatalf("CreateMember: %v", err)
	}
	jonas, err := s.CreateMember(t.Context(), store.CreateMemberParams{
		UID: "mem_jonas", Name: "Jonas", Email: "jonas@brunnen.lan",
		Role: store.RoleMember, PasswordHash: theHash, CreatedAt: testClock,
	})
	if err != nil {
		t.Fatalf("CreateMember: %v", err)
	}

	svc := NewMemberService(s, func() time.Time { return testClock }, func() (string, error) {
		return "grp-1", nil
	})

	// Read as the plain Member rather than the Admin: the weaker grant is the one the
	// whole household holds.
	res, err := svc.ListMembers(auth.WithMember(t.Context(), jonas), connect.NewRequest(&apiv1.ListMembersRequest{}))
	if err != nil {
		t.Fatalf("ListMembers: %v", err)
	}
	if len(res.Msg.GetMembers()) != 2 {
		t.Fatalf("listed %d Members, want 2", len(res.Msg.GetMembers()))
	}

	if written := prototext.Format(res.Msg); strings.Contains(written, "ThisIsTheStoredBcryptHash") {
		t.Errorf("a listed Member carries the stored hash:\n%s", written)
	}
}
