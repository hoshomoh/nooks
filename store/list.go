package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/uptrace/bun"
)

// Sharing is who can reach a List.
type Sharing string

const (
	// SharingPrivate means only the owner. Nobody else sees it in their sidebar.
	SharingPrivate Sharing = "PRIVATE"
	// SharingInstance means everyone on the Instance, including whoever joins later.
	SharingInstance Sharing = "INSTANCE"
	// SharingSpecific means named Members and Groups.
	SharingSpecific Sharing = "SPECIFIC"
)

// List is a named, ordered collection of Items — the only container in Nooks.
type List struct {
	ID      int64
	UID     string
	Name    string
	OwnerID int64
	Sharing Sharing
	// CanEdit is false when whoever the List is shared with may see and print it but
	// not tick or add.
	CanEdit   bool
	CreatedAt time.Time
	UpdatedAt time.Time
	// DeletedAt is the zero value for a live List.
	DeletedAt time.Time
	// ArchivedAt is the zero value for a List still in the sidebar. Archiving is not
	// deleting: what is archived is out of the way, not gone.
	ArchivedAt time.Time
	// ArchivedByID is who put it away, so a row can say so. Zero while it is not.
	ArchivedByID int64
	// OpenCount and DoneCount are how many Items are not yet ticked and how many are.
	// Kept on the List by the store rather than added up on read — see the 0015
	// migration for why.
	OpenCount int
	DoneCount int
}

// Archived reports whether the List has been put away.
func (l List) Archived() bool { return !l.ArchivedAt.IsZero() }

// Shared reports whether anyone besides the owner can reach the List. The sidebar shows
// a dot against a shared List.
func (l List) Shared() bool { return l.Sharing != SharingPrivate }

// CreateListParams is everything needed to add a List.
type CreateListParams struct {
	UID     string
	Name    string
	OwnerID int64
	Sharing Sharing
	CanEdit bool
	At      time.Time
}

func (s *sqlStore) CreateList(ctx context.Context, params CreateListParams) (List, error) {
	if params.Name == "" {
		return List{}, errors.New("store: list name is required")
	}
	if params.Sharing == "" {
		params.Sharing = SharingPrivate
	}

	row := &listModel{
		UID:       params.UID,
		Name:      params.Name,
		OwnerID:   params.OwnerID,
		Sharing:   string(params.Sharing),
		CanEdit:   params.CanEdit,
		CreatedAt: formatTime(params.At),
		UpdatedAt: formatTime(params.At),
	}
	if _, err := s.db.NewInsert().Model(row).Returning("*").Exec(ctx); err != nil {
		return List{}, fmt.Errorf("create list: %w", err)
	}

	list, err := row.toList()
	if err != nil {
		return List{}, err
	}
	// Indexing happens with the write, so nothing can exist without being findable.
	if err := s.indexList(ctx, list); err != nil {
		return List{}, err
	}
	return list, nil
}

// ListByID reads one List by internal identity, for a caller holding an Item that names
// its List that way.
func (s *sqlStore) ListByID(ctx context.Context, id int64) (List, error) {
	row := new(listModel)
	err := s.db.NewSelect().Model(row).Where("id = ? AND deleted_at = ''", id).Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return List{}, ErrNotFound
		}
		return List{}, fmt.Errorf("read list: %w", err)
	}
	return row.toList()
}

// ListByUID finds a live List by its public identifier.
func (s *sqlStore) ListByUID(ctx context.Context, uid string) (List, error) {
	row := new(listModel)
	err := s.db.NewSelect().Model(row).Where("uid = ? AND deleted_at = ''", uid).Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return List{}, ErrNotFound
		}
		return List{}, fmt.Errorf("read list: %w", err)
	}
	return row.toList()
}

// whereListVisible narrows a query to the Lists a Member can reach: their own,
// everything shared with the whole Instance, and everything shared with them by name.
//
// Every query that returns rows belonging to Lists goes through this, so a new kind of
// sharing is added in one place rather than found by grep. Columns are qualified with
// the list table, which is how it also works for queries that join to it.
func whereListVisible(query *bun.SelectQuery, memberID int64, named []int64) *bun.SelectQuery {
	return query.WhereGroup(" AND ", func(q *bun.SelectQuery) *bun.SelectQuery {
		q = q.Where("list.owner_id = ?", memberID).
			WhereOr("list.sharing = ?", string(SharingInstance))
		if len(named) > 0 {
			q = q.WhereOr("list.id IN (?)", bun.In(named))
		}
		return q
	})
}

