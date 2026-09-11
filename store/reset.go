package store

import (
	"context"
	"fmt"
)

/*
tablesInDeleteOrder is every table holding an Instance's own data, children first.

Written out rather than discovered, and children before parents even though the schema
cascades: a list the reader can check against the migrations is the only way to be sure
nothing is missed, and relying on cascade would mean a future table with no foreign key
silently surviving a delete that promised to take everything.

The search index goes with them. It is derived, and an index of things that no longer
exist would answer queries about them.
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
the Instance. What is left is a database with a schema and nothing in it, which is what
the app reads as "not set up yet".

One transaction, so an Instance is never half-deleted. A reset that failed in the middle
would leave somebody signed in to an Instance with no Lists and no way to explain it.
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
