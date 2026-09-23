package conventions

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"unicode/utf8"
)

// anIdentifier is the shape of the identifiers other systems hand out: a UUID.
var anIdentifier = regexp.MustCompile(`\b[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-` +
	`[0-9a-fA-F]{4}-[0-9a-fA-F]{12}\b`)

/*
excused are the tracked files allowed to carry one, with the reason.

Empty, and that is the point. An entry here is somebody deciding that a particular
identifier belongs in a public repository, which is a decision worth making once and
writing down rather than a thing that arrives in a paste.
*/
var excused = map[string]string{}

/*
Nothing tracked carries an identifier from somewhere else.

This repository is public, and the rule for it is that only what is genuinely for the
project and for everybody working on it gets pushed. Identifiers are how that rule is
broken without anybody meaning to: a project id, a workspace, a board, pasted into a
comment while working from it and then committed, useful to nobody reading the code and
pointing at an account that is not theirs.

`reference/` and the working notes are kept out by .gitignore and, for imports,
by TestDependenciesPointInwards. Neither sees prose. This does, and it reads what `git`
says is tracked rather than what is on disk, because what is tracked is what is public.

Written as a shape rather than as a list of the things not to say. A guard that spelled
them out would put them in the repository itself, which is the thing it is for.
*/
func TestNothingTrackedCarriesSomebodyElsesIdentifier(t *testing.T) {
	// safe.directory because this also runs inside a container, where the checkout
	// belongs to somebody other than the user running the tests and git refuses to look
	// at it. Nothing here writes, so there is nothing for that check to protect.
	listing := exec.CommandContext(t.Context(),
		"git", "-c", "safe.directory=*", "ls-files", "-z")
	listing.Dir = filepath.Join("..", "..")

	var refused bytes.Buffer
	listing.Stderr = &refused
	listed, err := listing.Output()
	if err != nil {
		// With what git said, because an exit status on its own names no cause.
		t.Fatalf("git ls-files: %v: %s", err, strings.TrimSpace(refused.String()))
	}

	tracked := strings.FieldsFunc(string(listed), func(r rune) bool { return r == 0 })
	if len(tracked) < 200 {
		t.Fatalf("git tracks %d files, too few to be reading the repository", len(tracked))
	}

	read := 0
	for _, name := range tracked {
		body, err := readIfText(filepath.Join("..", "..", name))
		if err != nil || body == "" {
			continue
		}
		read++
		if why, allowed := excused[name]; allowed {
			if len(why) < 20 {
				t.Errorf("%s is excused without a reason worth reading", name)
			}
			continue
		}
		for _, found := range anIdentifier.FindAllString(body, -1) {
			t.Errorf("%s carries %s, which belongs to something outside this project",
				name, found)
		}
	}
	if read < 150 {
		t.Fatalf("read %d tracked files, so most of the repository went unread", read)
	}

	for name := range excused {
		if !contains(tracked, name) {
			t.Errorf("excused names %s, which git does not track", name)
		}
	}
}

// contains reports whether a listing holds one name.
func contains(names []string, name string) bool {
	for _, one := range names {
		if one == name {
			return true
		}
	}
	return false
}

// readIfText reads a file and reports nothing for anything that is not text, so that an
// icon or a font is not searched for words.
func readIfText(path string) (string, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	if bytes.IndexByte(raw, 0) >= 0 || !utf8.Valid(raw) {
		return "", nil
	}
	return string(raw), nil
}
