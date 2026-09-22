package v1

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"
	"unicode/utf8"

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
	// announce is nil when nobody is watching. A join request is still recorded.
	announce Announcer
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

// WithAnnouncer wires live updates in and returns the service, so the server can build
// and wire it in one expression.
func (s *AuthService) WithAnnouncer(a Announcer) *AuthService {
	s.announce = a
	return s
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

	/*
	 * Claimed before the password is hashed, which is where the race lived.
	 *
	 * checkSetupIsOpen counts Members and finds none. Hashing is bcrypt and
	 * deliberately slow, so a couple of hundred milliseconds used to pass between that
	 * answer and the insert, and two requests arriving inside it both passed. With
	 * different emails both became Admins, and the one who lost was never told.
	 *
	 * The claim is a row nothing can insert twice, so the second caller is refused here
	 * and the slow part happens once the race is already decided.
	 */
	claimed, err := s.store.ClaimFirstRun(ctx, s.now())
	if err != nil {
		return nil, internalError("claim first run", err)
	}
	if !claimed {
		return nil, errAlreadySetUp
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
		Name: msg.GetInstanceName(),
		// Open, which is what a fresh Instance has always done. The approval is still
		// what decides an account; this decides whether anybody may ask for one.
		PublicSignup:     true,
		SetupCompletedAt: s.now(),
	}
	if err := s.store.SaveInstanceSettings(ctx, settings); err != nil {
		return nil, internalError("save instance settings", err)
	}

	res := connect.NewResponse(&apiv1.CompleteSetupResponse{Member: memberToProto(member)})
	access, err := s.startSession(ctx, res.Header(), member)
	if err != nil {
		return nil, err
	}
	res.Msg.AccessToken = access.Token
	res.Msg.AccessTokenExpiresAt = formatMoment(access.ExpiresAt)
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

	res := connect.NewResponse(&apiv1.SignInResponse{Member: memberToProto(member)})
	access, err := s.startSession(ctx, res.Header(), member)
	if err != nil {
		return nil, err
	}
	res.Msg.AccessToken = access.Token
	res.Msg.AccessTokenExpiresAt = formatMoment(access.ExpiresAt)
	return res, nil
}

