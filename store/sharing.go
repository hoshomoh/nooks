package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/uptrace/bun"
)

// Group is a named set of Members that exists only as a shortcut for sharing.
//
// It carries no permissions of its own: sharing a List with a Group reaches everyone in
// it, and removing someone from a Group takes away the Lists they got through it and
// nothing else.
type Group struct {
	ID        int64
	UID       string
	Name      string
	CreatedAt time.Time
}

// Share is one row of who a List is shared with by name — a Member or a Group, never
// both.
type Share struct {
	ID       int64
	ListID   int64
	MemberID int64
	GroupID  int64
}

// CreateGroup adds a Group.
func (s *sqlStore) CreateGroup(ctx context.Context, uid, name string, at time.Time) (Group, error) {
	if name == "" {
		return Group{}, errors.New("store: group name is required")
	}
	row := &groupModel{UID: uid, Name: name, CreatedAt: formatTime(at)}
	if _, err := s.db.NewInsert().Model(row).Returning("*").Exec(ctx); err != nil {
		return Group{}, fmt.Errorf("create group: %w", err)
	}
	return row.toGroup()
}

// GroupByUID finds a Group.
func (s *sqlStore) GroupByUID(ctx context.Context, uid string) (Group, error) {
	row := new(groupModel)
	if err := s.db.NewSelect().Model(row).Where("uid = ?", uid).Scan(ctx); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Group{}, ErrNotFound
		}
		return Group{}, fmt.Errorf("read group: %w", err)
	}
	return row.toGroup()
}

// Groups lists every Group on the Instance, by name.
func (s *sqlStore) Groups(ctx context.Context) ([]Group, error) {
	var rows []groupModel
	if err := s.db.NewSelect().Model(&rows).Order("name ASC").Scan(ctx); err != nil {
		return nil, fmt.Errorf("read groups: %w", err)
	}
	groups := make([]Group, 0, len(rows))
	for _, row := range rows {
		group, err := row.toGroup()
		if err != nil {
			return nil, err
		}
		groups = append(groups, group)
	}
	return groups, nil
}

// AddToGroup puts a Member in a Group. Adding twice is not an error.
func (s *sqlStore) AddToGroup(ctx context.Context, groupID, memberID int64) error {
	_, err := s.db.NewInsert().
		Model(&groupMemberModel{GroupID: groupID, MemberID: memberID}).
		On("CONFLICT (group_id, member_id) DO NOTHING").
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("add to group: %w", err)
	}
	return nil
}

// RemoveFromGroup takes a Member out of a Group, and with it the Lists they reached
// through it.
func (s *sqlStore) RemoveFromGroup(ctx context.Context, groupID, memberID int64) error {
	_, err := s.db.NewDelete().
		Model((*groupMemberModel)(nil)).
		Where("group_id = ? AND member_id = ?", groupID, memberID).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("remove from group: %w", err)
	}
	return nil
}

// ReplaceGroupMembers sets exactly who is in a Group, replacing whoever was there.
//
// Membership is one decision for the same reason sharing is: an Admin picks the people
// who are in a Group, rather than adding and removing them one at a time and hoping the
// result is what they meant.
func (s *sqlStore) ReplaceGroupMembers(ctx context.Context, groupID int64, memberIDs []int64) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.NewDelete().
		Model((*groupMemberModel)(nil)).
		Where("group_id = ?", groupID).
		Exec(ctx); err != nil {
		return fmt.Errorf("clear group members: %w", err)
	}

	rows := make([]groupMemberModel, 0, len(memberIDs))
	for _, id := range memberIDs {
		rows = append(rows, groupMemberModel{GroupID: groupID, MemberID: id})
	}
	if len(rows) > 0 {
		if _, err := tx.NewInsert().Model(&rows).Exec(ctx); err != nil {
			return fmt.Errorf("write group members: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit group members: %w", err)
	}
	return nil
}

// GroupMemberIDs lists who is in a Group.
func (s *sqlStore) GroupMemberIDs(ctx context.Context, groupID int64) ([]int64, error) {
	var ids []int64
	err := s.db.NewSelect().
		Model((*groupMemberModel)(nil)).
		Column("member_id").
		Where("group_id = ?", groupID).
		Scan(ctx, &ids)
	if err != nil {
		return nil, fmt.Errorf("read group members: %w", err)
	}
	return ids, nil
}

