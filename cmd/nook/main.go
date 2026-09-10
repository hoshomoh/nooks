// Command nook runs a Nook instance: one binary serving both the API and the app.
package main

import (
	"fmt"
	"os"

	"github.com/hoshomoh/nooks/internal/version"
)

func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "nook:", err)
		os.Exit(1)
	}
}

// run holds the real entry point so that it can be tested: it takes its arguments
// and its output rather than reaching for the process globals.
func run(args []string, stdout *os.File) error {
	if len(args) > 0 && (args[0] == "--version" || args[0] == "version") {
		fmt.Fprintln(stdout, version.String())
		return nil
	}
	// The server arrives in M1; until then the binary only reports what it is.
	fmt.Fprintln(stdout, "nook "+version.String())
	return nil
}
