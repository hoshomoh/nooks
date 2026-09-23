// Package live streams changes to the browsers watching for them.
//
// Server-sent events rather than websockets: everything here goes one way, from the
// Instance to the browser, and a browser already knows how to reconnect an EventSource
// on its own. A websocket would be a second protocol to keep alive for no second
// direction of traffic.
package live

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/hoshomoh/nooks/server/auth"
	"github.com/hoshomoh/nooks/server/events"
	v1 "github.com/hoshomoh/nooks/server/router/api/v1"
	"github.com/hoshomoh/nooks/store"
)

// heartbeat keeps the connection open through anything in between that would otherwise
// close an idle one — a reverse proxy, a phone's radio going to sleep.
const heartbeat = 25 * time.Second

// Handler streams events to one signed-in Member.
type Handler struct {
	store    store.Store
	broker   *events.Broker
	resolver *auth.Resolver
	// beat is how often the connection is kept alive and the session looked at again.
	// A field rather than the constant so a test can reach that tick without waiting
	// out a real one.
	beat time.Duration
}

func NewHandler(s store.Store, broker *events.Broker, resolver *auth.Resolver) *Handler {
	return &Handler{store: s, broker: broker, resolver: resolver, beat: heartbeat}
}

// wireEvent is what a browser receives.
//
// It says what changed rather than what it changed to: the browser re-reads through the
// same API it always uses, so a live update can never show somebody something the API
// would not have given them.
type wireEvent struct {
	Kind string `json:"kind"`
	// ListUID is the List it concerns, when it concerns one.
	ListUID string `json:"listUid,omitempty"`
	// Watchers is who is looking at that List, for a presence event.
	Watchers []string `json:"watchers,omitempty"`
	// Editing is who has a Note open, by Item, so everybody else reads that one rather
	// than writing over them.
	Editing []wireEditor `json:"editing,omitempty"`
}

// wireEditor is one person with one Note open.
type wireEditor struct {
	ItemUID string `json:"itemUid"`
	Name    string `json:"name"`
}

// ServeHTTP opens the stream and holds it until the browser goes away.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	member, ok := h.member(r)
	if !ok {
		http.Error(w, "sign in first", http.StatusUnauthorized)
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming is not available here", http.StatusInternalServerError)
		return
	}

	listUID, ok := h.watchableList(r, member)
	if !ok {
		// A List they cannot reach is a List that does not exist, so the stream opens
		// without it rather than telling them the difference.
		listUID = ""
	}

	writeStreamHeaders(w)
	flusher.Flush()

	/*
	 * The Note this connection has open, if any.
	 *
	 * Held on the watch rather than claimed and released by calls of its own, so the
	 * stream ending is what lets it go. A laptop closed mid-edit cannot hold a Note:
	 * the thing holding it is the connection, and that is already bounded by the
	 * heartbeat and by the browser going away.
	 *
	 * Not checked against the List here. A Note on a List this Member cannot reach is
	 * not a Note anybody is told about, because presence only ever goes to watchers of
	 * the List it names, and this one has already been checked.
	 */
	sub := h.broker.Watch(events.WatchParams{
		MemberID:   member.ID,
		Name:       member.Name,
		ListUID:    listUID,
		EditingUID: editingFrom(r, listUID),
	})
	defer sub.Close()

	ticker := time.NewTicker(h.beat)
	defer ticker.Stop()

	for {
		select {
		case <-r.Context().Done():
			return
		case event, open := <-sub.Events:
			if !open {
				return
			}
			if err := writeEvent(w, event); err != nil {
				return
			}
			flusher.Flush()
		case <-ticker.C:
			// The session is checked again, not only when the stream opened. A stream
			// outlives almost everything: nothing on this side closes one, and the
			// heartbeat below keeps it through any proxy that would have timed it out,
			// so without this a signed-out browser goes on being told what changed and
			// who is reading it. CompletePasswordReset says what that is worth when
			// it ends every session: "whoever else is signed in may well be why they
			// are here".
			if _, ok := h.member(r); !ok {
				return
			}
			if _, err := fmt.Fprint(w, ": beat\n\n"); err != nil {
				return
			}
			flusher.Flush()
		}
	}
}

