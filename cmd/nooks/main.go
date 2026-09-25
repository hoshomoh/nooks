// Command nooks runs a nooks instance: one binary serving both the API and the app.
package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
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

// run takes its arguments, environment and output as parameters so it can be exercised
// without the process globals.
func run(ctx context.Context, args []string, env func(string) string, stdout, stderr io.Writer) error {
	if len(args) > 0 && (args[0] == "--version" || args[0] == "version") {
		fmt.Fprintln(stdout, version.String())
		return nil
	}

	// Before the flags are parsed for a server, because this is not one: it asks a
	// server that is already running and says what it found.
	if healthRequested(args) {
		cfg, err := profile.Parse(args[1:], env, stderr)
		if err != nil {
			return err
		}
		return checkHealth(ctx, cfg.Addr, http.DefaultClient)
	}

	cfg, err := profile.Parse(args, env, stderr)
	if err != nil {
		return err
	}

	log := slog.New(slog.NewTextHandler(stderr, &slog.HandlerOptions{Level: cfg.LogLevel}))
	// Also the default, so the one place that turns a failure into an answer can say
	// what it was without a logger being threaded to it. See internalError.
	slog.SetDefault(log)

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

func openStore(ctx context.Context, cfg profile.Config) (store.Store, error) {
	switch cfg.Driver {
	case profile.DriverSQLite:
		return store.OpenSQLite(ctx, cfg.SQLitePath())
	case profile.DriverPostgres:
		return store.OpenPostgres(ctx, cfg.DSN, cfg.DBMaxConns)
	default:
		return nil, fmt.Errorf("unknown driver %q", cfg.Driver)
	}
}
