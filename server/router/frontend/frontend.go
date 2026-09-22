// Package frontend serves the built app from inside the binary.
//
// `pnpm --filter @nooks/web release` writes the build into dist/, which go:embed bakes
// in. Only .gitkeep is committed, so a binary built without that step says so rather
// than serving an index.html whose scripts all 404.
//
// `all:` so .gitkeep is embedded too: without it a fresh clone has no matching files
// and does not compile. The release script writes the marker back afterwards, because
// `--emptyOutDir` takes the whole directory with it.
package frontend

import (
	"bytes"
	"crypto/sha256"
	"embed"
	"encoding/base64"
	"fmt"
	"io/fs"
	"net/http"
	"regexp"
	"strings"
	"time"
)

//go:embed all:dist
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
	return handlerFor(dist), nil
}

/*
handlerFor serves one app, from wherever it is.

Taking the files rather than reaching for the embedded ones is what makes the rules
above testable. The embedded copy is written by the app build and is not committed, so a
test that read it would pass on a machine that had built the app and fail on one that
had not — which is what CI is, every time.
*/
func handlerFor(dist fs.FS) http.Handler {
	files := http.FileServer(http.FS(dist))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name := strings.TrimPrefix(r.URL.Path, "/")
		if name == "" || !fileExists(dist, name) {
			serveIndex(w, r, dist)
			return
		}
		w.Header().Set("Cache-Control", cacheControlFor(name))
		files.ServeHTTP(w, r)
	})
}

// serveIndex writes index.html for a client-side route.
func serveIndex(w http.ResponseWriter, r *http.Request, dist fs.FS) {
	page, err := fs.ReadFile(dist, "index.html")
	if err != nil {
		http.Error(w, "app is not built into this binary", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Security-Policy", policyFor(page))
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

/*
The policy the app is served under.

Everything comes from this origin. Nooks loads no script, style, font or image from
anywhere else and talks to nothing but itself, so saying so costs nothing and means that
a way to inject a script into a page would still have nowhere to load one from.

style-src allows inline because the menus and popovers position themselves by writing a
style attribute, which is what style-src governs. Pinning those would mean pinning a
number that moves with the pointer.

frame-ancestors is 'none' rather than 'self'. Nothing here frames this page, so 'self'
would allow something no part of the app needs, and one frame is all the attack takes: a
page loads this one invisibly, lines its own buttons up over the Member's, and collects
the presses. What they think they are dismissing is what deletes a List.
*/
const policyRules = "default-src 'self'; " +
	"style-src 'self' 'unsafe-inline'; " +
	"img-src 'self' data: blob:; " +
	"font-src 'self' data:; " +
	"connect-src 'self'; " +
	"object-src 'none'; " +
	"base-uri 'self'; " +
	"frame-ancestors 'none'; " +
	"form-action 'self'"

// inlineScript finds a script written into the page rather than loaded by it.
var inlineScript = regexp.MustCompile(`(?s)<script(?:\s[^>]*)?>(.*?)</script>`)

/*
policyFor writes the policy for one page, naming the scripts written into it.

The app carries one: it reads the chosen theme and sets it before anything is painted,
which is a thing that has to happen inline or the first frame is the wrong colour. A
policy that refused it would leave that frame wrong on every load, and nothing would say
why.

Hashed from the page being served rather than written down somewhere else. A hash kept
apart from what it describes is a hash that goes stale, and the failure is a blank app
in somebody's house with a console message they will never see.
*/
func policyFor(page []byte) string {
	policy := policyRules + "; script-src 'self'"

	for _, found := range inlineScript.FindAllSubmatch(page, -1) {
		if len(bytes.TrimSpace(found[1])) == 0 {
			continue
		}
		sum := sha256.Sum256(found[1])
		policy += " 'sha256-" + base64.StdEncoding.EncodeToString(sum[:]) + "'"
	}
	return policy
}