/*
TokenReach narrows a read to the Lists an Access token was cut for.

Limited is the whole of it: a token that names nothing reaches nothing, which is not
the same as a caller that names nothing because it is not a token.
*/
type TokenReach struct {
	Limited bool
	ListIDs []int64
}

// narrow applies the token's reach, if there is one.
func (r TokenReach) narrow(query *bun.SelectQuery) *bun.SelectQuery {
	if !r.Limited {
		return query
	}
	if len(r.ListIDs) == 0 {
		return query.Where("1 = 0")
	}
	return query.Where("list.id IN (?)", bun.In(r.ListIDs))
}

/** ListStatus narrows a page to Lists in one state. */
type ListStatus string

const (
	StatusAny       ListStatus = ""
	StatusActive    ListStatus = "ACTIVE"
	StatusCompleted ListStatus = "COMPLETED"
	StatusArchived  ListStatus = "ARCHIVED"
)

/** ListOrder is how a page of Lists is arranged. */
type ListOrder string

const (
	OrderUpdated ListOrder = "UPDATED"
	OrderName    ListOrder = "NAME"
	OrderOpen    ListOrder = "OPEN"
)

// ListQuery is one page of the Lists a Member can reach.
type ListQuery struct {
	MemberID int64
	Status   ListStatus
	Order    ListOrder
	// Reach narrows to the Lists an Access token names. The zero value does not narrow,
	// which is what a browser and a token cut for everything both want.
	Reach TokenReach
	// Offset and Limit are rows, not pages. Limit zero means no limit, which only the
	// sidebar's own bounded reads use.
	Offset int
	Limit  int
}

// ListPage is the rows a query asked for, and how many there were in total.
type ListPage struct {
	Lists []List
	Total int
	// AtLeast says the count stopped at countCeiling and there are more than Total.
	AtLeast bool
}

/*
countCeiling is how far a total is counted before it is called "at least this many".

Reading a page stops depending on how much there is; counting one does not. An exact
total visits every row that matched, so on a very large Instance it is the one part of
drawing All lists that grows without limit — a fifth of a second on a hundred thousand
Lists, where the page itself is two milliseconds.

A thousand is past anything a household will reach and far short of what it costs to
count a million, and the row says "of 1,000+" rather than a number it did not earn.
*/
const countCeiling = 1000

/*
countUpTo counts what a query matched, giving up at countCeiling.

The limit is inside the subquery, so the database stops reading rather than counting
them all and rounding down afterwards.
*/
func (s *sqlStore) countUpTo(ctx context.Context, matching *bun.SelectQuery) (int, bool, error) {
	var found int
	err := s.db.NewSelect().
		ColumnExpr("COUNT(*)").
		TableExpr("(?) AS counted", matching.Limit(countCeiling+1)).
		Scan(ctx, &found)
	if err != nil {
		return 0, false, fmt.Errorf("count lists: %w", err)
	}
	return min(found, countCeiling), found > countCeiling, nil
}

/*
ListsPage reads one page of the Lists a Member can reach, with their counts.

Filtering, ordering and paging all happen here rather than in the caller. A page sorted
after it was cut is sorted within itself and wrong about everything else, and a caller
that filters what it was given has already paid for the rows it throws away.
*/
func (s *sqlStore) ListsPage(ctx context.Context, q ListQuery) (ListPage, error) {
	named, err := s.SharedListIDs(ctx, q.MemberID)
	if err != nil {
		return ListPage{}, err
	}

	// Built twice rather than once and reused: counting wants no columns, no order and
	// a limit of its own, and the page wants all three.
	matching := func(query *bun.SelectQuery) *bun.SelectQuery {
		query = query.ModelTableExpr("list AS list").Where("list.deleted_at = ''")
		query = whereListVisible(query, q.MemberID, named)
		query = q.Reach.narrow(query)
		return whereListStatus(query, q.Status)
	}

	counting := matching(s.db.NewSelect().Model((*listModel)(nil)).ColumnExpr("1"))
	total, atLeast, err := s.countUpTo(ctx, counting)
	if err != nil {
		return ListPage{}, err
	}

	rows := []listModel{}
	reading := orderListsBy(matching(s.db.NewSelect().Model(&rows).ColumnExpr("list.*")), q.Order)
	if q.Limit > 0 {
		reading = reading.Limit(q.Limit).Offset(q.Offset)
	}
	if err := reading.Scan(ctx); err != nil {
		return ListPage{}, fmt.Errorf("read lists: %w", err)
	}

	return toListPage(rows, total, atLeast)
}

