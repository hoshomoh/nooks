package v1

import (
	"context"
	"errors"
	"strings"
	"time"

	"connectrpc.com/connect"

	"github.com/hoshomoh/nooks/internal/password"
	apiv1 "github.com/hoshomoh/nooks/proto/gen/nooks/api/v1"
	"github.com/hoshomoh/nooks/store"
)

// MemberService covers who is here and how they are grouped.
//
// A Group carries no permissions of its own: it is a shortcut for sharing, so that "the
// flatmates" is one thing to pick rather than three.
type MemberService struct {
	store  store.Store
	now    func() time.Time
	newUID func() (string, error)
}

// NewMemberService builds the service, with the clock and identifiers injected so a
// test needs neither a real clock nor randomness.
func NewMemberService(s store.Store, now func() time.Time, newUID func() (string, error)) *MemberService {
	if now == nil {
		now = time.Now
	}
	if newUID == nil {
		newUID = newMemberUID
	}
	return &MemberService{store: s, now: now, newUID: newUID}
}

// ListMembers returns everyone on the Instance.
//
// Any Member may read it: they share an Instance, and the share dialog has to offer
// somebody to share with.
func (s *MemberService) ListMembers(
	ctx context.Context,
	_ *connect.Request[apiv1.ListMembersRequest],
) (*connect.Response[apiv1.ListMembersResponse], error) {
	if _, err := requireMember(ctx); err != nil {
		return nil, err
	}

	members, err := s.store.Members(ctx)
	if err != nil {
		return nil, internalError("read members", err)
	}

	out := make([]*apiv1.Member, 0, len(members))
	for _, member := range members {
		out = append(out, memberToProto(member))
	}
	return connect.NewResponse(&apiv1.ListMembersResponse{Members: out}), nil
}

// ListGroups returns every Group with who is in it.
func (s *MemberService) ListGroups(
	ctx context.Context,
	_ *connect.Request[apiv1.ListGroupsRequest],
) (*connect.Response[apiv1.ListGroupsResponse], error) {
	if _, err := requireMember(ctx); err != nil {
		return nil, err
	}

	groups, err := s.store.Groups(ctx)
	if err != nil {
		return nil, internalError("read groups", err)
	}

	out := make([]*apiv1.Group, 0, len(groups))
	for _, group := range groups {
		filled, err := s.groupToProto(ctx, group)
		if err != nil {
			return nil, err
		}
		out = append(out, filled)
	}
	return connect.NewResponse(&apiv1.ListGroupsResponse{Groups: out}), nil
}

// errGroupNameRequired reports a Group with nothing to call it.
var errGroupNameRequired = connect.NewError(connect.CodeInvalidArgument,
	errors.New("give the group a name"))

// CreateGroup adds a Group.
func (s *MemberService) CreateGroup(
	ctx context.Context,
	req *connect.Request[apiv1.CreateGroupRequest],
) (*connect.Response[apiv1.CreateGroupResponse], error) {
	if _, err := requireAdmin(ctx); err != nil {
		return nil, err
	}

	name := strings.TrimSpace(req.Msg.GetName())
	if name == "" {
		return nil, errGroupNameRequired
	}

	uid, err := s.newUID()
	if err != nil {
		return nil, internalError("make an identifier", err)
	}
	group, err := s.store.CreateGroup(ctx, uid, name, s.now())
	if err != nil {
		return nil, internalError("create group", err)
	}

	filled, err := s.groupToProto(ctx, group)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&apiv1.CreateGroupResponse{Group: filled}), nil
}

// SetGroupMembers replaces who is in a Group.
//
// Removing someone takes away the Lists they reached through it, and nothing else.
func (s *MemberService) SetGroupMembers(
	ctx context.Context,
	req *connect.Request[apiv1.SetGroupMembersRequest],
) (*connect.Response[apiv1.SetGroupMembersResponse], error) {
	if _, err := requireAdmin(ctx); err != nil {
		return nil, err
	}

	group, err := s.store.GroupByUID(ctx, req.Msg.GetGroupUid())
	if errors.Is(err, store.ErrNotFound) {
		return nil, errNoSuchGroup
	}
	if err != nil {
		return nil, internalError("read group", err)
	}

	ids := make([]int64, 0, len(req.Msg.GetMemberUids()))
	for _, uid := range req.Msg.GetMemberUids() {
		member, err := s.store.MemberByUID(ctx, uid)
		if errors.Is(err, store.ErrNotFound) {
			return nil, errNoSuchMember
		}
		if err != nil {
			return nil, internalError("read member", err)
		}
		ids = append(ids, member.ID)
	}

	if err := s.store.ReplaceGroupMembers(ctx, group.ID, ids); err != nil {
		return nil, internalError("set group members", err)
	}

	filled, err := s.groupToProto(ctx, group)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&apiv1.SetGroupMembersResponse{Group: filled}), nil
}

// groupToProto converts a Group and fills in who is in it.
func (s *MemberService) groupToProto(ctx context.Context, group store.Group) (*apiv1.Group, error) {
	ids, err := s.store.GroupMemberIDs(ctx, group.ID)
	if err != nil {
		return nil, internalError("read group members", err)
	}

	members := make([]*apiv1.Member, 0, len(ids))
	for _, id := range ids {
		member, err := s.store.MemberByID(ctx, id)
		if err != nil {
			return nil, internalError("read member", err)
		}
		members = append(members, memberToProto(member))
	}

	lists, err := s.store.ListsSharedWithGroup(ctx, group.ID)
	if err != nil {
		return nil, internalError("read lists shared with group", err)
	}
	names := make([]string, 0, len(lists))
	for _, list := range lists {
		names = append(names, list.Name)
	}

	return &apiv1.Group{
		Uid: group.UID, Name: group.Name, Members: members, ListNames: names,
	}, nil
}

