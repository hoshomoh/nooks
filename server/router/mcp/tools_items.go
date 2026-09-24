package mcp

import (
	"context"
	"fmt"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"

	apiv1 "github.com/hoshomoh/nooks/proto/gen/nooks/api/v1"
	v1 "github.com/hoshomoh/nooks/server/router/api/v1"
)

type addItemArgs struct {
	ListUID  string `json:"list_uid" jsonschema:"the list to add to, from list_lists"`
	Label    string `json:"label" jsonschema:"what to add, in the words a person would use"`
	Quantity string `json:"quantity,omitempty" jsonschema:"free text like 2 or 1 kg, not a number"`
	DueOn    string `json:"due_on,omitempty" jsonschema:"a day as YYYY-MM-DD, or empty for no date"`
}

// A pointer means "say nothing about this field". Sending an empty string clears the
// value, which is a different instruction from leaving it alone, and an assistant
// changing a quantity should not have to restate the label to do it.
type updateItemArgs struct {
	ItemUID  string  `json:"item_uid" jsonschema:"the item to change, from get_list"`
	Label    *string `json:"label,omitempty" jsonschema:"new wording, or omit to leave it"`
	Quantity *string `json:"quantity,omitempty" jsonschema:"new quantity as free text, empty string to clear it, or omit to leave it"`
	DueOn    *string `json:"due_on,omitempty" jsonschema:"new day as YYYY-MM-DD, empty string to clear it, or omit to leave it"`
	Note     *string `json:"note,omitempty" jsonschema:"the note as markdown, empty string to remove it, or omit to leave it. This replaces the whole note rather than adding to it, so read the item first with get_item. Only what the app draws is rendered: paragraphs, one level of heading (### ), checklists (- [ ] and - [x] ), quotes (> ), fenced code blocks, tables, horizontal rules (---), and inline bold, italic, strikethrough, code spans and links. Bullet and numbered lists are not rendered and survive as the literal characters typed, so use a checklist instead"`

	// Conflict detection, which is offered for text and nothing else: a tick says what
	// an Item should be rather than what it was, so there is nothing to compare.
	ExpectedLabel *string `json:"expected_label,omitempty" jsonschema:"the label as you last read it. Set it and the change is refused if somebody else has since changed the label, rather than overwriting them"`
	ExpectedNote  *string `json:"expected_note,omitempty" jsonschema:"the note as you last read it, in full. Set it and the change is refused if somebody else has since changed the note. Use this whenever you are rewriting a note you read earlier"`
}

type getItemArgs struct {
	ItemUID string `json:"item_uid" jsonschema:"the item to read, from get_list, search or list_dated_items"`
}

type completeItemArgs struct {
	ItemUID string `json:"item_uid" jsonschema:"the item to tick off, from get_list"`
	Done    bool   `json:"done" jsonschema:"true to tick it off, false to put it back"`
}

type itemUIDArgs struct {
	ItemUID string `json:"item_uid" jsonschema:"the item to act on, from get_list"`
}

type moveItemArgs struct {
	ItemUID      string `json:"item_uid" jsonschema:"the item to move, from get_list"`
	AfterItemUID string `json:"after_item_uid,omitempty" jsonschema:"the item to place it after, or empty to put it at the top"`
}