// toListPage converts the rows a read returned.
func toListPage(rows []listModel, total int, atLeast bool) (ListPage, error) {
	out := make([]List, 0, len(rows))
	for _, row := range rows {
		list, err := row.toList()
		if err != nil {
			return ListPage{}, err
		}
		out = append(out, list)
	}
	return ListPage{Lists: out, Total: total, AtLeast: atLeast}, nil
}

/*
whereListStatus narrows to one state.

Completed is "has had Items and none are open", not "nothing open": a List nobody has
put anything on yet has nothing open either, and it is not finished. Archived is its
own answer rather than a flavour of the others, so a List put away is out of active and
completed alike.
*/
func whereListStatus(query *bun.SelectQuery, status ListStatus) *bun.SelectQuery {
	switch status {
	case StatusArchived:
		return query.Where("list.archived_at <> ''")
	case StatusActive:
		return query.Where("list.archived_at = ''").
			Where("(list.open_count > 0 OR list.done_count = 0)")
	case StatusCompleted:
		return query.Where("list.archived_at = ''").
			Where("list.open_count = 0").
			Where("list.done_count > 0")
	default:
		return query.Where("list.archived_at = ''")
	}
}

// orderListsBy arranges a page the way the caller asked.
func orderListsBy(query *bun.SelectQuery, order ListOrder) *bun.SelectQuery {
	switch order {
	case OrderName:
		return query.OrderExpr(byName)
	case OrderOpen:
		// Fullest first, and by name between equals, so the order does not shuffle
		// every time something is ticked.
		return query.OrderExpr("list.open_count DESC").OrderExpr(byName)
	default:
		return query.OrderExpr("list.updated_at DESC")
	}
}

/*
SidebarGroup is one group of the sidebar: what it draws, and how many there are.

Capped, because a sidebar is what somebody is working in rather than everything they
can reach. Total is what the "see all" beside it says.
*/
type SidebarGroup struct {
	Lists []List
	Total int
	// AtLeast says the count gave up and there are more than Total. See countCeiling.
	AtLeast bool
}

// Sidebar is the four groups the sidebar draws, read in one go.
type Sidebar struct {
	Pinned    SidebarGroup
	Mine      SidebarGroup
	Shared    SidebarGroup
	Completed SidebarGroup
}

// SidebarCaps is how many rows each group draws before it says how many there are.
type SidebarCaps struct {
	Pinned    int
	Mine      int
	Shared    int
	Completed int
}

