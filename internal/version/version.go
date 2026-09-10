// Package version reports which build of Nook is running.
package version

// Version is the semantic version of this build. Release builds override it via
// -ldflags; a build from source reports "dev".
var Version = "dev"

// Commit is the git revision this build came from, when the build sets it.
var Commit = ""

// String renders the version for the CLI and the About screen, e.g. "1.0.2 (a1b2c3d)".
func String() string {
	if Commit == "" {
		return Version
	}
	return Version + " (" + Commit + ")"
}
