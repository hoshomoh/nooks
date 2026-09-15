package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/hoshomoh/nooks/internal/profile"
)

/*
`nooks --health` asks a running Instance whether it is well, and exits 0 if it is.

The image ships a static binary and a CA bundle and nothing else, so there is no shell
and no curl to make the request. The binary has to check itself.
*/

const healthTimeout = 3 * time.Second

// healthRequested is read before the flags, because a health check is not a way of
// configuring a run.
func healthRequested(args []string) bool {
	return len(args) > 0 && (args[0] == "--health" || args[0] == "health")
}

// checkHealth reads the same flag and environment the server does, so a container
// started on a different port is checked on that port without being told twice.
func checkHealth(ctx context.Context, addr string, client *http.Client) error {
	ctx, cancel := context.WithTimeout(ctx, healthTimeout)
	defer cancel()

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, profile.URLFor(addr, "/healthz"), nil)
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
