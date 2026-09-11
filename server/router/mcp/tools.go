package mcp

import (
	"context"
	"fmt"
	"strings"

	"connectrpc.com/connect"
	sdk "github.com/modelcontextprotocol/go-sdk/mcp"

	apiv1 "github.com/hoshomoh/nooks/proto/gen/nooks/api/v1"
	v1 "github.com/hoshomoh/nooks/server/router/api/v1"
)

// Arguments are described here rather than in a schema file: the SDK reads these
// structs and writes the JSON Schema, so the description an assistant reads and the
// field a handler uses cannot drift apart.

type listListsArgs struct{}

type getListArgs struct {
	ListUID string `json:"list_uid" jsonschema:"the list to read, from list_lists"`
}

type addItemArgs struct {
	ListUID  string `json:"list_uid" jsonschema:"the list to add to, from list_lists"`
	Label    string `json:"label" jsonschema:"what to add, in the words a person would use"`
	Quantity string `json:"quantity,omitempty" jsonschema:"free text like 2 or 1 kg, not a number"`
	DueOn    string `json:"due_on,omitempty" jsonschema:"a day as YYYY-MM-DD, or empty for no date"`
}

type completeItemArgs struct {
	ItemUID string `json:"item_uid" jsonschema:"the item to tick off, from get_list"`
	Done    bool   `json:"done" jsonschema:"true to tick it off, false to put it back"`
}

type searchArgs struct {
	Query string `json:"query" jsonschema:"what to look for across every list the caller can reach"`
}

/*
addTools registers what an assistant may do.

Five, covering the whole of what a household actually asks for: what is on the lists,
what is on one list, add something, tick something off, find something. Every one of
them goes through ListService, so the answer to "may I?" is the same answer the app
gets.
*/
func addTools(server *sdk.Server, lists *v1.ListService) {
	sdk.AddTool(server, &sdk.Tool{
		Name:        "list_lists",
		Description: "Every list the caller can reach, with how many items are still open on each.",
	}, func(ctx context.Context, _ *sdk.CallToolRequest, _ listListsArgs) (*sdk.CallToolResult, any, error) {
		res, err := lists.ListLists(ctx, connect.NewRequest(&apiv1.ListListsRequest{}))
		if err != nil {
			return errorText(err), nil, nil
		}

		var said strings.Builder
		for _, list := range res.Msg.GetLists() {
			fmt.Fprintf(&said, "%s (%s) — %d open\n", list.GetName(), list.GetUid(), list.GetOpenCount())
		}
		if said.Len() == 0 {
			return text("No lists."), nil, nil
		}
		return text(said.String()), nil, nil
	})

	sdk.AddTool(server, &sdk.Tool{
		Name:        "get_list",
		Description: "What is on one list, ticked and unticked, with each item's identifier.",
	}, func(ctx context.Context, _ *sdk.CallToolRequest, args getListArgs) (*sdk.CallToolResult, any, error) {
		res, err := lists.GetList(ctx, connect.NewRequest(&apiv1.GetListRequest{
			ListUid: args.ListUID,
		}))
		if err != nil {
			return errorText(err), nil, nil
		}

		var said strings.Builder
		fmt.Fprintf(&said, "%s\n", res.Msg.GetList().GetName())
		for _, item := range res.Msg.GetItems() {
			mark := " "
			if item.GetDone() {
				mark = "x"
			}
			fmt.Fprintf(&said, "[%s] %s", mark, item.GetLabel())
			if quantity := item.GetQuantity(); quantity != "" {
				fmt.Fprintf(&said, " (%s)", quantity)
			}
			if due := item.GetDueOn(); due != "" {
				fmt.Fprintf(&said, " due %s", due)
			}
			fmt.Fprintf(&said, " — %s\n", item.GetUid())
		}
		return text(said.String()), nil, nil
	})

	sdk.AddTool(server, &sdk.Tool{
		Name:        "add_item",
		Description: "Put something on a list. Refused when the caller's token may only read.",
	}, func(ctx context.Context, _ *sdk.CallToolRequest, args addItemArgs) (*sdk.CallToolResult, any, error) {
		res, err := lists.CreateItem(ctx, connect.NewRequest(&apiv1.CreateItemRequest{
			ListUid: args.ListUID, Label: args.Label,
			Quantity: args.Quantity, DueOn: args.DueOn,
		}))
		if err != nil {
			return errorText(err), nil, nil
		}
		return text(fmt.Sprintf("Added %q — %s", args.Label, res.Msg.GetItem().GetUid())), nil, nil
	})

	sdk.AddTool(server, &sdk.Tool{
		Name:        "complete_item",
		Description: "Tick an item off, or put it back. Refused when the caller's token may only read.",
	}, func(ctx context.Context, _ *sdk.CallToolRequest, args completeItemArgs) (*sdk.CallToolResult, any, error) {
		if _, err := lists.SetItemDone(ctx, connect.NewRequest(&apiv1.SetItemDoneRequest{
			ItemUid: args.ItemUID, Done: args.Done,
		})); err != nil {
			return errorText(err), nil, nil
		}
		if args.Done {
			return text("Ticked off."), nil, nil
		}
		return text("Put back."), nil, nil
	})

	sdk.AddTool(server, &sdk.Tool{
		Name:        "search",
		Description: "Find lists and items by what somebody wrote in them, across everything the caller can reach.",
	}, func(ctx context.Context, _ *sdk.CallToolRequest, args searchArgs) (*sdk.CallToolResult, any, error) {
		res, err := lists.Search(ctx, connect.NewRequest(&apiv1.SearchRequest{Query: args.Query}))
		if err != nil {
			return errorText(err), nil, nil
		}

		var said strings.Builder
		for _, hit := range res.Msg.GetHits() {
			fmt.Fprintf(&said, "%s — on %s\n", hit.GetText(), hit.GetListName())
		}
		if said.Len() == 0 {
			return text(fmt.Sprintf("Nothing matching %q.", args.Query)), nil, nil
		}
		return text(said.String()), nil, nil
	})
}
