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
	"github.com/hoshomoh/nooks/server/router/mcp"
	"github.com/hoshomoh/nooks/store"
)

// shutdownGrace is how long in-flight requests get to finish once a stop is asked for.
const shutdownGrace = 10 * time.Second

/*
maxRequestBody is the most any one request may send.

Four mebibytes is far past anything a person types. A Note is markdown somebody wrote,
an Item is a line, a List is a name; the largest honest request here is a few kilobytes,
and the ceiling is only meant to be out of the way of it.

What it stops is one request being enormous: an unbounded body is read into memory
before anything looks at it, so a single call could exhaust a home server's RAM, and
whatever it carried would land in the database. It does not stop somebody sending many
ordinary requests — bounding what an Instance holds in total is a different question,
and not one a limit here answers.
*/
const maxRequestBody = 4 << 20

/*
tidyEvery is how often the Instance tidies up after itself.

Hours rather than minutes: both jobs are about a shape that changes slowly, and neither
is urgent enough to wake a home server for.
*/
const tidyEvery = 6 * time.Hour

// Server is one Instance's HTTP server.
type Server struct {
	http  *http.Server
	log   *slog.Logger
	store store.Store
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
		mode:  cfg.Mode,
		store: s,
		http: &http.Server{
			Addr:              cfg.Addr,
			Handler:           requestLogger(log, withDefences(boundBodies(mux))),
			ReadHeaderTimeout: 10 * time.Second,
			// A body may be capped at four mebibytes and still arrive a byte a minute,
			// which holds a connection and the goroutine reading it for as long as the
			// sender likes. Requests here are small, so half a minute is generous.
			ReadTimeout: 30 * time.Second,
			// Between requests, not during one. Without it a keep-alive connection is
			// held until whichever side gives up first, which may be neither.
			IdleTimeout: 2 * time.Minute,
			// WriteTimeout is deliberately unset. The event stream writes for as long as
			// a browser has the app open, and a deadline on writing would end it on a
			// timer. What bounds a stalled reader there is the broker dropping events it
			// cannot hand over, not the clock.
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
	services := v1.Services{
		Activity: v1.NewActivityService(s, nil),
		Auth:     authService,
		Instance: v1.NewInstanceService(s),
		List:     v1.NewListService(s, nil, nil).WithAnnouncer(publisher),
		Member:   v1.NewMemberService(s, nil, nil).WithAnnouncer(publisher),
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

	// The same tools an assistant gets, through the same permission rules. A token that
	// may only read is refused a write here exactly as it is everywhere else.
	mux.Handle("/mcp", mcp.Handler(services, resolver))

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

/*
withDefences sets the headers every answer should carry.

Two of them, and both are free. nosniff stops a browser deciding for itself what a
response is: the REST API answers some things over GET, so a Member can be sent straight
to one, and what comes back is JSON with their own words in it. The type is always
declared, and this says not to second-guess it.

same-origin keeps a Nooks address off other people's servers. A path here names a List
and an Item, so a request that left carrying one would be handing a stranger the shape
of somebody's household.

Two more are deliberately absent, and the defense log says why: framing, which
self-hosters do on purpose in dashboards, and a content policy, which needs checking
against the built app rather than guessing.
*/
func withDefences(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := w.Header()
		header.Set("X-Content-Type-Options", "nosniff")
		header.Set("Referrer-Policy", "same-origin")
		next.ServeHTTP(w, r)
	})
}

/*
boundBodies refuses a request that is trying to send too much.

Wrapped around the whole mux rather than configured per handler: Connect, the REST
gateway, the MCP endpoint and the event stream are four different kinds of handler, and
a ceiling that only some of them have is one somebody will find the gap in.
*/
func boundBodies(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Body != nil {
			r.Body = http.MaxBytesReader(w, r.Body, maxRequestBody)
		}
		next.ServeHTTP(w, r)
	})
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
	// The URL and not only the address: a listener reports "[::]:8081", which is true
	// and is not something anybody can open.
	s.log.Info("nooks listening",
		"url", profile.URLFor(listener.Addr().String(), "/"),
		"addr", listener.Addr().String(),
		"mode", string(s.mode),
	)
	if s.mode == profile.ModeDev {
		s.log.Info("this mode serves the api only; run `cd apps/web && pnpm dev` for the app",
			"app", "http://localhost:3001")
	}

	go s.keepHouse(ctx)

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

/*
keepHouse does the two jobs nothing else was going to do.

Once at startup as well as on the tick, because an Instance that was off for a month
comes back with a month of expired sessions in it and should not wait six hours to say
so.
*/
func (s *Server) keepHouse(ctx context.Context) {
	s.tidyUp(ctx)

	ticker := time.NewTicker(tidyEvery)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.tidyUp(ctx)
		}
	}
}

/*
tidyUp clears out what has expired and has the database look at itself again.

Both are logged and carried on with rather than returned. A session row that outlives
its expiry is already refused on sight, and statistics going stale makes an Instance
slower rather than wrong; neither is a reason to stop serving the shopping list.
*/
func (s *Server) tidyUp(ctx context.Context) {
	// Expiry is enforced when a session is read, so these rows change nothing about who
	// can get in. Left alone they accumulate for the life of the Instance, which is the
	// only reason to sweep them.
	switch gone, err := s.store.DeleteExpiredSessions(ctx, time.Now()); {
	case err != nil:
		s.log.Warn("could not clear expired sessions", "error", err)
	case gone > 0:
		s.log.Info("cleared expired sessions", "count", gone)
	}

	// Past what the panel can show, except a request nobody has answered. Those are kept
	// however old they get and the read hands them back with the newest fifty, because
	// Activity is the only place a join or a reset appears and sweeping one leaves
	// somebody waiting on a decision nobody can see.
	switch gone, err := s.store.DeleteUnreadableActivity(ctx); {
	case err != nil:
		s.log.Warn("could not clear unreadable activity", "error", err)
	case gone > 0:
		s.log.Info("cleared activity nothing could read", "count", gone)
	}

	if err := s.store.Analyse(ctx); err != nil {
		s.log.Warn("could not refresh database statistics", "error", err)
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
