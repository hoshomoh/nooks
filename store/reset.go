package store

import (
	"context"
	"fmt"
)

/*
tablesInDeleteOrder is every table holding an Instance's own data, children first.

Written out rather than left to cascade, so the list can be checked against the
migrations. A future table with no foreign key would otherwise survive a delete that
promised to take everything. The search index goes with them: it is derived, and an
index of things that no longer exist would answer queries about them.
*/
var tablesInDeleteOrder = []string{
	"access_token_list",
	"access_token",
	"list_share",
	"list_pin",
	"group_member",
	"member_group",
	"activity",
	"item",
	"list",
	"session",
	"reset_request",
	"join_request",
	"member",
	"setting",
	"search_index",
}

/*
ResetInstance empties the Instance and returns it to first run.

Everything goes: Members, Lists, Items, Notes, Access tokens and the settings that named
the Instance. What is left is a schema and nothing in it, which the app reads as "not
set up yet".

One transaction, so an Instance is never half-deleted.
*/
func (s *sqlStore) ResetInstance(ctx context.Context) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin instance reset: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	for _, table := range tablesInDeleteOrder {
		if _, err := tx.NewRaw("DELETE FROM " + table).Exec(ctx); err != nil {
			return fmt.Errorf("empty %s: %w", table, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit instance reset: %w", err)
	}
	return nil
}
