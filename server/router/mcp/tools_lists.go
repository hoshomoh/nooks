package mcp

import (
	"context"

	"fmt"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"

	apiv1 "github.com/hoshomoh/nooks/proto/gen/nooks/api/v1"
	v1 "github.com/hoshomoh/nooks/server/router/api/v1"
)

type listListsArgs struct{}

type getListArgs struct {
	ListUID string `json:"list_uid" jsonschema:"the list to read, from list_lists"`
}

type createListArgs struct {
	Name string `json:"name" jsonschema:"what to call it, in the words a person would use"`
}

type renameListArgs struct {
	ListUID string `json:"list_uid" jsonschema:"the list to rename, from list_lists"`
	Name    string `json:"name" jsonschema:"the new name"`
}

type listUIDArgs struct {
	ListUID string `json:"list_uid" jsonschema:"the list to act on, from list_lists"`
}

type pinListArgs struct {
	ListUID string `json:"list_uid" jsonschema:"the list to pin, from list_lists"`
	Pinned  bool   `json:"pinned" jsonschema:"true to pin it to the sidebar, false to unpin"`
}

// addListTools registers what can be done to a List itself, as opposed to what is on it.
func addListTools(server *sdk.Server, lists *v1.ListService) {
	sdk.AddTool(server, &sdk.Tool{
		Name:        "list_lists",
		Description: "Every list the caller can reach, with how many items are still open on each.",
	}, func(ctx context.Context, _ *sdk.CallToolRequest, _ listListsArgs) (*sdk.CallToolResult, any, error) {
		return answer(ctx, lists.ListLists, &apiv1.ListListsRequest{},
			func(res *apiv1.ListListsResponse) string {
				rows := make([]string, 0, len(res.GetLists()))
				for _, list := range res.GetLists() {
					rows = append(rows, listLine(list))
				}
				return lines(rows, "No lists.")
			})
	})

	sdk.AddTool(server, &sdk.Tool{
		Name:        "get_list",
		Description: "What is on one list, ticked and unticked, with each item's identifier.",
	}, func(ctx context.Context, _ *sdk.CallToolRequest, args getListArgs) (*sdk.CallToolResult, any, error) {
		return answer(ctx, lists.GetList, &apiv1.GetListRequest{ListUid: args.ListUID},
			func(res *apiv1.GetListResponse) string {
				rows := make([]string, 0, len(res.GetItems())+1)
				rows = append(rows, res.GetList().GetName())
				for _, item := range res.GetItems() {
					rows = append(rows, itemLine(item))
				}
				return lines(rows, "Empty list.")
			})
	})

	sdk.AddTool(server, &sdk.Tool{
		Name:        "create_list",
		Description: "Start a new list. Refused when the caller's token may only read.",
	}, func(ctx context.Context, _ *sdk.CallToolRequest, args createListArgs) (*sdk.CallToolResult, any, error) {
		return answer(ctx, lists.CreateList, &apiv1.CreateListRequest{Name: args.Name},
			func(res *apiv1.CreateListResponse) string {
				return fmt.Sprintf("Created %q — %s", res.GetList().GetName(), res.GetList().GetUid())
			})
	})

	sdk.AddTool(server, &sdk.Tool{
		Name:        "rename_list",
		Description: "Rename a list. Only its owner may, and only with a token that may write.",
	}, func(ctx context.Context, _ *sdk.CallToolRequest, args renameListArgs) (*sdk.CallToolResult, any, error) {
		return answer(ctx, lists.RenameList,
			&apiv1.RenameListRequest{ListUid: args.ListUID, Name: args.Name},
			func(res *apiv1.RenameListResponse) string {
				return fmt.Sprintf("Renamed to %q.", res.GetList().GetName())
			})
	})

	sdk.AddTool(server, &sdk.Tool{
		Name:        "duplicate_list",
		Description: "Copy a list and everything on it, unticked. Useful for a weekly shop.",
	}, func(ctx context.Context, _ *sdk.CallToolRequest, args listUIDArgs) (*sdk.CallToolResult, any, error) {
		return answer(ctx, lists.DuplicateList, &apiv1.DuplicateListRequest{ListUid: args.ListUID},
			func(res *apiv1.DuplicateListResponse) string {
				return fmt.Sprintf("Copied to %q — %s", res.GetList().GetName(), res.GetList().GetUid())
			})
	})

	sdk.AddTool(server, &sdk.Tool{
		Name:        "pin_list",
		Description: "Pin a list to the caller's own sidebar, or unpin it. Nobody else's sidebar changes.",
	}, func(ctx context.Context, _ *sdk.CallToolRequest, args pinListArgs) (*sdk.CallToolResult, any, error) {
		said := "Pinned."
		if !args.Pinned {
			said = "Unpinned."
		}
		return answer(ctx, lists.SetListPinned,
			&apiv1.SetListPinnedRequest{ListUid: args.ListUID, Pinned: args.Pinned},
			done[apiv1.SetListPinnedResponse](said))
	})

	sdk.AddTool(server, &sdk.Tool{
		Name:        "delete_list",
		Description: "Delete a list and everything on it. Refused unless the caller's token may delete.",
	}, func(ctx context.Context, _ *sdk.CallToolRequest, args listUIDArgs) (*sdk.CallToolResult, any, error) {
		return answer(ctx, lists.DeleteList, &apiv1.DeleteListRequest{ListUid: args.ListUID},
			done[apiv1.DeleteListResponse]("Deleted."))
	})
}
