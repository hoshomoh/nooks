// Package store persists everything an Instance knows.
//
// The package exposes one interface, Store, and two implementations of it that
// differ only in dialect. Nothing above this package knows which driver is in use,
// and nothing in this package knows about HTTP.
package store

import (
	"context"
	"time"
)

// InstanceSettings is an Instance's own configuration: the parts of it that the
// sign-in page and the Public list can see before anyone has signed in.
type InstanceSettings struct {
	// Name is the mutable display label chosen at first run, e.g. "Brunnen Street".
	// Empty until first run completes.
	Name string

	// PublicSignup, when false, makes every account a Join request an Admin approves.
	PublicSignup bool

	// SetupCompletedAt is when first run finished. The zero value means it has not,
	// and the app shows first run rather than sign in.
	SetupCompletedAt time.Time

	// Public is the one List anybody can read without an account, and what a Visitor
	// sees of it.
	Public PublicList
}

// PublicList is the Instance's one public page.
//
// At most one, because the page has a single address and a Visitor arriving at it must
// land somewhere definite. More than one would make "the public list" a question.
type PublicList struct {
	// ListUID is the List on the page, or empty when there is none.
	ListUID string
	// ShowNames puts contributor names on the rows. Off by default: a public page is
	// about what needs buying, not about who is in the household.
	ShowNames bool
	// ShowMeta puts quantities and dates on the rows — the metadata a shopper needs.
	ShowMeta bool
	// AllowJoin adds a quiet link at the foot of the page for a Visitor to ask for an
	// account.
	AllowJoin bool
}

// IsPublished reports whether the Instance has a public page at all.
func (p PublicList) IsPublished() bool { return p.ListUID != "" }

// NeedsSetup reports whether first run still has to happen.
func (s InstanceSettings) NeedsSetup() bool {
	return s.SetupCompletedAt.IsZero()
}

