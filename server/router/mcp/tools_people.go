package mcp

import (
	"context"
	"fmt"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"

	apiv1 "github.com/hoshomoh/nooks/proto/gen/nooks/api/v1"
	v1 "github.com/hoshomoh/nooks/server/router/api/v1"
)

type whoamiArgs struct{}

type listMembersArgs struct{}

type listGroupsArgs struct{}

type setSharingArgs struct {
	ListUID    string   `json:"list_uid" jsonschema:"the list to share, from list_lists"`
	Sharing    string   `json:"sharing" jsonschema:"private for the owner alone, instance for everyone here, or specific for named people and groups"`
	CanEdit    bool     `json:"can_edit" jsonschema:"true to let them tick and add, false to let them only read and print"`
	MemberUIDs []string `json:"member_uids,omitempty" jsonschema:"people it reaches by name, from list_members, for specific sharing"`
	GroupUIDs  []string `json:"group_uids,omitempty" jsonschema:"groups it reaches, from list_groups, for specific sharing"`
}

// sharingWords maps what an assistant can say onto what the API takes. The enum's own
// spelling is a protocol detail, and a tool that demanded SHARING_SPECIFIC would be
// asking a reader to learn one.
var sharingWords = map[string]apiv1.Sharing{
	"private":  apiv1.Sharing_SHARING_PRIVATE,
	"instance": apiv1.Sharing_SHARING_INSTANCE,
	"specific": apiv1.Sharing_SHARING_SPECIFIC,
}

// addPeopleTools registers who is here and who a List reaches.
func addPeopleTools(server *sdk.Server, services v1.Services) {
	sdk.AddTool(server, &sdk.Tool{
		Name:        "whoami",
		Description: "Who the caller is signed in as, and whether they are an admin.",
	}, func(ctx context.Context, _ *sdk.CallToolRequest, _ whoamiArgs) (*sdk.CallToolResult, any, error) {
		return answer(ctx, services.Auth.GetCurrentMember, &apiv1.GetCurrentMemberRequest{},
			func(res *apiv1.GetCurrentMemberResponse) string {
				return memberLine(res.GetMember())
			})
	})

	sdk.AddTool(server, &sdk.Tool{
		Name:        "list_members",
		Description: "Everyone on this instance, with the identifier set_list_sharing needs.",
	}, func(ctx context.Context, _ *sdk.CallToolRequest, _ listMembersArgs) (*sdk.CallToolResult, any, error) {
		return answer(ctx, services.Member.ListMembers, &apiv1.ListMembersRequest{},
			func(res *apiv1.ListMembersResponse) string {
				rows := make([]string, 0, len(res.GetMembers()))
				for _, member := range res.GetMembers() {
					rows = append(rows, memberLine(member))
				}
				return lines(rows, "Nobody here.")
			})
	})

	sdk.AddTool(server, &sdk.Tool{
		Name:        "list_groups",
		Description: "The groups on this instance and who is in them. A group is only ever a shortcut for sharing.",
	}, func(ctx context.Context, _ *sdk.CallToolRequest, _ listGroupsArgs) (*sdk.CallToolResult, any, error) {
		return answer(ctx, services.Member.ListGroups, &apiv1.ListGroupsRequest{},
			func(res *apiv1.ListGroupsResponse) string {
				rows := make([]string, 0, len(res.GetGroups()))
				for _, group := range res.GetGroups() {
					rows = append(rows, fmt.Sprintf("%s (%s) — %d people",
						group.GetName(), group.GetUid(), len(group.GetMembers())))
				}
				return lines(rows, "No groups.")
			})
	})

	sdk.AddTool(server, &sdk.Tool{
		Name:        "get_list_shares",
		Description: "Which people and groups a list reaches by name.",
	}, func(ctx context.Context, _ *sdk.CallToolRequest, args listUIDArgs) (*sdk.CallToolResult, any, error) {
		return answer(ctx, services.List.GetListShares,
			&apiv1.GetListSharesRequest{ListUid: args.ListUID},
			func(res *apiv1.GetListSharesResponse) string {
				return fmt.Sprintf("%d people, %d groups: %v %v",
					len(res.GetMemberUids()), len(res.GetGroupUids()),
					res.GetMemberUids(), res.GetGroupUids())
			})
	})

	sdk.AddTool(server, &sdk.Tool{
		Name:        "set_list_sharing",
		Description: "Change who a list reaches. Only its owner may. Named people and groups replace whatever was there, because sharing is one decision rather than a sequence of additions.",
	}, func(ctx context.Context, _ *sdk.CallToolRequest, args setSharingArgs) (*sdk.CallToolResult, any, error) {
		sharing, ok := sharingWords[args.Sharing]
		if !ok {
			return errorText(fmt.Errorf(
				"sharing must be private, instance or specific, not %q", args.Sharing)), nil, nil
		}
		return answer(ctx, services.List.SetListSharing, &apiv1.SetListSharingRequest{
			ListUid: args.ListUID, Sharing: sharing, CanEdit: args.CanEdit,
			MemberUids: args.MemberUIDs, GroupUids: args.GroupUIDs,
		}, func(res *apiv1.SetListSharingResponse) string {
			return fmt.Sprintf("%q is now shared: %s.", res.GetList().GetName(), args.Sharing)
		})
	})
}
