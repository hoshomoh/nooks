package mcp

import (
	"context"

	"fmt"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"

	apiv1 "github.com/hoshomoh/nooks/proto/gen/nooks/api/v1"
	v1 "github.com/hoshomoh/nooks/server/router/api/v1"
)

type listListsArgs struct {
	Page int `json:"page,omitempty" jsonschema:"which page to read, 1-based; leave it out for the first"`
	// Without these an archived list cannot be found again, which would make
	// archive_list a one way door.
	Status string `json:"status,omitempty" jsonschema:"which lists to show: active for ones still in play, completed for ones with everything ticked, archived for ones put away, or leave it out for everything except archived"`
	Order  string `json:"order,omitempty" jsonschema:"how to order them: updated for what changed last, name, or open for the fullest first. Leave it out for updated"`
}

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

type archiveListArgs struct {
	ListUID  string `json:"list_uid" jsonschema:"the list to archive, from list_lists"`
	Archived bool   `json:"archived" jsonschema:"true to put it out of the sidebar, false to bring it back"`
}

// statusWords and orderWords map what an assistant may say onto what the API takes, for
// the reason sharingWords gives. The empty string is "did not ask", not a bad answer.
var statusWords = map[string]apiv1.ListStatus{
	"":          apiv1.ListStatus_LIST_STATUS_UNSPECIFIED,
	"active":    apiv1.ListStatus_LIST_STATUS_ACTIVE,
	"completed": apiv1.ListStatus_LIST_STATUS_COMPLETED,
	"archived":  apiv1.ListStatus_LIST_STATUS_ARCHIVED,
}

var orderWords = map[string]apiv1.ListOrder{
	"":        apiv1.ListOrder_LIST_ORDER_UNSPECIFIED,
	"updated": apiv1.ListOrder_LIST_ORDER_UPDATED,
	"name":    apiv1.ListOrder_LIST_ORDER_NAME,
	"open":    apiv1.ListOrder_LIST_ORDER_OPEN,
}

// addListTools registers what can be done to a List itself, as opposed to what is on it.
func addListTools(server *sdk.Server, lists *v1.ListService) {
	sdk.AddTool(server, &sdk.Tool{
		Name: "list_lists",
		Description: "One page of the lists the caller can reach, with how many items are " +
			"still open on each. The last line says whether there are more and how to ask for them.",
	}, func(ctx context.Context, _ *sdk.CallToolRequest, args listListsArgs) (*sdk.CallToolResult, any, error) {
		status, ok := statusWords[args.Status]
		if !ok {
			return errorText(fmt.Errorf(
				"status must be active, completed or archived, not %q", args.Status)), nil, nil
		}
		order, ok := orderWords[args.Order]
		if !ok {
			return errorText(fmt.Errorf(
				"order must be updated, name or open, not %q", args.Order)), nil, nil
		}
		return answer(ctx, lists.ListLists, &apiv1.ListListsRequest{
			Page: int32(args.Page), Status: status, Order: order,
		},
			func(res *apiv1.ListListsResponse) string {
				rows := make([]string, 0, len(res.GetLists())+1)
				for _, list := range res.GetLists() {
					rows = append(rows, listLine(list))
				}
				if where := pageLine(res); where != "" {
					rows = append(rows, "", where)
				}
				return lines(rows, "No lists.")
			})
	})

	sdk.AddTool(server, &sdk.Tool{
		Name: "get_list",
		Description: "What is on one list, ticked and unticked, with each item's identifier. " +
			"A long note is shortened to fit its row; get_item has it in full.",
	}, func(ctx context.Context, _ *sdk.CallToolRequest, args getListArgs) (*sdk.CallToolResult, any, error) {
		return answer(ctx, lists.GetList, &apiv1.GetListRequest{ListUid: args.ListUID},
			func(res *apiv1.GetListResponse) string {
				rows := make([]string, 0, len(res.GetItems())+1)
				rows = append(rows, listLine(res.GetList()))
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
		Name:        "archive_list",
		Description: "Put a list out of the sidebar for everyone, or bring it back. It keeps its items and stays searchable — this is not deleting.",
	}, func(ctx context.Context, _ *sdk.CallToolRequest, args archiveListArgs) (*sdk.CallToolResult, any, error) {
		said := "Archived."
		if !args.Archived {
			said = "Restored to the sidebar."
		}
		return answer(ctx, lists.SetListArchived,
			&apiv1.SetListArchivedRequest{ListUid: args.ListUID, Archived: args.Archived},
			done[apiv1.SetListArchivedResponse](said))
	})

	sdk.AddTool(server, &sdk.Tool{
		Name:        "delete_list",
		Description: "Delete a list and everything on it. Refused unless the caller's token may delete.",
	}, func(ctx context.Context, _ *sdk.CallToolRequest, args listUIDArgs) (*sdk.CallToolResult, any, error) {
		return answer(ctx, lists.DeleteList, &apiv1.DeleteListRequest{ListUid: args.ListUID},
			done[apiv1.DeleteListResponse]("Deleted."))
	})
}
