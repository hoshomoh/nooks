package v1

import (
	"context"
	"errors"
	"fmt"
	"time"

	"connectrpc.com/connect"

	"github.com/hoshomoh/nooks/internal/password"
	apiv1 "github.com/hoshomoh/nooks/proto/gen/nooks/api/v1"
	"github.com/hoshomoh/nooks/server/auth"
	"github.com/hoshomoh/nooks/store"
)

// AuthService covers first run, signing in and out, and replacing a password.
type AuthService struct {
	store store.Store
	// now and newUID are injected so that sessions, timestamps and identifiers are all
	// deterministic in a test.
	now     func() time.Time
	newUID  func() (string, error)
	secure  bool
	newAuth func() (token string, hash string, err error)
}

// AuthServiceOptions carries what the service needs from its surroundings.
type AuthServiceOptions struct {
	// Secure marks session cookies as Secure. It is a deployment fact, so the server
	// passes it in rather than this package guessing.
	Secure bool
	// Now defaults to time.Now.
	Now func() time.Time
	// NewUID defaults to a random identifier.
	NewUID func() (string, error)
	// NewToken defaults to auth.NewToken.
	NewToken func() (token string, hash string, err error)
}

// NewAuthService builds the service.
func NewAuthService(s store.Store, opts AuthServiceOptions) *AuthService {
	svc := &AuthService{
		store:   s,
		now:     opts.Now,
		newUID:  opts.NewUID,
		secure:  opts.Secure,
		newAuth: opts.NewToken,
	}
	if svc.now == nil {
		svc.now = time.Now
	}
	if svc.newUID == nil {
		svc.newUID = newMemberUID
	}
	if svc.newAuth == nil {
		svc.newAuth = auth.NewToken
	}
	return svc
}

// CompleteSetup creates the first Admin and names the Instance.
//
// It is only available while there are no Members. Once one exists this is closed
// permanently — otherwise anyone who reached the Instance first could seize it.
func (s *AuthService) CompleteSetup(
	ctx context.Context,
	req *connect.Request[apiv1.CompleteSetupRequest],
) (*connect.Response[apiv1.CompleteSetupResponse], error) {
	msg := req.Msg
	if err := s.checkSetupIsOpen(ctx, msg); err != nil {
		return nil, err
	}

	member, err := s.createMember(ctx, createMemberInput{
		name:     msg.GetName(),
		email:    msg.GetEmail(),
		plain:    msg.GetPassword(),
		role:     store.RoleAdmin,
		mustPick: false,
	})
	if err != nil {
		return nil, err
	}

	settings := store.InstanceSettings{
		Name:             msg.GetInstanceName(),
		SetupCompletedAt: s.now(),
	}
	if err := s.store.SaveInstanceSettings(ctx, settings); err != nil {
		return nil, internalError("save instance settings", err)
	}

	res := connect.NewResponse(&apiv1.CompleteSetupResponse{Member: memberToProto(member)})
	if err := s.startSession(ctx, res.Header(), member); err != nil {
		return nil, err
	}
	return res, nil
}

// SignIn exchanges an email and password for a session cookie.
func (s *AuthService) SignIn(
	ctx context.Context,
	req *connect.Request[apiv1.SignInRequest],
) (*connect.Response[apiv1.SignInResponse], error) {
	member, err := s.store.MemberByEmail(ctx, req.Msg.GetEmail())
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return nil, errWrongCredentials
		}
		return nil, internalError("read member", err)
	}

	if err := password.Verify(member.PasswordHash, req.Msg.GetPassword()); err != nil {
		if errors.Is(err, password.ErrWrong) {
			return nil, errWrongCredentials
		}
		return nil, internalError("verify password", err)
	}

	if err := s.store.MarkMemberSignedIn(ctx, member.ID, s.now()); err != nil {
		return nil, internalError("mark member signed in", err)
	}

	res := connect.NewResponse(&apiv1.SignInResponse{Member: memberToProto(member)})
	if err := s.startSession(ctx, res.Header(), member); err != nil {
		return nil, err
	}
	return res, nil
}

// SignOut ends the current session.
func (s *AuthService) SignOut(
	ctx context.Context,
	req *connect.Request[apiv1.SignOutRequest],
) (*connect.Response[apiv1.SignOutResponse], error) {
	res := connect.NewResponse(&apiv1.SignOutResponse{})
	res.Header().Add("Set-Cookie", auth.ExpiredCookie(s.secure).String())

	// Signing out twice is not a failure, so a missing or unknown cookie is fine.
	if token := cookieValue(req.Header().Get("Cookie"), auth.CookieName); token != "" {
		if err := s.store.DeleteSession(ctx, auth.HashToken(token)); err != nil {
			return nil, internalError("delete session", err)
		}
	}
	return res, nil
}

// GetCurrentMember returns whoever the session cookie belongs to.
func (s *AuthService) GetCurrentMember(
	ctx context.Context,
	_ *connect.Request[apiv1.GetCurrentMemberRequest],
) (*connect.Response[apiv1.GetCurrentMemberResponse], error) {
	member, ok := auth.MemberFrom(ctx)
	if !ok {
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("not signed in"))
	}
	return connect.NewResponse(&apiv1.GetCurrentMemberResponse{Member: memberToProto(member)}), nil
}

