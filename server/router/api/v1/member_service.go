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
	store    store.Store
	now      func() time.Time
	newUID   func() (string, error)
	announce Announcer
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

// WithAnnouncer wires the live stream in. Without one the service still works and
// nobody is told anything, which is what every test but one wants.
func (s *MemberService) WithAnnouncer(a Announcer) *MemberService {
	s.announce = a
	return s
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
	// Resolved before the Group exists, so a bad identifier leaves nothing behind.
	ids, err := memberIDsOf(ctx, s.store, req.Msg.GetMemberUids())
	if err != nil {
		return nil, err
	}

	group, err := s.store.CreateGroup(ctx, uid, name, s.now())
	if err != nil {
		return nil, internalError("create group", err)
	}
	if len(ids) > 0 {
		if err := s.store.ReplaceGroupMembers(ctx, group.ID, ids); err != nil {
			return nil, internalError("set group members", err)
		}
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

	ids, err := memberIDsOf(ctx, s.store, req.Msg.GetMemberUids())
	if err != nil {
		return nil, err
	}

	// Read before the replace: afterwards there is no way to know who was in it.
	was, err := s.store.GroupMemberIDs(ctx, group.ID)
	if err != nil {
		return nil, internalError("read group members", err)
	}

	if err := s.store.ReplaceGroupMembers(ctx, group.ID, ids); err != nil {
		return nil, internalError("set group members", err)
	}

	/*
	 * Joining or leaving a Group changes which Lists somebody reaches, and nothing else
	 * tells them. A List shared with a Group is not itself changed by this, so the
	 * ListChanged that would normally carry the news is never sent.
	 *
	 * Both directions: whoever left has lost whatever the Group reached, and whoever
	 * joined has gained it. Somebody whose membership did not change is left alone.
	 * Announced whether or not the Group reaches a List, because the answer costs a
	 * read and being told to look again costs one refetch.
	 */
	if s.announce != nil {
		if changed := differing(was, ids); len(changed) > 0 {
			s.announce.ListsChanged(changed)
		}
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

	// Their browser read who it was signed in as when the tab opened and has held the
	// answer ever since, so this is the only thing that will tell it otherwise.
	if s.announce != nil {
		s.announce.MemberChanged(member.ID)
	}
	return connect.NewResponse(&apiv1.SetMemberRoleResponse{Member: memberToProto(member)}), nil
}

/*
errPickWhatBecomesOfTheirLists refuses a removal that did not say.

Naming nobody and not asking for deletion is a request that forgot, and the answer to a
request that forgot must not be the one that destroys a household's shopping.
*/
var errPickWhatBecomesOfTheirLists = connect.NewError(connect.CodeInvalidArgument,
	errors.New("say who the lists they started go to, or that they are to be deleted"))

// errCannotDoBoth refuses a removal that asked for both at once.
var errCannotDoBoth = connect.NewError(connect.CodeInvalidArgument,
	errors.New("their lists go to somebody or are deleted, not both"))

// errCannotInheritThemselves refuses handing somebody's Lists to the person leaving.
var errCannotInheritThemselves = connect.NewError(connect.CodeInvalidArgument,
	errors.New("their lists cannot be given to the member being removed"))

/*
settleTheirLists does what the Admin said with the Lists a removed Member started.

Their own Lists used to go by cascade, because removing a Member deleted the row. The
row now stays, so that what they added to everybody else's Lists is not carried off with
it, and that leaves these Lists nothing's job until somebody takes it.

Giving them away is one statement. Deleting them is soft, where the cascade was hard, so
what was on one is still in the database rather than gone from it.

Done before the Member is emptied, while the row is still theirs to find the Lists by.
*/
func (s *MemberService) settleTheirLists(
	ctx context.Context,
	member store.Member,
	msg *apiv1.RemoveMemberRequest,
) error {
	givingTo := msg.GetGiveListsToUid()
	deleting := msg.GetDeleteTheirLists()

	switch {
	case givingTo == "" && !deleting:
		return errPickWhatBecomesOfTheirLists
	case givingTo != "" && deleting:
		return errCannotDoBoth
	}

	if deleting {
		owned, err := s.store.ListsOwnedBy(ctx, member.ID)
		if err != nil {
			return internalError("read the lists a member owns", err)
		}
		for _, list := range owned {
			if err := s.store.DeleteList(ctx, list.UID, s.now()); err != nil {
				return internalError("delete a removed member's list", err)
			}
		}
		return nil
	}

	heir, err := s.memberByUID(ctx, givingTo)
	if err != nil {
		return err
	}
	if heir.ID == member.ID {
		return errCannotInheritThemselves
	}
	return s.store.GiveListsTo(ctx, member.ID, heir.ID)
}

/*
RemoveMember deletes an account and everything of theirs.

Their Lists go with them and so does everything they put on anybody else's — see
store.DeleteMember, which says which columns do it and why the alternatives each cost
something. The dialog tells the Admin as much before they press it.

Refused for the last Admin and for yourself: the first would leave an Instance nobody
can administer, and the second is the one removal nobody else can undo for you.
*/
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

	if err := s.settleTheirLists(ctx, member, req.Msg); err != nil {
		return nil, err
	}

	if err := s.store.RemoveMember(ctx, member.ID, s.now()); err != nil {
		return nil, internalError("remove member", err)
	}

	/*
	 * Everyone left, rather than a worked-out set.
	 *
	 * What goes with a removed Member could have been shared with anybody — the whole
	 * Instance, named Members, or a Group they were in — and their Lists are deleted
	 * outright rather than merely put out of reach. Working out exactly whose sidebar
	 * changed means reading every List they owned and every share on each.
	 *
	 * Removing somebody is a rare thing an Admin does on purpose, and a household is
	 * not a number that runs away. One refetch each is cheaper than a set that is
	 * subtly wrong.
	 */
	if s.announce != nil {
		if remaining, err := everyone(ctx, s.store); err == nil && len(remaining) > 0 {
			s.announce.ListsChanged(remaining)
		}
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

/*
UpdateOwnProfile changes the signed-in Member's own name and email.

Their own, and only their own. An Admin who wants somebody else's name changed asks
them: an account is a person, and Nooks does not let one person edit another.

A browser, never a token: a key that reaches somebody's Lists must not be a way to
change the address their account signs in with.
*/
func (s *MemberService) UpdateOwnProfile(
	ctx context.Context,
	req *connect.Request[apiv1.UpdateOwnProfileRequest],
) (*connect.Response[apiv1.UpdateOwnProfileResponse], error) {
	member, err := requireBrowser(ctx)
	if err != nil {
		return nil, err
	}

	name := strings.TrimSpace(req.Msg.GetName())
	if err := requireText(name, "a name"); err != nil {
		return nil, err
	}
	email := strings.TrimSpace(req.Msg.GetEmail())
	if err := requireText(email, "an email"); err != nil {
		return nil, err
	}

	if err := s.store.SetMemberProfile(ctx, member.ID, name, email); err != nil {
		if errors.Is(err, store.ErrEmailTaken) {
			return nil, errEmailTaken
		}
		return nil, internalError("update profile", err)
	}

	updated, err := s.store.MemberByID(ctx, member.ID)
	if err != nil {
		return nil, internalError("read member", err)
	}
	return connect.NewResponse(&apiv1.UpdateOwnProfileResponse{
		Member: memberToProto(updated),
	}), nil
}

// differing is the Members in one set and not the other, either way round.
func differing(was, now []int64) []int64 {
	in := func(ids []int64) map[int64]bool {
		held := make(map[int64]bool, len(ids))
		for _, id := range ids {
			held[id] = true
		}
		return held
	}
	before, after := in(was), in(now)

	changed := make([]int64, 0, len(was)+len(now))
	for _, id := range was {
		if !after[id] {
			changed = append(changed, id)
		}
	}
	for _, id := range now {
		if !before[id] {
			changed = append(changed, id)
		}
	}
	return changed
}
