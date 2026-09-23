package server

import (
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/hoshomoh/nooks/internal/profile"
	"github.com/hoshomoh/nooks/store/storetest"
)

/*
What bounds a connection, and the one that must stay unbounded.

Three of these stop a caller holding a connection open for nothing. The fourth is the
point of the test: WriteTimeout is deliberately absent, because the event stream writes
for as long as a browser has the app open and a deadline on writing would end it on a
timer. That is an easy thing to add in good faith while tightening the others, and live
updates would then stop working after a minute with nothing to say why.

ReadTimeout was checked against a real stream before it was set: it bounds reading the
request, and a handler that keeps writing afterwards is not cut off by it.
*/
func TestWhatBoundsAConnection(t *testing.T) {
	dir := t.TempDir()
	s := storetest.FreshIn(t, dir)

	server, err := New(profile.Config{
		Addr: ":0", Data: dir, Driver: profile.DriverSQLite, Mode: profile.ModeDev,
	}, s, slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	for _, one := range []struct {
		what string
		got  time.Duration
	}{
		{"ReadHeaderTimeout", server.http.ReadHeaderTimeout},
		{"ReadTimeout", server.http.ReadTimeout},
		{"IdleTimeout", server.http.IdleTimeout},
	} {
		if one.got <= 0 {
			t.Errorf("%s is unset, so a caller can hold a connection open for nothing", one.what)
		}
	}

	if server.http.WriteTimeout != 0 {
		t.Errorf("WriteTimeout is %v: the event stream writes for as long as the app is "+
			"open, and a deadline on writing ends it on a timer", server.http.WriteTimeout)
	}
}