// ReplaceListShares sets exactly who a List is shared with by name, replacing whatever
// was there. Sharing is one decision, not a sequence of additions.
func (s *sqlStore) ReplaceListShares(ctx context.Context, listID int64, memberIDs, groupIDs []int64) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.NewDelete().
		Model((*listShareModel)(nil)).
		Where("list_id = ?", listID).
		Exec(ctx); err != nil {
		return fmt.Errorf("clear list shares: %w", err)
	}

	rows := make([]listShareModel, 0, len(memberIDs)+len(groupIDs))
	for _, id := range memberIDs {
		member := id
		rows = append(rows, listShareModel{ListID: listID, MemberID: &member})
	}
	for _, id := range groupIDs {
		group := id
		rows = append(rows, listShareModel{ListID: listID, GroupID: &group})
	}
	if len(rows) > 0 {
		if _, err := tx.NewInsert().Model(&rows).Exec(ctx); err != nil {
			return fmt.Errorf("write list shares: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit list shares: %w", err)
	}
	return nil
}

// ListShares reads who a List is shared with by name.
func (s *sqlStore) ListShares(ctx context.Context, listID int64) ([]Share, error) {
	var rows []listShareModel
	if err := s.db.NewSelect().Model(&rows).Where("list_id = ?", listID).Scan(ctx); err != nil {
		return nil, fmt.Errorf("read list shares: %w", err)
	}
	shares := make([]Share, 0, len(rows))
	for _, row := range rows {
		shares = append(shares, row.toShare())
	}
	return shares, nil
}

// ListsSharedWithGroup is every live List a Group reaches.
//
// A Group is only ever a shortcut for sharing, so what it reaches is the whole of what
// it does — and the Groups page says so rather than making an Admin work it out.
func (s *sqlStore) ListsSharedWithGroup(ctx context.Context, groupID int64) ([]List, error) {
	var rows []listModel
	err := s.db.NewSelect().
		Model(&rows).
		Where("deleted_at = ''").
		Where("id IN (?)", s.db.NewSelect().
			Model((*listShareModel)(nil)).
			Column("list_id").
			Where("group_id = ?", groupID)).
		Order("name ASC").
		Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("read lists shared with group: %w", err)
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

// SharedListIDs is every List reaching a Member by name — directly, or through a Group
// they are in.
func (s *sqlStore) SharedListIDs(ctx context.Context, memberID int64) ([]int64, error) {
	var ids []int64
	err := s.db.NewSelect().
		Model((*listShareModel)(nil)).
		Column("list_id").
		Where("member_id = ?", memberID).
		WhereOr("group_id IN (?)",
			s.db.NewSelect().
				Model((*groupMemberModel)(nil)).
				Column("group_id").
				Where("member_id = ?", memberID)).
		Scan(ctx, &ids)
	if err != nil {
		return nil, fmt.Errorf("read shared lists: %w", err)
	}
	return ids, nil
}

// groupModel is the stored shape of a Group.
type groupModel struct {
	bun.BaseModel `bun:"table:member_group,alias:member_group"`

	ID        int64  `bun:"id,pk,autoincrement"`
	UID       string `bun:"uid,notnull"`
	Name      string `bun:"name,notnull"`
	CreatedAt string `bun:"created_at,notnull"`
}

func (m groupModel) toGroup() (Group, error) {
	createdAt, err := parseTime(m.CreatedAt)
	if err != nil {
		return Group{}, err
	}
	return Group{ID: m.ID, UID: m.UID, Name: m.Name, CreatedAt: createdAt}, nil
}

// groupMemberModel is one Member's membership of one Group.
type groupMemberModel struct {
	bun.BaseModel `bun:"table:group_member,alias:group_member"`

	GroupID  int64 `bun:"group_id,pk"`
	MemberID int64 `bun:"member_id,pk"`
}

// listShareModel is one row of named sharing. Exactly one of the two ids is set.
type listShareModel struct {
	bun.BaseModel `bun:"table:list_share,alias:list_share"`

	ID       int64  `bun:"id,pk,autoincrement"`
	ListID   int64  `bun:"list_id,notnull"`
	MemberID *int64 `bun:"member_id"`
	GroupID  *int64 `bun:"group_id"`
}

func (m listShareModel) toShare() Share {
	share := Share{ID: m.ID, ListID: m.ListID}
	if m.MemberID != nil {
		share.MemberID = *m.MemberID
	}
	if m.GroupID != nil {
		share.GroupID = *m.GroupID
	}
	return share
}
