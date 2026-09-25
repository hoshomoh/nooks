// Package mcp serves nooks' tools to an assistant over the Model Context Protocol.
package mcp

import (
	"net/http"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/hoshomoh/nooks/internal/version"
	"github.com/hoshomoh/nooks/server/auth"
	v1 "github.com/hoshomoh/nooks/server/router/api/v1"
)

/*
Handler answers MCP over HTTP at /mcp.

The tools call the same services the browser and the REST API call, with the same
Grant on the context. There is no separate path through the permission rules and no
in-process HTTP round trip to fake one: an assistant holding a read-only token is
refused a write by accessTo, exactly as a script would be.

It takes the whole set of services rather than one of them because an Access token
reaches more than Lists, and a door that reached less would be a second, quieter answer
to "what may this token do".
*/
func Handler(services v1.Services, resolver *auth.Resolver) http.Handler {
	server := sdk.NewServer(&sdk.Implementation{
		Name:    "nooks",
		Version: version.String(),
	}, &sdk.ServerOptions{Instructions: instructions})

	addTools(server, services)

	streamable := sdk.NewStreamableHTTPHandler(
		func(*http.Request) *sdk.Server { return server },
		nil,
	)
	return withGrant(streamable, resolver)
}

/*
withGrant resolves the caller before the protocol sees the request.

The same resolver.Grant the Connect interceptor and the REST gateway use. Three doors,
one function: the moment they disagree is the moment a token can do something through
one of them that it cannot through the others.

A request with no usable credential is refused here rather than inside a tool, because
an assistant should be told it is not signed in once, not once per tool call.
*/
func withGrant(next http.Handler, resolver *auth.Resolver) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		grant, ok := resolver.Grant(r.Context(), r.Header)
		if !ok {
			w.Header().Set("WWW-Authenticate", `Bearer realm="nooks"`)
			http.Error(w, "not signed in", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r.WithContext(auth.WithGrant(r.Context(), grant)))
	})
}

// errorText turns a failed call into something an assistant can act on, rather than a
// stack of protocol nouns. The services already write in plain words.
func errorText(err error) *sdk.CallToolResult {
	return &sdk.CallToolResult{
		IsError: true,
		Content: []sdk.Content{&sdk.TextContent{Text: err.Error()}},
	}
}

// text is the ordinary answer: one block an assistant can read back.
func text(said string) *sdk.CallToolResult {
	return &sdk.CallToolResult{Content: []sdk.Content{&sdk.TextContent{Text: said}}}
}

/*
instructions orient an assistant once, at the start, in what no single tool can say.

The protocol offers this and nothing was sent, so a client arrived at twenty-four tools
with no idea which one to call first or how they fit together. A tool description is
read when a tool is being considered; this is read before any of them are.

It says the shape of the thing and the two traps. The shape is that there is one
container and identifiers are handed out rather than guessed. The traps are that a note
is replaced whole rather than appended to, and that a refusal is usually the token's
edges rather than a fault worth retrying.
*/
const instructions = `Lists are the only container. Everything is an item on a list, and there are no folders, tags or projects to look for.

Start at list_lists. Every identifier the other tools take came from a tool that returned it, and none of them can be guessed.

get_list shortens a long note to keep a row a row. get_item has the whole of one, and is what to read before rewriting a note: update_item replaces a note rather than adding to it. Send expected_note with the rewrite and it is refused rather than overwriting somebody who changed it in between.

A date is a day. Nothing here is scheduled to a time.

What a token may do is decided when it is minted, so a refusal is usually the token's edges rather than a fault to retry. Reading, writing and deleting are separate, and the admin work, adding people, minting tokens and deciding requests, is not open to a token at all.`
