package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthRequested(t *testing.T) {
	for _, args := range [][]string{{"--health"}, {"health"}, {"--health", "--addr", ":9090"}} {
		if !healthRequested(args) {
			t.Errorf("healthRequested(%v) = false, want true", args)
		}
	}
	for _, args := range [][]string{{}, {"--addr", ":9090"}, {"--version"}} {
		if healthRequested(args) {
			t.Errorf("healthRequested(%v) = true, want false", args)
		}
	}
}

/*
A listen address is not a host a client can dial.

":8081" and "0.0.0.0:8081" are both how a server says "every interface", and both mean
"this machine" to the process asking. Getting this wrong is a health check that fails
on a perfectly well Instance, which is worse than none — an orchestrator restarts it.
*/
func TestHealthURLIsSomethingDialable(t *testing.T) {
	for addr, want := range map[string]string{
		":8081":          "http://127.0.0.1:8081/healthz",
		"0.0.0.0:8081":   "http://127.0.0.1:8081/healthz",
		"[::]:8081":      "http://127.0.0.1:8081/healthz",
		"127.0.0.1:8081": "http://127.0.0.1:8081/healthz",
		"localhost:9090": "http://localhost:9090/healthz",
		" :8081 ":        "http://127.0.0.1:8081/healthz",
	} {
		if got := healthURL(addr); got != want {
			t.Errorf("healthURL(%q) = %q, want %q", addr, got, want)
		}
	}
}

func TestHealthCheckReadsTheAnswer(t *testing.T) {
	well := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/healthz" {
			t.Errorf("asked for %q, want /healthz", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer well.Close()

	if err := checkHealth(t.Context(), hostOf(well.URL), well.Client()); err != nil {
		t.Errorf("checkHealth on a well Instance = %v, want nil", err)
	}
}

func TestHealthCheckFailsOnAnUnwellInstance(t *testing.T) {
	unwell := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer unwell.Close()

	if err := checkHealth(t.Context(), hostOf(unwell.URL), unwell.Client()); err == nil {
		t.Error("checkHealth on an unwell Instance = nil, want an error")
	}
}

func TestHealthCheckFailsWhenNothingIsListening(t *testing.T) {
	// Nothing is bound here, which is what a container that has not started yet looks
	// like — and the case a start_period exists for.
	closed := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	addr := hostOf(closed.URL)
	closed.Close()

	if err := checkHealth(context.Background(), addr, http.DefaultClient); err == nil {
		t.Error("checkHealth against nothing = nil, want an error")
	}
}

/** hostOf is the host:port out of a test server's URL. */
func hostOf(url string) string {
	return url[len("http://"):]
}
