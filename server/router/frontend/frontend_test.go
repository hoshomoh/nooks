package frontend

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// get exercises the handler and returns the recorded response.
func get(t *testing.T, path string) *httptest.ResponseRecorder {
	t.Helper()
	h, err := Handler()
	if err != nil {
		t.Fatalf("Handler: %v", err)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
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
