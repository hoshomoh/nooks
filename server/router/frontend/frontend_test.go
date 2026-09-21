package frontend

import (
	"crypto/sha256"
	"encoding/base64"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"
)

/*
The rules are tested against an app that is built here, not the embedded one.

The embedded copy is written by `pnpm --filter @nooks/web release` and is not committed,
so a test that read it would pass on a machine that had built the app and fail on one
that had not. CI is the second kind: it runs `go test` before it builds the app, which is
exactly the arrangement that should not decide whether these pass.
*/
func built() fs.FS {
	return fstest.MapFS{
		"index.html":               {Data: []byte("<!doctype html><title>nooks</title>")},
		"assets/index-a1b2c3d4.js": {Data: []byte("console.log(1)")},
		"favicon.svg":              {Data: []byte("<svg/>")},
	}
}

// get exercises the handler and returns the recorded response.
func get(t *testing.T, path string) *httptest.ResponseRecorder {
	t.Helper()
	rec := httptest.NewRecorder()
	handlerFor(built()).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
	return rec
}

func TestServesTheApp(t *testing.T) {
	rec := get(t, "/")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if rec.Body.Len() == 0 {
		t.Error("body is empty, want index.html")
	}
}

// TestClientRouteFallsBackToIndex is the property that makes a reload on a deep link
// work: an unknown path is the app's route, not a missing file.
func TestClientRouteFallsBackToIndex(t *testing.T) {
	rec := get(t, "/lists/groceries")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d for a client route, want %d", rec.Code, http.StatusOK)
	}
}

func TestIndexIsNotCached(t *testing.T) {
	rec := get(t, "/")
	if got := rec.Header().Get("Cache-Control"); got != htmlCacheControl {
		t.Errorf("Cache-Control = %q, want %q", got, htmlCacheControl)
	}
}

func TestHashedAssetsAreCachedForever(t *testing.T) {
	if got := cacheControlFor("assets/index-a1b2c3d4.js"); got != assetCacheControl {
		t.Errorf("cacheControlFor(hashed asset) = %q, want %q", got, assetCacheControl)
	}
	if got := cacheControlFor("favicon.svg"); got != htmlCacheControl {
		t.Errorf("cacheControlFor(unhashed file) = %q, want %q", got, htmlCacheControl)
	}
}

// A binary built without the app says so, rather than serving a page whose scripts are
// all missing. This is what a fresh clone produces, and the message is the whole point.
func TestSaysWhenTheAppWasNeverBuilt(t *testing.T) {
	rec := httptest.NewRecorder()
	handlerFor(fstest.MapFS{}).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
	if body := rec.Body.String(); !strings.Contains(body, "not built into this binary") {
		t.Errorf("body = %q, want it to say the app is missing", body)
	}
}

/*
The policy names the scripts the page actually carries.

The app writes one into the page — it sets the chosen theme before the first paint — and
a policy that did not name it would leave that frame the wrong colour on every load,
with only a console message to say why. Hashing the page being served is what stops the
two drifting apart, so this checks the hash is of that page and not of a copy of it.
*/
func TestThePolicyNamesThePageItIsServing(t *testing.T) {
	const page = `<!doctype html><html><head>` +
		`<script>document.documentElement.className = "dark"</script>` +
		`<script type="module" src="/assets/app.js"></script>` +
		`</head><body></body></html>`

	policy := policyFor([]byte(page))

	sum := sha256.Sum256([]byte(`document.documentElement.className = "dark"`))
	want := "'sha256-" + base64.StdEncoding.EncodeToString(sum[:]) + "'"
	if !strings.Contains(policy, want) {
		t.Errorf("policy = %q, want it to name %s", policy, want)
	}

	// The one with a src is loaded, not written in, so 'self' already covers it.
	if strings.Count(policy, "'sha256-") != 1 {
		t.Errorf("policy = %q, want exactly the one inline script named", policy)
	}
}

// Nothing may be loaded from anywhere else, and nothing may be written into the page
// that the page did not already carry.
func TestThePolicyShutsTheDoors(t *testing.T) {
	policy := policyFor([]byte(`<!doctype html><html></html>`))

	for _, want := range []string{
		"default-src 'self'",
		"script-src 'self'",
		"connect-src 'self'",
		"object-src 'none'",
		"base-uri 'self'",
		"form-action 'self'",
	} {
		if !strings.Contains(policy, want) {
			t.Errorf("policy = %q, want %q in it", policy, want)
		}
	}
	if strings.Contains(policy, "script-src 'self' 'unsafe-inline'") {
		t.Error("script-src allows any inline script, which is the whole thing this stops")
	}
}

// A page with nothing written into it names no hashes.
func TestAPageWithNoInlineScriptNamesNone(t *testing.T) {
	policy := policyFor([]byte(`<html><head><script src="/a.js"></script></head></html>`))

	if strings.Contains(policy, "sha256-") {
		t.Errorf("policy = %q, want no hash for a page that carries no inline script", policy)
	}
}

// The page is served under it, not merely able to produce one.
func TestTheIndexIsServedWithThePolicy(t *testing.T) {
	dist := fstest.MapFS{
		"index.html": &fstest.MapFile{Data: []byte(`<html><body>nooks</body></html>`)},
	}

	res := httptest.NewRecorder()
	handlerFor(dist).ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/lists/list_x", nil))

	if got := res.Header().Get("Content-Security-Policy"); !strings.Contains(got, "default-src 'self'") {
		t.Errorf("Content-Security-Policy = %q, want the page served under a policy", got)
	}
}
