// Package events carries changes from the Member who made them to the Members watching.
//
// Everything here is in memory and per process. Nooks is one binary on one machine, so
// there is nothing to coordinate between: a broker that survived a restart would be a
// message queue nobody asked to run.
package events

import (
	"slices"
	"strings"
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
	// Editing is who has a Note open for editing on that List, by Item. Everybody else
	// reads that Note rather than writing it, which is how two people are stopped from
	// losing each other's words.
	Editing []Editor
}

// Editor is one person with one Note open.
type Editor struct {
	ItemUID string
	Name    string
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

/*
watchesPerMember is how many streams one Member may have open at once.

A stream is a goroutine, a buffered channel and a row the presence loop walks, so
without a cap one Member with a script decides how much of all three the Instance
spends. It matters more since one Note gained one editor: a stream that lingers holds a
Note read-only for everybody else, and this is what bounds how long that can last.

Ten covers a laptop, a phone, a tablet and spare tabs with room over. Past ten the
oldest is closed rather than the newest refused, because a refused connection leaves the
tab somebody has just opened silently without live updates, which is far harder to
notice than a reconnect.
*/
const watchesPerMember = 10

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
	// editingUID is the Item whose Note this connection has open, or empty. Held on the
	// watch rather than anywhere else, so it is released by the stream ending: a closed
	// laptop cannot hold a Note for ever, because the thing holding it is the
	// connection.
	editingUID string
	events     chan Event
	closed     bool
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
	// EditingUID is the Item whose Note is open for editing, or empty.
	EditingUID string
}

// Watch opens a subscription. The caller closes it when the connection ends.
func (b *Broker) Watch(params WatchParams) Subscription {
	b.mu.Lock()
	b.next++
	id := b.next
	w := &watch{
		memberID:   params.MemberID,
		listUID:    params.ListUID,
		name:       params.Name,
		editingUID: params.EditingUID,
		events:     make(chan Event, queueDepth),
	}
	b.watches[id] = w
	closed := b.closeOldestOver(params.MemberID)
	b.mu.Unlock()

	// The List this one opened on, plus whatever List each closed stream was watching.
	// A closed stream is one fewer person standing there, and the others are told.
	closed[params.ListUID] = true
	for listUID := range closed {
		b.announcePresence(listUID)
	}

	var once sync.Once
	return Subscription{
		Events: w.events,
		Close: func() {
			once.Do(func() { b.drop(id) })
		},
	}
}