// addItemTools registers what can be done to what is on a List.
func addItemTools(server *sdk.Server, lists *v1.ListService) {
	sdk.AddTool(server, &sdk.Tool{
		Name: "add_item",
		Description: fmt.Sprintf("Put something on a list. A label is at most %d "+
			"characters and a quantity %d. Refused when the caller's token may only read.",
			v1.LimitItemLabel, v1.LimitItemQuantity),
	}, func(ctx context.Context, _ *sdk.CallToolRequest, args addItemArgs) (*sdk.CallToolResult, any, error) {
		return answer(ctx, lists.CreateItem, &apiv1.CreateItemRequest{
			ListUid: args.ListUID, Label: args.Label,
			Quantity: args.Quantity, DueOn: args.DueOn,
		}, func(res *apiv1.CreateItemResponse) string {
			return fmt.Sprintf("Added %q — %s", args.Label, res.GetItem().GetUid())
		})
	})

	sdk.AddTool(server, &sdk.Tool{
		Name: "update_item",
		Description: fmt.Sprintf("Change an item's wording, quantity, due date or note. "+
			"Omitted fields are left alone. A label is at most %d characters, a quantity "+
			"%d and a note %d. Notes are markdown, but only the blocks the app draws: "+
			"paragraphs, ### headings, - [ ] checklists, > quotes, fenced code, tables "+
			"and --- rules. Bullet and numbered lists are not rendered.",
			v1.LimitItemLabel, v1.LimitItemQuantity, v1.LimitItemNote),
	}, func(ctx context.Context, _ *sdk.CallToolRequest, args updateItemArgs) (*sdk.CallToolResult, any, error) {
		return answer(ctx, lists.UpdateItem, &apiv1.UpdateItemRequest{
			ItemUid: args.ItemUID, Label: args.Label,
			Quantity: args.Quantity, DueOn: args.DueOn, Note: args.Note,
			ExpectedLabel: args.ExpectedLabel, ExpectedNote: args.ExpectedNote,
		}, func(res *apiv1.UpdateItemResponse) string {
			return "Changed. " + itemLine(res.GetItem())
		})
	})

	sdk.AddTool(server, &sdk.Tool{
		Name: "get_item",
		Description: "One item with its note in full, and the list it is on. " +
			"get_list shortens a long note to a row, so this is where the rest of it is. " +
			"Read this before rewriting a note, because update_item replaces it whole.",
	}, func(ctx context.Context, _ *sdk.CallToolRequest, args getItemArgs) (*sdk.CallToolResult, any, error) {
		return answer(ctx, lists.GetItem, &apiv1.GetItemRequest{ItemUid: args.ItemUID},
			func(res *apiv1.GetItemResponse) string {
				return itemWhole(res.GetItem(), res.GetList())
			})
	})

	sdk.AddTool(server, &sdk.Tool{
		Name:        "complete_item",
		Description: "Tick an item off, or put it back. Refused when the caller's token may only read.",
	}, func(ctx context.Context, _ *sdk.CallToolRequest, args completeItemArgs) (*sdk.CallToolResult, any, error) {
		said := "Ticked off."
		if !args.Done {
			said = "Put back."
		}
		return answer(ctx, lists.SetItemDone,
			&apiv1.SetItemDoneRequest{ItemUid: args.ItemUID, Done: args.Done},
			done[apiv1.SetItemDoneResponse](said))
	})

	sdk.AddTool(server, &sdk.Tool{
		Name:        "move_item",
		Description: "Reorder an item within its list. Lists keep the order a person put them in, not an order Nooks chose.",
	}, func(ctx context.Context, _ *sdk.CallToolRequest, args moveItemArgs) (*sdk.CallToolResult, any, error) {
		return answer(ctx, lists.MoveItem,
			&apiv1.MoveItemRequest{ItemUid: args.ItemUID, AfterItemUid: args.AfterItemUID},
			done[apiv1.MoveItemResponse]("Moved."))
	})

	sdk.AddTool(server, &sdk.Tool{
		Name:        "delete_item",
		Description: "Take an item off a list for good. Refused unless the caller's token may delete. To tick something off instead, use complete_item.",
	}, func(ctx context.Context, _ *sdk.CallToolRequest, args itemUIDArgs) (*sdk.CallToolResult, any, error) {
		return answer(ctx, lists.DeleteItem, &apiv1.DeleteItemRequest{ItemUid: args.ItemUID},
			done[apiv1.DeleteItemResponse]("Deleted."))
	})
}
