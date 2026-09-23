// Package gateway serves the REST API by handing requests to the same services the
// browser talks to.
//
// grpc-gateway generates a proxy that calls a gRPC-shaped server interface, and Nooks'
// services are Connect-shaped. These adapters are that difference and nothing else:
// build the request the service expects, carrying what the call arrived with, then
// hand back the message. Only the request half travels — a Set-Cookie on the way out
// has nowhere to go, which is why the ceremonies that hand over a session have no REST
// route at all.
//
// It is boilerplate, and the compiler does not check it: each adapter embeds an
// UnimplementedXServiceServer, so an RPC with no line here still builds and answers
// "unimplemented" over REST while working perfectly in the app. scripts/check-gateway.sh
// is what actually catches that, and runs in CI.
package gateway

import (
	"context"

	apiv1 "github.com/hoshomoh/nooks/proto/gen/nooks/api/v1"
	v1 "github.com/hoshomoh/nooks/server/router/api/v1"
)

type activityService struct {
	apiv1.UnimplementedActivityServiceServer
	svc *v1.ActivityService
}

func (g activityService) ListActivity(ctx context.Context, req *apiv1.ListActivityRequest) (*apiv1.ListActivityResponse, error) {
	res, err := g.svc.ListActivity(ctx, requestFrom(ctx, req))
	if err != nil {
		return nil, asStatus(err)
	}
	return res.Msg, nil
}

func (g activityService) MarkActivityRead(ctx context.Context, req *apiv1.MarkActivityReadRequest) (*apiv1.MarkActivityReadResponse, error) {
	res, err := g.svc.MarkActivityRead(ctx, requestFrom(ctx, req))
	if err != nil {
		return nil, asStatus(err)
	}
	return res.Msg, nil
}

type authService struct {
	apiv1.UnimplementedAuthServiceServer
	svc *v1.AuthService
}

func (g authService) CompleteSetup(ctx context.Context, req *apiv1.CompleteSetupRequest) (*apiv1.CompleteSetupResponse, error) {
	res, err := g.svc.CompleteSetup(ctx, requestFrom(ctx, req))
	if err != nil {
		return nil, asStatus(err)
	}
	return res.Msg, nil
}

func (g authService) SignIn(ctx context.Context, req *apiv1.SignInRequest) (*apiv1.SignInResponse, error) {
	res, err := g.svc.SignIn(ctx, requestFrom(ctx, req))
	if err != nil {
		return nil, asStatus(err)
	}
	return res.Msg, nil
}

func (g authService) SignOut(ctx context.Context, req *apiv1.SignOutRequest) (*apiv1.SignOutResponse, error) {
	res, err := g.svc.SignOut(ctx, requestFrom(ctx, req))
	if err != nil {
		return nil, asStatus(err)
	}
	return res.Msg, nil
}

func (g authService) RefreshAccess(ctx context.Context, req *apiv1.RefreshAccessRequest) (*apiv1.RefreshAccessResponse, error) {
	res, err := g.svc.RefreshAccess(ctx, requestFrom(ctx, req))
	if err != nil {
		return nil, asStatus(err)
	}
	return res.Msg, nil
}

func (g authService) GetCurrentMember(ctx context.Context, req *apiv1.GetCurrentMemberRequest) (*apiv1.GetCurrentMemberResponse, error) {
	res, err := g.svc.GetCurrentMember(ctx, requestFrom(ctx, req))
	if err != nil {
		return nil, asStatus(err)
	}
	return res.Msg, nil
}

func (g authService) ReplacePassword(ctx context.Context, req *apiv1.ReplacePasswordRequest) (*apiv1.ReplacePasswordResponse, error) {
	res, err := g.svc.ReplacePassword(ctx, requestFrom(ctx, req))
	if err != nil {
		return nil, asStatus(err)
	}
	return res.Msg, nil
}