// SignOut ends the current session.
func (s *AuthService) SignOut(
	ctx context.Context,
	req *connect.Request[apiv1.SignOutRequest],
) (*connect.Response[apiv1.SignOutResponse], error) {
	res := connect.NewResponse(&apiv1.SignOutResponse{})
	for _, expired := range auth.ExpiredCookies(s.secure) {
		res.Header().Add("Set-Cookie", expired.String())
	}

	// The whole tree, not just the cookie: an access token that outlived the session it
	// came from is a credential nobody knows they still have.
	//
	// Signing out twice is not a failure, so a missing or unknown cookie is fine.
	if token := auth.SessionTokenFrom(req.Header()); token != "" {
		if err := s.store.DeleteSessionTree(ctx, auth.HashToken(token)); err != nil {
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
	member, err := requireMember(ctx)
	if err != nil {
		return nil, err
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
	// A browser, never a token: a key that reaches somebody's Lists must not be a way
	// to take the account those Lists belong to.
	grant, err := requireBrowserGrant(ctx)
	if err != nil {
		return nil, err
	}
	member := grant.Member

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
	/*
		Every other browser is signed out.

		Changing a password is what somebody does when they think another device has
		their account, so leaving those sessions alive would make the one remedy they
		reach for do nothing. The browser asking keeps its session: it has just proved
		it knows the old password, and throwing it out would answer a settings change
		with a sign-in page.
	*/
	if err := s.store.DeleteSessionsFor(ctx, member.ID, grant.Session); err != nil {
		return nil, internalError("end other sessions", err)
	}

	updated, err := s.store.MemberByID(ctx, member.ID)
	if err != nil {
		return nil, internalError("read member", err)
	}
	return connect.NewResponse(&apiv1.ReplacePasswordResponse{Member: memberToProto(updated)}), nil
}

// errAlreadySetUp refuses first run to everybody but whoever got there first, whether
// they are late by a Member or by a hundred milliseconds.
var errAlreadySetUp = connect.NewError(connect.CodeFailedPrecondition,
	errors.New("this instance has already been set up"))

// checkSetupIsOpen rejects first run when it has already happened, or when the form is
// incomplete. Once an Instance has an Admin this closes permanently — otherwise whoever
// reached it next could seize it.
func (s *AuthService) checkSetupIsOpen(ctx context.Context, msg *apiv1.CompleteSetupRequest) error {
	count, err := s.store.CountMembers(ctx)
	if err != nil {
		return internalError("count members", err)
	}
	if count > 0 {
		return errAlreadySetUp
	}
	if err := requireText(msg.GetName(), "a name"); err != nil {
		return err
	}
	if err := requireText(msg.GetEmail(), "an email"); err != nil {
		return err
	}
	if err := requireText(msg.GetInstanceName(), "a name for this instance"); err != nil {
		return err
	}
	if err := withinLimit(msg.GetName(), "a name", limitMemberName); err != nil {
		return err
	}
	if err := withinLimit(msg.GetEmail(), "an email", limitMemberEmail); err != nil {
		return err
	}
	return withinLimit(msg.GetInstanceName(), "an instance name", limitInstanceName)
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
func (s *AuthService) startSession(ctx context.Context, header interface{ Add(string, string) }, member store.Member) (issuedAccess, error) {
	token, hash, err := s.newAuth()
	if err != nil {
		return issuedAccess{}, internalError("make session token", err)
	}

	now := s.now()
	expires := now.Add(auth.SessionLifetime)
	session := store.Session{
		TokenHash: hash,
		MemberID:  member.ID,
		CreatedAt: now,
		ExpiresAt: expires,
		Kind:      store.SessionRefresh,
	}
	if err := s.store.CreateSession(ctx, session); err != nil {
		return issuedAccess{}, internalError("create session", err)
	}

	// Minted alongside, so a caller that cannot use cookies has something to hold from
	// the moment it signs in. A browser ignores it.
	access, err := s.mintAccess(ctx, member, hash)
	if err != nil {
		return issuedAccess{}, err
	}

	// Here rather than in SignIn, because every way a session begins is somebody
	// arriving: first run, signing in, finishing a join, replacing a password. Marking
	// it only on sign-in left the first Admin listed forever as not yet arrived.
	if err := s.store.MarkMemberSignedIn(ctx, member.ID, now); err != nil {
		return issuedAccess{}, internalError("mark member signed in", err)
	}

	header.Add("Set-Cookie", auth.NewCookie(token, expires, s.secure).String())
	return access, nil
}

// issuedAccess is the short-lived credential handed to a caller, and when it runs out.
type issuedAccess struct {
	Token     string
	ExpiresAt time.Time
}

/*
mintAccess makes a short-lived token belonging to a refresh token.

Recorded with its parent so that signing out takes it too: a credential that outlives
the session it came from is a credential nobody knows they still have.
*/
func (s *AuthService) mintAccess(
	ctx context.Context,
	member store.Member,
	refreshHash string,
) (issuedAccess, error) {
	token, hash, err := s.newAuth()
	if err != nil {
		return issuedAccess{}, internalError("make access token", err)
	}

	now := s.now()
	expires := now.Add(auth.AccessLifetime)
	if err := s.store.CreateSession(ctx, store.Session{
		TokenHash:  hash,
		MemberID:   member.ID,
		CreatedAt:  now,
		ExpiresAt:  expires,
		Kind:       store.SessionAccess,
		ParentHash: refreshHash,
	}); err != nil {
		return issuedAccess{}, internalError("create access token", err)
	}
	return issuedAccess{Token: token, ExpiresAt: expires}, nil
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

/*
internalError is what a caller is told when the Instance itself failed.

It logs, because nothing else does. The request log says a request happened and how long
it took; it does not say that it failed or why, so an Instance that cannot reach its
database writes one ordinary-looking line per attempt and the person running it has
nothing to read. Whoever is self-hosting has no other window into this.

The cause stays in that log and goes no further. It used to travel to the caller, which
meant a Member, or a stranger asking for a password reset, could be shown a table name,
the path to the database file, or `dial tcp 10.0.0.5:5432: connection refused`.

What travels instead is a kind and a reference. The kind is what the app looks up to
draw a sentence in the Member's own language, because a string written here would be
English wherever it was read. The reference is how whoever runs the Instance finds the
matching log line, which on a machine they own is a thing they can actually do.
*/
func internalError(what string, err error) error {
	ref := newErrorRef()
	kind := kindOf(what)
	slog.Error("request failed", "doing", what, "kind", kind, "ref", ref, "error", err)

	// The message carries no cause, and `what` is a verb and a noun with nothing of the
	// machine in it. A caller that is not the app, over REST or MCP, gets something it
	// can quote rather than a bare code.
	failure := connect.NewError(connect.CodeInternal,
		fmt.Errorf("could not %s (ref %s)", what, ref))
	failure.Meta().Set(errorKindHeader, kind)
	failure.Meta().Set(errorRefHeader, ref)
	return failure
}

// What the app reads off a failure to decide what to say, and what to quote back.
//
// Written lowercase, which is how the app reads them: Go canonicalises a header name on
// the way out and the browser lowercases it on the way in, so the two agree whatever
// case is written here, and agreeing in the source as well is worth more than matching
// the shape HTTP happens to put on the wire.
const (
	errorKindHeader = "nooks-error-kind"
	errorRefHeader  = "nooks-error-ref"
)

// The kinds, which are what a Member can do about it rather than what went wrong.
const (
	// kindSaveFailed means their change did not happen.
	kindSaveFailed = "save-failed"
	// kindLoadFailed means something could not be read, and nothing was changed.
	kindLoadFailed = "load-failed"
)

/*
loads are the words an internalError description opens with when nothing was written.

Read off the description rather than passed at each of the hundred and thirty-seven call
sites, which would be a hundred and thirty-seven chances to pass the wrong one. The
descriptions already lead with a verb because they were written to read as "could not
read the member"; this says which of those verbs mean a read.

A verb that is not here is treated as a write, which is the safe way round: telling
somebody their change may not have happened when it did is recoverable, and telling them
it happened when it did not is not. TestEveryFailureSaysWhichKindItIs holds every
description in the tree to using a verb this knows, so a new one is a decision somebody
makes rather than a default they fall into.
*/
var loads = map[string]bool{
	"read":   true,
	"count":  true,
	"list":   true,
	"search": true,
}

// kindOf says whether a failure doing this left the Member's work where it was.
func kindOf(what string) string {
	verb, _, _ := strings.Cut(what, " ")
	if loads[verb] {
		return kindLoadFailed
	}
	return kindSaveFailed
}

/*
newErrorRef is the short handle a failure is quoted by.

Short because somebody reads it off a screen and types it into a search. Random rather
than a counter: a counter would say how many times the Instance has failed, which is
nobody's business but the operator's and is not what this is for.
*/
func newErrorRef() string {
	raw := make([]byte, 3)
	if _, err := rand.Read(raw); err != nil {
		return "unknown"
	}
	return hex.EncodeToString(raw)
}

// requireText rejects a blank field, naming what was wanted.
func requireText(value, what string) error {
	if value == "" {
		return connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("%s is required", what))
	}
	return nil
}

/*
The most a Member may write into each field.

Nothing bounded any of these. A request is capped at four mebibytes, so a List could be
named with four mebibytes of text, and every screen that draws that name would then try
to. The numbers are what somebody could mean rather than what a column could hold: a
long shopping line is under sixty characters and a long recipe is about four thousand.

Counted in characters rather than bytes, because that is what the number means to
whoever is typing. Storage is already bounded by the request cap above.
*/
const (
	limitItemLabel    = 500
	limitItemQuantity = 50
	limitItemNote     = 64_000
	limitListName     = 200
	limitMemberName   = 100
	// The longest an address can be, from RFC 5321. Not a judgement.
	limitMemberEmail  = 254
	limitGroupName    = 200
	limitTokenName    = 200
	limitInstanceName = 200
	limitJoinMessage  = 1_000
)

// withinLimit rejects a field longer than a Member could have meant, naming the number
// so they can see how far over they are rather than guessing.
func withinLimit(value, what string, limit int) error {
	if utf8.RuneCountInString(value) > limit {
		return connect.NewError(connect.CodeInvalidArgument,
			fmt.Errorf("%s is longer than %d characters", what, limit))
	}
	return nil
}

// activity records entries in the panel, from the same clock and identifiers this
// service already has.
func (s *AuthService) activity() activityRecorder {
	return activityRecorder{store: s.store, now: s.now, newUID: s.newUID, announce: s.announce}
}

// memberToProto converts a stored Member to its wire form. It never copies the hash,
// which TestListMembersNeverCarriesAStoredHash holds it to.
func memberToProto(m store.Member) *apiv1.Member {
	out := &apiv1.Member{
		Uid:                m.UID,
		Name:               m.Name,
		Email:              m.Email,
		Role:               roleToProto(m.Role),
		MustChangePassword: m.MustChangePassword,
		CreatedAt:          m.CreatedAt.Format(time.RFC3339),
	}
	// Empty for somebody who has not arrived yet: the account is real, the person has
	// simply not used it.
	if !m.LastSignedInAt.IsZero() {
		out.LastSignedInAt = m.LastSignedInAt.Format(time.RFC3339)
	}
	return out
}

// roleFromProto returns an empty Role for an unspecified value, which callers treat as
// a missing argument.
func roleFromProto(role apiv1.Role) store.Role {
	switch role {
	case apiv1.Role_ROLE_ADMIN:
		return store.RoleAdmin
	case apiv1.Role_ROLE_MEMBER:
		return store.RoleMember
	default:
		return ""
	}
}

// errEmailTaken reports an email another Member already uses. It is what the caller
// did, not a server fault.
var errEmailTaken = connect.NewError(connect.CodeAlreadyExists,
	errors.New("somebody here already uses that email"))

// createMemberError turns a failed CreateMember into what the caller should read.
func createMemberError(err error) error {
	if errors.Is(err, store.ErrEmailTaken) {
		return errEmailTaken
	}
	return internalError("create member", err)
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

/*
RefreshAccess exchanges the refresh cookie for a new access token.

The refresh token is read from the cookie and never appears in a request or a response
body. That is the whole point of splitting them: the credential that lasts a month is
not one a caller can copy out of a log, and the one that does travel is worth an hour.
*/
func (s *AuthService) RefreshAccess(
	ctx context.Context,
	req *connect.Request[apiv1.RefreshAccessRequest],
) (*connect.Response[apiv1.RefreshAccessResponse], error) {
	presented := auth.SessionTokenFrom(req.Header())
	if presented == "" {
		return nil, errNotSignedIn
	}

	refreshHash := auth.HashToken(presented)
	session, err := s.store.SessionByTokenHash(ctx, refreshHash)
	if err != nil || session.Kind != store.SessionRefresh {
		return nil, errNotSignedIn
	}
	if session.Expired(s.now()) {
		_ = s.store.DeleteSessionTree(ctx, refreshHash)
		return nil, errNotSignedIn
	}

	member, err := s.store.MemberByID(ctx, session.MemberID)
	if err != nil {
		return nil, internalError("read member", err)
	}

	access, err := s.mintAccess(ctx, member, refreshHash)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&apiv1.RefreshAccessResponse{
		AccessToken:          access.Token,
		AccessTokenExpiresAt: formatMoment(access.ExpiresAt),
	}), nil
}

// errNotSignedIn is the one answer to every way of arriving without a usable session.
// Saying which way it failed would tell somebody probing which half they got right.
var errNotSignedIn = connect.NewError(connect.CodeUnauthenticated,
	errors.New("not signed in"))
