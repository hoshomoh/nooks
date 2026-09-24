package conventions

import (
	"go/ast"
	"strings"
	"testing"
)

/*
excusedGrants are the places outside the auth package that build a Grant with something
in it, and why that is right.

A Grant says who is asking and what they may do. Built by hand it says whatever the
author put in it, and a Grant with only a Member in it means "a browser session for this
person", which is the widest thing there is: no token, so none of the narrowing a token
carries applies.
*/
var excusedGrants = map[string]string{
	"reachableListIDs": "asks what the Member can reach, not what the caller can, so a " +
		"token is scoped inside its owner's reach rather than its own; every TokenService " +
		"method is behind requireBrowser, so no token reaches this at all",
}

/*
A Grant is resolved, not written.

Three doors let a caller in and a fourth reads a backup, and all four ask the same
`resolver.Grant` what a request carries. That is the whole of the rule: the moment two of
them disagree, a token can do something through one door it cannot through another.

The way that rule breaks quietly is not a new door. It is a method somewhere deciding it
needs a Grant, writing one, and acting on it: the fields are exported, so
`auth.Grant{Member: somebody}` compiles and says browser session with full reach. Every
ability check downstream reads it and agrees.

An empty one is left alone. `auth.Grant{}` beside an error is a caller saying it has
nothing to give back, and the caller checks the error.
*/
func TestAGrantIsResolvedRatherThanWritten(t *testing.T) {
	files, _ := goFilesIn(t)
	if len(files) < 50 {
		t.Fatalf("read %d files, so this is looking at less of the tree than it did", len(files))
	}

	seen := map[string]bool{}
	for path, file := range files {
		if strings.Contains(filepathSlash(path), "/server/auth/") {
			continue
		}
		for _, one := range file.Decls {
			fn, ok := one.(*ast.FuncDecl)
			if !ok || fn.Body == nil {
				continue
			}
			if !buildsAGrant(fn.Body) {
				continue
			}
			seen[fn.Name.Name] = true
			why, allowed := excusedGrants[fn.Name.Name]
			if !allowed {
				t.Errorf("%s writes a Grant rather than resolving one, which says "+
					"whatever it was given; see %s", fn.Name.Name, path)
				continue
			}
			if len(why) < 40 {
				t.Errorf("excusedGrants[%s] is too short to be a reason somebody thought about",
					fn.Name.Name)
			}
		}
	}

	for name := range excusedGrants {
		if !seen[name] {
			t.Errorf("excusedGrants names %s, which no longer writes one", name)
		}
	}
}

// buildsAGrant reports whether a body writes a Grant with anything in it. An empty one
// is a zero value handed back beside an error.
func buildsAGrant(body *ast.BlockStmt) bool {
	built := false
	ast.Inspect(body, func(node ast.Node) bool {
		lit, ok := node.(*ast.CompositeLit)
		if !ok || len(lit.Elts) == 0 {
			return true
		}
		named, ok := lit.Type.(*ast.SelectorExpr)
		if !ok || named.Sel.Name != "Grant" {
			return true
		}
		if pkg, ok := named.X.(*ast.Ident); ok && pkg.Name == "auth" {
			built = true
		}
		return true
	})
	return built
}

// filepathSlash reads a path the same way on any machine.
func filepathSlash(path string) string { return strings.ReplaceAll(path, "\\", "/") }