/*
closeOldestOver ends this Member's oldest streams until they are inside the cap, and
reports which Lists those streams were watching.

Under the caller's lock, and in the same one that added the new watch, so the count a
decision is taken on is the count after it. Counting first and closing second would let
through as many as arrive at once, which is the whole case this is for.

Oldest by identifier, which is the order they were opened in: `next` only goes up.

Walks every open watch rather than keeping an index per Member. This runs once when a
connection opens, where announcePresence already walks the same map, and an index would
be a second thing to keep true in drop and EndWatchesOn as well.
*/
func (b *Broker) closeOldestOver(memberID int64) map[string]bool {
	theirs := make([]int64, 0, watchesPerMember+1)
	for id, w := range b.watches {
		if w.memberID == memberID {
			theirs = append(theirs, id)
		}
	}
	if len(theirs) <= watchesPerMember {
		return make(map[string]bool, 1)
	}
	slices.Sort(theirs)

	closed := make(map[string]bool)
	for _, id := range theirs[:len(theirs)-watchesPerMember] {
		w := b.watches[id]
		delete(b.watches, id)
		w.closed = true
		close(w.events)
		if w.listUID != "" {
			closed[w.listUID] = true
		}
	}
	return closed
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

/*
EndWatchesOn closes the streams these Members have open on one List.

Presence is the one thing a stream keeps sending after access is lost. Publish works out
its audience at the moment of the event, so somebody unshared stops being told the List
changed; announcePresence sends to every open watch on that uid, and a watch is checked
when it opens and never again. Somebody unshared while their stream is open kept
learning who else was reading, and kept appearing to the others as present.

Ending the watch rather than re-reading who may see the List on every announce: that
read would sit on the hottest path there is, behind the one mutex every Publish waits
for, to catch a case that only happens when sharing changes. Sharing changing is where
this is called from.

Their connection ends and the browser opens another, which is what it does whenever a
stream drops. The new one is checked on the way in and finds the List gone.
*/
func (b *Broker) EndWatchesOn(listUID string, members []int64) {
	if listUID == "" || len(members) == 0 {
		return
	}
	losing := make(map[int64]bool, len(members))
	for _, id := range members {
		losing[id] = true
	}

	b.mu.Lock()
	ended := false
	for id, w := range b.watches {
		if w.listUID != listUID || !losing[w.memberID] {
			continue
		}
		delete(b.watches, id)
		w.closed = true
		close(w.events)
		ended = true
	}
	b.mu.Unlock()

	// Told after the lock is given up, because announcePresence takes it again. Only
	// when somebody actually went: the others are being told who is left.
	if ended {
		b.announcePresence(listUID)
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

/*
presentOn is who is on a List, each of them once, in a settled order.

Sorted because the answer is drawn. Walking the watches gives whatever order Go
randomised the map into this time, so "Anna and Jonas are here" became "Jonas and Anna
are here" on the next event with nobody having arrived or left, and the names swapped
places on screen. The same fault as the one that let a Note change hands, found beside
it and fixed with it.
*/
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
	slices.SortFunc(here, func(a, b person) int {
		return strings.Compare(a.name, b.name)
	})
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

// holder is somebody with a Note open, kept with their id so the answer each watcher
// gets can leave them out of it, and with the watch that claimed it so that the first
// to arrive is a fact rather than whatever a map was walked in.
type holder struct {
	memberID int64
	watchID  int64
	Editor
}

/*
editingOn is who has a Note open on a List.

One Note has one editor. Two people opening the same one is the case this exists to
stop, and the first to arrive keeps it: whoever else opens it is told somebody is
already there and reads instead.

First by watch identifier, which only goes up, so it is the order they opened in. It
used to be whichever the map was walked to first, which is not an order at all: Go
randomises it per run and per walk, so the Note could change hands between one presence
event and the next with nobody having touched anything. The race detector caught that as
a failing test; on this machine the map had simply been kind.

Sorted before it goes out for the same reason. Two announcements carrying the same
editors in a different order are two different events to whatever draws them.
*/
func (b *Broker) editingOn(listUID string) []holder {
	held := make(map[string]holder)
	for id, w := range b.watches {
		if w.listUID != listUID || w.editingUID == "" {
			continue
		}
		if first, taken := held[w.editingUID]; taken && first.watchID < id {
			continue
		}
		held[w.editingUID] = holder{
			memberID: w.memberID,
			watchID:  id,
			Editor:   Editor{ItemUID: w.editingUID, Name: w.name},
		}
	}

	editors := make([]holder, 0, len(held))
	for _, one := range held {
		editors = append(editors, one)
	}
	slices.SortFunc(editors, func(a, b holder) int {
		return strings.Compare(a.ItemUID, b.ItemUID)
	})
	return editors
}

// editorsExcept is those editors, leaving out the one being told. Somebody does not
// need telling that they have their own Note open.
func editorsExcept(editing []holder, except int64) []Editor {
	editors := make([]Editor, 0, len(editing))
	for _, one := range editing {
		if one.memberID == except {
			continue
		}
		editors = append(editors, one.Editor)
	}
	return editors
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
	editing := b.editingOn(listUID)

	for _, w := range b.watches {
		if w.listUID != listUID {
			continue
		}
		send(w, Event{
			Kind:     KindPresence,
			ListUID:  listUID,
			Watchers: namesExcept(here, w.memberID),
			Editing:  editorsExcept(editing, w.memberID),
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
