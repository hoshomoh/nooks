package v1

import (
	"net/http"
	"strings"
	"testing"
	"time"

	"connectrpc.com/connect"

	apiv1 "github.com/hoshomoh/nooks/proto/gen/nooks/api/v1"
	"github.com/hoshomoh/nooks/server/auth"
	"github.com/hoshomoh/nooks/store"
)

// bearer is the header a script presents a token in.
func bearer(secret string) http.Header {
	return http.Header{"Authorization": []string{"Bearer " + secret}}
}

// tokenActivityOf builds the recorder over the fixture's store, on the test clock.
func (f listFixture) tokenActivityOf() *TokenActivity {
	return NewTokenActivity(f.store, func() time.Time { return testClock }, nil)
}

// The resolver is the only thing that sees a token presented, so it is what has to
// notice the first time one is used.
func TestTheFirstUseOfATokenIsRecordedOnce(t *testing.T) {
	f := newListFixture(t)
	uid := f.createList(t, f.jonas, "Bike")

	// Real secrets, because the resolver hashes what it is handed: the fixture's
	// predictable pair would never match.
	svc := NewTokenService(f.store, func() time.Time { return testClock }, nil, nil)
	made, err := svc.CreateAccessToken(f.as(t, f.jonas), connect.NewRequest(
		&apiv1.CreateAccessTokenRequest{
			Name: "Kitchen tablet", ListUids: []string{uid},
			Permission: apiv1.Permission_PERMISSION_READ,
		},
	))
	if err != nil {
		t.Fatalf("CreateAccessToken: %v", err)
	}

	resolver := auth.NewResolver(f.store, func() time.Time { return testClock }).
		WithTokenWatcher(f.tokenActivityOf())
	secret := made.Msg.GetSecret()

	for range 2 {
		if _, ok := resolver.Grant(t.Context(), bearer(secret)); !ok {
			t.Fatal("the token was not accepted")
		}
	}

	entries, err := f.store.ActivityFor(t.Context(), f.jonas.ID)
	if err != nil {
		t.Fatalf("ActivityFor: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("got %d entries, want exactly one for the first use", len(entries))
	}
	if !strings.Contains(entries[0].Text, "Kitchen tablet") {
		t.Errorf("text = %q, want it to name the token", entries[0].Text)
	}
	if entries[0].Kind != store.ActivityTokenUsed {
		t.Errorf("kind = %q, want %q", entries[0].Kind, store.ActivityTokenUsed)
	}
}