// ReplacePassword sets a new password for the signed-in Member.
//
// The current password is required even when it was a temporary one, so that an
// unattended browser cannot be used to take the account over.
func (s *AuthService) ReplacePassword(
	ctx context.Context,
	req *connect.Request[apiv1.ReplacePasswordRequest],
) (*connect.Response[apiv1.ReplacePasswordResponse], error) {
	member, ok := auth.MemberFrom(ctx)
	if !ok {
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("not signed in"))
	}

	if err := password.Verify(member.PasswordHash, req.Msg.GetCurrentPassword()); err != nil {
		if errors.Is(err, password.ErrWrong) {
			return nil, connect.NewError(connect.CodeInvalidArgument,
				errors.New("that is not your current password"))
		}
		return nil, internalError("verify password", err)
	}

	hash, err := password.Hash(req.Msg.GetNewPassword())
	if err != nil {
		return nil, passwordError(err)
	}
	if err := s.store.SetMemberPassword(ctx, member.ID, hash); err != nil {
		return nil, internalError("set member password", err)
	}

	updated, err := s.store.MemberByID(ctx, member.ID)
	if err != nil {
		return nil, internalError("read member", err)
	}
	return connect.NewResponse(&apiv1.ReplacePasswordResponse{Member: memberToProto(updated)}), nil
}

// checkSetupIsOpen rejects first run when it has already happened, or when the form is
// incomplete. Once an Instance has an Admin this closes permanently — otherwise whoever
// reached it next could seize it.
func (s *AuthService) checkSetupIsOpen(ctx context.Context, msg *apiv1.CompleteSetupRequest) error {
	count, err := s.store.CountMembers(ctx)
	if err != nil {
		return internalError("count members", err)
	}
	if count > 0 {
		return connect.NewError(connect.CodeFailedPrecondition,
			errors.New("this instance has already been set up"))
	}
	if err := requireText(msg.GetName(), "a name"); err != nil {
		return err
	}
	if err := requireText(msg.GetEmail(), "an email"); err != nil {
		return err
	}
	return requireText(msg.GetInstanceName(), "a name for this instance")
}

// createMemberInput is what createMember needs, kept separate so the signature does not
// grow a row of same-typed arguments.
type createMemberInput struct {
	name     string
	email    string
	plain    string
	role     store.Role
	mustPick bool
}

// createMember validates and hashes a password, then stores the Member.
func (s *AuthService) createMember(ctx context.Context, in createMemberInput) (store.Member, error) {
	hash, err := password.Hash(in.plain)
	if err != nil {
		return store.Member{}, passwordError(err)
	}
	uid, err := s.newUID()
	if err != nil {
		return store.Member{}, internalError("make member uid", err)
	}

	member, err := s.store.CreateMember(ctx, store.CreateMemberParams{
		UID:                uid,
		Name:               in.name,
		Email:              in.email,
		Role:               in.role,
		PasswordHash:       hash,
		MustChangePassword: in.mustPick,
		CreatedAt:          s.now(),
	})
	if err != nil {
		if errors.Is(err, store.ErrEmailTaken) {
			return store.Member{}, connect.NewError(connect.CodeAlreadyExists, err)
		}
		return store.Member{}, internalError("create member", err)
	}
	return member, nil
}

// startSession mints a session and puts its cookie on the response.
func (s *AuthService) startSession(ctx context.Context, header interface{ Add(string, string) }, member store.Member) error {
	token, hash, err := s.newAuth()
	if err != nil {
		return internalError("make session token", err)
	}

	now := s.now()
	expires := now.Add(auth.SessionLifetime)
	session := store.Session{
		TokenHash: hash,
		MemberID:  member.ID,
		CreatedAt: now,
		ExpiresAt: expires,
	}
	if err := s.store.CreateSession(ctx, session); err != nil {
		return internalError("create session", err)
	}

	header.Add("Set-Cookie", auth.NewCookie(token, expires, s.secure).String())
	return nil
}

// errWrongCredentials is deliberately the same for an unknown email and a wrong
// password, so sign-in cannot be used to discover who has an account.
var errWrongCredentials = connect.NewError(connect.CodeUnauthenticated,
	errors.New("that email and password do not match"))

// passwordError turns a rule failure into an argument error, and anything else into an
// internal one.
func passwordError(err error) error {
	if errors.Is(err, password.ErrTooShort) || errors.Is(err, password.ErrTooLong) {
		return connect.NewError(connect.CodeInvalidArgument, err)
	}
	return internalError("hash password", err)
}

// internalError wraps a failure the caller can do nothing about.
func internalError(what string, err error) error {
	return connect.NewError(connect.CodeInternal, fmt.Errorf("%s: %w", what, err))
}

// requireText rejects a blank field, naming what was wanted.
func requireText(value, what string) error {
	if value == "" {
		return connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("%s is required", what))
	}
	return nil
}

// memberToProto converts a stored Member to its wire form. It never copies the hash.
func memberToProto(m store.Member) *apiv1.Member {
	return &apiv1.Member{
		Uid:                m.UID,
		Name:               m.Name,
		Email:              m.Email,
		Role:               roleToProto(m.Role),
		MustChangePassword: m.MustChangePassword,
	}
}

func roleToProto(role store.Role) apiv1.Role {
	switch role {
	case store.RoleAdmin:
		return apiv1.Role_ROLE_ADMIN
	case store.RoleMember:
		return apiv1.Role_ROLE_MEMBER
	default:
		return apiv1.Role_ROLE_UNSPECIFIED
	}
}
