// Package events carries changes from the Member who made them to the Members watching.
//
// Everything here is in memory and per process. Nooks is one binary on one machine, so
// there is nothing to coordinate between: a broker that survived a restart would be a
// message queue nobody asked to run.
package events

import (
	"sync"
)

// Kind says what changed.
type Kind string

const (
	// KindListChanged means a List or something on it changed: an Item was added,
	// ticked, renamed or removed.
	KindListChanged Kind = "list.changed"
	// KindListsChanged means which Lists a Member can reach changed.
	KindListsChanged Kind = "lists.changed"
	// KindActivity means something arrived in a Member's Activity panel.
	KindActivity Kind = "activity"
	// KindPresence means who is looking at a List changed.
	KindPresence Kind = "presence"
	// KindMemberChanged means the signed-in Member's own account changed under them,
	// which nothing else would tell their browser: it reads itself once and keeps the
	// answer for the life of the tab.
	KindMemberChanged Kind = "member.changed"
)

// Event is one thing that happened.
//
// It says what changed rather than what it changed to: a watcher re-reads through the
// same API it always uses, so a live update can never show somebody something the API
// would not have given them.
type Event struct {
	Kind Kind
	// ListUID is the List it concerns, when it concerns one.
	ListUID string
	// Watchers is who is looking at that List, for a presence event.
	Watchers []string
}

// Subscription is one Member's open connection.
type Subscription struct {
	// Events is closed when the subscription ends.
	Events <-chan Event
	// Close ends it. Calling it twice is safe.
	Close func()
}

// queueDepth is how far behind a watcher may fall before events are dropped.
//
// A browser that has stopped reading is a browser that has gone away; holding events
// for it would grow without bound. Dropping is safe because an event says only that
// something changed, and the next one says it again.
const queueDepth = 32

// Broker fans events out to the Members watching for them.
type Broker struct {
	mu      sync.Mutex
	next    int64
	watches map[int64]*watch
}

// watch is one open connection.
type watch struct {
	memberID int64
	// listUID is the List this connection is looking at, for presence. Empty when the
	// Member is somewhere that is not a List.
	listUID string
	name    string
	events  chan Event
	closed  bool
}

func NewBroker() *Broker {
	return &Broker{watches: make(map[int64]*watch)}
}

// WatchParams is who is watching, and what they are looking at.
type WatchParams struct {
	MemberID int64
	// Name is how presence names them to the others: "Jonas is here".
	Name string
	// ListUID is the List on screen, or empty.
	ListUID string
}

// Watch opens a subscription. The caller closes it when the connection ends.
func (b *Broker) Watch(params WatchParams) Subscription {
	b.mu.Lock()
	b.next++
	id := b.next
	w := &watch{
		memberID: params.MemberID,
		listUID:  params.ListUID,
		name:     params.Name,
		events:   make(chan Event, queueDepth),
	}
	b.watches[id] = w
	b.mu.Unlock()

	b.announcePresence(params.ListUID)

	var once sync.Once
	return Subscription{
		Events: w.events,
		Close: func() {
			once.Do(func() { b.drop(id) })
		},
	}
}

// drop ends one subscription and tells the List it was watching.
func (b *Broker) drop(id int64) {
	b.mu.Lock()
	w, ok := b.watches[id]
	if ok {
		delete(b.watches, id)
		w.closed = true
		close(w.events)
	}
	b.mu.Unlock()

	if ok && w.listUID != "" {
		b.announcePresence(w.listUID)
	}
}

// Publish sends an event to the Members named.
//
// The audience is worked out by the caller, which is the only place that knows who may
// see a List. The broker never decides who can see what.
func (b *Broker) Publish(event Event, audience []int64) {
	reaches := make(map[int64]bool, len(audience))
	for _, id := range audience {
		reaches[id] = true
	}

	b.mu.Lock()
	defer b.mu.Unlock()
	for _, w := range b.watches {
		if !reaches[w.memberID] {
			continue
		}
		send(w, event)
	}
}

// WatchersOf is who is looking at a List, by name, without repeats.
//
// A Member with two tabs open is one person standing there.
func (b *Broker) WatchersOf(listUID string) []string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return namesExcept(b.presentOn(listUID), 0)
}

// person is one Member standing on a List, counted once however many tabs they have.
type person struct {
	memberID int64
	name     string
}

// presentOn is who is on a List, each of them once.
func (b *Broker) presentOn(listUID string) []person {
	seen := make(map[int64]bool)
	here := make([]person, 0, len(b.watches))
	for _, w := range b.watches {
		if w.listUID != listUID || seen[w.memberID] {
			continue
		}
		seen[w.memberID] = true
		here = append(here, person{memberID: w.memberID, name: w.name})
	}
	return here
}

// namesExcept is those people by name, leaving one of them out.
//
// Presence answers "who else is here", so the person asking is never in the answer: a
// Member does not need telling that they are reading the List they are reading. Nobody
// has member 0, so that leaves everybody in.
func namesExcept(here []person, except int64) []string {
	names := make([]string, 0, len(here))
	for _, one := range here {
		if one.memberID == except {
			continue
		}
		names = append(names, one.name)
	}
	return names
}

// announcePresence tells everyone on a List who else is now standing there.
func (b *Broker) announcePresence(listUID string) {
	if listUID == "" {
		return
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	// Gathered once, not once per watcher. Each watcher is told a different answer,
	// because each of them is the one person their own answer leaves out. That is one
	// name dropped from a list they all share, and walking every open watch again for
	// each of them made joining a crowded List cost the square of the crowd. See
	// BenchmarkJoiningACrowdedList.
	here := b.presentOn(listUID)

	for _, w := range b.watches {
		if w.listUID != listUID {
			continue
		}
		send(w, Event{
			Kind:     KindPresence,
			ListUID:  listUID,
			Watchers: namesExcept(here, w.memberID),
		})
	}
}

// send hands an event to one watcher, or drops it if they are not keeping up.
func send(w *watch, event Event) {
	if w.closed {
		return
	}
	select {
	case w.events <- event:
	default:
		// Behind by a whole queue: the next event will say the same thing.
	}
}