// errLastAdmin refuses to leave an Instance with nobody who can administer it.
var errLastAdmin = connect.NewError(connect.CodeFailedPrecondition,
	errors.New("an instance needs at least one admin"))

// errCannotRemoveYourself refuses the one removal that cannot be undone by anyone.
var errCannotRemoveYourself = connect.NewError(connect.CodeFailedPrecondition,
	errors.New("you cannot remove your own account"))

// AddMember creates an account with a temporary password, read out once.
//
// Nooks has no mail server, so the Admin hands the password over however they already
// talk to this person. The Member must replace it before anything else.
func (s *MemberService) AddMember(
	ctx context.Context,
	req *connect.Request[apiv1.AddMemberRequest],
) (*connect.Response[apiv1.AddMemberResponse], error) {
	if _, err := requireAdmin(ctx); err != nil {
		return nil, err
	}

	name := strings.TrimSpace(req.Msg.GetName())
	email := strings.TrimSpace(req.Msg.GetEmail())
	if err := requireText(name, "a name"); err != nil {
		return nil, err
	}
	if err := requireText(email, "an email"); err != nil {
		return nil, err
	}

	uid, err := s.newUID()
	if err != nil {
		return nil, internalError("make an identifier", err)
	}

	temporary := password.NewTemporary()
	hash, err := password.Hash(temporary)
	if err != nil {
		return nil, internalError("hash password", err)
	}

	member, err := s.store.CreateMember(ctx, store.CreateMemberParams{
		UID: uid, Name: name, Email: email, Role: store.RoleMember,
		PasswordHash: hash, MustChangePassword: true, CreatedAt: s.now(),
	})
	if err != nil {
		return nil, createMemberError(err)
	}

	// The only moment the password can be read. Nooks keeps the hash and nothing else.
	return connect.NewResponse(&apiv1.AddMemberResponse{
		Member:            memberToProto(member),
		TemporaryPassword: temporary,
	}), nil
}

// SetMemberRole makes somebody an Admin, or stops them being one.
func (s *MemberService) SetMemberRole(
	ctx context.Context,
	req *connect.Request[apiv1.SetMemberRoleRequest],
) (*connect.Response[apiv1.SetMemberRoleResponse], error) {
	admin, err := requireAdmin(ctx)
	if err != nil {
		return nil, err
	}

	member, err := s.memberByUID(ctx, req.Msg.GetMemberUid())
	if err != nil {
		return nil, err
	}
	role := roleFromProto(req.Msg.GetRole())
	if role == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("say which role"))
	}

	// Standing down as the only Admin would lock the Instance for everyone, including
	// the person doing it.
	if role != store.RoleAdmin && member.ID == admin.ID {
		last, err := s.isLastAdmin(ctx, member)
		if err != nil {
			return nil, err
		}
		if last {
			return nil, errLastAdmin
		}
	}

	if err := s.store.SetMemberRole(ctx, member.ID, role); err != nil {
		return nil, internalError("set member role", err)
	}
	member.Role = role
	return connect.NewResponse(&apiv1.SetMemberRoleResponse{Member: memberToProto(member)}), nil
}

// RemoveMember deletes an account. What they added stays on its Lists.
func (s *MemberService) RemoveMember(
	ctx context.Context,
	req *connect.Request[apiv1.RemoveMemberRequest],
) (*connect.Response[apiv1.RemoveMemberResponse], error) {
	admin, err := requireAdmin(ctx)
	if err != nil {
		return nil, err
	}

	member, err := s.memberByUID(ctx, req.Msg.GetMemberUid())
	if err != nil {
		return nil, err
	}
	if member.ID == admin.ID {
		return nil, errCannotRemoveYourself
	}

	last, err := s.isLastAdmin(ctx, member)
	if err != nil {
		return nil, err
	}
	if last {
		return nil, errLastAdmin
	}

	if err := s.store.DeleteMember(ctx, member.ID); err != nil {
		return nil, internalError("remove member", err)
	}
	return connect.NewResponse(&apiv1.RemoveMemberResponse{}), nil
}

// memberByUID finds a Member, reading a bad identifier as a bad argument rather than a
// server fault.
func (s *MemberService) memberByUID(ctx context.Context, uid string) (store.Member, error) {
	member, err := s.store.MemberByUID(ctx, uid)
	if errors.Is(err, store.ErrNotFound) {
		return store.Member{}, errNoSuchMember
	}
	if err != nil {
		return store.Member{}, internalError("read member", err)
	}
	return member, nil
}

// isLastAdmin reports whether this Member is the only one who can administer the
// Instance.
func (s *MemberService) isLastAdmin(ctx context.Context, member store.Member) (bool, error) {
	if !member.IsAdmin() {
		return false, nil
	}
	admins, err := s.store.AdminIDs(ctx)
	if err != nil {
		return false, internalError("read admins", err)
	}
	return len(admins) <= 1, nil
}
