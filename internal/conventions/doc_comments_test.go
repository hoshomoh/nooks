package conventions

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// theTree is the Go in this repository, from this package's own directory.
var theTree = []string{"../../cmd", "../../internal", "../../server", "../../store"}

/*
A doc comment says the name of the thing it documents.

Go's convention, and here it is load-bearing rather than tidy: inserting a function
directly beneath an existing doc comment strands that comment on the new function and
leaves the old one with none. Nothing complains. It compiles, the tests pass, and the
only place it shows is the rendered documentation, where one function is described as
another.

It had happened four times before this test existed — `DeleteList` documenting
`SetListArchived`, `ActivityFor` documenting `DeleteUnreadableActivity`, `ListByUID`
documenting `ListByID`, `ListsForMember` documenting `CanReachList` — every one of them
where a function had been added under a comment that already belonged to something else.

It is not only functions. The same insertion under a comment strands it on whatever
came next, and `ListLists` spent the paging rework documenting a const block — with its
old text, still claiming it returned every List.

A declaration with no doc comment is not the rule: the gateway adapters are fifty-three
lines of the same thing and the package comment says so once, which is better than
saying it fifty-three times. Nor is a comment that opens with an ordinary word. What is
caught is a comment opening with the name of something else that is declared somewhere
in this tree, which is what a stranded one always does.
*/
func TestADocCommentNamesWhatItDocuments(t *testing.T) {
	files := goFilesIn(t)
	if len(files) < 50 {
		t.Fatalf("read %d files, so this is looking at less of the tree than it did", len(files))
	}

	// Every name this repository declares, so "opens with somebody else's name" is a
	// question that can be asked at all.
	declared := make(map[string]bool)
	for _, file := range files {
		for _, one := range file.Decls {
			for _, name := range namesOf(one) {
				declared[name] = true
			}
		}
	}
	if len(declared) < 200 {
		t.Fatalf("found %d declared names, too few to recognise a stranded comment", len(declared))
	}

	for path, file := range files {
		for _, one := range file.Decls {
			doc := docOf(one)
			if doc == nil {
				continue
			}
			opens := firstWordOf(doc.Text())
			if opens == "" || !declared[opens] {
				continue
			}
			names := namesOf(one)
			if slices.Contains(names, opens) {
				continue
			}
			t.Errorf("%s: the comment on %s opens with %q — a comment stranded by an "+
				"insertion leaves whatever it belonged to with none",
				path, strings.Join(names, ", "), opens)
		}
	}
}

// goFilesIn parses the tree once, by path.
func goFilesIn(t *testing.T) map[string]*ast.File {
	t.Helper()

	files := make(map[string]*ast.File)
	for _, root := range theTree {
		err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
			if err != nil || entry.IsDir() || !strings.HasSuffix(path, ".go") {
				return err
			}
			if strings.HasSuffix(path, "_test.go") {
				return nil
			}
			parsed, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.ParseComments)
			if err != nil {
				t.Errorf("parse %s: %v", path, err)
				return nil
			}
			files[path] = parsed
			return nil
		})
		if err != nil {
			t.Fatalf("walk %s: %v", root, err)
		}
	}
	return files
}

// docOf is the comment above a declaration, whichever kind it is.
func docOf(one ast.Decl) *ast.CommentGroup {
	switch declared := one.(type) {
	case *ast.FuncDecl:
		return declared.Doc
	case *ast.GenDecl:
		return declared.Doc
	default:
		return nil
	}
}

// namesOf is everything a declaration introduces: one for a func, and for a grouped
// const or var every name in the group, since the comment above one covers all of them.
func namesOf(one ast.Decl) []string {
	switch declared := one.(type) {
	case *ast.FuncDecl:
		return []string{declared.Name.Name}
	case *ast.GenDecl:
		names := make([]string, 0, len(declared.Specs))
		for _, spec := range declared.Specs {
			switch held := spec.(type) {
			case *ast.TypeSpec:
				names = append(names, held.Name.Name)
			case *ast.ValueSpec:
				for _, name := range held.Names {
					names = append(names, name.Name)
				}
			}
		}
		return names
	default:
		return nil
	}
}

// firstWordOf is what a doc comment opens with, without the markup a name may wear.
func firstWordOf(doc string) string {
	fields := strings.Fields(doc)
	if len(fields) == 0 {
		return ""
	}
	return strings.Trim(fields[0], "`*_")
}