/*
SidebarLists reads the four groups a sidebar draws.

Four bounded queries rather than everything a Member can reach: the groups are what
somebody is working in, and the cost of drawing them should not grow with how much they
have ever made.

Pinned wins over finished, and both win over the plain groups, so a List appears once.
*/
func (s *sqlStore) SidebarLists(ctx context.Context, memberID int64, caps SidebarCaps, reach TokenReach) (Sidebar, error) {
	named, err := s.SharedListIDs(ctx, memberID)
	if err != nil {
		return Sidebar{}, err
	}
	pins, err := s.PinnedListIDs(ctx, memberID)
	if err != nil {
		return Sidebar{}, err
	}

	/*
		Each group says the whole of what it is, rather than narrowing a shared query.

		A group drawn as "everything reachable, minus what the other three took" has to
		read everything reachable to find out it is empty — and a Member with no pinned
		Lists and a great many of their own would pay for a whole scan to draw nothing.
		Said in full, each of these starts from the index that answers it.
	*/
	unfinished := func(q *bun.SelectQuery) *bun.SelectQuery {
		return q.Where("NOT (list.open_count = 0 AND list.done_count > 0)")
	}
	notPinned := func(q *bun.SelectQuery) *bun.SelectQuery {
		if len(pins) == 0 {
			return q
		}
		return q.Where("list.id NOT IN (?)", bun.In(pins))
	}
	// Shared with me is what somebody else made and I can reach: whoever shared it with
	// everybody, or named me. Not "anything I do not own" — that is every List here.
	sharedWithMe := func(q *bun.SelectQuery) *bun.SelectQuery {
		return q.Where("list.owner_id <> ?", memberID).
			WhereGroup(" AND ", func(q *bun.SelectQuery) *bun.SelectQuery {
				q = q.Where("list.sharing = ?", string(SharingInstance))
				if len(named) > 0 {
					q = q.WhereOr("list.id IN (?)", bun.In(named))
				}
				return q
			})
	}

	group := func(limit int, narrow func(*bun.SelectQuery) *bun.SelectQuery) (SidebarGroup, error) {
		page, err := s.listsWhere(ctx, limit, reach, narrow)
		if err != nil {
			return SidebarGroup{}, err
		}
		return SidebarGroup{Lists: page.Lists, Total: page.Total, AtLeast: page.AtLeast}, nil
	}

	var out Sidebar

	// The pins are already in hand, so this is a read of a handful by identifier.
	if out.Pinned, err = group(caps.Pinned, func(q *bun.SelectQuery) *bun.SelectQuery {
		if len(pins) == 0 {
			return q.Where("1 = 0")
		}
		return whereListVisible(q.Where("list.id IN (?)", bun.In(pins)), memberID, named)
	}); err != nil {
		return Sidebar{}, err
	}

	if out.Mine, err = group(caps.Mine, func(q *bun.SelectQuery) *bun.SelectQuery {
		return unfinished(notPinned(q.Where("list.owner_id = ?", memberID)))
	}); err != nil {
		return Sidebar{}, err
	}

	if out.Shared, err = group(caps.Shared, func(q *bun.SelectQuery) *bun.SelectQuery {
		return unfinished(notPinned(sharedWithMe(q)))
	}); err != nil {
		return Sidebar{}, err
	}

	if out.Completed, err = group(caps.Completed, func(q *bun.SelectQuery) *bun.SelectQuery {
		q = notPinned(q).Where("list.open_count = 0").Where("list.done_count > 0")
		return whereListVisible(q, memberID, named)
	}); err != nil {
		return Sidebar{}, err
	}

	return out, nil
}

/*
listsWhere reads one bounded, name-ordered group of Lists for the sidebar.

It joins the pins as well as the counts, so a group can ask whether this Member pinned
the List without a second read.
*/
func (s *sqlStore) listsWhere(
	ctx context.Context,
	limit int,
	reach TokenReach,
	narrow func(*bun.SelectQuery) *bun.SelectQuery,
) (ListPage, error) {
	matching := func(query *bun.SelectQuery) *bun.SelectQuery {
		query = query.ModelTableExpr("list AS list").
			Where("list.deleted_at = ''").
			Where("list.archived_at = ''")
		return narrow(reach.narrow(query))
	}

	counting := matching(s.db.NewSelect().Model((*listModel)(nil)).ColumnExpr("1"))
	total, atLeast, err := s.countUpTo(ctx, counting)
	if err != nil {
		return ListPage{}, err
	}

	rows := []listModel{}
	reading := matching(s.db.NewSelect().Model(&rows).ColumnExpr("list.*")).OrderExpr(byName)
	if limit > 0 {
		reading = reading.Limit(limit)
	}
	if err := reading.Scan(ctx); err != nil {
		return ListPage{}, fmt.Errorf("read sidebar lists: %w", err)
	}

	return toListPage(rows, total, atLeast)
}

/*
CanReachList reports whether a Member may see one List at all.

The same question ListsForMember answers about every List, asked about one. A caller
holding a uid and wanting a yes or no should not read everything a Member has ever made
to find out: that is a full pass over their Lists to decide one thing, and it was seven
seconds on a hundred thousand of them.

Membership only, with nothing about Access tokens in it. The one caller is the event
stream, which takes a session cookie and never a token.
*/
func (s *sqlStore) CanReachList(ctx context.Context, memberID int64, listUID string) (bool, error) {
	named, err := s.SharedListIDs(ctx, memberID)
	if err != nil {
		return false, err
	}

	query := s.db.NewSelect().
		Model((*listModel)(nil)).
		ModelTableExpr("list AS list").
		ColumnExpr("1").
		Where("list.uid = ? AND list.deleted_at = ''", listUID)

	found, err := whereListVisible(query, memberID, named).Limit(1).Exists(ctx)
	if err != nil {
		return false, fmt.Errorf("read list: %w", err)
	}
	return found, nil
}

