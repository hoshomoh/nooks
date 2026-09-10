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
}

// NewHandler builds the endpoint.
func NewHandler(s store.Store, broker *events.Broker, resolver *auth.Resolver) *Handler {
	return &Handler{store: s, broker: broker, resolver: resolver}
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

	sub := h.broker.Watch(events.WatchParams{
		MemberID: member.ID,
		Name:     member.Name,
		ListUID:  listUID,
	})
	defer sub.Close()

	ticker := time.NewTicker(heartbeat)
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
			if _, err := fmt.Fprint(w, ": beat\n\n"); err != nil {
				return
			}
			flusher.Flush()
		}
	}
}

// member resolves the session cookie, the same way every other request does.
func (h *Handler) member(r *http.Request) (store.Member, bool) {
	cookie, err := r.Cookie(auth.CookieName)
	if err != nil {
		return store.Member{}, false
	}
	return h.resolver.Member(r.Context(), cookie.Value)
}

// watchableList is the List the browser says it is looking at, if the Member may see it.
func (h *Handler) watchableList(r *http.Request, member store.Member) (string, bool) {
	uid := r.URL.Query().Get("list")
	if uid == "" {
		return "", false
	}

	lists, err := h.store.ListsForMember(r.Context(), member.ID)
	if err != nil {
		return "", false
	}
	for _, list := range lists {
		if list.UID == uid {
			return uid, true
		}
	}
	return "", false
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
	payload, err := json.Marshal(wireEvent{
		Kind:     string(event.Kind),
		ListUID:  event.ListUID,
		Watchers: event.Watchers,
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

// NewPublisher builds one.
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

// ActivityArrived announces that something is waiting in a Member's panel.
func (p *Publisher) ActivityArrived(memberID int64) {
	p.broker.Publish(events.Event{Kind: events.KindActivity}, []int64{memberID})
}
