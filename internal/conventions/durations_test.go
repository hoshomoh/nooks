package conventions

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// statedDuration is one lifetime constant and every place that says it out loud.
type statedDuration struct {
	// where the constant is declared, and the expression it is declared as.
	file string
	name string
	is   string

	// what the prose calls it, and the files that call it that.
	words  string
	stated []string
}

/*
A lifetime the prose states is still the lifetime the code uses.

Three constants decide how long a credential or an approval lasts, and none of them is
only in the code. They are written out in words in the published API reference, in
CONTEXT.md's glossary, and in a string a Member reads on the screen while deciding
whether to hurry. Changing `time.Hour` to six of them is one character and leaves nine
sentences lying, one of them to whoever is reading the API reference and one of them to
whoever is standing at the screen.

The other half of this test is the list of places itself. A reworded sentence that no
longer says the duration drops out of the check silently, so each file is required to
still contain the words — the list cannot quietly become shorter than the prose is.

Parsed rather than imported: `internal/` knows nothing of `store` or `server`, which
TestDependenciesPointInwards holds, and a test import would break it.
*/
func TestTheStatedLifetimesAreTheRealOnes(t *testing.T) {
	durations := []statedDuration{{
		file:   "../../store/request.go",
		name:   "ResetApprovalLifetime",
		is:     "time.Hour",
		words:  "an hour",
		stated: []string{"../../proto/nooks/api/v1/request_service.proto", "../../CONTEXT.md", "../../apps/web/src/i18n/locales/en.json"},
	}, {
		file:   "../../server/auth/cookie.go",
		name:   "AccessLifetime",
		is:     "time.Hour",
		words:  "an hour",
		stated: []string{"../../store/session.go", "../../server/router/api/v1/auth.go"},
	}, {
		file:   "../../server/auth/cookie.go",
		name:   "SessionLifetime",
		is:     "30 * 24 * time.Hour",
		words:  "a month",
		stated: []string{"../../store/session.go", "../../server/auth/cookie.go", "../../proto/nooks/api/v1/auth_service.proto", "../../server/router/api/v1/auth.go"},
	}}

	for _, one := range durations {
		t.Run(one.name, func(t *testing.T) {
			got, found := constExpression(t, one.file, one.name)
			if !found {
				t.Fatalf("%s is not declared in %s, so nothing here is checking it",
					one.name, filepath.Base(one.file))
			}
			if got != one.is {
				t.Errorf("%s is now %s, and %d places still say %q: %s",
					one.name, got, len(one.stated), one.words, strings.Join(one.stated, ", "))
			}

			for _, path := range one.stated {
				text, err := os.ReadFile(path)
				if err != nil {
					t.Fatalf("read %s: %v", path, err)
				}
				if !strings.Contains(string(text), one.words) {
					t.Errorf("%s no longer says %q, so it has stopped holding %s to anything",
						path, one.words, one.name)
				}
			}
		})
	}
}

// constExpression returns the expression a named constant is declared as, verbatim.
func constExpression(t *testing.T, path, name string) (string, bool) {
	t.Helper()

	source, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	fileSet := token.NewFileSet()
	file, err := parser.ParseFile(fileSet, path, source, parser.SkipObjectResolution)
	if err != nil {
		t.Fatalf("parse %s: %v", path, err)
	}

	for _, decl := range file.Decls {
		group, ok := decl.(*ast.GenDecl)
		if !ok || group.Tok != token.CONST {
			continue
		}
		for _, spec := range group.Specs {
			value, ok := spec.(*ast.ValueSpec)
			if !ok {
				continue
			}
			for at, ident := range value.Names {
				if ident.Name != name || at >= len(value.Values) {
					continue
				}
				from := fileSet.Position(value.Values[at].Pos()).Offset
				to := fileSet.Position(value.Values[at].End()).Offset
				return string(source[from:to]), true
			}
		}
	}
	return "", false
}
