package store

import (
	"database/sql"
	"io/fs"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

/*
An Instance that has been running upgrades without losing what is on it.

Every other test here builds the schema from nothing, which is the one case an upgrade
is not. A migration that is fine against an empty database and wrong against a full one
— a backfill that misses, a column added NOT NULL with no default, a rename that drops
what was in it — passes all of them and is found by whoever runs it on a household that
has been going for a year.

So this builds the database as it stood at an older version, puts a household in it, and
starts the current one against it.

The Access token is the case the documentation makes a promise about: when one
permission became three abilities, tokens that already existed kept everything they had.
`operations/upgrade.mdx` says so under "What a migration will not do", and a key that
quietly stops being able to do what it did is a key somebody debugs for an afternoon.
*/
func TestUpgradingAnInstanceKeepsWhatIsOnIt(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nooks.db")
	upTo(t, path, "0011_token_all_lists.sql")
	seedTheOldWay(t, path)

	// What an upgrade is: the new version, against the directory that was already there.
	s, err := OpenSQLite(t.Context(), path)
	if err != nil {
		t.Fatalf("open the upgraded instance: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })

	member, err := s.MemberByEmail(t.Context(), "anna@brunnen.lan")
	if err != nil {
		t.Fatalf("the Member did not survive the upgrade: %v", err)
	}
	if member.Name != "Anna" {
		t.Errorf("Name = %q, want the one that was there", member.Name)
	}

	list, err := s.ListByUID(t.Context(), "lst_groceries")
	if err != nil {
		t.Fatalf("the List did not survive the upgrade: %v", err)
	}
	items, err := s.ItemsOnList(t.Context(), list.ID)
	if err != nil {
		t.Fatalf("ItemsOnList: %v", err)
	}
	if len(items) != 2 {
		t.Errorf("the List holds %d Items, want the 2 that were on it", len(items))
	}

	// The counts were added by a later migration than the Items, so they have to be
	// right for rows nothing has touched since.
	if list.OpenCount != 1 || list.DoneCount != 1 {
		t.Errorf("counts are %d open and %d done, want 1 and 1", list.OpenCount, list.DoneCount)
	}

	for _, want := range []struct {
		hash  string
		read  bool
		write bool
		gone  bool
	}{
		// A WRITE token could already delete, so it keeps that too.
		{hash: "hash_writer", read: true, write: true, gone: true},
		{hash: "hash_reader", read: true, write: false, gone: false},
	} {
		token, err := s.AccessTokenByHash(t.Context(), want.hash)
		if err != nil {
			t.Fatalf("the token did not survive the upgrade: %v", err)
		}
		got := token.Abilities
		if got.Read != want.read || got.Write != want.write || got.Delete != want.gone {
			t.Errorf("%s came out read=%v write=%v delete=%v, want %v %v %v",
				token.Name, got.Read, got.Write, got.Delete, want.read, want.write, want.gone)
		}
	}
}

// upTo builds a database with every migration through the named one, and records them
// the way the runner would, so that opening it afterwards applies only what is left.
func upTo(t *testing.T, path, last string) {
	t.Helper()

	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer func() { _ = db.Close() }()

	if _, err := db.ExecContext(t.Context(),
		`CREATE TABLE schema_migration (name TEXT PRIMARY KEY)`); err != nil {
		t.Fatalf("create the ledger: %v", err)
	}

	applied := 0
	for _, name := range migrationNames(t) {
		if name > last {
			break
		}
		statements, err := migrations.ReadFile("migration/sqlite/" + name)
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		if _, err := db.ExecContext(t.Context(), string(statements)); err != nil {
			t.Fatalf("apply %s: %v", name, err)
		}
		if _, err := db.ExecContext(t.Context(),
			`INSERT INTO schema_migration (name) VALUES (?)`, name); err != nil {
			t.Fatalf("record %s: %v", name, err)
		}
		applied++
	}
	if applied < 5 {
		t.Fatalf("applied %d migrations, so this is not building an older database", applied)
	}
}

// migrationNames is every sqlite migration, in the order the runner applies them.
func migrationNames(t *testing.T) []string {
	t.Helper()

	entries, err := fs.ReadDir(migrations, "migration/sqlite")
	if err != nil {
		t.Fatalf("read the migrations: %v", err)
	}
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || name == "LATEST.sql" || !strings.HasSuffix(name, ".sql") {
			continue
		}
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// seedTheOldWay writes rows in the shape the older schema had, which is the only way to
// get a database that looks like one somebody has been using.
func seedTheOldWay(t *testing.T, path string) {
	t.Helper()

	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer func() { _ = db.Close() }()

	at := time.Date(2026, 3, 1, 9, 0, 0, 0, time.UTC).Format(time.RFC3339)
	for _, statement := range []string{
		`INSERT INTO member (id, uid, name, email, role, password_hash, created_at)
		 VALUES (1, 'mem_anna', 'Anna', 'anna@brunnen.lan', 'ADMIN', 'hash', '` + at + `')`,
		`INSERT INTO list (id, uid, name, owner_id, created_at, updated_at)
		 VALUES (1, 'lst_groceries', 'Groceries', 1, '` + at + `', '` + at + `')`,
		`INSERT INTO item (id, uid, list_id, label, position, added_by_id, created_at, updated_at)
		 VALUES (1, 'itm_milk', 1, 'Milk', 1.0, 1, '` + at + `', '` + at + `')`,
		`INSERT INTO item (id, uid, list_id, label, position, added_by_id, created_at, updated_at, done_at)
		 VALUES (2, 'itm_bread', 1, 'Bread', 2.0, 1, '` + at + `', '` + at + `', '` + at + `')`,
		`INSERT INTO access_token (id, uid, member_id, name, token_hash, permission, created_at)
		 VALUES (1, 'tok_w', 1, 'Kitchen tablet', 'hash_writer', 'WRITE', '` + at + `')`,
		`INSERT INTO access_token (id, uid, member_id, name, token_hash, permission, created_at)
		 VALUES (2, 'tok_r', 1, 'Recipe importer', 'hash_reader', 'READ', '` + at + `')`,
	} {
		if _, err := db.ExecContext(t.Context(), statement); err != nil {
			t.Fatalf("seed: %v\n%s", err, statement)
		}
	}
}