func (g authService) RequestJoin(ctx context.Context, req *apiv1.RequestJoinRequest) (*apiv1.RequestJoinResponse, error) {
	res, err := g.svc.RequestJoin(ctx, requestFrom(ctx, req))
	if err != nil {
		return nil, asStatus(err)
	}
	return res.Msg, nil
}

func (g authService) GetJoinRequest(ctx context.Context, req *apiv1.GetJoinRequestRequest) (*apiv1.GetJoinRequestResponse, error) {
	res, err := g.svc.GetJoinRequest(ctx, requestFrom(ctx, req))
	if err != nil {
		return nil, asStatus(err)
	}
	return res.Msg, nil
}

func (g authService) CompleteJoin(ctx context.Context, req *apiv1.CompleteJoinRequest) (*apiv1.CompleteJoinResponse, error) {
	res, err := g.svc.CompleteJoin(ctx, requestFrom(ctx, req))
	if err != nil {
		return nil, asStatus(err)
	}
	return res.Msg, nil
}

func (g authService) RequestPasswordReset(ctx context.Context, req *apiv1.RequestPasswordResetRequest) (*apiv1.RequestPasswordResetResponse, error) {
	res, err := g.svc.RequestPasswordReset(ctx, requestFrom(ctx, req))
	if err != nil {
		return nil, asStatus(err)
	}
	return res.Msg, nil
}

func (g authService) GetResetRequest(ctx context.Context, req *apiv1.GetResetRequestRequest) (*apiv1.GetResetRequestResponse, error) {
	res, err := g.svc.GetResetRequest(ctx, requestFrom(ctx, req))
	if err != nil {
		return nil, asStatus(err)
	}
	return res.Msg, nil
}

func (g authService) CompletePasswordReset(ctx context.Context, req *apiv1.CompletePasswordResetRequest) (*apiv1.CompletePasswordResetResponse, error) {
	res, err := g.svc.CompletePasswordReset(ctx, requestFrom(ctx, req))
	if err != nil {
		return nil, asStatus(err)
	}
	return res.Msg, nil
}

type instanceService struct {
	apiv1.UnimplementedInstanceServiceServer
	svc *v1.InstanceService
}

func (g instanceService) GetInstance(ctx context.Context, req *apiv1.GetInstanceRequest) (*apiv1.GetInstanceResponse, error) {
	res, err := g.svc.GetInstance(ctx, requestFrom(ctx, req))
	if err != nil {
		return nil, asStatus(err)
	}
	return res.Msg, nil
}

func (g instanceService) GetInstanceSettings(ctx context.Context, req *apiv1.GetInstanceSettingsRequest) (*apiv1.GetInstanceSettingsResponse, error) {
	res, err := g.svc.GetInstanceSettings(ctx, requestFrom(ctx, req))
	if err != nil {
		return nil, asStatus(err)
	}
	return res.Msg, nil
}

func (g instanceService) UpdateInstanceSettings(ctx context.Context, req *apiv1.UpdateInstanceSettingsRequest) (*apiv1.UpdateInstanceSettingsResponse, error) {
	res, err := g.svc.UpdateInstanceSettings(ctx, requestFrom(ctx, req))
	if err != nil {
		return nil, asStatus(err)
	}
	return res.Msg, nil
}

func (g instanceService) GetInstanceAbout(ctx context.Context, req *apiv1.GetInstanceAboutRequest) (*apiv1.GetInstanceAboutResponse, error) {
	res, err := g.svc.GetInstanceAbout(ctx, requestFrom(ctx, req))
	if err != nil {
		return nil, asStatus(err)
	}
	return res.Msg, nil
}

func (g instanceService) DeleteInstance(ctx context.Context, req *apiv1.DeleteInstanceRequest) (*apiv1.DeleteInstanceResponse, error) {
	res, err := g.svc.DeleteInstance(ctx, requestFrom(ctx, req))
	if err != nil {
		return nil, asStatus(err)
	}
	return res.Msg, nil
}

type listService struct {
	apiv1.UnimplementedListServiceServer
	svc *v1.ListService
}

