package main

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"
)

/*
Asking a running Instance whether it is well, from inside its own image.

The image ships a static binary and a CA bundle: no shell, no curl, nothing that could
make an HTTP request on a health check's behalf. That is the point of it — there is
nothing in there to keep patched — but it does mean the only thing that can check the
Instance is the Instance's own binary.

	nooks --health

Exits 0 when /healthz answers, and non-zero otherwise, which is the whole of what
Docker and every orchestrator want from a health check.
*/

/** HEALTH_TIMEOUT is how long a check waits before calling it unwell. */
const healthTimeout = 3 * time.Second

// healthRequested reports whether the arguments ask for a health check rather than a
// server. Read before the flags, because it is not a way of configuring a run.
func healthRequested(args []string) bool {
	return len(args) > 0 && (args[0] == "--health" || args[0] == "health")
}

/*
checkHealth asks the address the Instance would be listening on.

It reads the same flag and environment the server does, so a check inside a container
started with a different port asks the right one without being told twice.
*/
func checkHealth(ctx context.Context, addr string, client *http.Client) error {
	ctx, cancel := context.WithTimeout(ctx, healthTimeout)
	defer cancel()

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, healthURL(addr), nil)
	if err != nil {
		return err
	}

	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer func() {
		_, _ = io.Copy(io.Discard, response.Body)
		_ = response.Body.Close()
	}()

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("health: %s", response.Status)
	}
	return nil
}

/*
healthURL turns a listen address into something to ask.

A listen address is often ":8081" or "0.0.0.0:8081", and neither is a host a client can
dial. Both mean "this machine" to the process doing the asking.
*/
func healthURL(addr string) string {
	host, port, err := net.SplitHostPort(strings.TrimSpace(addr))
	if err != nil || host == "" || host == "0.0.0.0" || host == "::" {
		host = "127.0.0.1"
	}
	if err != nil {
		port = strings.TrimPrefix(strings.TrimSpace(addr), ":")
	}
	return fmt.Sprintf("http://%s/healthz", net.JoinHostPort(host, port))
}
