package conventions

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strings"
	"testing"
)

// where the flags are declared, and the two tables that write them out for a reader.
const flagSource = "../../internal/profile/profile.go"

var flagTables = []string{
	"../../README.md",
	"../../apps/website/content/docs/configure.mdx",
}

/*
Every flag the binary takes is in both tables that claim to list them.

There are two, and neither can import anything: the README's "Options" and the
documentation site's "configure" page. A flag added to `Parse` and to neither is a flag
only somebody reading Go finds.

It had already happened. `--secure-cookies` and `--log-level` reached the site's table
and not the README's, which had gone two rows stale. The first of those is the one a
self-hoster behind TLS needs: without it the session cookie is not marked Secure, and
the README is where somebody putting this on a server starts.

The flag names are parsed from the calls that register them rather than listed here. A
list here would be a third copy, and the thing being tested is that copies drift.
*/
func TestEveryFlagIsWrittenDown(t *testing.T) {
	flags := flagsRegisteredIn(t, flagSource)
	if len(flags) < 5 {
		t.Fatalf("found %d flags in %s, too few to be reading the ones it registers",
			len(flags), flagSource)
	}

	for _, table := range flagTables {
		text, err := os.ReadFile(table)
		if err != nil {
			t.Fatalf("read %s: %v", table, err)
		}
		for _, flag := range flags {
			// As a reader types it, in the backticks both tables use.
			if !strings.Contains(string(text), fmt.Sprintf("`--%s`", flag)) {
				t.Errorf("%s does not list --%s, which the binary takes", table, flag)
			}
		}
	}
}

// flagsRegisteredIn is every flag name passed to a flag-set registration.
func flagsRegisteredIn(t *testing.T, path string) []string {
	t.Helper()

	fileSet := token.NewFileSet()
	file, err := parser.ParseFile(fileSet, path, nil, parser.SkipObjectResolution)
	if err != nil {
		t.Fatalf("parse %s: %v", path, err)
	}

	var flags []string
	ast.Inspect(file, func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if !ok || len(call.Args) == 0 {
			return true
		}
		method, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		switch method.Sel.Name {
		case "String", "Bool", "Int", "Duration":
		default:
			return true
		}
		name, ok := call.Args[0].(*ast.BasicLit)
		if !ok || name.Kind != token.STRING {
			return true
		}
		flags = append(flags, strings.Trim(name.Value, `"`))
		return true
	})
	return flags
}
