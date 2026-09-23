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
readBounds names every store read that pulls rows into a slice, and what stops it
growing.

A read of one row is bounded by being one row. A read of a collection is bounded by
whatever the collection counts, and that is a fact about the product rather than about
the query: "how many people live here" is a number an Admin controls, and "how much this
household has ever written" is not. The difference is the whole of what makes one of
these safe and another a page that takes five seconds.

Written down because it kept being worked out again. Three passes have counted these by
hand, the count has been wrong twice, and one line of the log went stale the moment the
join queue gained a ceiling. A read added with no thought about this is now a test
failure rather than something the next pass rediscovers.

Both directions are checked. A read that is not here fails, and a name here that is no
longer a collection read fails too, because a stale entry is exactly the drift this is
for.
*/
var readBounds = map[string]string{
	// Bounded by the query itself.
	"ActivityFor": "ActivityLimit, plus the requests nobody has answered, which the queue rules bound",
	"ListsPage":   "the page size the caller asks for",
	"Search":      "the result limit",
	"listsWhere":  "the limit its callers pass",

	// Bounded by how many people, groups, keys or shares a household has, which is a
	// number an Admin decides rather than one that runs away.
	"AdminIDs":             "how many Admins there are",
	"Members":              "how many people live here",
	"Groups":               "how many Groups an Admin made",
	"GroupMemberIDs":       "how many people are in one Group",
	"ListShares":           "how many people and Groups one List is shared with",
	"ListsSharedWithGroup": "how many Lists are shared with one Group",
	"SharedListIDs":        "how many Lists reach one Member",
	"ListsForMember":       "how many Lists a household has",
	"ListsOwnedBy":         "how many Lists one Member made",
	"PinnedListIDs":        "how many Lists one Member pinned",
	"accessTokens":         "how many keys a household made",
	"TokenListIDs":         "how many Lists one key names",
	"InstanceSettings":     "the settings keys, which are a fixed list in that file",
	"PendingJoinRequests":  "PendingJoinLimit",
	"PendingResetRequests": "one waiting reset per Member, so by how many Members there are",

	// Bounded by what the caller hands over, which the API caps before it gets here.
	"CreateItems": "how many Items one call may add",

	// Bounded by nothing, and known.
	"ItemsOnList":         "nothing, and that is the parked question 8",
	"DatedItemsForMember": "nothing, and that is the parked question 9: 4.7s at 30,083 rows",
}

// TestEveryCollectionReadSaysWhatBoundsIt holds readBounds to the store as it is.
func TestEveryCollectionReadSaysWhatBoundsIt(t *testing.T) {
	reads := collectionReadsIn(t)
	if len(reads) < 20 {
		t.Fatalf("found %d collection reads, too few to be reading the whole store", len(reads))
	}

	for _, name := range reads {
		if _, said := readBounds[name]; !said {
			t.Errorf("%s reads a collection and readBounds does not say what stops it growing",
				name)
		}
	}
	for name := range readBounds {
		if !slices.Contains(reads, name) {
			t.Errorf("readBounds still names %s, which no longer reads a collection", name)
		}
	}
}

/*
collectionReadsIn is every function in the store that runs a select and scans the rows
into a slice.

Scanning into a slice is the property that matters, not the shape of the query: a read
of one row cannot return more than one whatever is in the table. So a function counts
when the value handed to Model or Scan is a variable this function declared as a slice,
which is a question about the code rather than about a name.
*/
func collectionReadsIn(t *testing.T) []string {
	t.Helper()

	paths, err := filepath.Glob(filepath.Join("..", "..", "store", "*.go"))
	if err != nil {
		t.Fatalf("glob the store: %v", err)
	}

	var reads []string
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
			if readsACollection(fn) {
				reads = append(reads, fn.Name.Name)
			}
		}
	}
	slices.Sort(reads)
	return reads
}

// readsACollection reports whether this function selects rows into a slice of its own.
func readsACollection(fn *ast.FuncDecl) bool {
	slicesHere := sliceNamesIn(fn.Body)

	selects, intoASlice := false, false
	ast.Inspect(fn.Body, func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if !ok {
			return true
		}
		method, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		if method.Sel.Name == "NewSelect" {
			selects = true
		}
		if method.Sel.Name == "Model" || method.Sel.Name == "Scan" {
			for _, arg := range call.Args {
				if named, ok := addressOf(arg); ok && slicesHere[named] {
					intoASlice = true
				}
			}
		}
		return true
	})
	return selects && intoASlice
}

// sliceNamesIn is every local declared as a slice, however it was written.
func sliceNamesIn(body *ast.BlockStmt) map[string]bool {
	named := make(map[string]bool)
	ast.Inspect(body, func(node ast.Node) bool {
		switch declared := node.(type) {
		case *ast.ValueSpec:
			if _, isSlice := declared.Type.(*ast.ArrayType); isSlice {
				for _, name := range declared.Names {
					named[name.Name] = true
				}
			}
		case *ast.AssignStmt:
			for at, value := range declared.Rhs {
				if at >= len(declared.Lhs) || !makesASlice(value) {
					continue
				}
				if name, ok := declared.Lhs[at].(*ast.Ident); ok {
					named[name.Name] = true
				}
			}
		}
		return true
	})
	return named
}

// makesASlice reports whether an expression produces a slice: make([]T, …) or []T{…}.
func makesASlice(value ast.Expr) bool {
	switch made := value.(type) {
	case *ast.CompositeLit:
		_, isSlice := made.Type.(*ast.ArrayType)
		return isSlice
	case *ast.CallExpr:
		name, ok := made.Fun.(*ast.Ident)
		if !ok || name.Name != "make" || len(made.Args) == 0 {
			return false
		}
		_, isSlice := made.Args[0].(*ast.ArrayType)
		return isSlice
	default:
		return false
	}
}

// addressOf is the name behind &x, and whether the expression was that shape.
func addressOf(arg ast.Expr) (string, bool) {
	taken, ok := arg.(*ast.UnaryExpr)
	if !ok || taken.Op != token.AND {
		return "", false
	}
	name, ok := taken.X.(*ast.Ident)
	if !ok {
		return "", false
	}
	return name.Name, true
}
