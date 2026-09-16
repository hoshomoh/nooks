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

// ListWithCounts is a List and how much is on it, read together.
type ListWithCounts struct {
	List
	OpenCount int
	DoneCount int
}

// ListPage is the rows a query asked for, and how many there were in total.
type ListPage struct {
	Lists []ListWithCounts
	Total int
}

/*
countsSubquery is how many Items are open and done on each List, in one pass.

Counting by reading a List's Items, one query per List, made drawing a sidebar cost a
read of every Item in the database. This is a single grouped aggregate the page joins
against, so the work is one scan rather than one query per row.
*/
func (s *sqlStore) countsSubquery() *bun.SelectQuery {
	return s.db.NewSelect().
		Model((*itemModel)(nil)).
		Column("list_id").
		ColumnExpr("SUM(CASE WHEN done_at = '' THEN 1 ELSE 0 END) AS open_count").
		ColumnExpr("SUM(CASE WHEN done_at <> '' THEN 1 ELSE 0 END) AS done_count").
		Where("deleted_at = ''").
		GroupExpr("list_id")
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

	rows := []listWithCountsRow{}

	query := s.db.NewSelect().
		Model(&rows).
		ModelTableExpr("list AS list").
		ColumnExpr("list.*").
		ColumnExpr("COALESCE(counts.open_count, 0) AS open_count").
		ColumnExpr("COALESCE(counts.done_count, 0) AS done_count").
		Join("LEFT JOIN (?) AS counts ON counts.list_id = list.id", s.countsSubquery()).
		Where("list.deleted_at = ''")

	query = whereListVisible(query, q.MemberID, named)
	query = q.Reach.narrow(query)
	query = whereListStatus(query, q.Status)

	total, err := query.Count(ctx)
	if err != nil {
		return ListPage{}, fmt.Errorf("count lists: %w", err)
	}

	query = orderListsBy(query, q.Order)
	if q.Limit > 0 {
		query = query.Limit(q.Limit).Offset(q.Offset)
	}

	if err := query.Scan(ctx); err != nil {
		return ListPage{}, fmt.Errorf("read lists: %w", err)
	}

	return toListPage(rows, total)
}

// listWithCountsRow is a List row with its two counts joined on.
type listWithCountsRow struct {
	listModel `bun:",extend"`
	OpenCount int `bun:"open_count"`
	DoneCount int `bun:"done_count"`
}

// toListPage converts the rows a read returned.
func toListPage(rows []listWithCountsRow, total int) (ListPage, error) {
	out := make([]ListWithCounts, 0, len(rows))
	for _, row := range rows {
		list, err := row.listModel.toList()
		if err != nil {
			return ListPage{}, err
		}
		out = append(out, ListWithCounts{List: list, OpenCount: row.OpenCount, DoneCount: row.DoneCount})
	}
	return ListPage{Lists: out, Total: total}, nil
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
			Where("(COALESCE(counts.open_count, 0) > 0 OR COALESCE(counts.done_count, 0) = 0)")
	case StatusCompleted:
		return query.Where("list.archived_at = ''").
			Where("COALESCE(counts.open_count, 0) = 0").
			Where("COALESCE(counts.done_count, 0) > 0")
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
		return query.OrderExpr("COALESCE(counts.open_count, 0) DESC").OrderExpr(byName)
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
	Lists []ListWithCounts
	Total int
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
	group := func(limit int, narrow func(*bun.SelectQuery) *bun.SelectQuery) (SidebarGroup, error) {
		page, err := s.listsWhere(ctx, memberID, limit, reach, narrow)
		if err != nil {
			return SidebarGroup{}, err
		}
		return SidebarGroup{Lists: page.Lists, Total: page.Total}, nil
	}

	var out Sidebar
	var err error

	if out.Pinned, err = group(caps.Pinned, func(q *bun.SelectQuery) *bun.SelectQuery {
		return q.Where("pin.member_id IS NOT NULL")
	}); err != nil {
		return Sidebar{}, err
	}

	if out.Mine, err = group(caps.Mine, func(q *bun.SelectQuery) *bun.SelectQuery {
		return q.Where("pin.member_id IS NULL").
			Where("list.owner_id = ?", memberID).
			Where("NOT (COALESCE(counts.open_count, 0) = 0 AND COALESCE(counts.done_count, 0) > 0)")
	}); err != nil {
		return Sidebar{}, err
	}

	if out.Shared, err = group(caps.Shared, func(q *bun.SelectQuery) *bun.SelectQuery {
		return q.Where("pin.member_id IS NULL").
			Where("list.owner_id <> ?", memberID).
			Where("NOT (COALESCE(counts.open_count, 0) = 0 AND COALESCE(counts.done_count, 0) > 0)")
	}); err != nil {
		return Sidebar{}, err
	}

	if out.Completed, err = group(caps.Completed, func(q *bun.SelectQuery) *bun.SelectQuery {
		return q.Where("pin.member_id IS NULL").
			Where("COALESCE(counts.open_count, 0) = 0").
			Where("COALESCE(counts.done_count, 0) > 0")
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
	memberID int64,
	limit int,
	reach TokenReach,
	narrow func(*bun.SelectQuery) *bun.SelectQuery,
) (ListPage, error) {
	named, err := s.SharedListIDs(ctx, memberID)
	if err != nil {
		return ListPage{}, err
	}

	rows := []listWithCountsRow{}
	query := s.db.NewSelect().
		Model(&rows).
		ModelTableExpr("list AS list").
		ColumnExpr("list.*").
		ColumnExpr("COALESCE(counts.open_count, 0) AS open_count").
		ColumnExpr("COALESCE(counts.done_count, 0) AS done_count").
		Join("LEFT JOIN (?) AS counts ON counts.list_id = list.id", s.countsSubquery()).
		Join("LEFT JOIN list_pin AS pin ON pin.list_id = list.id AND pin.member_id = ?", memberID).
		Where("list.deleted_at = ''").
		Where("list.archived_at = ''")

	query = whereListVisible(query, memberID, named)
	query = reach.narrow(query)
	query = narrow(query)

	total, err := query.Count(ctx)
	if err != nil {
		return ListPage{}, fmt.Errorf("count sidebar lists: %w", err)
	}

	query = query.OrderExpr(byName)
	if limit > 0 {
		query = query.Limit(limit)
	}
	if err := query.Scan(ctx); err != nil {
		return ListPage{}, fmt.Errorf("read sidebar lists: %w", err)
	}

	return toListPage(rows, total)
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
	}, nil
}

type listPinModel struct {
	bun.BaseModel `bun:"table:list_pin,alias:list_pin"`

	MemberID int64 `bun:"member_id,pk"`
	ListID   int64 `bun:"list_id,pk"`
}
