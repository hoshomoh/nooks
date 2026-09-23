package live

import (
	"bufio"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/hoshomoh/nooks/server/auth"
	"github.com/hoshomoh/nooks/server/events"
	"github.com/hoshomoh/nooks/store"
)

var testClock = time.Date(2026, 8, 25, 9, 0, 0, 0, time.UTC)

// instance is a store with one Member signed in, and the handler over it.
type instance struct {
	t       *testing.T
	store   store.Store
	broker  *events.Broker
	handler *Handler
	member  store.Member
	session string
}

func newInstance(t *testing.T) *instance {
	t.Helper()
	s, err := store.OpenSQLite(t.Context(), filepath.Join(t.TempDir(), "nooks.db"))
	if err != nil {
		t.Fatalf("OpenSQLite: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })

	member, err := s.CreateMember(t.Context(), store.CreateMemberParams{
		UID: "mem_anna", Name: "Anna", Email: "anna@brunnen.lan", Role: store.RoleAdmin,
		PasswordHash: "hash", CreatedAt: testClock,
	})
	if err != nil {
		t.Fatalf("CreateMember: %v", err)
	}

	token, hash, err := auth.NewToken()
	if err != nil {
		t.Fatalf("NewToken: %v", err)
	}
	if err := s.CreateSession(t.Context(), store.Session{
		TokenHash: hash, MemberID: member.ID,
		CreatedAt: testClock, ExpiresAt: testClock.Add(24 * time.Hour),
	}); err != nil {
		t.Fatalf("CreateSession: %v", err)
	}

	broker := events.NewBroker()
	resolver := auth.NewResolver(s, func() time.Time { return testClock })
	return &instance{
		t: t, store: s, broker: broker,
		handler: NewHandler(s, broker, resolver),
		member:  member, session: token,
	}
}

// list adds a List, so a stream has something to say it is watching.
func (i *instance) list(uid, name string, sharing store.Sharing, owner store.Member) store.List {
	i.t.Helper()
	list, err := i.store.CreateList(i.t.Context(), store.CreateListParams{
		UID: uid, Name: name, OwnerID: owner.ID, Sharing: sharing, CanEdit: true, At: testClock,
	})
	if err != nil {
		i.t.Fatalf("CreateList %s: %v", name, err)
	}
	return list
}

/*
open starts the stream and answers the lines a browser would read.

The handler holds the connection until its request is cancelled, so the request carries
a context the test ends, and the body is read on another goroutine the way a browser
reads it.
*/
func (i *instance) open(query, session string) (lines chan string, stop func(), code chan int) {
	i.t.Helper()
	ctx, cancel := context.WithCancel(i.t.Context())

	req := httptest.NewRequest(http.MethodGet, "/api/v1/events"+query, nil).WithContext(ctx)
	if session != "" {
		req.AddCookie(&http.Cookie{Name: auth.CookieName, Value: session})
	}

	read, write := io.Pipe()
	res := &streamRecorder{body: write, header: http.Header{}, code: http.StatusOK}
	code = make(chan int, 1)
	go func() {
		i.handler.ServeHTTP(res, req)
		code <- res.code
		_ = write.Close()
	}()

	lines = make(chan string, 64)
	go func() {
		scanner := bufio.NewScanner(read)
		for scanner.Scan() {
			lines <- scanner.Text()
		}
		close(lines)
	}()
	return lines, cancel, code
}

/*
keepPublishing offers the same event until somebody is there to hear it.

The broker holds nothing for a watcher that has not arrived, and a handler subscribes on
its own goroutine, so a test that publishes once is racing it. Repeating is honest here:
the event says only that something changed, and saying it twice says the same thing.
*/
func (i *instance) keepPublishing(event events.Event, to int64) {
	i.t.Helper()
	done := make(chan struct{})
	i.t.Cleanup(func() { close(done) })

	go func() {
		for {
			select {
			case <-done:
				return
			default:
				i.broker.Publish(event, []int64{to})
				time.Sleep(20 * time.Millisecond)
			}
		}
	}()
}

// waitFor reads until a line contains want, or gives up.
func waitFor(t *testing.T, lines chan string, want string) string {
	t.Helper()
	deadline := time.After(3 * time.Second)
	for {
		select {
		case line, open := <-lines:
			if !open {
				t.Fatalf("the stream closed before saying %q", want)
			}
			if strings.Contains(line, want) {
				return line
			}
		case <-deadline:
			t.Fatalf("nothing said %q in time", want)
		}
	}
}

// A stream is for somebody who is signed in, and for nobody else.
func TestTheStreamNeedsASession(t *testing.T) {
	i := newInstance(t)

	res := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/events", nil)
	i.handler.ServeHTTP(res, req)

	if res.Code != http.StatusUnauthorized {
		t.Errorf("code = %d, want 401 without a session", res.Code)
	}
}

/*
A change to a List reaches the browsers that may see it.

This is what the whole package is for, and none of it was covered: the stream held a
connection open and nothing said whether anything ever came down it.
*/
func TestAChangeReachesAWatchingBrowser(t *testing.T) {
	i := newInstance(t)
	list := i.list("list_groceries", "Groceries", store.SharingInstance, i.member)

	lines, stop, _ := i.open("?list="+list.UID, i.session)
	defer stop()

	// Presence lands as soon as somebody is watching, which is how we know it is open.
	waitFor(t, lines, "presence")

	i.broker.Publish(events.Event{Kind: events.KindListChanged, ListUID: list.UID},
		[]int64{i.member.ID})

	got := waitFor(t, lines, string(events.KindListChanged))
	if !strings.Contains(got, list.UID) {
		t.Errorf("event = %q, want it to name the List that changed", got)
	}
}

/*
A List the Member cannot see is not watched, and is not denied either.

Telling them apart would make the stream a way to ask whether a List exists. The stream
opens either way; it simply has nothing to say about that one.
*/
func TestAListTheMemberCannotSeeIsNotWatched(t *testing.T) {
	i := newInstance(t)
	jonas, err := i.store.CreateMember(t.Context(), store.CreateMemberParams{
		UID: "mem_jonas", Name: "Jonas", Email: "jonas@brunnen.lan", Role: store.RoleMember,
		PasswordHash: "hash", CreatedAt: testClock,
	})
	if err != nil {
		t.Fatalf("CreateMember: %v", err)
	}
	theirs := i.list("list_private", "Private", store.SharingPrivate, jonas)

	lines, stop, code := i.open("?list="+theirs.UID, i.session)
	defer stop()

	// It opens: a refusal here would say the List is there. Watching nothing means no
	// presence event to wait on, and nothing holds an event for a watcher that has not
	// subscribed yet, so this keeps offering one until the stream is listening.
	i.keepPublishing(events.Event{Kind: events.KindActivity}, i.member.ID)
	waitFor(t, lines, string(events.KindActivity))

	// And says nothing about a List it is not watching.
	i.broker.Publish(events.Event{Kind: events.KindListChanged, ListUID: theirs.UID},
		[]int64{jonas.ID})

	select {
	case line, open := <-lines:
		if open && strings.Contains(line, theirs.UID) {
			t.Errorf("the stream carried a List the Member cannot see: %q", line)
		}
	case <-time.After(300 * time.Millisecond):
	}

	stop()
	select {
	case <-code:
	case <-time.After(3 * time.Second):
		t.Error("the handler did not return when the browser went away")
	}
}

// A browser that goes away takes its subscription with it, or the broker fills up with
// connections nobody is reading.
func TestClosingTheBrowserEndsTheSubscription(t *testing.T) {
	i := newInstance(t)
	list := i.list("list_groceries", "Groceries", store.SharingInstance, i.member)

	lines, stop, code := i.open("?list="+list.UID, i.session)
	waitFor(t, lines, "presence")

	if watching := i.broker.WatchersOf(list.UID); len(watching) != 1 {
		t.Fatalf("watchers = %v, want the one that just opened", watching)
	}

	stop()
	select {
	case <-code:
	case <-time.After(3 * time.Second):
		t.Fatal("the handler did not return")
	}

	if watching := i.broker.WatchersOf(list.UID); len(watching) != 0 {
		t.Errorf("watchers = %v after the browser went, want none", watching)
	}
}

/*
A stream does not outlive the session that opened it.

Nothing on this side ends a stream: the handler holds the connection until the browser
goes away, and the heartbeat keeps it through any proxy that would have closed an idle
one. So a stream authorised once at open would go on saying what changed and who is
reading it to a browser that had signed out, for as long as it stayed connected.

That matters most where the code says so itself. CompletePasswordReset deletes every
session, sparing none, because "whoever else is signed in may well be why they are
here". The one thing it could not end was the stream they already had open.

What arrives is only that something changed, never what it changed to, plus the first
names of whoever is reading a List. An activity oracle rather than a content leak, on
an account its owner believes they have taken back.

The beat is shortened rather than waited out, and the session deleted the way
CompletePasswordReset ends them.
*/
func TestAStreamEndsWhenTheSessionDoes(t *testing.T) {
	i := newInstance(t)
	i.handler.beat = 10 * time.Millisecond

	lines, stop, code := i.open("", i.session)
	defer stop()

	i.keepPublishing(events.Event{Kind: events.KindActivity}, i.member.ID)
	waitFor(t, lines, "activity")

	if err := i.store.DeleteSessionsFor(t.Context(), i.member.ID, ""); err != nil {
		t.Fatalf("DeleteSessionsFor: %v", err)
	}

	select {
	case <-code:
	case <-time.After(3 * time.Second):
		t.Fatal("the stream is still open after every session was deleted")
	}
}

/*
streamRecorder is a ResponseWriter that streams.

httptest.ResponseRecorder buffers, so nothing can be read from it until the handler has
returned — and this handler does not return until the browser goes away. A pipe is the
other way round: what the handler writes is readable straight away, which is what a
browser sees.
*/
type streamRecorder struct {
	body   *io.PipeWriter
	header http.Header
	code   int
}

func (r *streamRecorder) Header() http.Header { return r.header }

func (r *streamRecorder) Write(b []byte) (int, error) { return r.body.Write(b) }

func (r *streamRecorder) WriteHeader(code int) { r.code = code }

// Flush is what makes this a stream rather than a response. A pipe has no buffer of its
// own, so there is nothing to push.
func (r *streamRecorder) Flush() {}
