// Package server wires an Instance's HTTP surface: the Connect API and, in prod, the
// embedded app. It owns transport concerns only — no persistence, no policy.
package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"time"

	"connectrpc.com/connect"

	"github.com/hoshomoh/nooks/internal/profile"
	apiv1 "github.com/hoshomoh/nooks/proto/gen/nooks/api/v1/apiv1connect"
	"github.com/hoshomoh/nooks/server/auth"
	"github.com/hoshomoh/nooks/server/events"
	v1 "github.com/hoshomoh/nooks/server/router/api/v1"
	"github.com/hoshomoh/nooks/server/router/backup"
	"github.com/hoshomoh/nooks/server/router/frontend"
	"github.com/hoshomoh/nooks/server/router/gateway"
	"github.com/hoshomoh/nooks/server/router/live"
	"github.com/hoshomoh/nooks/store"
)

// shutdownGrace is how long in-flight requests get to finish once a stop is asked for.
const shutdownGrace = 10 * time.Second

// Server is one Instance's HTTP server.
type Server struct {
	http *http.Server
	log  *slog.Logger
	// mode is kept so the startup lines can say what this process is actually serving.
	mode profile.Mode
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

	mux, err := newMux(cfg, s)
	if err != nil {
		return nil, err
	}

	return &Server{
		mode: cfg.Mode,
		http: &http.Server{
			Addr:              cfg.Addr,
			Handler:           requestLogger(log, mux),
			ReadHeaderTimeout: 10 * time.Second,
		},
		log: log,
	}, nil
}

// newMux wires every route an Instance answers.
func newMux(cfg profile.Config, s store.Store) (*http.ServeMux, error) {
	// Every request passes through the resolver, which attaches the signed-in Member
	// when there is one. It never rejects: first run, sign-in and the Public list are
	// all legitimately anonymous.
	// One broker per process. Nooks is one binary on one machine, so there is nothing
	// to coordinate between.
	broker := events.NewBroker()
	publisher := live.NewPublisher(s, broker)

	// The resolver is the only thing that sees a token presented, so it is what tells a
	// Member their key has started being used.
	tokenActivity := v1.NewTokenActivity(s, nil, nil).WithAnnouncer(publisher)
	resolver := auth.NewResolver(s, nil).WithTokenWatcher(tokenActivity)
	interceptors := connect.WithInterceptors(resolver.Interceptor())

	authService := v1.NewAuthService(s, v1.AuthServiceOptions{
		Secure: cfg.SecureCookies,
	}).WithAnnouncer(publisher)

	// Built once and served twice: the browser reaches them over Connect, a script or an
	// alternative client over REST. One instance of each, so the two surfaces cannot
	// drift into behaving differently.
	services := gateway.Services{
		Activity: v1.NewActivityService(s, nil),
		Auth:     authService,
		Instance: v1.NewInstanceService(s),
		List:     v1.NewListService(s, nil, nil).WithAnnouncer(publisher),
		Member:   v1.NewMemberService(s, nil, nil),
		Public:   v1.NewPublicService(s),
		Request:  v1.NewRequestService(s, nil),
		Token:    v1.NewTokenService(s, nil, nil, nil).WithActivity(tokenActivity),
	}

	mux := http.NewServeMux()
	mux.Handle(apiv1.NewInstanceServiceHandler(services.Instance, interceptors))
	mux.Handle(apiv1.NewAuthServiceHandler(services.Auth, interceptors))
	mux.Handle(apiv1.NewRequestServiceHandler(services.Request, interceptors))
	mux.Handle(apiv1.NewListServiceHandler(services.List, interceptors))
	mux.Handle(apiv1.NewMemberServiceHandler(services.Member, interceptors))
	mux.Handle(apiv1.NewActivityServiceHandler(services.Activity, interceptors))
	mux.Handle(apiv1.NewPublicServiceHandler(services.Public, interceptors))
	mux.Handle(apiv1.NewTokenServiceHandler(services.Token, interceptors))
	mux.Handle("GET /api/v1/events", live.NewHandler(s, broker, resolver))
	// The REST API, generated from the same protos the app is built against. Mounted
	// under its own prefix so the Connect routes, the event stream and the backup
	// download keep theirs.
	rest, err := gateway.New(services, resolver)
	if err != nil {
		return nil, fmt.Errorf("build the rest api: %w", err)
	}
	mux.Handle("/api/v1/", rest)

	mux.Handle("GET /api/v1/backup", backup.NewHandler(s, resolver, nil))
	mux.HandleFunc("GET /healthz", handleHealthz)

	// In dev the Vite server serves the app and proxies here, so the binary serves only
	// the API. In prod one binary serves both.
	if cfg.Mode == profile.ModeProd {
		app, err := frontend.Handler()
		if err != nil {
			return nil, err
		}
		mux.Handle("/", app)
	}
	return mux, nil
}

// handleHealthz answers the liveness check a container runtime asks.
func handleHealthz(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

// Serve listens until the context is cancelled, then shuts down gracefully.
//
// The port is taken before anything is logged, so "listening" is only ever said by a
// process that is. Announcing it first and binding afterwards means a server that
// failed to start still reports that it started, which is the one line somebody reads
// before deciding the problem is somewhere else.
func (s *Server) Serve(ctx context.Context) error {
	listener, err := net.Listen("tcp", s.http.Addr)
	if err != nil {
		return fmt.Errorf("listen on %s: %w", s.http.Addr, err)
	}
	s.log.Info("nooks listening", "addr", listener.Addr().String(), "mode", string(s.mode))
	if s.mode == profile.ModeDev {
		s.log.Info("serving the api only; the app is served by vite", "app", "http://localhost:3001")
	}

	errs := make(chan error, 1)
	go func() {
		if err := s.http.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errs <- fmt.Errorf("serve on %s: %w", s.http.Addr, err)
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

	s.log.Info("nooks stopping")
	if err := s.http.Shutdown(ctx); err != nil {
		return fmt.Errorf("shut down: %w", err)
	}
	return nil
}

/*
requestLogger records one line per request.

At info, not debug. Somebody self-hosting has no other window into what their Instance
is doing, and a server that says nothing between starting and stopping is one you cannot
tell apart from a server that has hung. Turning it down is a flag away.

Two paths are left out, because both would drown the rest: the health check, which
something may be polling every few seconds, and the event stream, which is one request
that stays open for as long as a browser is watching.
*/
func requestLogger(log *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		if quietPath(r.URL.Path) {
			return
		}
		log.Info("request", "method", r.Method, "path", r.URL.Path, "took", time.Since(start))
	})
}

// quietPath reports whether a path is one that would fill the log by itself.
func quietPath(path string) bool {
	return path == "/healthz" || path == "/api/v1/events"
}
