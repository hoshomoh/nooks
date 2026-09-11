// Package mcp serves Nooks' tools to an assistant over the Model Context Protocol.
package mcp

import (
	"context"
	"net/http"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/hoshomoh/nooks/internal/version"
	"github.com/hoshomoh/nooks/server/auth"
	v1 "github.com/hoshomoh/nooks/server/router/api/v1"
)

/*
Handler answers MCP over HTTP at /mcp.

The tools call the same ListService the browser and the REST API call, with the same
Grant on the context. There is no separate path through the permission rules and no
in-process HTTP round trip to fake one: an assistant holding a read-only token is
refused a write by accessTo, exactly as a script would be.

The set of tools is deliberately small. A tool for every RPC would be a menu an
assistant has to read before it can do the one thing it was asked to do, and most of
those RPCs are about running an Instance rather than about somebody's shopping.
*/
func Handler(lists *v1.ListService, resolver *auth.Resolver) http.Handler {
	server := sdk.NewServer(&sdk.Implementation{
		Name:    "nooks",
		Version: version.String(),
	}, nil)

	addTools(server, lists)

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

// ctxGrant is unused here but documents the contract: the tools read the caller from
// the context, which withGrant put there.
var _ = func(ctx context.Context) { _, _ = auth.GrantFrom(ctx) }
