package conventions

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"strings"
	"testing"
)

// theBuilder is the one function allowed to make a Connect request out of a REST call.
const theBuilder = "requestFrom"

/*
Every REST adapter hands the service what the call arrived with.

The services are written for Connect and read headers off the request, so an adapter that
builds an empty one leaves its service seeing nothing. That is not a guess: the adapter
for refreshing a session did exactly this, and the documented POST to refresh answered
"not signed in" to a caller holding a perfectly good cookie. Every adapter goes through
one builder now.

Nothing held that. `check-gateway.sh` asks whether an RPC has an adapter at all, which is
a different question: an adapter written the wrong way has one. A new adapter reaching
for `connect.NewRequest` compiles, passes that check, and drops the headers for its own
RPC and no other, which is the kind of fault that is found by somebody's session not
working rather than by a test.
*/
func TestEveryRestAdapterCarriesWhatArrived(t *testing.T) {
	paths, err := filepath.Glob(filepath.Join("..", "..", "server", "router", "gateway", "*.go"))
	if err != nil {
		t.Fatalf("glob the gateway: %v", err)
	}

	built := 0
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
			made, uses := requestsMadeIn(fn.Body)
			built += uses
			if fn.Name.Name == theBuilder {
				// The builder is where a request is allowed to be made from nothing.
				continue
			}
			for range made {
				t.Errorf("%s builds a Connect request itself, so the headers the REST "+
					"call arrived with do not reach its service; use %s",
					fn.Name.Name, theBuilder)
			}
		}
	}

	// A rename of the builder would leave every adapter looking clean while carrying
	// nothing, so the count of adapters going through it is checked as well.
	if built < 40 {
		t.Errorf("%d adapters go through %s, and there are more RPCs than that, so "+
			"either it has been renamed or they have stopped using it", built, theBuilder)
	}
}

// requestsMadeIn counts requests built from nothing, and requests built by the builder.
func requestsMadeIn(body *ast.BlockStmt) (made, uses int) {
	ast.Inspect(body, func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if !ok {
			return true
		}
		switch fun := call.Fun.(type) {
		case *ast.SelectorExpr:
			pkg, ok := fun.X.(*ast.Ident)
			if ok && pkg.Name == "connect" && fun.Sel.Name == "NewRequest" {
				made++
			}
		case *ast.Ident:
			if fun.Name == theBuilder {
				uses++
			}
		case *ast.IndexExpr:
			// requestFrom[T](...), which is how a generic is written at the call site.
			if name, ok := fun.X.(*ast.Ident); ok && name.Name == theBuilder {
				uses++
			}
		}
		return true
	})
	return made, uses
}
