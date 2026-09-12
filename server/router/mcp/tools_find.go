package mcp

import (
	"context"
	"fmt"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"

	apiv1 "github.com/hoshomoh/nooks/proto/gen/nooks/api/v1"
	v1 "github.com/hoshomoh/nooks/server/router/api/v1"
)

type searchArgs struct {
	Query string `json:"query" jsonschema:"what to look for across every list the caller can reach"`
}

type datedItemsArgs struct {
	From string `json:"from,omitempty" jsonschema:"earliest day as YYYY-MM-DD, or empty to include everything overdue"`
	To   string `json:"to" jsonschema:"latest day as YYYY-MM-DD"`
}

// addFindTools registers the two ways of reaching an Item without knowing which List
// it is on: by what it says, and by when it is due.
func addFindTools(server *sdk.Server, lists *v1.ListService) {
	sdk.AddTool(server, &sdk.Tool{
		Name:        "search",
		Description: "Find lists and items by what somebody wrote in them, across everything the caller can reach.",
	}, func(ctx context.Context, _ *sdk.CallToolRequest, args searchArgs) (*sdk.CallToolResult, any, error) {
		return answer(ctx, lists.Search, &apiv1.SearchRequest{Query: args.Query},
			func(res *apiv1.SearchResponse) string {
				rows := make([]string, 0, len(res.GetHits()))
				for _, hit := range res.GetHits() {
					rows = append(rows, fmt.Sprintf("%s — on %s", hit.GetText(), hit.GetListName()))
				}
				return lines(rows, fmt.Sprintf("Nothing matching %q.", args.Query))
			})
	})

	sdk.AddTool(server, &sdk.Tool{
		Name:        "list_dated_items",
		Description: "Items due in a range of days, across every list. Leave 'from' empty to sweep up whatever is overdue as well, which is what Today does.",
	}, func(ctx context.Context, _ *sdk.CallToolRequest, args datedItemsArgs) (*sdk.CallToolResult, any, error) {
		return answer(ctx, lists.ListDatedItems,
			&apiv1.ListDatedItemsRequest{From: args.From, To: args.To},
			func(res *apiv1.ListDatedItemsResponse) string {
				rows := make([]string, 0, len(res.GetItems()))
				for _, dated := range res.GetItems() {
					rows = append(rows, fmt.Sprintf("%s — on %s",
						itemLine(dated.GetItem()), dated.GetListName()))
				}
				return lines(rows, "Nothing due.")
			})
	})
}