func (g listService) ListLists(ctx context.Context, req *apiv1.ListListsRequest) (*apiv1.ListListsResponse, error) {
	res, err := g.svc.ListLists(ctx, requestFrom(ctx, req))
	if err != nil {
		return nil, asStatus(err)
	}
	return res.Msg, nil
}

func (g listService) GetSidebar(ctx context.Context, req *apiv1.GetSidebarRequest) (*apiv1.GetSidebarResponse, error) {
	res, err := g.svc.GetSidebar(ctx, requestFrom(ctx, req))
	if err != nil {
		return nil, asStatus(err)
	}
	return res.Msg, nil
}

func (g listService) GetList(ctx context.Context, req *apiv1.GetListRequest) (*apiv1.GetListResponse, error) {
	res, err := g.svc.GetList(ctx, requestFrom(ctx, req))
	if err != nil {
		return nil, asStatus(err)
	}
	return res.Msg, nil
}

func (g listService) GetItem(ctx context.Context, req *apiv1.GetItemRequest) (*apiv1.GetItemResponse, error) {
	res, err := g.svc.GetItem(ctx, requestFrom(ctx, req))
	if err != nil {
		return nil, asStatus(err)
	}
	return res.Msg, nil
}

func (g listService) CreateList(ctx context.Context, req *apiv1.CreateListRequest) (*apiv1.CreateListResponse, error) {
	res, err := g.svc.CreateList(ctx, requestFrom(ctx, req))
	if err != nil {
		return nil, asStatus(err)
	}
	return res.Msg, nil
}

func (g listService) RenameList(ctx context.Context, req *apiv1.RenameListRequest) (*apiv1.RenameListResponse, error) {
	res, err := g.svc.RenameList(ctx, requestFrom(ctx, req))
	if err != nil {
		return nil, asStatus(err)
	}
	return res.Msg, nil
}

func (g listService) SetListArchived(ctx context.Context, req *apiv1.SetListArchivedRequest) (*apiv1.SetListArchivedResponse, error) {
	res, err := g.svc.SetListArchived(ctx, requestFrom(ctx, req))
	if err != nil {
		return nil, asStatus(err)
	}
	return res.Msg, nil
}

func (g listService) SetListSharing(ctx context.Context, req *apiv1.SetListSharingRequest) (*apiv1.SetListSharingResponse, error) {
	res, err := g.svc.SetListSharing(ctx, requestFrom(ctx, req))
	if err != nil {
		return nil, asStatus(err)
	}
	return res.Msg, nil
}

func (g listService) GetListShares(ctx context.Context, req *apiv1.GetListSharesRequest) (*apiv1.GetListSharesResponse, error) {
	res, err := g.svc.GetListShares(ctx, requestFrom(ctx, req))
	if err != nil {
		return nil, asStatus(err)
	}
	return res.Msg, nil
}

func (g listService) DeleteList(ctx context.Context, req *apiv1.DeleteListRequest) (*apiv1.DeleteListResponse, error) {
	res, err := g.svc.DeleteList(ctx, requestFrom(ctx, req))
	if err != nil {
		return nil, asStatus(err)
	}
	return res.Msg, nil
}

func (g listService) RestoreList(ctx context.Context, req *apiv1.RestoreListRequest) (*apiv1.RestoreListResponse, error) {
	res, err := g.svc.RestoreList(ctx, requestFrom(ctx, req))
	if err != nil {
		return nil, asStatus(err)
	}
	return res.Msg, nil
}

func (g listService) DuplicateList(ctx context.Context, req *apiv1.DuplicateListRequest) (*apiv1.DuplicateListResponse, error) {
	res, err := g.svc.DuplicateList(ctx, requestFrom(ctx, req))
	if err != nil {
		return nil, asStatus(err)
	}
	return res.Msg, nil
}

func (g listService) SetListPinned(ctx context.Context, req *apiv1.SetListPinnedRequest) (*apiv1.SetListPinnedResponse, error) {
	res, err := g.svc.SetListPinned(ctx, requestFrom(ctx, req))
	if err != nil {
		return nil, asStatus(err)
	}
	return res.Msg, nil
}

