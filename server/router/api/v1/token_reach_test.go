package v1

import (
	"path/filepath"
	"testing"
	"time"

	"connectrpc.com/connect"

	"github.com/hoshomoh/nooks/server/auth"
	"github.com/hoshomoh/nooks/store"
)

/*
An Access token reaches Lists and nothing else.

requireBrowser is one line inside a method body, and a method that forgets it is wrong
in a way nothing shows: the request works, and it works for a key somebody pasted into
a shell script. anonymous_test.go makes adding an RPC a decision about whether a
stranger may reach it; this makes it a decision about whether a leaked key may.

The token below belongs to an Admin and was cut to do everything a token can, so the
only thing left that can refuse it is being a token. An RPC missing from the list below
fails until somebody says which side of the line it is on.
*/
var reachableByToken = map[string]string{
	// The Lists and Items: what a token is for.
	"ListService.CreateList":      "an assistant makes lists",
	"ListService.CreateItem":      "and puts things on them",
	"ListService.UpdateItem":      "and edits them",
	"ListService.DeleteItem":      "if it was cut to delete",
	"ListService.MoveItem":        "reordering is an edit",
	"ListService.SetItemDone":     "ticking is the most common thing an assistant does",
	"ListService.GetItem":         "reading one",
	"ListService.GetList":         "reading a List",
	"ListService.ListLists":       "finding one",
	"ListService.GetSidebar":      "the same grouping the app shows",
	"ListService.Search":          "finding an Item by what it says",
	"ListService.ListDatedItems":  "what is due",
	"ListService.RenameList":      "an edit to a List it reaches",
	"ListService.DeleteList":      "if it was cut to delete",
	"ListService.DuplicateList":   "a List made from one it reaches",
	"ListService.SetListPinned":   "an edit to a List it reaches",
	"ListService.SetListArchived": "putting a finished List away",
	"ListService.RestoreList":     "bringing back one its Member deleted, which is a write",
	"ListService.SetListSharing":  "sharing a List is about a List, not an account",
	"ListService.GetListShares":   "reading who a List reaches",

	// What is going on, and who is here.
	"ActivityService.ListActivity":     "what is waiting, narrowed to the Lists it reaches",
	"ActivityService.MarkActivityRead": "clearing the count it just read",
	"MemberService.ListMembers":        "names, to say who added what and to share with",
	"MemberService.ListGroups":         "the same, for sharing with several at once",
	"AuthService.GetCurrentMember":     "whose key this is",
	"InstanceService.GetInstance":      "the name of the place",
	"InstanceService.GetInstanceAbout": "version and size, which the about tool reports",

	// Open to anyone, so open to a token as well. anonymous_test.go is where these are
	// decided; a token presenting one is no further in than a stranger.
	"AuthService.SignIn":                "answers a stranger by design",
	"AuthService.SignOut":               "acts on the cookie in the request, which a token has none of",
	"AuthService.RefreshAccess":         "the refresh cookie is the credential",
	"AuthService.CompleteSetup":         "first run",
	"AuthService.RequestJoin":           "a stranger asking for an account",
	"AuthService.GetJoinRequest":        "read by its own secret",
	"AuthService.CompleteJoin":          "accepting one",
	"AuthService.RequestPasswordReset":  "somebody locked out",
	"AuthService.GetResetRequest":       "read by its own secret",
	"AuthService.CompletePasswordReset": "finishing one",
	"PublicService.GetPublicList":       "published to the world",
}

func TestNoRPCAnswersALeakedKeyByAccident(t *testing.T) {
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
	token := store.AccessToken{
		ID: 1, UID: "tok_1", MemberID: anna.ID, Name: "the leaked one",
		Abilities: store.TokenAbilities{Read: true, Write: true, Delete: true},
		AllLists:  true, CreatedAt: testClock,
	}
	ctx := auth.WithGrant(t.Context(), auth.NewTokenGrant(anna, token, nil))

	services := Services{
		Activity: NewActivityService(s, nil),
		Auth:     NewAuthService(s, AuthServiceOptions{}),
		Instance: NewInstanceService(s),
		List:     NewListService(s, now, nil),
		Member:   NewMemberService(s, now, nil),
		Public:   NewPublicService(s),
		Request:  NewRequestService(s, now),
		Token:    NewTokenService(s, now, nil, nil),
	}

	for _, svc := range servicesIn(services) {
		for _, call := range rpcsOn(svc) {
			t.Run(call.name, func(t *testing.T) {
				err := call.refuse(ctx)

				if _, reachable := reachableByToken[call.name]; reachable {
					return
				}
				if err == nil {
					t.Fatalf("%s answered an access token", call.name)
				}
				if got := connect.CodeOf(err); got != connect.CodePermissionDenied {
					t.Errorf("%s refused an access token with %v, want permission_denied",
						call.name, got)
				}
			})
		}
	}
}
