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
	handler := Handler(testServices(s, now), auth.NewResolver(s, now))
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

// testServices builds the same set server.go does, so a tool that reaches past
// ListService is exercised here rather than only in production.
func testServices(s store.Store, now func() time.Time) v1.Services {
	return v1.Services{
		Activity: v1.NewActivityService(s, nil),
		Auth:     v1.NewAuthService(s, v1.AuthServiceOptions{}),
		Instance: v1.NewInstanceService(s),
		List:     v1.NewListService(s, now, nil),
		Member:   v1.NewMemberService(s, now, nil),
		Public:   v1.NewPublicService(s),
		Request:  v1.NewRequestService(s, now),
		Token:    v1.NewTokenService(s, now, nil, nil),
	}
}

// addMilk puts one Item on the List and answers its identifier, so a test about
// changing or removing an Item does not have to reach past the door to set itself up.
func (i *instance) addMilk(session *sdk.ClientSession) string {
	i.t.Helper()
	res, err := session.CallTool(i.t.Context(), &sdk.CallToolParams{
		Name:      "add_item",
		Arguments: map[string]any{"list_uid": i.listUID, "label": "Milk"},
	})
	if err != nil {
		i.t.Fatalf("CallTool: %v", err)
	}
	if res.IsError {
		i.t.Fatalf("add_item: %s", said(res))
	}
	_, uid, found := strings.Cut(said(res), "— ")
	if !found {
		i.t.Fatalf("add_item said %q, want an identifier after an em dash", said(res))
	}
	return strings.TrimSpace(uid)
}

// Deleting is its own ability, so a token that may write is still refused it. The same
// property as the read-only case, one door further in.
func TestAWriteTokenCannotDeleteOverMCP(t *testing.T) {
	i := newInstance(t)
	session := i.connect(i.tokenFor(store.TokenAbilities{Read: true, Write: true}))
	itemUID := i.addMilk(session)

	res, err := session.CallTool(t.Context(), &sdk.CallToolParams{
		Name:      "delete_item",
		Arguments: map[string]any{"item_uid": itemUID},
	})
	if err != nil {
		t.Fatalf("CallTool: %v", err)
	}
	if !res.IsError {
		t.Fatalf("deleting succeeded without the delete ability: %q", said(res))
	}
}

func TestADeleteTokenDeletesOverMCP(t *testing.T) {
	i := newInstance(t)
	session := i.connect(i.tokenFor(store.TokenAbilities{Read: true, Write: true, Delete: true}))
	itemUID := i.addMilk(session)

	res, err := session.CallTool(t.Context(), &sdk.CallToolParams{
		Name:      "delete_item",
		Arguments: map[string]any{"item_uid": itemUID},
	})
	if err != nil {
		t.Fatalf("CallTool: %v", err)
	}
	if res.IsError {
		t.Fatalf("delete_item: %s", said(res))
	}
}

// An omitted field is not an empty one: changing a quantity must leave the wording
// where it was, or an assistant has to restate the whole Item to touch any of it.
func TestUpdateItemLeavesOmittedFieldsAlone(t *testing.T) {
	i := newInstance(t)
	session := i.connect(i.tokenFor(store.TokenAbilities{Read: true, Write: true}))
	itemUID := i.addMilk(session)

	res, err := session.CallTool(t.Context(), &sdk.CallToolParams{
		Name:      "update_item",
		Arguments: map[string]any{"item_uid": itemUID, "quantity": "2 pints"},
	})
	if err != nil {
		t.Fatalf("CallTool: %v", err)
	}
	if res.IsError {
		t.Fatalf("update_item: %s", said(res))
	}
	if answer := said(res); !strings.Contains(answer, "Milk") || !strings.Contains(answer, "2 pints") {
		t.Errorf("answer = %q, want the old label and the new quantity", answer)
	}
}

