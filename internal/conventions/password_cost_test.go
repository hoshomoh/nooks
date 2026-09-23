package conventions

import (
	"go/ast"
	"strings"
	"testing"
)

/*
Only a test may make hashing cheap.

`password.Cost` is a variable for one reason: bcrypt is slow on purpose, and the suite
hashes often enough that the real cost took two packages past the ten minutes `go test`
allows, under the race detector CI runs and this machine cannot. A commit touching no Go
at all failed on it.

A knob that exists for the suite is a knob production can reach, and lowered in a running
Instance this one is a real weakness: it is the difference between a stolen database
being a slow problem and an immediate one. So the knob is guarded rather than trusted.

Written against the parse rather than a search for the text, because `Cost =` appears in
the prose above the variable itself, and a grep that reads comments is a grep that
reports the thing it is checking.
*/
func TestNothingButATestLowersTheCostOfAPassword(t *testing.T) {
	files, tests := goFilesIn(t)
	if len(files) < 50 {
		t.Fatalf("read %d files, so this is looking at less of the tree than it did", len(files))
	}

	for path, file := range files {
		for _, assigned := range costAssignmentsIn(file) {
			t.Errorf("%s assigns %s outside a test, which weakens every password on the "+
				"Instance", path, assigned)
		}
	}

	// A guard whose subject has been renamed away passes for the wrong reason, so the
	// tests that do lower it have to still be doing so.
	lowering := 0
	for path, file := range tests {
		if !strings.HasSuffix(path, "main_test.go") {
			continue
		}
		lowering += len(costAssignmentsIn(file))
	}
	if lowering < 6 {
		t.Errorf("%d test packages lower the cost, and six did, so either it is spelt "+
			"differently now or one has gone back to paying for bcrypt", lowering)
	}
}

// costAssignmentsIn is every assignment to the password cost in one file, named as
// written. The declaration of the variable itself is not an assignment to it.
func costAssignmentsIn(file *ast.File) []string {
	var assigned []string
	ast.Inspect(file, func(node ast.Node) bool {
		statement, ok := node.(*ast.AssignStmt)
		if !ok {
			return true
		}
		for _, target := range statement.Lhs {
			if name, ok := namesTheCost(target); ok {
				assigned = append(assigned, name)
			}
		}
		return true
	})
	return assigned
}

// namesTheCost reports whether an expression is the password cost, written either as
// `password.Cost` from outside that package or as `Cost` from within it.
func namesTheCost(target ast.Expr) (string, bool) {
	switch named := target.(type) {
	case *ast.SelectorExpr:
		pkg, ok := named.X.(*ast.Ident)
		if ok && pkg.Name == "password" && named.Sel.Name == "Cost" {
			return "password.Cost", true
		}
	case *ast.Ident:
		if named.Name == "Cost" {
			return "Cost", true
		}
	}
	return "", false
}
