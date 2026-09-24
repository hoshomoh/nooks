package conventions

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"strings"
	"testing"
	"unicode"
)

/*
Nothing unexported is left behind where nobody calls it.

Go compiles an unexported function nobody calls without a word, and `go vet` says
nothing, so code outlives whatever replaced it. Three were found at once by measuring
coverage, which is a roundabout way to learn that something is never run: `nextPosition`,
superseded when an Item's place moved inside the insert and left in place by the change
that superseded it, and `addToGroup` with `removeFromGroup`, a second way to change a
Group that contradicts the comment on the one way there is.

Dead code is not only clutter. Those two were a working pair somebody could reach for,
and they skip everything `ReplaceGroupMembers` does around the write.

staticcheck's U1000 is the tool for this and does not run here: the version that builds
against this toolchain cannot read Go 1.27's export data, so it fails before it starts.
This is smaller and enough. A use from a test counts, which is lenient on purpose: a
helper a test drives is a decision somebody made, and the thing worth failing on is code
nothing mentions at all.
*/
func TestNothingUnexportedIsLeftBehind(t *testing.T) {
	looked := 0
	for _, root := range theTree {
		err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
			if err != nil || !entry.IsDir() {
				return err
			}
			looked += unreferencedIn(t, path)
			return nil
		})
		if err != nil {
			t.Fatalf("walk %s: %v", root, err)
		}
	}
	if looked < 60 {
		t.Fatalf("read %d unexported functions, which is less of the tree than it holds", looked)
	}
}

// unreferencedIn reports every unexported function in one directory that nothing names,
// and returns how many it considered.
func unreferencedIn(t *testing.T, dir string) int {
	t.Helper()

	fset := token.NewFileSet()
	packages, err := parser.ParseDir(fset, dir, nil, 0)
	if err != nil {
		// A directory with no Go in it is not a fault.
		return 0
	}

	considered := 0
	for _, pkg := range packages {
		declared := map[string]token.Pos{}
		named := map[string]int{}

		for path, file := range pkg.Files {
			if !strings.HasSuffix(path, "_test.go") {
				for _, one := range file.Decls {
					fn, ok := one.(*ast.FuncDecl)
					if !ok || !isPrivate(fn.Name.Name) {
						continue
					}
					declared[fn.Name.Name] = fn.Pos()
				}
			}
			// Names are counted from the tests as well: a helper a test drives is not
			// something nobody mentions.
			ast.Inspect(file, func(node ast.Node) bool {
				if ident, ok := node.(*ast.Ident); ok {
					named[ident.Name]++
				}
				return true
			})
		}

		considered += len(declared)
		for name, at := range declared {
			// One mention is the declaration naming itself.
			if named[name] <= 1 {
				t.Errorf("%s: %s is declared and nothing calls it", fset.Position(at), name)
			}
		}
	}
	return considered
}

// isPrivate reports whether a name is one this package keeps to itself, leaving out the
// two the toolchain calls rather than any code.
func isPrivate(name string) bool {
	if name == "" || name == "init" || name == "main" {
		return false
	}
	return unicode.IsLower(rune(name[0]))
}
