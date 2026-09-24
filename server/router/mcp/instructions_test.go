package mcp

import (
	"regexp"
	"testing"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/hoshomoh/nooks/store"
)

// named matches the words in the instructions that are meant to be something a client
// can call or send. Nothing else in the text carries an underscore.
var named = regexp.MustCompile(`[a-z]+(?:_[a-z]+)+`)

/*
What an assistant is told at the start names things that are really there.

The protocol sends this once, before any tool is considered, and nothing was sent at
all: a client arrived at twenty-four tools with no idea which to call first. Text that
orients somebody is worse than no text the moment it is wrong, and the way it goes wrong
is a rename, because nothing compiles against prose.

So every tool and argument the instructions name is checked against what the server
actually serves. Renaming get_item without touching the sentence that sends an assistant
to it fails here.
*/
func TestTheInstructionsNameWhatIsThere(t *testing.T) {
	i := newInstance(t)
	session := i.connect(i.tokenFor(store.TokenAbilities{Read: true, Write: true, Delete: true}))

	got := session.InitializeResult().Instructions
	if got != instructions {
		t.Fatalf("a client is given %q, want the instructions this package holds", got)
	}

	res, err := session.ListTools(t.Context(), &sdk.ListToolsParams{})
	if err != nil {
		t.Fatalf("ListTools: %v", err)
	}

	served := map[string]bool{}
	for _, tool := range res.Tools {
		served[tool.Name] = true
		for _, argument := range argumentsOf(t, tool) {
			served[argument] = true
		}
	}

	found := named.FindAllString(instructions, -1)
	for _, word := range found {
		if !served[word] {
			t.Errorf("the instructions send an assistant to %q, which the server does not serve", word)
		}
	}

	// list_lists, get_list, get_item, update_item and expected_note. Text this stopped
	// reading would pass by checking nothing.
	if len(found) < 5 {
		t.Fatalf("read %d names out of the instructions, fewer than they carry", len(found))
	}
}

// argumentsOf is every argument a tool takes, whatever its type.
func argumentsOf(t *testing.T, tool *sdk.Tool) []string {
	t.Helper()

	schema, ok := tool.InputSchema.(map[string]any)
	if !ok {
		t.Fatalf("%s: input schema is %T, not an object", tool.Name, tool.InputSchema)
	}
	properties, ok := schema["properties"].(map[string]any)
	if !ok {
		return nil
	}

	names := make([]string, 0, len(properties))
	for name := range properties {
		names = append(names, name)
	}
	return names
}