// Store is the persistence boundary for an Instance.
//
// Implementations are obtained from OpenSQLite or OpenPostgres. Callers accept this
// interface rather than a concrete type so that a fake can stand in for tests.
type Store interface {
	// InstanceSettings reads the Instance's own configuration. It returns the zero
	// value, not an error, when first run has not happened.
	InstanceSettings(ctx context.Context) (InstanceSettings, error)

	// SaveInstanceSettings writes the Instance's own configuration in full.
	SaveInstanceSettings(ctx context.Context, settings InstanceSettings) error

	// CreateMember adds a Member, returning ErrEmailTaken if the email is in use.
	CreateMember(ctx context.Context, params CreateMemberParams) (Member, error)

	// MemberByEmail, MemberByUID and MemberByID each return ErrNotFound when there is
	// no such Member.
	MemberByEmail(ctx context.Context, email string) (Member, error)
	MemberByUID(ctx context.Context, uid string) (Member, error)
	// Members returns everyone on the Instance, by name.
	Members(ctx context.Context) ([]Member, error)

	// SetMemberRole makes somebody an Admin, or stops them being one.
	SetMemberRole(ctx context.Context, id int64, role Role) error

	// DeleteMember removes an account. What they added stays on its Lists.
	DeleteMember(ctx context.Context, id int64) error
	MemberByID(ctx context.Context, id int64) (Member, error)

	// CountMembers reports how many Members exist.
	CountMembers(ctx context.Context) (int, error)

	// SetMemberPassword replaces a password and clears the must-change flag.
	SetMemberPassword(ctx context.Context, id int64, hash string) error

	// MarkMemberSignedIn records that a Member has just signed in.
	MarkMemberSignedIn(ctx context.Context, id int64, at time.Time) error

	// CreateSession records a signed-in browser.
	CreateSession(ctx context.Context, session Session) error

	// SessionByTokenHash returns ErrNotFound when there is no such Session. An expired
	// Session is still returned; expiry is the caller's decision.
	SessionByTokenHash(ctx context.Context, tokenHash string) (Session, error)

	// DeleteSession signs one browser out, and is not an error if it was already gone.
	DeleteSession(ctx context.Context, tokenHash string) error

	// DeleteExpiredSessions clears out Sessions past their expiry.
	DeleteExpiredSessions(ctx context.Context, now time.Time) (int64, error)

	// CreateJoinRequest records a Visitor's request for an account.
	CreateJoinRequest(ctx context.Context, params CreateJoinRequestParams) (JoinRequest, error)

	// PendingJoinRequests lists what is waiting for an Admin, oldest first.
	PendingJoinRequests(ctx context.Context) ([]JoinRequest, error)

	// JoinRequestByUID returns ErrNotFound when there is no such request.
	JoinRequestByUID(ctx context.Context, uid string) (JoinRequest, error)

	// DecideJoinRequest records an Admin's decision on a pending request.
	DecideJoinRequest(ctx context.Context, uid string, status RequestStatus, at time.Time) error

	// UseJoinRequest spends an approved request, so one approval creates one account.
	UseJoinRequest(ctx context.Context, uid string, at time.Time) error

	// CreateResetRequest records a Member's request to replace a forgotten password.
	CreateResetRequest(ctx context.Context, uid string, memberID int64, at time.Time) (ResetRequest, error)

	// PendingResetRequests lists what is waiting for an Admin, oldest first.
	PendingResetRequests(ctx context.Context) ([]ResetRequest, error)

	// ResetRequestByUID returns ErrNotFound when there is no such request.
	ResetRequestByUID(ctx context.Context, uid string) (ResetRequest, error)

	// DecideResetRequest records an Admin's decision. An approval carries an expiry.
	DecideResetRequest(ctx context.Context, uid string, status RequestStatus, at time.Time) error

	// UseResetRequest spends an approved request, so one approval sets one password.
	UseResetRequest(ctx context.Context, uid string) error

	// CreateList adds a List.
	CreateList(ctx context.Context, params CreateListParams) (List, error)

	// ListByUID returns ErrNotFound when there is no such live List.
	ListByUID(ctx context.Context, uid string) (List, error)

	// ListsForMember returns every live List a Member can reach.
	ListsForMember(ctx context.Context, memberID int64) ([]List, error)

	// RenameList changes a List's name.
	RenameList(ctx context.Context, uid string, name string, at time.Time) error

	// SetListSharing changes who can reach a List and whether they may edit it.
	SetListSharing(ctx context.Context, uid string, sharing Sharing, canEdit bool, at time.Time) error

	// DeleteList removes a List. The removal is soft.
	DeleteList(ctx context.Context, uid string, at time.Time) error

	// PinList and UnpinList change one Member's own sidebar, and nobody else's.
	PinList(ctx context.Context, memberID, listID int64) error
	UnpinList(ctx context.Context, memberID, listID int64) error

	// PinnedListIDs returns the Lists one Member has pinned.
	PinnedListIDs(ctx context.Context, memberID int64) ([]int64, error)

	// CreateItem appends an Item to a List.
	CreateItem(ctx context.Context, params CreateItemParams) (Item, error)

	// ItemsOnList returns a List's live Items in their manual order.
	ItemsOnList(ctx context.Context, listID int64) ([]Item, error)

	// ItemByUID returns ErrNotFound when there is no such live Item.
	ItemByUID(ctx context.Context, uid string) (Item, error)

	// DatedItemsForMember returns unticked, dated Items across every reachable List.
	DatedItemsForMember(ctx context.Context, memberID int64, from, to string) ([]Item, error)

	// UpdateItem changes an Item's own fields; a nil field is left alone.
	UpdateItem(ctx context.Context, uid string, params UpdateItemParams, at time.Time) error

	// SetItemDone and SetItemNotDone tick and untick. Ticking never conflicts.
	SetItemDone(ctx context.Context, uid string, doneBy int64, at time.Time) error
	SetItemNotDone(ctx context.Context, uid string, at time.Time) error

	// MoveItem places an Item at a position worked out from its new neighbours.
	MoveItem(ctx context.Context, uid string, position float64, at time.Time) error

	// DeleteItem removes an Item. The removal is soft.
	DeleteItem(ctx context.Context, uid string, at time.Time) error

	// Search returns hits for what a Member typed, most relevant first.
	//
	// It does not filter by who may see what: that is accessTo's job in the API layer,
	// and keeping one place in charge is what stops search becoming a way around
	// permissions.
	Search(ctx context.Context, query string) ([]SearchHit, error)

	// Index and Unindex maintain the search index. The store calls them on its own
	// writes; a caller should not need to.
	Index(ctx context.Context, entry IndexEntry) error
	Unindex(ctx context.Context, kind SearchKind, uid string) error

	// CreateGroup adds a Group — a shortcut for sharing, with no permissions of its own.
	CreateGroup(ctx context.Context, uid, name string, at time.Time) (Group, error)

	// GroupByUID returns ErrNotFound when there is no such Group.
	GroupByUID(ctx context.Context, uid string) (Group, error)

	// Groups lists every Group on the Instance.
	Groups(ctx context.Context) ([]Group, error)

	// AddToGroup and RemoveFromGroup change who is in a Group. Removing someone takes
	// away the Lists they reached through it, and nothing else.
	AddToGroup(ctx context.Context, groupID, memberID int64) error
	RemoveFromGroup(ctx context.Context, groupID, memberID int64) error

	// GroupMemberIDs lists who is in a Group.
	GroupMemberIDs(ctx context.Context, groupID int64) ([]int64, error)

	// ReplaceGroupMembers sets exactly who is in a Group.
	ReplaceGroupMembers(ctx context.Context, groupID int64, memberIDs []int64) error

	// ListsSharedWithGroup is every live List a Group reaches.
	ListsSharedWithGroup(ctx context.Context, groupID int64) ([]List, error)

	// ReplaceListShares sets exactly who a List is shared with by name.
	ReplaceListShares(ctx context.Context, listID int64, memberIDs, groupIDs []int64) error

	// ListShares reads who a List is shared with by name.
	ListShares(ctx context.Context, listID int64) ([]Share, error)

	// SharedListIDs is every List reaching a Member by name, directly or via a Group.
	SharedListIDs(ctx context.Context, memberID int64) ([]int64, error)

	// CreateActivity records something for a Member to see. Nooks sends no email, so
	// this is the only place it surfaces.
	CreateActivity(ctx context.Context, params CreateActivityParams) (Activity, error)

	// ActivityFor returns what is waiting for one Member, newest first.
	ActivityFor(ctx context.Context, memberID int64) ([]Activity, error)

	// MarkActivityRead marks everything a Member has now seen.
	MarkActivityRead(ctx context.Context, memberID int64, at time.Time) error

	// ResolveActivity records what became of everything pointing at one request.
	ResolveActivity(ctx context.Context, targetUID string, outcome Outcome) error

	// AdminIDs lists the Members who can act on a request.
	AdminIDs(ctx context.Context) ([]int64, error)

	// Close releases the underlying database handle.
	Close() error
}
