package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/uptrace/bun"
)

// positionGap is the space left between Items when they are appended.
//
// Positions are floats so an Item can be dropped between two others by averaging their
// positions, without renumbering the rest. A wide gap keeps that possible for a very
// long time before the floats run out of room between neighbours.
const positionGap = 1024.0

// Item is one line on a List.
type Item struct {
	ID     int64
	UID    string
	ListID int64
	// Label is the text the Member typed.
	Label string
	// Quantity is free text — "2", "1 kg" — never a number.
	Quantity string
	// DueOn is a date, not a time: an Item is due on a day. Empty for no due date.
	DueOn string
	// Note is the optional document attached to the Item, as markdown.
	Note string
	// Position orders the Item within its List.
	Position float64
	// DoneAt is the zero value until the Item is ticked.
	DoneAt time.Time
	// DoneByID is who ticked it, if anyone.
	DoneByID int64
	// AddedByID is who put it on the List. Shown at the right of the row.
	AddedByID int64
	// AddedByTokenID is the Access token it came through, or zero for a browser. The
	// Member is still recorded: a token is somebody's access narrowed, not an identity.
	AddedByTokenID int64
	CreatedAt      time.Time
	UpdatedAt      time.Time
	DeletedAt      time.Time
}

func (i Item) Done() bool { return !i.DoneAt.IsZero() }

// CreateItemParams is everything needed to add an Item. Position is worked out by the
// store, because "put it at the end" is a fact about the List, not the caller.
type CreateItemParams struct {
	UID      string
	ListID   int64
	Label    string
	Quantity string
	DueOn    string
	// Note is the document the Item starts with. Empty for one that is simply added;
	// a copy of an Item carries the Note it was copied from.
	Note      string
	AddedByID int64
	// AddedByTokenID is the token it came through, or zero for a browser.
	AddedByTokenID int64
	At             time.Time
}

func (s *sqlStore) CreateItem(ctx context.Context, params CreateItemParams) (Item, error) {
	if params.Label == "" {
		return Item{}, errors.New("store: item label is required")
	}

	row := &itemModel{
		UID:       params.UID,
		ListID:    params.ListID,
		Label:     params.Label,
		Quantity:  params.Quantity,
		DueOn:     params.DueOn,
		Note:      params.Note,
		AddedByID: params.AddedByID,
		CreatedAt: formatTime(params.At),
		UpdatedAt: formatTime(params.At),
	}
	if params.AddedByTokenID != 0 {
		row.AddedByTokenID = &params.AddedByTokenID
	}
	if err := s.insertAtTheEnd(ctx, row, params.ListID); err != nil {
		return Item{}, err
	}

	item, err := row.toItem()
	if err != nil {
		return Item{}, err
	}
	if err := s.recount(ctx, item.ListID); err != nil {
		return Item{}, err
	}
	if err := s.indexItem(ctx, item); err != nil {
		return Item{}, err
	}
	return item, nil
}

/*
insertAtTheEnd writes one Item after every Item already on its List.

The position is worked out inside the insert rather than read and then written. Reading
the highest and then inserting is two statements with a gap between them, so two people
adding to one List at the same moment both read the same number and both land on it.
Reads break the tie on id, so a List still comes back in arrival order, but the tie is
made and "put this after that one" has no answer while two Items share a place.

One statement is enough on SQLite, which has one writer, so the second insert cannot
begin until the first has finished. It is not enough on Postgres, where a statement
reads from a snapshot taken when it started: two inserts running together both see the
List as it was before either of them, both take the same maximum, and both land on it.
This is what the comment here used to claim the engine prevented, and
TestTwoItemsAddedAtOnceGetDifferentPositions found it the first time the suite ran
against Postgres with the race detector on.

So on Postgres the List's own row is locked first, which makes adding to one List
one-at-a-time and leaves adding to different Lists as parallel as it was. SQLite needs
no such thing and has no FOR UPDATE to do it with.
*/
func (s *sqlStore) insertAtTheEnd(ctx context.Context, row *itemModel, listID int64) error {
	if s.name != postgresDriver {
		return appendItem(ctx, s.db, row, listID)
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin create item: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if err := lockList(ctx, tx, listID); err != nil {
		return err
	}
	if err := appendItem(ctx, tx, row, listID); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit create item: %w", err)
	}
	return nil
}

// appendItem is the insert itself, which is the same statement whether or not it is
// running inside the transaction that holds the List.
func appendItem(ctx context.Context, db bun.IDB, row *itemModel, listID int64) error {
	insert := db.NewInsert().Model(row).Returning("*").
		Value("position", "(SELECT COALESCE(MAX(position), 0.0) + ? FROM item "+
			"WHERE list_id = ? AND deleted_at = '')", positionGap, listID)
	if _, err := insert.Exec(ctx); err != nil {
		return fmt.Errorf("create item: %w", err)
	}
	return nil
}

