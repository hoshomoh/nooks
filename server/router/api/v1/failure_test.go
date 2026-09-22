package v1

import (
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"connectrpc.com/connect"
)

/*
Every failure in this package says which kind it is.

`internalError` reads the kind off the description it is given, because passing one at
each of the call sites would be that many chances to pass the wrong one. The cost of
reading it off is that a description opening with a verb nothing has classified falls
through to the safe default silently, and nobody finds out.

So the descriptions are parsed out of the source and each one's verb is required to be
known. A new one is then a decision somebody makes here rather than a default they fall
into somewhere else.
*/
func TestEveryFailureSaysWhichKindItIs(t *testing.T) {
	// Verbs that mean nothing was written. Everything else is treated as a write, which
	// is the safe way round.
	known := map[string]bool{
		"read": true, "count": true, "list": true, "search": true,
		"create": true, "make": true, "save": true, "set": true, "update": true,
		"delete": true, "remove": true, "share": true, "hash": true, "verify": true,
		"claim": true, "resolve": true, "spend": true, "end": true, "give": true,
		"revoke": true, "index": true, "announce": true, "tell": true, "add": true,
		"open": true, "close": true, "write": true, "start": true, "decide": true,
		"record": true, "mark": true, "copy": true, "tick": true, "move": true,
		"rename": true, "archive": true, "pin": true,
	}

	descriptions := failureDescriptions(t)
	if len(descriptions) < 50 {
		t.Fatalf("found %d failure descriptions, too few to be reading the package",
			len(descriptions))
	}

	for _, what := range descriptions {
		verb, _, _ := strings.Cut(what, " ")
		if !known[verb] {
			t.Errorf("%q opens with %q, which nothing has classified: add it above and "+
				"say whether it means the Member's change happened", what, verb)
		}
	}
}

// failureDescriptions is the first argument of every internalError call in the package.
func failureDescriptions(t *testing.T) []string {
	t.Helper()

	var found []string
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("read the package: %v", err)
	}

	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		file, err := parser.ParseFile(token.NewFileSet(), filepath.Join(".", entry.Name()), nil,
			parser.SkipObjectResolution)
		if err != nil {
			t.Fatalf("parse %s: %v", entry.Name(), err)
		}

		ast.Inspect(file, func(node ast.Node) bool {
			call, ok := node.(*ast.CallExpr)
			if !ok || len(call.Args) == 0 {
				return true
			}
			name, ok := call.Fun.(*ast.Ident)
			if !ok || name.Name != "internalError" {
				return true
			}
			text, ok := call.Args[0].(*ast.BasicLit)
			if !ok || text.Kind != token.STRING {
				return true
			}
			found = append(found, strings.Trim(text.Value, `"`))
			return true
		})
	}
	return found
}

/*
The cause of a failure stays in the log.

It used to travel. `internalError` wrapped the underlying error with %w and Connect sends
the message, so a Member — or a stranger asking for a password reset, which is reachable
without signing in — could be handed a table name, the path to the database file, or
`dial tcp 10.0.0.5:5432: connection refused`.

What is asserted is the absence of the cause rather than the presence of the right
words, because the mistake this catches is somebody putting the cause back for the
reason it was there in the first place: it is genuinely useful, to the wrong audience.
*/
func TestAFailureCarriesNoCause(t *testing.T) {
	cause := errors.New("dial tcp 10.0.0.5:5432: connection refused, /var/lib/nooks/nooks.db")

	failure := internalError("read the member", cause)

	said := failure.Error()
	for _, secret := range []string{"10.0.0.5", "/var/lib/nooks", "connection refused"} {
		if strings.Contains(said, secret) {
			t.Errorf("a caller is told %q, which carries %q", said, secret)
		}
	}

	connectErr := new(connect.Error)
	if !errors.As(failure, &connectErr) {
		t.Fatalf("internalError did not answer a connect error: %v", failure)
	}
	if got := connectErr.Meta().Get(errorKindHeader); got != kindLoadFailed {
		t.Errorf("kind = %q, want %q: reading the member changes nothing", got, kindLoadFailed)
	}
	if connectErr.Meta().Get(errorRefHeader) == "" {
		t.Error("no reference travelled, so nobody can find the log line")
	}
}