// ListsForMember returns every live List a Member can reach: their own, everything
// shared with the whole Instance, and everything shared with them by name — directly or
// through a Group they are in.
func (s *sqlStore) ListsForMember(ctx context.Context, memberID int64) ([]List, error) {
	named, err := s.SharedListIDs(ctx, memberID)
	if err != nil {
		return nil, err
	}

	query := s.db.NewSelect().
		Model((*listModel)(nil)).
		Where("deleted_at = ''").
		OrderExpr(byName)

	query = whereListVisible(query, memberID, named)

	var rows []listModel
	if err := query.Model(&rows).Scan(ctx); err != nil {
		return nil, fmt.Errorf("read lists: %w", err)
	}

	lists := make([]List, 0, len(rows))
	for _, row := range rows {
		list, err := row.toList()
		if err != nil {
			return nil, err
		}
		lists = append(lists, list)
	}
	return lists, nil
}

func (s *sqlStore) RenameList(ctx context.Context, uid string, name string, at time.Time) error {
	if name == "" {
		return errors.New("store: list name is required")
	}
	if err := s.updateList(ctx, uid, at, func(q *bun.UpdateQuery) *bun.UpdateQuery {
		return q.Set("name = ?", name)
	}); err != nil {
		return err
	}

	list, err := s.ListByUID(ctx, uid)
	if err != nil {
		return err
	}
	return s.indexList(ctx, list)
}

func (s *sqlStore) SetListSharing(ctx context.Context, uid string, sharing Sharing, canEdit bool, at time.Time) error {
	return s.updateList(ctx, uid, at, func(q *bun.UpdateQuery) *bun.UpdateQuery {
		return q.Set("sharing = ?", string(sharing)).Set("can_edit = ?", canEdit)
	})
}

/*
SetListArchived puts a List away, or brings it back.

Archived Lists are still returned here: they are reachable, searchable and restorable,
and only the sidebar leaves them out. A query that hid them would also hide them from
the page somebody goes to in order to find one.

memberID is recorded on the way in so a row can say who did it, and cleared on the way
back out: "archived by Anna" is only true while it is.
*/
func (s *sqlStore) SetListArchived(ctx context.Context, uid string, memberID int64, at time.Time) error {
	return s.updateList(ctx, uid, at, func(q *bun.UpdateQuery) *bun.UpdateQuery {
		if memberID == 0 {
			return q.Set("archived_at = ?", "").Set("archived_by_id = ?", 0)
		}
		return q.Set("archived_at = ?", formatTime(at)).Set("archived_by_id = ?", memberID)
	})
}

// DeleteList removes a List, and with it the Items on it. The removal is soft, so a
// List deleted by mistake is recoverable.
func (s *sqlStore) DeleteList(ctx context.Context, uid string, at time.Time) error {
	list, err := s.ListByUID(ctx, uid)
	if err != nil {
		return err
	}
	if err := s.updateList(ctx, uid, at, func(q *bun.UpdateQuery) *bun.UpdateQuery {
		return q.Set("deleted_at = ?", formatTime(at))
	}); err != nil {
		return err
	}

	// A deleted List and its Items must stop being findable, or search would hand back
	// things the Member can no longer open.
	items, err := s.ItemsOnList(ctx, list.ID)
	if err != nil {
		return err
	}
	for _, item := range items {
		if err := s.Unindex(ctx, KindItem, item.UID); err != nil {
			return err
		}
		if err := s.Unindex(ctx, KindNote, item.UID); err != nil {
			return err
		}
	}
	return s.Unindex(ctx, KindList, uid)
}

/*
recount rewrites how much is on one List.

Counted rather than adjusted. A counter kept by adding and subtracting drifts the first
time a path forgets to adjust it or a write is retried, and the number a Member reads
beside a name is then wrong until somebody notices.

What makes that affordable is an index per count holding only the rows that count is
about — idx_item_open_on_list and idx_item_done_on_list — so each is a range scan that
never opens a row. Without them this was proportional to everything on the List, twice,
on every add, tick and delete. See BenchmarkTickOnACrowdedList.
*/
func (s *sqlStore) recount(ctx context.Context, listID int64) error {
	const open = `(SELECT COUNT(*) FROM item WHERE item.list_id = ? AND item.deleted_at = '' AND item.done_at = '')`
	const done = `(SELECT COUNT(*) FROM item WHERE item.list_id = ? AND item.deleted_at = '' AND item.done_at <> '')`

	_, err := s.db.NewUpdate().
		Model((*listModel)(nil)).
		Set("open_count = "+open, listID).
		Set("done_count = "+done, listID).
		Where("id = ?", listID).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("recount list: %w", err)
	}
	return nil
}