/*
lockList holds one List against anybody else working out a position on it, until the
transaction ends.

Postgres only: it is the driver that needs it and the only one with the statement for
it. The row it selects is thrown away, because it is the lock that is wanted and not the
List. A List that does not exist takes no lock and needs none, since the insert that
follows has a foreign key to fail on.
*/
func lockList(ctx context.Context, tx bun.Tx, listID int64) error {
	_, err := tx.NewRaw("SELECT id FROM list WHERE id = ? FOR UPDATE", listID).Exec(ctx)
	if err != nil {
		return fmt.Errorf("hold list %d: %w", listID, err)
	}
	return nil
}

/*
CreateItems adds several Items to one List at once.

One position read, one insert and one transaction for the batch, so copying a List is a
commit rather than a commit per row. The Items keep the order they arrive in.
*/
func (s *sqlStore) CreateItems(ctx context.Context, params []CreateItemParams) ([]Item, error) {
	if len(params) == 0 {
		return nil, nil
	}
	for _, p := range params {
		if p.Label == "" {
			return nil, errors.New("store: item label is required")
		}
		// One List, which the position read and the recount below both take from the
		// first of them. A batch spanning two would order the Items against the wrong
		// List and leave the other's counts behind, neither of them visibly.
		if p.ListID != params[0].ListID {
			return nil, errors.New("store: items must be added to one list at a time")
		}
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin create items: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	// Read then write, with a gap, which is the shape insertAtTheEnd explains. Here the
	// gap is inside a transaction and that is still not enough on Postgres, so the List
	// is held for the same reason and in the same way.
	if s.name == postgresDriver {
		if err := lockList(ctx, tx, params[0].ListID); err != nil {
			return nil, err
		}
	}

	var last sql.NullFloat64
	err = tx.NewSelect().
		Model((*itemModel)(nil)).
		ColumnExpr("MAX(position)").
		Where("list_id = ? AND deleted_at = ''", params[0].ListID).
		Scan(ctx, &last)
	if err != nil {
		return nil, fmt.Errorf("read last position: %w", err)
	}

	rows := make([]itemModel, 0, len(params))
	for i, p := range params {
		row := itemModel{
			UID:       p.UID,
			ListID:    p.ListID,
			Label:     p.Label,
			Quantity:  p.Quantity,
			DueOn:     p.DueOn,
			Note:      p.Note,
			Position:  last.Float64 + positionGap*float64(i+1),
			AddedByID: p.AddedByID,
			CreatedAt: formatTime(p.At),
			UpdatedAt: formatTime(p.At),
		}
		if p.AddedByTokenID != 0 {
			row.AddedByTokenID = &p.AddedByTokenID
		}
		rows = append(rows, row)
	}

	if _, err := tx.NewInsert().Model(&rows).Returning("*").Exec(ctx); err != nil {
		return nil, fmt.Errorf("create items: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit create items: %w", err)
	}

	items, err := toItems(rows)
	if err != nil {
		return nil, err
	}
	if err := s.recount(ctx, params[0].ListID); err != nil {
		return nil, err
	}
	// Indexed after the commit: a Note that fails to index is a Note that cannot be
	// searched for, which is worth reporting and is not worth losing the Items over.
	for _, item := range items {
		if err := s.indexItem(ctx, item); err != nil {
			return nil, err
		}
	}
	return items, nil
}

// ItemsOnList returns a List's live Items in their manual order.
func (s *sqlStore) ItemsOnList(ctx context.Context, listID int64) ([]Item, error) {
	var rows []itemModel
	err := s.db.NewSelect().
		Model(&rows).
		Where("list_id = ? AND deleted_at = ''", listID).
		// id breaks a tie, because positions are not unique. Appending reads the
		// highest position and then inserts, so two people adding to the same List at
		// the same moment both land on it. Without a second key the two swap places
		// between reads, and a shared List is exactly where that happens.
		Order("position ASC", "id ASC").
		Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("read items: %w", err)
	}
	return toItems(rows)
}

// ItemByUID finds a live Item.
func (s *sqlStore) ItemByUID(ctx context.Context, uid string) (Item, error) {
	row := new(itemModel)
	err := s.db.NewSelect().Model(row).Where("uid = ? AND deleted_at = ''", uid).Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Item{}, ErrNotFound
		}
		return Item{}, fmt.Errorf("read item: %w", err)
	}
	return row.toItem()
}

