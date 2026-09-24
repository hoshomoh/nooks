package mcp

import (
	"slices"
	"strconv"
	"strings"
	"testing"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"

	v1 "github.com/hoshomoh/nooks/server/router/api/v1"
	"github.com/hoshomoh/nooks/store"
)

// capped is the argument name an assistant writes into, against the number the Instance
// refuses past. Named by argument rather than by tool so a new tool taking a label is
// held to the same answer without anybody adding it here.
var capped = map[string]int{
	"label":    v1.LimitItemLabel,
	"quantity": v1.LimitItemQuantity,
	"note":     v1.LimitItemNote,
	"name":     v1.LimitListName,
}

/*
A tool says how much a field holds before it is asked.

Nothing in these tools carried a length. An assistant writing a six hundred character
label learned the cap of five hundred by being refused, and the refusal is the first
time the number appears anywhere it can read, which leaves it guessing what to trim to.

The web was given this when the limits went in and the tools were not, which is the
shape of the gap: a change lands on the surface somebody was looking at.

Read off a real session rather than the source, because the description an assistant
gets is the one the SDK serves. The tools are not listed here on purpose. Every
registered tool is asked what strings it takes, and any argument this package knows a
cap for has to have its number in the description, so the next tool with a label in it
fails this until somebody says how long a label may be.
*/
func TestAToolSaysHowMuchAFieldHolds(t *testing.T) {
	i := newInstance(t)
	session := i.connect(i.tokenFor(store.TokenAbilities{Read: true, Write: true, Delete: true}))

	res, err := session.ListTools(t.Context(), &sdk.ListToolsParams{})
	if err != nil {
		t.Fatalf("ListTools: %v", err)
	}

	checked := 0
	for _, tool := range res.Tools {
		for _, argument := range stringArgumentsOf(t, tool) {
			limit, ok := capped[argument]
			if !ok {
				continue
			}
			checked++
			if !strings.Contains(tool.Description, strconv.Itoa(limit)) {
				t.Errorf("%s takes %s and never says it stops at %d: %q",
					tool.Name, argument, limit, tool.Description)
			}
		}
	}

	// Seven between four tools: a label and a quantity on add_item, those two and a
	// note on update_item, and a name on each of create_list and rename_list. A schema
	// this stopped reading would otherwise pass by finding nothing.
	if checked < 7 {
		t.Fatalf("found %d capped arguments across the tools, which is fewer than there are", checked)
	}
}

// stringArgumentsOf is the names of the string arguments a tool takes, as the client
// sees them: the SDK hands the schema back as plain JSON rather than as its own type.
func stringArgumentsOf(t *testing.T, tool *sdk.Tool) []string {
	t.Helper()

	schema, ok := tool.InputSchema.(map[string]any)
	if !ok {
		t.Fatalf("%s: input schema is %T, not an object", tool.Name, tool.InputSchema)
	}
	properties, ok := schema["properties"].(map[string]any)
	if !ok {
		// A tool taking nothing, such as whoami.
		return nil
	}

	names := make([]string, 0, len(properties))
	for name, described := range properties {
		if holdsAString(described) {
			names = append(names, name)
		}
	}
	return names
}

// holdsAString reports whether one argument takes text.
//
// An argument that may be left out is a pointer on the struct, and the schema writes
// that as ["null", "string"] rather than as "string". Reading only the plain spelling
// found four arguments where there are eight, and said nothing.
func holdsAString(described any) bool {
	property, ok := described.(map[string]any)
	if !ok {
		return false
	}
	switch kind := property["type"].(type) {
	case string:
		return kind == "string"
	case []any:
		return slices.Contains(kind, any("string"))
	}
	return false
}