// The tools reach past ListService, which is the whole point of handing the package
// every service rather than one.
func TestAnAssistantReachesBeyondLists(t *testing.T) {
	i := newInstance(t)
	session := i.connect(i.tokenFor(store.TokenAbilities{Read: true}))

	for _, tool := range []string{"whoami", "list_members", "about_instance"} {
		res, err := session.CallTool(t.Context(), &sdk.CallToolParams{Name: tool})
		if err != nil {
			t.Fatalf("CallTool %s: %v", tool, err)
		}
		if res.IsError {
			t.Errorf("%s: %s", tool, said(res))
		}
		if !strings.Contains(said(res), "Anna") && tool != "about_instance" {
			t.Errorf("%s said %q, want the signed-in Member", tool, said(res))
		}
	}
}

// withItem puts one Item, with a Note, on the fixture's List and answers its uid.
func (i *instance) withItem(label, note string) string {
	i.t.Helper()
	item, err := i.store.CreateItem(i.t.Context(), store.CreateItemParams{
		UID: "item_" + label, ListID: i.listID, Label: label,
		Note: note, AddedByID: i.member.ID, At: testClock,
	})
	if err != nil {
		i.t.Fatalf("CreateItem: %v", err)
	}
	return item.UID
}

/*
Every tool that shows a thing shows what to call about it next.

The shape of a whole family of bugs: a renderer holding the identifier and printing the
words without it, which leaves an assistant able to find something and unable to touch
it. Driven through a real session rather than read off the source, because what reaches
the assistant is the thing that matters.
*/
func TestEveryToolSaysWhatToCallNext(t *testing.T) {
	i := newInstance(t)
	itemUID := i.withItem("Coffee", "### Where\nSaturday market.")
	session := i.connect(i.tokenFor(store.TokenAbilities{Read: true}))

	for _, one := range []struct {
		tool string
		args map[string]any
		want []string
	}{
		{tool: "list_lists", want: []string{i.listUID}},
		{tool: "get_list", args: map[string]any{"list_uid": i.listUID}, want: []string{i.listUID, itemUID}},
		{tool: "get_item", args: map[string]any{"item_uid": itemUID}, want: []string{itemUID, i.listUID}},
		{tool: "search", args: map[string]any{"query": "Coffee"}, want: []string{itemUID, i.listUID}},
		{tool: "list_members", want: []string{i.member.UID}},
		{tool: "whoami", want: []string{i.member.UID}},
	} {
		t.Run(one.tool, func(t *testing.T) {
			res, err := session.CallTool(t.Context(), &sdk.CallToolParams{
				Name: one.tool, Arguments: one.args,
			})
			if err != nil {
				t.Fatalf("CallTool %s: %v", one.tool, err)
			}
			for _, want := range one.want {
				if !strings.Contains(said(res), want) {
					t.Errorf("%s answered %q, want it to carry %q", one.tool, said(res), want)
				}
			}
		})
	}
}

/*
A Note is read whole by asking for it, and never by accident.

get_list shortens one to keep a row a row, so the two have to disagree: if get_list ever
starts printing Notes in full it is back to dumping a document into a menu, and if
get_item ever stops there is nowhere left to read one.
*/
func TestOnlyGetItemGivesAWholeNote(t *testing.T) {
	i := newInstance(t)
	const note = "### Where\nSaturday market.\nThey pack up around two."
	itemUID := i.withItem("Coffee", note)
	session := i.connect(i.tokenFor(store.TokenAbilities{Read: true}))

	listed, err := session.CallTool(t.Context(), &sdk.CallToolParams{
		Name: "get_list", Arguments: map[string]any{"list_uid": i.listUID},
	})
	if err != nil {
		t.Fatalf("get_list: %v", err)
	}
	if strings.Contains(said(listed), "They pack up around two") {
		t.Errorf("get_list said %q, want the Note kept to its row", said(listed))
	}
	if !strings.Contains(said(listed), "get_item") {
		t.Errorf("get_list said %q, want it to say where the rest is", said(listed))
	}

	whole, err := session.CallTool(t.Context(), &sdk.CallToolParams{
		Name: "get_item", Arguments: map[string]any{"item_uid": itemUID},
	})
	if err != nil {
		t.Fatalf("get_item: %v", err)
	}
	if !strings.Contains(said(whole), note) {
		t.Errorf("get_item said %q, want the Note byte for byte", said(whole))
	}
}