// DatedItemsForMember returns every live, unticked Item with a due date on a List the
// Member can reach, earliest first. It is what Today, Upcoming and the calendar are all
// built from.
//
// An empty from means no lower bound, which is how Today gathers everything overdue
// rather than only what is due on the day itself.
func (s *sqlStore) DatedItemsForMember(ctx context.Context, memberID int64, from, to string) ([]Item, error) {
	named, err := s.SharedListIDs(ctx, memberID)
	if err != nil {
		return nil, err
	}

	query := s.db.NewSelect().
		Model((*itemModel)(nil)).
		Join("JOIN list ON list.id = item.list_id").
		Where("item.deleted_at = '' AND item.done_at = '' AND item.due_on <> ''").
		Where("list.deleted_at = ''").
		Where("item.due_on <= ?", to).
		Order("item.due_on ASC", "item.position ASC", "item.id ASC")
	query = whereListVisible(query, memberID, named)

	if from != "" {
		query = query.Where("item.due_on >= ?", from)
	}

	var rows []itemModel
	if err := query.Model(&rows).Scan(ctx); err != nil {
		return nil, fmt.Errorf("read dated items: %w", err)
	}
	return toItems(rows)
}

// UpdateItemParams carries the fields an edit may change. A nil field is left alone, so
// ticking an Item does not have to restate its label.
type UpdateItemParams struct {
	Label    *string
	Quantity *string
	DueOn    *string
	Note     *string
}

// UpdateItem changes an Item's own fields.
func (s *sqlStore) UpdateItem(ctx context.Context, uid string, params UpdateItemParams, at time.Time) error {
	query := s.db.NewUpdate().
		Model((*itemModel)(nil)).
		Set("updated_at = ?", formatTime(at)).
		Where("uid = ? AND deleted_at = ''", uid)

	if params.Label != nil {
		if *params.Label == "" {
			return errors.New("store: item label is required")
		}
		query = query.Set("label = ?", *params.Label)
	}
	if params.Quantity != nil {
		query = query.Set("quantity = ?", *params.Quantity)
	}
	if params.DueOn != nil {
		query = query.Set("due_on = ?", *params.DueOn)
	}
	if params.Note != nil {
		query = query.Set("note = ?", *params.Note)
	}

	result, err := query.Exec(ctx)
	if err != nil {
		return fmt.Errorf("update item: %w", err)
	}
	if err := requireOneRow(result, "item"); err != nil {
		return err
	}

	item, err := s.ItemByUID(ctx, uid)
	if err != nil {
		return err
	}
	return s.indexItem(ctx, item)
}

// SetItemDone ticks or unticks an Item.
//
// Ticking is last-write-wins: a tick is a tick whoever made it, so this never conflicts
// and never asks a question — see DESIGN.md §11.
func (s *sqlStore) SetItemDone(ctx context.Context, uid string, doneBy int64, at time.Time) error {
	return s.changeItemCount(ctx, uid, "tick item", func(q *bun.UpdateQuery) *bun.UpdateQuery {
		return q.Set("done_at = ?", formatTime(at)).
			Set("done_by_id = ?", doneBy).
			Set("updated_at = ?", formatTime(at))
	})
}

func (s *sqlStore) SetItemNotDone(ctx context.Context, uid string, at time.Time) error {
	return s.changeItemCount(ctx, uid, "untick item", func(q *bun.UpdateQuery) *bun.UpdateQuery {
		return q.Set("done_at = ?", "").
			Set("done_by_id = ?", nil).
			Set("updated_at = ?", formatTime(at))
	})
}

// MoveItem places an Item at a position, which the caller works out from its new
// neighbours — usually the midpoint between them.
func (s *sqlStore) MoveItem(ctx context.Context, uid string, position float64, at time.Time) error {
	result, err := s.db.NewUpdate().
		Model((*itemModel)(nil)).
		Set("position = ?", position).
		Set("updated_at = ?", formatTime(at)).
		Where("uid = ? AND deleted_at = ''", uid).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("move item: %w", err)
	}
	return requireOneRow(result, "item")
}

// DeleteItem removes an Item. The removal is soft.
func (s *sqlStore) DeleteItem(ctx context.Context, uid string, at time.Time) error {
	err := s.changeItemCount(ctx, uid, "delete item", func(q *bun.UpdateQuery) *bun.UpdateQuery {
		return q.Set("deleted_at = ?", formatTime(at)).
			Set("updated_at = ?", formatTime(at))
	})
	if err != nil {
		return err
	}
	if err := s.unindex(ctx, KindItem, uid); err != nil {
		return err
	}
	return s.unindex(ctx, KindNote, uid)
}

