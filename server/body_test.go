package server

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hoshomoh/nooks/internal/profile"
	"github.com/hoshomoh/nooks/store"
)

/*
A request that sends too much is refused before anything reads it.

Nothing bounded a body, so one call could carry as much as a caller cared to send: read
into memory before a handler saw it, and then stored. A home server is the whole
deployment, and its memory and its disk are the ones being spent.

Tested through the wrapper rather than the server, because what is being asserted is
that the ceiling is applied at all — every handler behind it is a different kind.
*/
func TestABodyPastTheCeilingIsRefused(t *testing.T) {
	var read int
	handler := boundBodies(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		read = len(body)
		if err != nil {
			http.Error(w, "too much", http.StatusRequestEntityTooLarge)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))

	tooMuch := httptest.NewRequest(http.MethodPost, "/api/v1/lists",
		strings.NewReader(strings.Repeat("x", maxRequestBody+1024)))
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, tooMuch)

	if res.Code != http.StatusRequestEntityTooLarge {
		t.Errorf("code = %d, want 413 for a body past the ceiling", res.Code)
	}
	if read > maxRequestBody {
		t.Errorf("the handler read %d bytes, want no more than the ceiling", read)
	}
}

// An ordinary request is nowhere near it, and nothing about it changes.
func TestAnOrdinaryBodyIsUntouched(t *testing.T) {
	const note = "### Where\nSaturday market. They pack up around two."

	handler := boundBodies(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("reading an ordinary body failed: %v", err)
		}
		if string(body) != note {
			t.Errorf("body = %q, want it through unchanged", string(body))
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/items", strings.NewReader(note))
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Errorf("code = %d, want 200", res.Code)
	}
}

/*
The ceiling is actually on the server, not merely written.

The two tests above call boundBodies directly, so they pass whether or not anything
wires it in. This one asks the handler the server was built with, which is the thing
that has to have it.
*/
func TestTheServerAppliesTheCeiling(t *testing.T) {
	dir := t.TempDir()
	s, err := store.OpenSQLite(t.Context(), filepath.Join(dir, "nooks.db"))
	if err != nil {
		t.Fatalf("OpenSQLite: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })

	server, err := New(profile.Config{
		Addr: ":0", Data: dir, Driver: profile.DriverSQLite, Mode: profile.ModeDev,
	}, s, slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	// Valid for the call it is making, and enormous. Nonsense of the same size would be
	// refused for being nonsense, which is not what is being asked here: without the
	// ceiling this body parses, and the answer is the ordinary "not signed in".
	body := `{"name":"` + strings.Repeat("x", maxRequestBody+1024) + `"}`
	req := httptest.NewRequest(http.MethodPost, "/nooks.api.v1.ListService/CreateList",
		strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()
	server.http.Handler.ServeHTTP(res, req)

	if !strings.Contains(res.Body.String(), "resource_exhausted") {
		t.Errorf("answered %d %s, want the request refused for its size",
			res.Code, res.Body.String())
	}
}

/*
Every answer carries the two headers that cost nothing.

Asserted on the server rather than the wrapper, because a header written in a function
nothing wraps with is a header nobody gets. The event stream is checked alongside an
ordinary request: it takes a different path out and would be an easy one to miss.
*/
func TestEveryAnswerCarriesItsDefences(t *testing.T) {
	dir := t.TempDir()
	s, err := store.OpenSQLite(t.Context(), filepath.Join(dir, "nooks.db"))
	if err != nil {
		t.Fatalf("OpenSQLite: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })

	server, err := New(profile.Config{
		Addr: ":0", Data: dir, Driver: profile.DriverSQLite, Mode: profile.ModeDev,
	}, s, slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	for _, path := range []string{"/healthz", "/api/v1/lists", "/api/v1/events"} {
		t.Run(path, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, path, nil)
			res := httptest.NewRecorder()
			server.http.Handler.ServeHTTP(res, req)

			for header, want := range map[string]string{
				"X-Content-Type-Options": "nosniff",
				"Referrer-Policy":        "same-origin",
			} {
				if got := res.Header().Get(header); got != want {
					t.Errorf("%s = %q, want %q", header, got, want)
				}
			}
		})
	}
}
