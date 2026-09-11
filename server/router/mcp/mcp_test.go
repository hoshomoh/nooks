package mcp

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/hoshomoh/nooks/server/auth"
	v1 "github.com/hoshomoh/nooks/server/router/api/v1"
	"github.com/hoshomoh/nooks/store"
)

var testClock = time.Date(2026, 8, 25, 9, 0, 0, 0, time.UTC)

// instance is an Instance with one Member, one List, and an MCP server over it.
type instance struct {
	t       *testing.T
	store   store.Store
	server  *httptest.Server
	member  store.Member
	listID  int64
	listUID string
}

func newInstance(t *testing.T) *instance {
	t.Helper()
	s, err := store.OpenSQLite(t.Context(), filepath.Join(t.TempDir(), "nooks.db"))
	if err != nil {
		t.Fatalf("OpenSQLite: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })

	member, err := s.CreateMember(t.Context(), store.CreateMemberParams{
		UID: "mem_anna", Name: "Anna", Email: "anna@brunnen.lan", Role: store.RoleAdmin,
		PasswordHash: "hash", CreatedAt: testClock,
	})
	if err != nil {
		t.Fatalf("CreateMember: %v", err)
	}
	list, err := s.CreateList(t.Context(), store.CreateListParams{
		UID: "list_groceries", Name: "Groceries", OwnerID: member.ID, At: testClock,
	})
	if err != nil {
		t.Fatalf("CreateList: %v", err)
	}

	now := func() time.Time { return testClock }
	handler := Handler(v1.NewListService(s, now, nil), auth.NewResolver(s, now))
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	return &instance{t: t, store: s, server: server, member: member, listID: list.ID, listUID: list.UID}
}

// tokenFor cuts an Access token with the given abilities and answers the secret.
func (i *instance) tokenFor(abilities store.TokenAbilities) string {
	i.t.Helper()
	secret, hash, err := auth.NewToken()
	if err != nil {
		i.t.Fatalf("NewToken: %v", err)
	}
	if _, err := i.store.CreateAccessToken(i.t.Context(), store.CreateAccessTokenParams{
		UID: "tok_1", MemberID: i.member.ID, Name: "Assistant", TokenHash: hash,
		Abilities: abilities, ListIDs: []int64{i.listID}, At: testClock,
	}); err != nil {
		i.t.Fatalf("CreateAccessToken: %v", err)
	}
	return secret
}

// connect opens an MCP session holding a token, the way an assistant would.
func (i *instance) connect(secret string) *sdk.ClientSession {
	i.t.Helper()
	transport := &sdk.StreamableClientTransport{
		Endpoint:   i.server.URL,
		HTTPClient: &http.Client{Transport: bearer{secret: secret}},
	}
	client := sdk.NewClient(&sdk.Implementation{Name: "test"}, nil)
	session, err := client.Connect(i.t.Context(), transport, nil)
	if err != nil {
		i.t.Fatalf("connect: %v", err)
	}
	i.t.Cleanup(func() { _ = session.Close() })
	return session
}

// bearer puts the token on every request, which is what an MCP client is configured with.
type bearer struct{ secret string }

func (b bearer) RoundTrip(r *http.Request) (*http.Response, error) {
	r.Header.Set("Authorization", "Bearer "+b.secret)
	return http.DefaultTransport.RoundTrip(r)
}

// said flattens a tool's answer so a test can read it.
func said(res *sdk.CallToolResult) string {
	var out strings.Builder
	for _, content := range res.Content {
		if block, ok := content.(*sdk.TextContent); ok {
			out.WriteString(block.Text)
		}
	}
	return out.String()
}

func TestAnAssistantSeesTheToolsItCanUse(t *testing.T) {
	i := newInstance(t)
	session := i.connect(i.tokenFor(store.TokenAbilities{Read: true}))

	tools, err := session.ListTools(t.Context(), nil)
	if err != nil {
		t.Fatalf("ListTools: %v", err)
	}

	want := map[string]bool{
		"list_lists": false, "get_list": false, "add_item": false,
		"complete_item": false, "search": false,
	}
	for _, tool := range tools.Tools {
		want[tool.Name] = true
	}
	for name, found := range want {
		if !found {
			t.Errorf("tool %q was not offered", name)
		}
	}
}

func TestAnAssistantReadsAList(t *testing.T) {
	i := newInstance(t)
	session := i.connect(i.tokenFor(store.TokenAbilities{Read: true}))

	res, err := session.CallTool(t.Context(), &sdk.CallToolParams{Name: "list_lists"})
	if err != nil {
		t.Fatalf("CallTool: %v", err)
	}
	if !strings.Contains(said(res), "Groceries") {
		t.Errorf("answer = %q, want the List it can reach", said(res))
	}
}

// The property the whole design rests on: the same rules, whichever door.
func TestAReadOnlyTokenCannotAddOverMCP(t *testing.T) {
	i := newInstance(t)
	session := i.connect(i.tokenFor(store.TokenAbilities{Read: true}))

	res, err := session.CallTool(t.Context(), &sdk.CallToolParams{
		Name:      "add_item",
		Arguments: map[string]any{"list_uid": i.listUID, "label": "Milk"},
	})
	if err != nil {
		t.Fatalf("CallTool: %v", err)
	}
	if !res.IsError {
		t.Fatalf("adding succeeded with a read-only token: %q", said(res))
	}
}

func TestAWriteTokenAddsOverMCP(t *testing.T) {
	i := newInstance(t)
	session := i.connect(i.tokenFor(store.TokenAbilities{Read: true, Write: true}))

	res, err := session.CallTool(t.Context(), &sdk.CallToolParams{
		Name:      "add_item",
		Arguments: map[string]any{"list_uid": i.listUID, "label": "Milk"},
	})
	if err != nil {
		t.Fatalf("CallTool: %v", err)
	}
	if res.IsError {
		t.Fatalf("adding failed with a write token: %q", said(res))
	}
	if !strings.Contains(said(res), "Milk") {
		t.Errorf("answer = %q, want it to name what it added", said(res))
	}
}

// An assistant should be told once that it is not signed in, not once per tool call.
func TestNoTokenIsRefusedBeforeTheProtocolStarts(t *testing.T) {
	i := newInstance(t)

	res, err := http.Post(i.server.URL, "application/json", strings.NewReader("{}"))
	if err != nil {
		t.Fatalf("post: %v", err)
	}
	defer func() { _ = res.Body.Close() }()

	if res.StatusCode != http.StatusUnauthorized {
		t.Errorf("code = %d, want 401", res.StatusCode)
	}
}
