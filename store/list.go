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
}

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

// CreateList adds a List.
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

// ListsForMember returns every live List a Member can reach: their own, and everything
// shared with the whole Instance. Named sharing arrives with Groups in M6.
func (s *sqlStore) ListsForMember(ctx context.Context, memberID int64) ([]List, error) {
	var rows []listModel
	err := s.db.NewSelect().
		Model(&rows).
		Where("deleted_at = ''").
		Where("owner_id = ? OR sharing = ?", memberID, string(SharingInstance)).
		Order("name ASC").
		Scan(ctx)
	if err != nil {
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

// RenameList changes a List's name.
func (s *sqlStore) RenameList(ctx context.Context, uid string, name string, at time.Time) error {
	if name == "" {
		return errors.New("store: list name is required")
	}
	return s.updateList(ctx, uid, at, func(q *bun.UpdateQuery) *bun.UpdateQuery {
		return q.Set("name = ?", name)
	})
}

// SetListSharing changes who can reach a List.
func (s *sqlStore) SetListSharing(ctx context.Context, uid string, sharing Sharing, canEdit bool, at time.Time) error {
	return s.updateList(ctx, uid, at, func(q *bun.UpdateQuery) *bun.UpdateQuery {
		return q.Set("sharing = ?", string(sharing)).Set("can_edit = ?", canEdit)
	})
}

// DeleteList removes a List, and with it the Items on it. The removal is soft, so a
// List deleted by mistake is recoverable.
func (s *sqlStore) DeleteList(ctx context.Context, uid string, at time.Time) error {
	return s.updateList(ctx, uid, at, func(q *bun.UpdateQuery) *bun.UpdateQuery {
		return q.Set("deleted_at = ?", formatTime(at))
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

// listModel is the stored shape of a List.
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
	return List{
		ID: m.ID, UID: m.UID, Name: m.Name, OwnerID: m.OwnerID,
		Sharing: Sharing(m.Sharing), CanEdit: m.CanEdit,
		CreatedAt: createdAt, UpdatedAt: updatedAt, DeletedAt: deletedAt,
	}, nil
}

// listPinModel is one Member's pin on one List.
type listPinModel struct {
	bun.BaseModel `bun:"table:list_pin,alias:list_pin"`

	MemberID int64 `bun:"member_id,pk"`
	ListID   int64 `bun:"list_id,pk"`
}