/*
changeItemCount applies a write that moves an Item between open, done and gone.

The three of them all name the Item by uid and all leave its List's counts out of step,
so the recount happens here rather than in each of them. Which List it is on is read
first: after a delete the Item can no longer be found by uid, and the List still has to
be told.
*/
func (s *sqlStore) changeItemCount(
	ctx context.Context,
	uid string,
	what string,
	apply func(*bun.UpdateQuery) *bun.UpdateQuery,
) error {
	var listID int64
	err := s.db.NewSelect().
		Model((*itemModel)(nil)).
		Column("list_id").
		Where("uid = ? AND deleted_at = ''", uid).
		Scan(ctx, &listID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}
		return fmt.Errorf("read item: %w", err)
	}

	query := s.db.NewUpdate().Model((*itemModel)(nil)).Where("uid = ? AND deleted_at = ''", uid)
	result, err := apply(query).Exec(ctx)
	if err != nil {
		return fmt.Errorf("%s: %w", what, err)
	}
	if err := requireOneRow(result, "item"); err != nil {
		return err
	}
	return s.recount(ctx, listID)
}

// indexItem makes an Item findable by its label and quantity — both are text a Member
// wrote, and "1 kg" is as searchable as "Tomatoes".
//
// Its Note is indexed separately, so a hit can say whether the words were in the Item
// or in what was written about it.
func (s *sqlStore) indexItem(ctx context.Context, item Item) error {
	text := item.Label
	if item.Quantity != "" {
		text += " " + item.Quantity
	}
	if err := s.index(ctx, IndexEntry{
		Kind: KindItem, UID: item.UID, ListID: item.ListID, Text: text,
	}); err != nil {
		return err
	}
	return s.indexNote(ctx, item)
}

// indexNote makes an Item's Note findable, and removes it from the index when the Note
// is emptied.
func (s *sqlStore) indexNote(ctx context.Context, item Item) error {
	if item.Note == "" {
		return s.unindex(ctx, KindNote, item.UID)
	}
	return s.index(ctx, IndexEntry{
		Kind: KindNote, UID: item.UID, ListID: item.ListID, Text: item.Note,
	})
}

// nextPosition is one gap past the last Item on the List.
//
// MAX is scanned as nullable rather than wrapped in COALESCE: SQLite types the literal
// in COALESCE(MAX(position), 0) as an integer and then refuses to scan it into a float,
// while an empty result is exactly what NULL already means.
func (s *sqlStore) nextPosition(ctx context.Context, listID int64) (float64, error) {
	var last sql.NullFloat64
	err := s.db.NewSelect().
		Model((*itemModel)(nil)).
		ColumnExpr("MAX(position)").
		Where("list_id = ? AND deleted_at = ''", listID).
		Scan(ctx, &last)
	if err != nil {
		return 0, fmt.Errorf("read last position: %w", err)
	}
	return last.Float64 + positionGap, nil
}

func toItems(rows []itemModel) ([]Item, error) {
	items := make([]Item, 0, len(rows))
	for _, row := range rows {
		item, err := row.toItem()
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

type itemModel struct {
	bun.BaseModel `bun:"table:item,alias:item"`

	ID        int64   `bun:"id,pk,autoincrement"`
	UID       string  `bun:"uid,notnull"`
	ListID    int64   `bun:"list_id,notnull"`
	Label     string  `bun:"label,notnull"`
	Quantity  string  `bun:"quantity,notnull"`
	Note      string  `bun:"note,notnull"`
	DueOn     string  `bun:"due_on,notnull"`
	Position  float64 `bun:"position,notnull"`
	DoneAt    string  `bun:"done_at,notnull"`
	DoneByID  *int64  `bun:"done_by_id"`
	AddedByID int64   `bun:"added_by_id,notnull"`
	// AddedByTokenID is null for anything a browser added, which is most things.
	AddedByTokenID *int64 `bun:"added_by_token_id"`
	CreatedAt      string `bun:"created_at,notnull"`
	UpdatedAt      string `bun:"updated_at,notnull"`
	DeletedAt      string `bun:"deleted_at,notnull"`
}

func (m itemModel) toItem() (Item, error) {
	doneAt, err := parseTime(m.DoneAt)
	if err != nil {
		return Item{}, err
	}
	createdAt, err := parseTime(m.CreatedAt)
	if err != nil {
		return Item{}, err
	}
	updatedAt, err := parseTime(m.UpdatedAt)
	if err != nil {
		return Item{}, err
	}
	deletedAt, err := parseTime(m.DeletedAt)
	if err != nil {
		return Item{}, err
	}

	item := Item{
		ID: m.ID, UID: m.UID, ListID: m.ListID, Label: m.Label, Quantity: m.Quantity,
		DueOn: m.DueOn, Note: m.Note, Position: m.Position, DoneAt: doneAt, AddedByID: m.AddedByID,
		CreatedAt: createdAt, UpdatedAt: updatedAt, DeletedAt: deletedAt,
	}
	if m.DoneByID != nil {
		item.DoneByID = *m.DoneByID
	}
	if m.AddedByTokenID != nil {
		item.AddedByTokenID = *m.AddedByTokenID
	}
	return item, nil
}
