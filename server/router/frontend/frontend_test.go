package frontend

import (
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
