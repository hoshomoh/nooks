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
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt time.Time
}

// Done reports whether the Item has been ticked.
func (i Item) Done() bool { return !i.DoneAt.IsZero() }

// CreateItemParams is everything needed to add an Item. Position is worked out by the
// store, because "put it at the end" is a fact about the List, not the caller.
type CreateItemParams struct {
	UID       string
	ListID    int64
	Label     string
	Quantity  string
	DueOn     string
	AddedByID int64
	At        time.Time
}

// CreateItem appends an Item to a List.
func (s *sqlStore) CreateItem(ctx context.Context, params CreateItemParams) (Item, error) {
	if params.Label == "" {
		return Item{}, errors.New("store: item label is required")
	}

	position, err := s.nextPosition(ctx, params.ListID)
	if err != nil {
		return Item{}, err
	}

	row := &itemModel{
		UID:       params.UID,
		ListID:    params.ListID,
		Label:     params.Label,
		Quantity:  params.Quantity,
		DueOn:     params.DueOn,
		Position:  position,
		AddedByID: params.AddedByID,
		CreatedAt: formatTime(params.At),
		UpdatedAt: formatTime(params.At),
	}
	if _, err := s.db.NewInsert().Model(row).Returning("*").Exec(ctx); err != nil {
		return Item{}, fmt.Errorf("create item: %w", err)
	}

	item, err := row.toItem()
	if err != nil {
		return Item{}, err
	}
	if err := s.indexItem(ctx, item); err != nil {
		return Item{}, err
	}
	return item, nil
}

// ItemsOnList returns a List's live Items in their manual order.
func (s *sqlStore) ItemsOnList(ctx context.Context, listID int64) ([]Item, error) {
	var rows []itemModel
	err := s.db.NewSelect().
		Model(&rows).
		Where("list_id = ? AND deleted_at = ''", listID).
		Order("position ASC").
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
		Order("item.due_on ASC", "item.position ASC")
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
	result, err := s.db.NewUpdate().
		Model((*itemModel)(nil)).
		Set("done_at = ?", formatTime(at)).
		Set("done_by_id = ?", doneBy).
		Set("updated_at = ?", formatTime(at)).
		Where("uid = ? AND deleted_at = ''", uid).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("tick item: %w", err)
	}
	return requireOneRow(result, "item")
}

// SetItemNotDone unticks an Item.
func (s *sqlStore) SetItemNotDone(ctx context.Context, uid string, at time.Time) error {
	result, err := s.db.NewUpdate().
		Model((*itemModel)(nil)).
		Set("done_at = ?", "").
		Set("done_by_id = ?", nil).
		Set("updated_at = ?", formatTime(at)).
		Where("uid = ? AND deleted_at = ''", uid).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("untick item: %w", err)
	}
	return requireOneRow(result, "item")
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
	result, err := s.db.NewUpdate().
		Model((*itemModel)(nil)).
		Set("deleted_at = ?", formatTime(at)).
		Set("updated_at = ?", formatTime(at)).
		Where("uid = ? AND deleted_at = ''", uid).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("delete item: %w", err)
	}
	if err := requireOneRow(result, "item"); err != nil {
		return err
	}
	if err := s.Unindex(ctx, KindItem, uid); err != nil {
		return err
	}
	return s.Unindex(ctx, KindNote, uid)
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
	if err := s.Index(ctx, IndexEntry{
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
		return s.Unindex(ctx, KindNote, item.UID)
	}
	return s.Index(ctx, IndexEntry{
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

// toItems converts a page of rows.
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

// itemModel is the stored shape of an Item.
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
	CreatedAt string  `bun:"created_at,notnull"`
	UpdatedAt string  `bun:"updated_at,notnull"`
	DeletedAt string  `bun:"deleted_at,notnull"`
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
	return item, nil
}
