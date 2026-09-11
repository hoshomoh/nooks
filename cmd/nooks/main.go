// Command nooks runs a Nooks instance: one binary serving both the API and the app.
package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/hoshomoh/nooks/internal/profile"
	"github.com/hoshomoh/nooks/internal/version"
	"github.com/hoshomoh/nooks/server"
	"github.com/hoshomoh/nooks/store"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := run(ctx, os.Args[1:], os.Getenv, os.Stdout, os.Stderr); err != nil {
		if errors.Is(err, profile.ErrHelp) {
			return
		}
		fmt.Fprintln(os.Stderr, "nooks:", err)
		os.Exit(1)
	}
}

// run is the real entry point. It takes its arguments, environment and output as
// parameters so that it can be exercised without the process globals.
func run(ctx context.Context, args []string, env func(string) string, stdout, stderr io.Writer) error {
	if len(args) > 0 && (args[0] == "--version" || args[0] == "version") {
		fmt.Fprintln(stdout, version.String())
		return nil
	}

	cfg, err := profile.Parse(args, env, stderr)
	if err != nil {
		return err
	}

	log := slog.New(slog.NewTextHandler(stderr, &slog.HandlerOptions{Level: cfg.LogLevel}))

	s, err := openStore(ctx, cfg)
	if err != nil {
		return err
	}
	defer func() {
		if err := s.Close(); err != nil {
			log.Error("close store", "error", err)
		}
	}()

	srv, err := server.New(cfg, s, log)
	if err != nil {
		return err
	}
	return srv.Serve(ctx)
}

// openStore opens the database the configuration selects.
func openStore(ctx context.Context, cfg profile.Config) (store.Store, error) {
	switch cfg.Driver {
	case profile.DriverSQLite:
		return store.OpenSQLite(ctx, cfg.SQLitePath())
	case profile.DriverPostgres:
		return store.OpenPostgres(ctx, cfg.DSN)
	default:
		return nil, fmt.Errorf("unknown driver %q", cfg.Driver)
	}
}
