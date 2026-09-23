package conventions

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

/*
writeSafety names every store write whose answer depends on what it can see, and says
what makes it right on Postgres.

SQLite has one writer, so a statement there decides against everything that has happened
and a write shaped as one statement is atomic by construction. Postgres decides against a
snapshot taken when the statement began, so two callers in that gap can both find the
world as it was before either of them. A rule written as a condition inside a write is
therefore true on one driver and only usually true on the other.

That is not a hypothetical. It has been found three times: two Items taking the same
position, a twenty-sixth request passing a ceiling of twenty-five, and a List reporting
five open when it had eight. Each was written as one statement, each was believed atomic
for that reason, and the comment above each said so.

So the question is asked of every one of them, once, here. An entry is a sentence
somebody had to think about, not a box to tick: "nothing holds it" is a perfectly good
answer where it is true, and it has to say why.
*/
var writeSafety = map[string]string{
	"CreateItems": "onOneList holds the List, and the recount is inside the same " +
		"transaction as the insert",
	"changeItemCount": "onOneList holds the List for the update and the recount together",
	"appendItem": "its one caller runs it inside onOneList, which is where the List is " +
		"held; it takes the transaction to write through rather than reaching for the pool",
	"recount": "never called on its own: every caller has the List held, and the count " +
		"is written with the change that made it wrong",

	"CreateJoinRequest":  "oneAtATime holds the whole join queue, which is what both the per-email rule and the ceiling are about",
	"CreateResetRequest": "oneAtATime holds the reset queue, for the one-per-Member rule",

	"RestoreList": "nothing holds it, and nothing needs to: the read only feeds the " +
		"search index, and what decides is the update's own WHERE, which requireOneRow " +
		"turns into ErrNotFound for whoever came second",
	"DeleteUnreadableActivity": "nothing holds it, and nothing needs to: it is a sweep " +
		"of what is already unreachable, it never removes a request nobody has answered, " +
		"and an entry it misses because somebody wrote during it is swept the next time",
}

// TestEveryWriteThatLooksSaysWhyItIsSafe holds writeSafety to the store as it is.
func TestEveryWriteThatLooksSaysWhyItIsSafe(t *testing.T) {
	writes := writesThatLookIn(t)
	if len(writes) < 6 {
		t.Fatalf("found %d writes that read, too few to be reading the whole store", len(writes))
	}

	for _, name := range writes {
		why, said := writeSafety[name]
		if !said {
			t.Errorf("%s decides a write from something it read, and writeSafety does not "+
				"say what makes that right on Postgres", name)
			continue
		}
		if len(why) < 40 {
			t.Errorf("writeSafety[%s] is too short to be a reason somebody thought about", name)
		}
	}
	for name := range writeSafety {
		if !slices.Contains(writes, name) {
			t.Errorf("writeSafety still names %s, which no longer decides a write from a read", name)
		}
	}
}

// writesThatLookIn is every store function that changes rows and also reads them, or
// tests a condition against them, in deciding what to write.
func writesThatLookIn(t *testing.T) []string {
	t.Helper()

	paths, err := filepath.Glob(filepath.Join("..", "..", "store", "*.go"))
	if err != nil {
		t.Fatalf("glob the store: %v", err)
	}

	var found []string
	for _, path := range paths {
		if strings.HasSuffix(path, "_test.go") {
			continue
		}
		parsed, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
		if err != nil {
			t.Fatalf("parse %s: %v", path, err)
		}
		for _, one := range parsed.Decls {
			fn, ok := one.(*ast.FuncDecl)
			if !ok || fn.Body == nil {
				continue
			}
			writes, looks := whatItDoes(fn.Body)
			if writes && looks {
				found = append(found, fn.Name.Name)
			}
		}
	}
	slices.Sort(found)
	return found
}

/*
whatItDoes reports whether a body changes rows and whether it reads them.

Both the query builder and the SQL written out by hand count, because the faults this is
for were written both ways: the Item positions were a sub-select inside a builder call,
and the request queue was a string.
*/
func whatItDoes(body *ast.BlockStmt) (writes, looks bool) {
	ast.Inspect(body, func(node ast.Node) bool {
		switch part := node.(type) {
		case *ast.SelectorExpr:
			switch part.Sel.Name {
			case "NewInsert", "NewUpdate", "NewDelete":
				writes = true
			case "NewSelect":
				looks = true
			}
		case *ast.BasicLit:
			if part.Kind != token.STRING {
				return true
			}
			sql := strings.ToUpper(part.Value)
			if strings.Contains(sql, "INSERT ") || strings.Contains(sql, "UPDATE ") ||
				strings.Contains(sql, "DELETE ") {
				writes = true
			}
			if strings.Contains(sql, "SELECT ") {
				looks = true
			}
		}
		return true
	})
	return writes, looks
}
