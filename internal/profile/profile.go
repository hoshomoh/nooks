// Package profile holds the configuration one Nooks instance runs with.
//
// Parsing is a pure function of its arguments and an environment lookup, so the
// whole surface can be tested without touching the process or the filesystem.
package profile

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"path/filepath"
	"strings"
)

// Driver names a supported database backend.
type Driver string

const (
	// DriverSQLite keeps the Instance in one file the owner can copy. The default.
	DriverSQLite Driver = "sqlite"
	// DriverPostgres keeps the Instance in an existing Postgres server.
	DriverPostgres Driver = "postgres"
)

// Mode is the environment an Instance runs in.
type Mode string

const (
	// ModeProd serves the embedded app and is the default.
	ModeProd Mode = "prod"
	// ModeDev expects the Vite dev server to serve the app separately.
	ModeDev Mode = "dev"
)

// Config is everything an Instance needs to start.
type Config struct {
	// Addr is the host:port to listen on, e.g. ":8081".
	Addr string
	// Data is the directory holding the SQLite file and any other instance data.
	Data string
	// Driver selects the database backend.
	Driver Driver
	// DSN is the Postgres connection string. Unused, and must be empty, for SQLite.
	DSN string
	// Mode is prod or dev.
	Mode Mode
}

// ErrHelp reports that the caller asked for usage rather than a running server.
var ErrHelp = flag.ErrHelp

// DefaultAddr is the port Nooks listens on when nothing says otherwise.
const DefaultAddr = ":8081"

// Parse builds a Config from command-line arguments and an environment lookup.
//
// Flags win over environment variables, which win over defaults. Every variable is
// prefixed NOOKS_, e.g. NOOKS_ADDR. Usage is written to out.
func Parse(args []string, env func(string) string, out io.Writer) (Config, error) {
	if env == nil {
		return Config{}, errors.New("profile: env lookup is required")
	}

	set := flag.NewFlagSet("nooks", flag.ContinueOnError)
	set.SetOutput(out)

	addr := set.String("addr", envOr(env, "NOOKS_ADDR", DefaultAddr), "host:port to listen on")
	data := set.String("data", envOr(env, "NOOKS_DATA", "./data"), "directory for instance data")
	driver := set.String("driver", envOr(env, "NOOKS_DRIVER", string(DriverSQLite)), "database driver: sqlite or postgres")
	dsn := set.String("dsn", envOr(env, "NOOKS_DSN", ""), "postgres connection string")
	mode := set.String("mode", envOr(env, "NOOKS_MODE", string(ModeProd)), "prod or dev")

	if err := set.Parse(args); err != nil {
		return Config{}, err
	}

	cfg := Config{
		Addr:   strings.TrimSpace(*addr),
		Data:   filepath.Clean(strings.TrimSpace(*data)),
		Driver: Driver(strings.TrimSpace(*driver)),
		DSN:    strings.TrimSpace(*dsn),
		Mode:   Mode(strings.TrimSpace(*mode)),
	}
	if err := cfg.validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

// validate rejects a Config that cannot produce a working Instance. It fails on the
// first problem so that the operator fixes one thing at a time.
func (c Config) validate() error {
	if c.Addr == "" {
		return errors.New("addr is required")
	}
	switch c.Driver {
	case DriverSQLite:
		if c.DSN != "" {
			return errors.New("dsn is for postgres; sqlite uses data instead")
		}
		if c.Data == "" || c.Data == "." {
			return errors.New("data directory is required for sqlite")
		}
	case DriverPostgres:
		if c.DSN == "" {
			return errors.New("dsn is required for postgres")
		}
	default:
		return fmt.Errorf("unknown driver %q: want sqlite or postgres", c.Driver)
	}
	switch c.Mode {
	case ModeProd, ModeDev:
	default:
		return fmt.Errorf("unknown mode %q: want prod or dev", c.Mode)
	}
	return nil
}

// SQLitePath is where the Instance keeps its single file.
func (c Config) SQLitePath() string {
	return filepath.Join(c.Data, "nooks.db")
}

// envOr returns the environment value for key, or fallback when it is unset or blank.
func envOr(env func(string) string, key, fallback string) string {
	if v := strings.TrimSpace(env(key)); v != "" {
		return v
	}
	return fallback
}
