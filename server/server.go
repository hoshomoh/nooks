// Package server wires an Instance's HTTP surface: the Connect API and, in prod, the
// embedded app. It owns transport concerns only — no persistence, no policy.
package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/hoshomoh/nooks/internal/profile"
	apiv1 "github.com/hoshomoh/nooks/proto/gen/nook/api/v1/apiv1connect"
	v1 "github.com/hoshomoh/nooks/server/router/api/v1"
	"github.com/hoshomoh/nooks/store"
)

// shutdownGrace is how long in-flight requests get to finish once a stop is asked for.
const shutdownGrace = 10 * time.Second

// Server is one Instance's HTTP server.
type Server struct {
	http *http.Server
	log  *slog.Logger
}

// New builds a Server. Its dependencies are passed in rather than constructed here, so
// a test can supply a fake store and a discarding logger.
func New(cfg profile.Config, s store.Store, log *slog.Logger) (*Server, error) {
	if s == nil {
		return nil, errors.New("server: store is required")
	}
	if log == nil {
		return nil, errors.New("server: logger is required")
	}

	mux := http.NewServeMux()
	mux.Handle(apiv1.NewInstanceServiceHandler(v1.NewInstanceService(s)))
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	return &Server{
		http: &http.Server{
			Addr:              cfg.Addr,
			Handler:           requestLogger(log, mux),
			ReadHeaderTimeout: 10 * time.Second,
		},
		log: log,
	}, nil
}

// Serve listens until the context is cancelled, then shuts down gracefully.
func (s *Server) Serve(ctx context.Context) error {
	errs := make(chan error, 1)
	go func() {
		s.log.Info("nook listening", "addr", s.http.Addr)
		if err := s.http.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errs <- fmt.Errorf("listen on %s: %w", s.http.Addr, err)
			return
		}
		errs <- nil
	}()

	select {
	case err := <-errs:
		return err
	case <-ctx.Done():
		return s.shutdown()
	}
}

// shutdown stops the server, giving in-flight requests a moment to finish.
func (s *Server) shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), shutdownGrace)
	defer cancel()

	s.log.Info("nook stopping")
	if err := s.http.Shutdown(ctx); err != nil {
		return fmt.Errorf("shut down: %w", err)
	}
	return nil
}

// requestLogger records one line per request at debug level. It is deliberately quiet:
// a home server's log should be readable a week later.
func requestLogger(log *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Debug("request", "method", r.Method, "path", r.URL.Path, "took", time.Since(start))
	})
}
