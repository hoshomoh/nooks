// Package frontend serves the built app from inside the binary.
//
// `pnpm --filter @nooks/web release` writes the build into dist/, which go:embed bakes
// in. A placeholder index.html is committed so that a fresh clone compiles before the
// app has ever been built.
package frontend

import (
	"embed"
	"fmt"
	"io/fs"
	"net/http"
	"strings"
	"time"
)

//go:embed dist
var embedded embed.FS

const (
	// htmlCacheControl keeps index.html uncached, so a Member picks up a new build on
	// their next load rather than after a hard refresh.
	htmlCacheControl = "no-cache, no-store, must-revalidate"
	// assetCacheControl is for Vite's content-hashed files, whose names change when
	// their contents do.
	assetCacheControl = "public, max-age=31536000, immutable"
)

// Handler serves the embedded app with a single-page fallback: a request that does not
// name a real file is answered with index.html, so a client-side route survives a
// reload.
func Handler() (http.Handler, error) {
	dist, err := fs.Sub(embedded, "dist")
	if err != nil {
		return nil, fmt.Errorf("open embedded app: %w", err)
	}
	files := http.FileServer(http.FS(dist))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name := strings.TrimPrefix(r.URL.Path, "/")
		if name == "" || !fileExists(dist, name) {
			serveIndex(w, r, dist)
			return
		}
		w.Header().Set("Cache-Control", cacheControlFor(name))
		files.ServeHTTP(w, r)
	}), nil
}

// serveIndex writes index.html for a client-side route.
func serveIndex(w http.ResponseWriter, r *http.Request, dist fs.FS) {
	page, err := fs.ReadFile(dist, "index.html")
	if err != nil {
		http.Error(w, "app is not built into this binary", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", htmlCacheControl)
	// A zero modtime tells ServeContent not to negotiate freshness; the
	// Cache-Control header above is the whole policy.
	http.ServeContent(w, r, "index.html", time.Time{}, strings.NewReader(string(page)))
}

// cacheControlFor picks a cache policy from the path. Vite writes hashed files under
// assets/, and only those are safe to cache forever.
func cacheControlFor(name string) string {
	if strings.HasPrefix(name, "assets/") {
		return assetCacheControl
	}
	return htmlCacheControl
}

// fileExists reports whether name is a regular file in the embedded app.
func fileExists(dist fs.FS, name string) bool {
	info, err := fs.Stat(dist, name)
	return err == nil && !info.IsDir()
}
