package v1

import (
	"database/sql"
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"connectrpc.com/connect"

	apiv1 "github.com/hoshomoh/nooks/proto/gen/nooks/api/v1"
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
		"rename": true, "archive": true, "pin": true, "restore": true, "purge": true,
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

/*
Every field a Member writes into is bounded, and the bound is enforced.

Nothing bounded any of them. A request is capped at four mebibytes, so a List could be
named with four mebibytes of text and every screen that draws that name would try to.

Each case sends one character past the limit through the path that field actually
arrives on, because a constant nothing reads is a limit in name only. The paths differ
enough that this cannot be a loop over a table of field names: an Item's label arrives
on CreateItem, an instance name on CompleteSetup, a message on RequestJoin.
*/
func TestEveryWrittenFieldIsBounded(t *testing.T) {
	over := func(limit int) string { return strings.Repeat("a", limit+1) }

	t.Run("item label, quantity and note", func(t *testing.T) {
		f := newListFixture(t)
		uid := f.createList(t, f.anna, "Groceries")

		for _, one := range []struct {
			what string
			req  *apiv1.CreateItemRequest
		}{
			{"label", &apiv1.CreateItemRequest{ListUid: uid, Label: over(LimitItemLabel)}},
			{"quantity", &apiv1.CreateItemRequest{
				ListUid: uid, Label: "Bread", Quantity: over(LimitItemQuantity),
			}},
		} {
			t.Run(one.what, func(t *testing.T) {
				_, err := f.svc.CreateItem(f.as(t, f.anna), connect.NewRequest(one.req))
				if got := connect.CodeOf(err); got != connect.CodeInvalidArgument {
					t.Errorf("code = %v, want invalid_argument", got)
				}
			})
		}

		t.Run("note", func(t *testing.T) {
			item, err := f.svc.CreateItem(f.as(t, f.anna), connect.NewRequest(
				&apiv1.CreateItemRequest{ListUid: uid, Label: "Bread"},
			))
			if err != nil {
				t.Fatalf("CreateItem: %v", err)
			}
			note := over(LimitItemNote)
			_, err = f.svc.UpdateItem(f.as(t, f.anna), connect.NewRequest(
				&apiv1.UpdateItemRequest{ItemUid: item.Msg.GetItem().GetUid(), Note: &note},
			))
			if got := connect.CodeOf(err); got != connect.CodeInvalidArgument {
				t.Errorf("code = %v, want invalid_argument", got)
			}
		})
	})

	t.Run("list name", func(t *testing.T) {
		f := newListFixture(t)
		_, err := f.svc.CreateList(f.as(t, f.anna), connect.NewRequest(
			&apiv1.CreateListRequest{Name: over(LimitListName)},
		))
		if got := connect.CodeOf(err); got != connect.CodeInvalidArgument {
			t.Errorf("code = %v, want invalid_argument", got)
		}
	})

	t.Run("join request name, email and message", func(t *testing.T) {
		for _, one := range []struct {
			what string
			req  *apiv1.RequestJoinRequest
		}{
			{"name", &apiv1.RequestJoinRequest{
				Name: over(LimitMemberName), Email: "til@example.com",
			}},
			{"email", &apiv1.RequestJoinRequest{
				Name: "Til", Email: over(LimitMemberEmail),
			}},
			{"message", &apiv1.RequestJoinRequest{
				Name: "Til", Email: "til@example.com", Message: over(LimitJoinMessage),
			}},
		} {
			t.Run(one.what, func(t *testing.T) {
				svc, _ := newAuthService(t)
				completeSetup(t, svc)

				_, err := svc.RequestJoin(t.Context(), connect.NewRequest(one.req))
				if got := connect.CodeOf(err); got != connect.CodeInvalidArgument {
					t.Errorf("code = %v, want invalid_argument", got)
				}
			})
		}
	})

	t.Run("instance name at first run", func(t *testing.T) {
		svc, _ := newAuthService(t)
		_, err := svc.CompleteSetup(t.Context(), connect.NewRequest(&apiv1.CompleteSetupRequest{
			Name: "Anna", Email: "anna@brunnen.lan", Password: goodPassword,
			InstanceName: over(LimitInstanceName),
		}))
		if got := connect.CodeOf(err); got != connect.CodeInvalidArgument {
			t.Errorf("code = %v, want invalid_argument", got)
		}
	})
}

/*
A lock held for a moment is not the Instance being broken.

CodeInternal means do not retry, and a caller that believes it will drop the write
rather than try again a second later. So the one failure worth retrying has to arrive
saying so, or every non-browser caller does the wrong thing with it.

Driven with a real SQLITE_BUSY rather than a stand-in. `*sqlite.Error` keeps its code
unexported and only the driver builds one, and a test that fabricates the error it is
checking for proves the fabrication.
*/
func TestAContendedWriteIsToldToTryAgain(t *testing.T) {
	failure := internalError("update item", heldLock(t))

	var connectErr *connect.Error
	if !errors.As(failure, &connectErr) {
		t.Fatalf("internalError returned %T, want a connect error", failure)
	}
	if connectErr.Code() != connect.CodeUnavailable {
		t.Errorf("a contended write came back as %v, want %v",
			connectErr.Code(), connect.CodeUnavailable)
	}
	if !strings.Contains(connectErr.Message(), "try again") {
		t.Errorf("a contended write says %q, which does not tell the caller to try again",
			connectErr.Message())
	}
	// The app reads the kind, not the code, and it should read what it read before.
	if got := connectErr.Meta().Get(errorKindHeader); got != kindSaveFailed {
		t.Errorf("a contended write is kind %q, want %q", got, kindSaveFailed)
	}

	// The same funnel still calls everything else broken.
	ordinary := internalError("update item", errors.New("disk on fire"))
	if !errors.As(ordinary, &connectErr) {
		t.Fatalf("internalError returned %T, want a connect error", ordinary)
	}
	if connectErr.Code() != connect.CodeInternal {
		t.Errorf("an ordinary failure came back as %v, want %v",
			connectErr.Code(), connect.CodeInternal)
	}
}

// heldLock is a real SQLITE_BUSY, made by writing to a database another connection is
// already writing to with nothing willing to wait.
func heldLock(t *testing.T) error {
	t.Helper()

	path := filepath.Join(t.TempDir(), "contended.db")
	dsn := "file:" + path + "?_pragma=busy_timeout(0)"

	held, err := sql.Open("sqlite", dsn)
	if err != nil {
		t.Fatalf("open the holder: %v", err)
	}
	t.Cleanup(func() { _ = held.Close() })

	tx, err := held.BeginTx(t.Context(), nil)
	if err != nil {
		t.Fatalf("begin: %v", err)
	}
	t.Cleanup(func() { _ = tx.Rollback() })
	if _, err := tx.ExecContext(t.Context(), "CREATE TABLE probe_held (x)"); err != nil {
		t.Fatalf("take the write lock: %v", err)
	}

	other, err := sql.Open("sqlite", dsn)
	if err != nil {
		t.Fatalf("open the second connection: %v", err)
	}
	t.Cleanup(func() { _ = other.Close() })

	_, err = other.ExecContext(t.Context(), "CREATE TABLE probe_waiting (x)")
	if err == nil {
		t.Fatal("the second write succeeded while the lock was held")
	}
	return err
}
