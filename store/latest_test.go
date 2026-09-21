package store

import (
	"database/sql"
	"fmt"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"

	_ "modernc.org/sqlite"
)

/*
LATEST.sql says what the schema is at head, and nothing made it true.

It is never run — the numbered migrations build the database — so it is documentation,
and documentation nothing checks drifts the first time somebody adds a migration in a
hurry. What it drifts into is worse than nothing: a file that states the schema, is
read by whoever is trying to understand the database, and is wrong.

Both are built here and compared by shape rather than by spelling. A migration appends
a column with ALTER TABLE and LATEST.sql declares it where it belongs, so the order the
two arrive in is not a difference and neither is a comment. A table, a column, a type,
a default or an index that exists in one and not the other is.

SQLite only. The Postgres LATEST.sql is the same file for the other driver and is still
unchecked, because checking it needs a server the suite skips without. The two are
edited together, so drift in one is a fair warning about the other.
*/
func TestLatestSQLSaysWhatTheMigrationsBuild(t *testing.T) {
	migrated := shapeOf(t, fromMigrations(t))
	documented := shapeOf(t, fromLatest(t))

	for name, stated := range documented {
		built, ok := migrated[name]
		if !ok {
			t.Errorf("LATEST.sql has %s and the migrations do not build it", name)
			continue
		}
		if stated != built {
			t.Errorf("%s differs.\n LATEST.sql: %s\n migrations: %s", name, stated, built)
		}
	}
	for name := range migrated {
		if _, ok := documented[name]; !ok {
			t.Errorf("the migrations build %s and LATEST.sql does not say so", name)
		}
	}
}

// fromMigrations opens a database the ordinary way, which applies every migration.
func fromMigrations(t *testing.T) *sql.DB {
	t.Helper()

	path := filepath.Join(t.TempDir(), "migrated.db")
	s, err := OpenSQLite(t.Context(), path)
	if err != nil {
		t.Fatalf("OpenSQLite: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })

	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatalf("reopen migrated: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

// fromLatest builds a database by running the file that claims to state the schema.
func fromLatest(t *testing.T) *sql.DB {
	t.Helper()

	statements, err := migrations.ReadFile("migration/sqlite/LATEST.sql")
	if err != nil {
		t.Fatalf("read LATEST.sql: %v", err)
	}

	db, err := sql.Open("sqlite", filepath.Join(t.TempDir(), "latest.db"))
	if err != nil {
		t.Fatalf("open latest: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	if _, err := db.ExecContext(t.Context(), string(statements)); err != nil {
		t.Fatalf("run LATEST.sql: %v", err)
	}
	return db
}

/*
shapeOf reads what a database holds, as one line per object.

A table is its columns sorted by name, so appending one and declaring it in place read
the same. Everything else — indexes, triggers, the search tables — is its own statement
with the comments taken out.

schema_migration is left out: it is how migrating works rather than part of the schema,
so LATEST.sql has no business stating it.
*/
func shapeOf(t *testing.T, db *sql.DB) map[string]string {
	t.Helper()

	objects, err := db.QueryContext(t.Context(),
		`SELECT type, name, COALESCE(sql, '') FROM sqlite_master
		 WHERE name NOT LIKE 'sqlite_%' AND name != 'schema_migration'`)
	if err != nil {
		t.Fatalf("read schema: %v", err)
	}
	defer func() { _ = objects.Close() }()

	found := make(map[string]string)
	for objects.Next() {
		var kind, name, statement string
		if err := objects.Scan(&kind, &name, &statement); err != nil {
			t.Fatalf("scan schema: %v", err)
		}
		if kind == "table" {
			found[kind+" "+name] = columnsOf(t, db, name)
			continue
		}
		found[kind+" "+name] = flatten(statement)
	}
	if err := objects.Err(); err != nil {
		t.Fatalf("read schema: %v", err)
	}
	if len(found) == 0 {
		t.Fatal("read an empty schema, so this test is checking nothing")
	}
	return found
}

// columnsOf is a table's columns sorted by name, each with what it holds.
func columnsOf(t *testing.T, db *sql.DB, table string) string {
	t.Helper()

	rows, err := db.QueryContext(t.Context(), `SELECT name, type, "notnull", COALESCE(dflt_value, ''), pk
		FROM pragma_table_info(?)`, table)
	if err != nil {
		t.Fatalf("read %s: %v", table, err)
	}
	defer func() { _ = rows.Close() }()

	columns := make([]string, 0, 8)
	for rows.Next() {
		var name, kind, fallback string
		var notNull, primary int
		if err := rows.Scan(&name, &kind, &notNull, &fallback, &primary); err != nil {
			t.Fatalf("scan %s: %v", table, err)
		}
		columns = append(columns, fmt.Sprintf("%s %s notnull=%d default=%q pk=%d",
			name, kind, notNull, fallback, primary))
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("read %s: %v", table, err)
	}
	if len(columns) == 0 {
		t.Fatalf("%s has no columns, so this test is checking nothing", table)
	}
	sort.Strings(columns)
	return strings.Join(columns, ", ")
}

// lineComment is a `--` comment to the end of its line.
var lineComment = regexp.MustCompile(`--[^\n]*`)

// flatten makes one line of a statement, so indentation and commentary are not
// differences between a file written to be read and one written to be run.
func flatten(statement string) string {
	return strings.Join(strings.Fields(lineComment.ReplaceAllString(statement, "")), " ")
}