// indexList makes a List findable by its name.
func (s *sqlStore) indexList(ctx context.Context, list List) error {
	return s.Index(ctx, IndexEntry{
		Kind: KindList, UID: list.UID, ListID: list.ID, Text: list.Name,
	})
}

// updateList applies a change to one live List, touching updated_at with it.
func (s *sqlStore) updateList(
	ctx context.Context,
	uid string,
	at time.Time,
	apply func(*bun.UpdateQuery) *bun.UpdateQuery,
) error {
	query := s.db.NewUpdate().
		Model((*listModel)(nil)).
		Set("updated_at = ?", formatTime(at)).
		Where("uid = ? AND deleted_at = ''", uid)

	result, err := apply(query).Exec(ctx)
	if err != nil {
		return fmt.Errorf("update list: %w", err)
	}
	return requireOneRow(result, "list")
}

// PinList pins a List to one Member's own sidebar. Pinning twice is not an error.
func (s *sqlStore) PinList(ctx context.Context, memberID, listID int64) error {
	_, err := s.db.NewInsert().
		Model(&listPinModel{MemberID: memberID, ListID: listID}).
		On("CONFLICT (member_id, list_id) DO NOTHING").
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("pin list: %w", err)
	}
	return nil
}

// UnpinList removes a pin. Unpinning something that was not pinned is not an error.
func (s *sqlStore) UnpinList(ctx context.Context, memberID, listID int64) error {
	_, err := s.db.NewDelete().
		Model((*listPinModel)(nil)).
		Where("member_id = ? AND list_id = ?", memberID, listID).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("unpin list: %w", err)
	}
	return nil
}

// PinnedListIDs returns the Lists one Member has pinned.
func (s *sqlStore) PinnedListIDs(ctx context.Context, memberID int64) ([]int64, error) {
	var ids []int64
	err := s.db.NewSelect().
		Model((*listPinModel)(nil)).
		Column("list_id").
		Where("member_id = ?", memberID).
		Scan(ctx, &ids)
	if err != nil {
		return nil, fmt.Errorf("read pins: %w", err)
	}
	return ids, nil
}

type listModel struct {
	bun.BaseModel `bun:"table:list,alias:list"`

	ID        int64  `bun:"id,pk,autoincrement"`
	UID       string `bun:"uid,notnull"`
	Name      string `bun:"name,notnull"`
	OwnerID   int64  `bun:"owner_id,notnull"`
	Sharing   string `bun:"sharing,notnull"`
	CanEdit   bool   `bun:"can_edit,notnull"`
	CreatedAt string `bun:"created_at,notnull"`
	UpdatedAt string `bun:"updated_at,notnull"`
	DeletedAt string `bun:"deleted_at,notnull"`

	ArchivedAt   string `bun:"archived_at,notnull"`
	ArchivedByID int64  `bun:"archived_by_id,notnull"`

	OpenCount int `bun:"open_count,notnull"`
	DoneCount int `bun:"done_count,notnull"`
}

func (m listModel) toList() (List, error) {
	createdAt, err := parseTime(m.CreatedAt)
	if err != nil {
		return List{}, err
	}
	updatedAt, err := parseTime(m.UpdatedAt)
	if err != nil {
		return List{}, err
	}
	deletedAt, err := parseTime(m.DeletedAt)
	if err != nil {
		return List{}, err
	}
	archivedAt, err := parseTime(m.ArchivedAt)
	if err != nil {
		return List{}, err
	}
	return List{
		ID: m.ID, UID: m.UID, Name: m.Name, OwnerID: m.OwnerID,
		Sharing: Sharing(m.Sharing), CanEdit: m.CanEdit,
		CreatedAt: createdAt, UpdatedAt: updatedAt, DeletedAt: deletedAt,
		ArchivedAt: archivedAt, ArchivedByID: m.ArchivedByID,
		OpenCount: m.OpenCount, DoneCount: m.DoneCount,
	}, nil
}

type listPinModel struct {
	bun.BaseModel `bun:"table:list_pin,alias:list_pin"`

	MemberID int64 `bun:"member_id,pk"`
	ListID   int64 `bun:"list_id,pk"`
}
