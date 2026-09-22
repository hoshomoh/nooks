package conventions

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"strings"
	"testing"
)

/*
A superseded doc comment is deleted rather than left above the one that replaced it.

Go attaches only the last comment group above a declaration, so a one-line comment left
above a rewritten block keeps compiling, keeps reading as though it belonged to
something, and is rendered nowhere. TestADocCommentNamesWhatItDocuments cannot see it:
that asks what a comment opens with, and this kind opens with the name of something no
longer declared anywhere.

All three that existed said something the block beneath them had stopped being true of.
One described Permission, a type that had become TokenAbilities, and called it "two
levels, not a matrix" directly above three separate answers. One said an Item's row
shows the Note's first line and a count of the rest, directly above the paragraph
explaining that the guessing had been taken out. The third was a plainer duplicate of
the line below it.

What is caught is narrow on purpose: a line comment group with a block comment group
opening on the very next line. Nothing legitimate is written that way, because the
second group is the declaration's own comment and the first therefore belongs to
nothing. Test files are read too — that is where one of the three was.
*/
func TestASupersededDocCommentIsGone(t *testing.T) {
	read := 0
	for _, root := range theTree {
		err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
			if err != nil || entry.IsDir() || !strings.HasSuffix(path, ".go") {
				return err
			}
			read++

			fileSet := token.NewFileSet()
			file, err := parser.ParseFile(fileSet, path, nil, parser.ParseComments)
			if err != nil {
				t.Errorf("parse %s: %v", path, err)
				return nil
			}
			for _, one := range strandedIn(file, fileSet) {
				t.Errorf("%s:%d %q sits directly above the block that replaced it",
					path, one.line, one.opens)
			}
			return nil
		})
		if err != nil {
			t.Fatalf("walk %s: %v", root, err)
		}
	}
	if read < 100 {
		t.Fatalf("read %d files, so this is looking at less of the tree than it did", read)
	}
}

// stranded is one orphaned comment: where it is and what it opens with.
type stranded struct {
	line  int
	opens string
}

// strandedIn finds every line comment a block comment opens directly under.
//
// One group, not two: the parser puts comments with no blank line between them into the
// same group whatever their style, so this is a `//` comment followed inside one group
// by a block comment rather than two groups in a row.
func strandedIn(file *ast.File, fileSet *token.FileSet) []stranded {
	var found []stranded
	for _, group := range file.Comments {
		for at := 0; at+1 < len(group.List); at++ {
			this, next := group.List[at], group.List[at+1]
			if !strings.HasPrefix(this.Text, "//") || !strings.HasPrefix(next.Text, "/*") {
				continue
			}
			// Everything above the block comment is the orphan, so report where the
			// run of line comments started rather than its last line.
			opensAt := at
			for opensAt > 0 && strings.HasPrefix(group.List[opensAt-1].Text, "//") {
				opensAt--
			}
			found = append(found, stranded{
				line:  fileSet.Position(group.List[opensAt].Pos()).Line,
				opens: firstWordOf(strings.TrimPrefix(group.List[opensAt].Text, "//")),
			})
		}
	}
	return found
}