func (g listService) CreateItem(ctx context.Context, req *apiv1.CreateItemRequest) (*apiv1.CreateItemResponse, error) {
	res, err := g.svc.CreateItem(ctx, requestFrom(ctx, req))
	if err != nil {
		return nil, asStatus(err)
	}
	return res.Msg, nil
}

func (g listService) UpdateItem(ctx context.Context, req *apiv1.UpdateItemRequest) (*apiv1.UpdateItemResponse, error) {
	res, err := g.svc.UpdateItem(ctx, requestFrom(ctx, req))
	if err != nil {
		return nil, asStatus(err)
	}
	return res.Msg, nil
}

func (g listService) SetItemDone(ctx context.Context, req *apiv1.SetItemDoneRequest) (*apiv1.SetItemDoneResponse, error) {
	res, err := g.svc.SetItemDone(ctx, requestFrom(ctx, req))
	if err != nil {
		return nil, asStatus(err)
	}
	return res.Msg, nil
}

func (g listService) MoveItem(ctx context.Context, req *apiv1.MoveItemRequest) (*apiv1.MoveItemResponse, error) {
	res, err := g.svc.MoveItem(ctx, requestFrom(ctx, req))
	if err != nil {
		return nil, asStatus(err)
	}
	return res.Msg, nil
}

func (g listService) DeleteItem(ctx context.Context, req *apiv1.DeleteItemRequest) (*apiv1.DeleteItemResponse, error) {
	res, err := g.svc.DeleteItem(ctx, requestFrom(ctx, req))
	if err != nil {
		return nil, asStatus(err)
	}
	return res.Msg, nil
}

func (g listService) ListDatedItems(ctx context.Context, req *apiv1.ListDatedItemsRequest) (*apiv1.ListDatedItemsResponse, error) {
	res, err := g.svc.ListDatedItems(ctx, requestFrom(ctx, req))
	if err != nil {
		return nil, asStatus(err)
	}
	return res.Msg, nil
}

func (g listService) Search(ctx context.Context, req *apiv1.SearchRequest) (*apiv1.SearchResponse, error) {
	res, err := g.svc.Search(ctx, requestFrom(ctx, req))
	if err != nil {
		return nil, asStatus(err)
	}
	return res.Msg, nil
}

type memberService struct {
	apiv1.UnimplementedMemberServiceServer
	svc *v1.MemberService
}

func (g memberService) ListMembers(ctx context.Context, req *apiv1.ListMembersRequest) (*apiv1.ListMembersResponse, error) {
	res, err := g.svc.ListMembers(ctx, requestFrom(ctx, req))
	if err != nil {
		return nil, asStatus(err)
	}
	return res.Msg, nil
}

func (g memberService) ListGroups(ctx context.Context, req *apiv1.ListGroupsRequest) (*apiv1.ListGroupsResponse, error) {
	res, err := g.svc.ListGroups(ctx, requestFrom(ctx, req))
	if err != nil {
		return nil, asStatus(err)
	}
	return res.Msg, nil
}

func (g memberService) CreateGroup(ctx context.Context, req *apiv1.CreateGroupRequest) (*apiv1.CreateGroupResponse, error) {
	res, err := g.svc.CreateGroup(ctx, requestFrom(ctx, req))
	if err != nil {
		return nil, asStatus(err)
	}
	return res.Msg, nil
}

func (g memberService) SetGroupMembers(ctx context.Context, req *apiv1.SetGroupMembersRequest) (*apiv1.SetGroupMembersResponse, error) {
	res, err := g.svc.SetGroupMembers(ctx, requestFrom(ctx, req))
	if err != nil {
		return nil, asStatus(err)
	}
	return res.Msg, nil
}

func (g memberService) AddMember(ctx context.Context, req *apiv1.AddMemberRequest) (*apiv1.AddMemberResponse, error) {
	res, err := g.svc.AddMember(ctx, requestFrom(ctx, req))
	if err != nil {
		return nil, asStatus(err)
	}
	return res.Msg, nil
}

