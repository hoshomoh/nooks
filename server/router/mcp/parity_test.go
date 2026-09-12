package mcp

import (
	"sort"
	"testing"

	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/hoshomoh/nooks/store"
)

/*
The three doors must agree about what an Access token may do.

Connect serves the browser, the gateway serves REST, and this package serves an
assistant. A tool missing from here is a thing a household can do from a script and not
from an assistant, which makes the same instance behave differently depending on which
door it came through. That gap was shipped once, deliberately, and this is what stops
it being shipped again quietly.

Every RPC is classified below, and an unclassified one fails the test: adding an RPC is
a decision about whether an assistant can reach it, and leaving that decision unmade is
not one of the options.
*/

// reach says how an assistant reaches an RPC, or why it cannot.
type reach struct {
	// tool is the MCP tool that covers it. Empty when a token cannot reach the RPC.
	tool string

	// why records the reason, for an RPC with no tool. It is prose rather than an
	// enum because the reasons do not form a set worth naming.
	why string
}

// browserOnly is the reason for every RPC behind requireBrowser or requireAdmin: an
// Access token is refused these over REST too, so their absence here is the token's
// boundary rather than this package's. See access.go.
const browserOnly = "requireBrowser: an Access token is refused this over REST too"

// ceremony is the reason for the RPCs that establish a session or serve a stranger.
// A caller already holding a token has, by definition, finished with them.
const ceremony = "signing in, joining, or reading a public list: not a token's work"

var reaches = map[string]reach{
	// Lists and what is on them.
	"ListService.ListLists":            {tool: "list_lists"},
	"ListService.GetList":              {tool: "get_list"},
	"ListService.CreateList":           {tool: "create_list"},
	"ListService.RenameList":           {tool: "rename_list"},
	"ListService.DuplicateList":        {tool: "duplicate_list"},
	"ListService.DeleteList":           {tool: "delete_list"},
	"ListService.SetListPinned":        {tool: "pin_list"},
	"ListService.SetListSharing":       {tool: "set_list_sharing"},
	"ListService.GetListShares":        {tool: "get_list_shares"},
	"ListService.CreateItem":           {tool: "add_item"},
	"ListService.UpdateItem":           {tool: "update_item"},
	"ListService.SetItemDone":          {tool: "complete_item"},
	"ListService.MoveItem":             {tool: "move_item"},
	"ListService.DeleteItem":           {tool: "delete_item"},
	"ListService.Search":               {tool: "search"},
	"ListService.ListDatedItems":       {tool: "list_dated_items"},
	"MemberService.ListMembers":        {tool: "list_members"},
	"MemberService.ListGroups":         {tool: "list_groups"},
	"AuthService.GetCurrentMember":     {tool: "whoami"},
	"ActivityService.ListActivity":     {tool: "list_activity"},
	"ActivityService.MarkActivityRead": {tool: "mark_activity_read"},
	"InstanceService.GetInstanceAbout": {tool: "about_instance"},

	// Running the Instance. A token cannot, anywhere.
	"AuthService.ReplacePassword":            {why: browserOnly},
	"MemberService.UpdateOwnProfile":         {why: browserOnly},
	"MemberService.AddMember":                {why: browserOnly},
	"MemberService.RemoveMember":             {why: browserOnly},
	"MemberService.SetMemberRole":            {why: browserOnly},
	"MemberService.CreateGroup":              {why: browserOnly},
	"MemberService.SetGroupMembers":          {why: browserOnly},
	"TokenService.CreateAccessToken":         {why: browserOnly},
	"TokenService.ListAccessTokens":          {why: browserOnly},
	"TokenService.RevokeAccessToken":         {why: browserOnly},
	"InstanceService.GetInstanceSettings":    {why: browserOnly},
	"InstanceService.UpdateInstanceSettings": {why: browserOnly},
	"InstanceService.DeleteInstance":         {why: browserOnly},
	"RequestService.ListPendingRequests":     {why: browserOnly},
	"RequestService.DecideJoinRequest":       {why: browserOnly},
	"RequestService.DecideResetRequest":      {why: browserOnly},

	// Getting in, and what a stranger sees.
	"AuthService.SignIn":                {why: ceremony},
	"AuthService.SignOut":               {why: ceremony},
	"AuthService.RefreshAccess":         {why: ceremony},
	"AuthService.CompleteSetup":         {why: ceremony},
	"AuthService.RequestJoin":           {why: ceremony},
	"AuthService.GetJoinRequest":        {why: ceremony},
	"AuthService.CompleteJoin":          {why: ceremony},
	"AuthService.RequestPasswordReset":  {why: ceremony},
	"AuthService.GetResetRequest":       {why: ceremony},
	"AuthService.CompletePasswordReset": {why: ceremony},
	"InstanceService.GetInstance":       {why: ceremony},
	"PublicService.GetPublicList":       {why: ceremony},
}

func TestEveryRPCIsClassified(t *testing.T) {
	for _, rpc := range declaredRPCs(t) {
		if _, ok := reaches[rpc]; !ok {
			t.Errorf("%s has no entry in reaches: give it a tool, or say why a token cannot reach it", rpc)
		}
	}
}

func TestClassificationDescribesRealRPCs(t *testing.T) {
	declared := make(map[string]bool)
	for _, rpc := range declaredRPCs(t) {
		declared[rpc] = true
	}
	for rpc := range reaches {
		if !declared[rpc] {
			t.Errorf("reaches names %s, which is not an RPC any more", rpc)
		}
	}
}

func TestEveryClassifiedToolIsRegistered(t *testing.T) {
	registered := registeredTools(t)

	wanted := map[string]bool{}
	for rpc, how := range reaches {
		if how.tool == "" {
			if how.why == "" {
				t.Errorf("%s has neither a tool nor a reason", rpc)
			}
			continue
		}
		wanted[how.tool] = true
		if !registered[how.tool] {
			t.Errorf("%s is classified as tool %q, which is not registered", rpc, how.tool)
		}
	}

	for tool := range registered {
		if !wanted[tool] {
			t.Errorf("tool %q is registered but covers no RPC", tool)
		}
	}
}

// declaredRPCs reads the service definitions rather than a list kept here, so an RPC
// added to a proto shows up without anybody remembering to mention it.
func declaredRPCs(t *testing.T) []string {
	t.Helper()

	var found []string
	protoregistry.GlobalFiles.RangeFilesByPackage("nooks.api.v1",
		func(file protoreflect.FileDescriptor) bool {
			services := file.Services()
			for i := range services.Len() {
				service := services.Get(i)
				methods := service.Methods()
				for j := range methods.Len() {
					found = append(found, string(service.Name())+"."+string(methods.Get(j).Name()))
				}
			}
			return true
		})

	if len(found) == 0 {
		t.Fatal("no RPCs found: the generated descriptors are not linked in")
	}
	sort.Strings(found)
	return found
}

// registeredTools asks the server what it serves, over a real session, so a tool with
// a schema the SDK refuses is caught here rather than by whoever tries to use it.
func registeredTools(t *testing.T) map[string]bool {
	t.Helper()

	i := newInstance(t)
	session := i.connect(i.tokenFor(store.TokenAbilities{Read: true, Write: true, Delete: true}))

	res, err := session.ListTools(t.Context(), &sdk.ListToolsParams{})
	if err != nil {
		t.Fatalf("ListTools: %v", err)
	}

	tools := make(map[string]bool, len(res.Tools))
	for _, tool := range res.Tools {
		tools[tool.Name] = true
	}
	return tools
}
