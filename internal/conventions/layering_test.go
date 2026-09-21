package conventions

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

/*
The dependencies point inwards, which STANDARDS §1 says and nothing checked.

`internal/` knows nothing of `store` or `server`. `store` knows nothing of `server`.
Nothing imports `reference/`.

A layering rule erodes one import at a time, and each one is reasonable on its own: a
helper in `internal/` wants a type that happens to live in `store`, and the quickest way
to have it is to import it. Nothing fails. The next person finds the two packages can no
longer be moved or read apart, and by then undoing it is a refactor rather than a
decision.

`reference/` is the third for a different reason. It is somebody's local reading
material — gitignored, its own module, not in this build — so an import of it would fail
for every other person who checked the repository out, and would carry whatever is in
there into a public repository if the file were ever committed.

Test imports count. A package's tests are part of it, and a test reaching across a layer
is the same erosion with a shorter fuse.
*/
func TestDependenciesPointInwards(t *testing.T) {
	packages := packagesWithImports(t)
	if len(packages) < 10 {
		t.Fatalf("read %d packages, too few to be looking at the module", len(packages))
	}

	for _, one := range packages {
		for _, imported := range one.imports {
			if strings.Contains(imported, "/reference/") {
				t.Errorf("%s imports %s, which is not in this repository for anybody else",
					one.path, imported)
			}
			if under(one.path, "internal") && (under(imported, "store") || under(imported, "server")) {
				t.Errorf("%s imports %s — internal/ knows nothing of store or server", one.path, imported)
			}
			if under(one.path, "store") && under(imported, "server") {
				t.Errorf("%s imports %s — store knows nothing of server", one.path, imported)
			}
		}
	}

	// A replace pointing at it would put reference/ in the build without any import
	// naming it, which is the other way the same thing happens.
	gomod, err := os.ReadFile(filepath.Join("..", "..", "go.mod"))
	if err != nil {
		t.Fatalf("read go.mod: %v", err)
	}
	for _, line := range strings.Split(string(gomod), "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "replace") && strings.Contains(line, "reference") {
			t.Errorf("go.mod puts reference/ in the build: %s", strings.TrimSpace(line))
		}
	}
}

// under reports whether a package path is inside one of this module's top directories.
func under(path, top string) bool {
	return strings.HasPrefix(path, module+"/"+top+"/") || path == module+"/"+top
}

// oneImporter is a package and everything it reaches for, its tests included.
type oneImporter struct {
	path    string
	imports []string
}

// packagesWithImports asks the toolchain rather than reading the files, so that a build
// tag or a generated file cannot hide an import from this.
func packagesWithImports(t *testing.T) []oneImporter {
	t.Helper()

	const format = `{{.ImportPath}}|{{join .Imports ","}},{{join .TestImports ","}},{{join .XTestImports ","}}`
	listing := exec.CommandContext(t.Context(), "go", "list", "-f", format, "./...")
	listing.Dir = filepath.Join("..", "..")
	listed, err := listing.Output()
	if err != nil {
		t.Fatalf("go list: %v", err)
	}

	packages := make([]oneImporter, 0, 32)
	for _, line := range strings.Split(strings.TrimSpace(string(listed)), "\n") {
		path, rest, found := strings.Cut(line, "|")
		if !found {
			continue
		}
		imports := make([]string, 0, 8)
		for _, imported := range strings.Split(rest, ",") {
			if imported != "" {
				imports = append(imports, imported)
			}
		}
		packages = append(packages, oneImporter{path: path, imports: imports})
	}
	return packages
}
