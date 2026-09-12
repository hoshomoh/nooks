package mcp

import (
	sdk "github.com/modelcontextprotocol/go-sdk/mcp"

	v1 "github.com/hoshomoh/nooks/server/router/api/v1"
)

/*
addTools registers what an assistant may do.

Everything an Access token can reach, which is everything the REST API gives one. The
set is not curated down to what somebody guessed an assistant would want: a household
that can tick an item off from a script and not from an assistant has been told the
same instance does two different things depending on which door it came through.

What is missing from here is missing from a token everywhere. Minting a token, changing
a password, adding a Member, deciding a join request and every other Admin action go
through requireBrowser, so a token is refused them over REST too — see access.go. The
boundary is the token's, not this package's, which is why parity_test.go checks it
against the service methods rather than against a list kept here.

Arguments are described on the structs rather than in a schema file: the SDK reads
those and writes the JSON Schema, so the description an assistant reads and the field a
handler uses cannot drift apart.
*/
func addTools(server *sdk.Server, services v1.Services) {
	addListTools(server, services.List)
	addItemTools(server, services.List)
	addFindTools(server, services.List)
	addPeopleTools(server, services)
	addInstanceTools(server, services)
}