/*
Archiving a List must not be a way of losing it.

archive_list is offered, so the listing that finds an archived one again has to be too.
Without the status argument an assistant could put a List away and never see it again.
*/
func TestAnAssistantCanFindWhatItArchived(t *testing.T) {
	i := newInstance(t)
	session := i.connect(i.tokenFor(store.TokenAbilities{Read: true, Write: true}))

	if _, err := session.CallTool(t.Context(), &sdk.CallToolParams{
		Name: "archive_list", Arguments: map[string]any{"list_uid": i.listUID, "archived": true},
	}); err != nil {
		t.Fatalf("archive_list: %v", err)
	}

	plain, err := session.CallTool(t.Context(), &sdk.CallToolParams{Name: "list_lists"})
	if err != nil {
		t.Fatalf("list_lists: %v", err)
	}
	if strings.Contains(said(plain), i.listUID) {
		t.Errorf("an archived List is still in the plain listing: %q", said(plain))
	}

	archived, err := session.CallTool(t.Context(), &sdk.CallToolParams{
		Name: "list_lists", Arguments: map[string]any{"status": "archived"},
	})
	if err != nil {
		t.Fatalf("list_lists archived: %v", err)
	}
	if !strings.Contains(said(archived), i.listUID) {
		t.Errorf("archived listing = %q, want the List it put away", said(archived))
	}
}

/*
An archived List stays searchable, which archive_list tells an assistant it does.

"It keeps its items and stays searchable, this is not deleting" is what the tool says,
and an assistant repeats it to whoever asked. It is true by nothing: neither the store's
Search nor the service filters archived Lists out, so the claim holds because no one has
written the filter rather than because anything keeps them from it.

Filtering archived Lists out of search is a reasonable thing for somebody to want, and
doing it would make the sentence above false in the one place a machine reads it and
acts. So the claim is held here instead.
*/
func TestAnArchivedListStaysSearchable(t *testing.T) {
	i := newInstance(t)
	i.withItem("Tarpaulin", "")
	session := i.connect(i.tokenFor(store.TokenAbilities{Read: true, Write: true}))

	if _, err := session.CallTool(t.Context(), &sdk.CallToolParams{
		Name: "archive_list", Arguments: map[string]any{"list_uid": i.listUID, "archived": true},
	}); err != nil {
		t.Fatalf("archive_list: %v", err)
	}

	found, err := session.CallTool(t.Context(), &sdk.CallToolParams{
		Name: "search", Arguments: map[string]any{"query": "Tarpaulin"},
	})
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if !strings.Contains(said(found), "Tarpaulin") {
		t.Errorf("search after archiving = %q, want the Item still found: archive_list "+
			"tells an assistant that archiving is not deleting", said(found))
	}
}

/*
A rewritten Note can be refused rather than overwriting somebody.

The API offers this and the tool did not, which left an assistant as the one caller that
could only ever clobber. It is the other half of shortening a Note on read.
*/
func TestARewrittenNoteCanBeRefused(t *testing.T) {
	i := newInstance(t)
	itemUID := i.withItem("Coffee", "Saturday market.")
	session := i.connect(i.tokenFor(store.TokenAbilities{Read: true, Write: true}))

	stale := "What the assistant thought it said."
	mine := "Sunday market."
	res, err := session.CallTool(t.Context(), &sdk.CallToolParams{
		Name: "update_item",
		Arguments: map[string]any{
			"item_uid": itemUID, "note": mine, "expected_note": stale,
		},
	})
	if err != nil {
		t.Fatalf("update_item: %v", err)
	}
	if !res.IsError {
		t.Errorf("update_item answered %q, want it refused", said(res))
	}
	// The refusal says who got there first rather than naming a code, because that is
	// what an assistant has to relay to whoever asked it.
	if !strings.Contains(said(res), "somebody else wrote in this note") {
		t.Errorf("refusal said %q, want it to say what happened", said(res))
	}

	item, err := i.store.ItemByUID(t.Context(), itemUID)
	if err != nil {
		t.Fatalf("ItemByUID: %v", err)
	}
	if item.Note != "Saturday market." {
		t.Errorf("Note = %q, want the one that was there", item.Note)
	}
}
