package store

import (
	"context"
	"fmt"
)

/*
InstanceStats is what the About page reports: how much this Instance is holding.

Counts rather than a dashboard: somebody self-hosting wants to know roughly how big it
is before deciding whether to back it up.
*/
type InstanceStats struct {
	Members int
	Lists   int
	Items   int
	// StorageBytes is how much room the database is taking, or zero where the driver
	// cannot say. Zero renders as nothing rather than as "0 B": a wrong number about
	// somebody's disk is worse than no number.
	StorageBytes int64
	// Driver is what is holding it, e.g. "sqlite". Shown beside the size, because
	// "one file you can copy" and "a server you back up" are different answers.
	Driver string
}

// Stats counts what the Instance holds.
//
// Deleted Lists and Items are not counted: a Member asking how much is here means what
// they can see, and a soft-deleted row is not that.
func (s *sqlStore) Stats(ctx context.Context) (InstanceStats, error) {
	members, err := s.CountMembers(ctx)
	if err != nil {
		return InstanceStats{}, err
	}

	lists, err := s.db.NewSelect().Model((*listModel)(nil)).Where("deleted_at = ''").Count(ctx)
	if err != nil {
		return InstanceStats{}, fmt.Errorf("count lists: %w", err)
	}
	// Deleting a List leaves its Items as they were, so restoring it brings them back.
	// That makes an Item's own deleted_at only half the answer: one on a List in the bin
	// is gone as far as anybody can see, and counting it reported Items with no List.
	items, err := s.db.NewSelect().
		Model((*itemModel)(nil)).
		Join("JOIN list ON list.id = item.list_id").
		Where("item.deleted_at = '' AND list.deleted_at = ''").
		Count(ctx)
	if err != nil {
		return InstanceStats{}, fmt.Errorf("count items: %w", err)
	}

	return InstanceStats{
		Members:      members,
		Lists:        int(lists),
		Items:        int(items),
		StorageBytes: s.storageBytes(ctx),
		Driver:       s.name,
	}, nil
}

/*
storageBytes asks the database how much room it is taking.

Driver-specific and best effort: SQLite multiplies page count by page size, Postgres
has a function. Anything that goes wrong answers zero, because the size of the database
is never a reason to fail the page.
*/
func (s *sqlStore) storageBytes(ctx context.Context) int64 {
	var size int64
	query := "SELECT page_count * page_size FROM pragma_page_count(), pragma_page_size()"
	if s.name == postgresDriver {
		query = "SELECT pg_database_size(current_database())"
	}
	if err := s.db.QueryRowContext(ctx, query).Scan(&size); err != nil {
		return 0
	}
	return size
}
