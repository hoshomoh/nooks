package mcp

import (
	"context"
	"fmt"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"

	apiv1 "github.com/hoshomoh/nooks/proto/gen/nooks/api/v1"
	v1 "github.com/hoshomoh/nooks/server/router/api/v1"
)

type listActivityArgs struct{}

type markActivityReadArgs struct{}

type aboutArgs struct{}

// addInstanceTools registers what has been happening and what is running.
func addInstanceTools(server *sdk.Server, services v1.Services) {
	sdk.AddTool(server, &sdk.Tool{
		Name:        "list_activity",
		Description: "What has happened recently on the lists the caller can reach, newest first.",
	}, func(ctx context.Context, _ *sdk.CallToolRequest, _ listActivityArgs) (*sdk.CallToolResult, any, error) {
		return answer(ctx, services.Activity.ListActivity, &apiv1.ListActivityRequest{},
			func(res *apiv1.ListActivityResponse) string {
				rows := make([]string, 0, len(res.GetActivity()))
				for _, entry := range res.GetActivity() {
					rows = append(rows, fmt.Sprintf("%s — %s", entry.GetCreatedAt(), entry.GetText()))
				}
				return lines(rows, "Nothing has happened yet.")
			})
	})

	sdk.AddTool(server, &sdk.Tool{
		Name:        "mark_activity_read",
		Description: "Clear the caller's unread activity count.",
	}, func(ctx context.Context, _ *sdk.CallToolRequest, _ markActivityReadArgs) (*sdk.CallToolResult, any, error) {
		return answer(ctx, services.Activity.MarkActivityRead, &apiv1.MarkActivityReadRequest{},
			done[apiv1.MarkActivityReadResponse]("Marked read."))
	})

	sdk.AddTool(server, &sdk.Tool{
		Name:        "about_instance",
		Description: "What this instance is running: version, licence, what is holding the data, and how much of it there is.",
	}, func(ctx context.Context, _ *sdk.CallToolRequest, _ aboutArgs) (*sdk.CallToolResult, any, error) {
		return answer(ctx, services.Instance.GetInstanceAbout, &apiv1.GetInstanceAboutRequest{},
			func(res *apiv1.GetInstanceAboutResponse) string {
				return fmt.Sprintf("%s, Nooks %s under %s on %s — %d people, %d lists, %d items.",
					res.GetInstanceName(), res.GetVersion(), res.GetLicence(),
					res.GetStorageDriver(), res.GetMemberCount(),
					res.GetListCount(), res.GetItemCount())
			})
	})
}