func (g memberService) SetMemberRole(ctx context.Context, req *apiv1.SetMemberRoleRequest) (*apiv1.SetMemberRoleResponse, error) {
	res, err := g.svc.SetMemberRole(ctx, requestFrom(ctx, req))
	if err != nil {
		return nil, asStatus(err)
	}
	return res.Msg, nil
}

func (g memberService) RemoveMember(ctx context.Context, req *apiv1.RemoveMemberRequest) (*apiv1.RemoveMemberResponse, error) {
	res, err := g.svc.RemoveMember(ctx, requestFrom(ctx, req))
	if err != nil {
		return nil, asStatus(err)
	}
	return res.Msg, nil
}

func (g memberService) UpdateOwnProfile(ctx context.Context, req *apiv1.UpdateOwnProfileRequest) (*apiv1.UpdateOwnProfileResponse, error) {
	res, err := g.svc.UpdateOwnProfile(ctx, requestFrom(ctx, req))
	if err != nil {
		return nil, asStatus(err)
	}
	return res.Msg, nil
}

type publicService struct {
	apiv1.UnimplementedPublicServiceServer
	svc *v1.PublicService
}

func (g publicService) GetPublicList(ctx context.Context, req *apiv1.GetPublicListRequest) (*apiv1.GetPublicListResponse, error) {
	res, err := g.svc.GetPublicList(ctx, requestFrom(ctx, req))
	if err != nil {
		return nil, asStatus(err)
	}
	return res.Msg, nil
}

type requestService struct {
	apiv1.UnimplementedRequestServiceServer
	svc *v1.RequestService
}

func (g requestService) ListPendingRequests(ctx context.Context, req *apiv1.ListPendingRequestsRequest) (*apiv1.ListPendingRequestsResponse, error) {
	res, err := g.svc.ListPendingRequests(ctx, requestFrom(ctx, req))
	if err != nil {
		return nil, asStatus(err)
	}
	return res.Msg, nil
}

func (g requestService) DecideJoinRequest(ctx context.Context, req *apiv1.DecideJoinRequestRequest) (*apiv1.DecideJoinRequestResponse, error) {
	res, err := g.svc.DecideJoinRequest(ctx, requestFrom(ctx, req))
	if err != nil {
		return nil, asStatus(err)
	}
	return res.Msg, nil
}

func (g requestService) DecideResetRequest(ctx context.Context, req *apiv1.DecideResetRequestRequest) (*apiv1.DecideResetRequestResponse, error) {
	res, err := g.svc.DecideResetRequest(ctx, requestFrom(ctx, req))
	if err != nil {
		return nil, asStatus(err)
	}
	return res.Msg, nil
}

type tokenService struct {
	apiv1.UnimplementedTokenServiceServer
	svc *v1.TokenService
}

func (g tokenService) ListAccessTokens(ctx context.Context, req *apiv1.ListAccessTokensRequest) (*apiv1.ListAccessTokensResponse, error) {
	res, err := g.svc.ListAccessTokens(ctx, requestFrom(ctx, req))
	if err != nil {
		return nil, asStatus(err)
	}
	return res.Msg, nil
}

func (g tokenService) CreateAccessToken(ctx context.Context, req *apiv1.CreateAccessTokenRequest) (*apiv1.CreateAccessTokenResponse, error) {
	res, err := g.svc.CreateAccessToken(ctx, requestFrom(ctx, req))
	if err != nil {
		return nil, asStatus(err)
	}
	return res.Msg, nil
}

func (g tokenService) RevokeAccessToken(ctx context.Context, req *apiv1.RevokeAccessTokenRequest) (*apiv1.RevokeAccessTokenResponse, error) {
	res, err := g.svc.RevokeAccessToken(ctx, requestFrom(ctx, req))
	if err != nil {
		return nil, asStatus(err)
	}
	return res.Msg, nil
}