// member resolves the session cookie, the same way every other request does.
func (h *Handler) member(r *http.Request) (store.Member, bool) {
	token := auth.SessionTokenFrom(r.Header)
	if token == "" {
		return store.Member{}, false
	}
	return h.resolver.Member(r.Context(), token)
}

// watchableList is the List the browser says it is looking at, if the Member may see it.
func (h *Handler) watchableList(r *http.Request, member store.Member) (string, bool) {
	uid := r.URL.Query().Get("list")
	if uid == "" {
		return "", false
	}

	reaches, err := h.store.CanReachList(r.Context(), member.ID, uid)
	if err != nil || !reaches {
		return "", false
	}
	return uid, true
}

// editingFrom is the Item whose Note the browser says it has open, or nothing.
//
// Only meaningful alongside a List, since that is what presence is announced on: an
// editor with no List to be announced to is nobody.
func editingFrom(r *http.Request, listUID string) string {
	if listUID == "" {
		return ""
	}
	return r.URL.Query().Get("editing")
}

// writeStreamHeaders says this is a stream, and asks everything in between not to hold
// it in a buffer waiting for more.
func writeStreamHeaders(w http.ResponseWriter) {
	header := w.Header()
	header.Set("Content-Type", "text/event-stream")
	header.Set("Cache-Control", "no-cache")
	header.Set("Connection", "keep-alive")
	// nginx buffers proxied responses by default, which would hold every event until
	// the connection closed. Self-hosting means somebody else's proxy is in the way.
	header.Set("X-Accel-Buffering", "no")
}

// writeEvent sends one event in the format an EventSource reads.
func writeEvent(w http.ResponseWriter, event events.Event) error {
	editing := make([]wireEditor, 0, len(event.Editing))
	for _, one := range event.Editing {
		editing = append(editing, wireEditor{ItemUID: one.ItemUID, Name: one.Name})
	}

	payload, err := json.Marshal(wireEvent{
		Kind:     string(event.Kind),
		ListUID:  event.ListUID,
		Watchers: event.Watchers,
		Editing:  editing,
	})
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(w, "data: %s\n\n", payload)
	return err
}

// Publisher is how the services announce a change.
//
// It resolves who may see a List before sending, so the broker never has to decide who
// can see what — that decision stays beside the rest of the permissions.
type Publisher struct {
	store  store.Store
	broker *events.Broker
}

func NewPublisher(s store.Store, broker *events.Broker) *Publisher {
	return &Publisher{store: s, broker: broker}
}

// ListChanged announces that a List, or something on it, changed.
func (p *Publisher) ListChanged(ctx context.Context, list store.List) {
	audience, err := v1.AudienceOf(ctx, p.store, list)
	if err != nil {
		// A change that was made and saved is not undone by failing to announce it.
		// The browser re-reads on its own next time it asks for anything.
		return
	}
	p.broker.Publish(events.Event{Kind: events.KindListChanged, ListUID: list.UID}, audience)
}

// ListsChanged announces that which Lists somebody can reach has changed.
func (p *Publisher) ListsChanged(audience []int64) {
	p.broker.Publish(events.Event{Kind: events.KindListsChanged}, audience)
}

/*
MemberChanged announces that a Member's own account changed.

An Admin changing somebody's role is the case: the app reads who it is signed in as once
and holds it for the life of the tab, so without this a Member promoted to Admin never
sees the screens they were just given, and one demoted goes on being offered buttons
that the Instance refuses.
*/
func (p *Publisher) MemberChanged(memberID int64) {
	p.broker.Publish(events.Event{Kind: events.KindMemberChanged}, []int64{memberID})
}

/*
ReachLost ends the streams of Members a List was just taken away from.

Their browsers are told the List is gone by ListsChanged, and reconnect on their own.
What this closes is the stream itself, because presence is checked when a watch opens
and never again: without it they would go on being told who else is reading a List they
can no longer open.
*/
func (p *Publisher) ReachLost(listUID string, members []int64) {
	p.broker.EndWatchesOn(listUID, members)
}

// ActivityArrived announces that something is waiting in a Member's panel.
func (p *Publisher) ActivityArrived(memberID int64) {
	p.broker.Publish(events.Event{Kind: events.KindActivity}, []int64{memberID})
}
