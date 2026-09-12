package mcp

import (
	"context"

	"connectrpc.com/connect"
	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

/*
answer runs one service call and renders what came back.

Every tool is the same three steps — wrap the arguments, call the service, say what
happened — and the interesting part is only ever the last one. Writing those steps out
twenty-two times would be twenty-two chances to drop an error on the floor, so they are
written once and each tool supplies the sentence.

A refusal is content rather than a transport error on purpose: an assistant told "only
an admin can do that" can say so, where a failed call leaves it with nothing to relay.
*/
func answer[Req, Res any](
	ctx context.Context,
	call func(context.Context, *connect.Request[Req]) (*connect.Response[Res], error),
	msg *Req,
	say func(*Res) string,
) (*sdk.CallToolResult, any, error) {
	res, err := call(ctx, connect.NewRequest(msg))
	if err != nil {
		return errorText(err), nil, nil
	}
	return text(say(res.Msg)), nil, nil
}

// done is the rendering for a call that returns nothing: the sentence is fixed, and
// the only news is that it worked.
func done[Res any](said string) func(*Res) string {
	return func(*Res) string { return said }
}
