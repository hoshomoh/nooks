package events

import (
	"testing"
	"time"
)

// waitFor reads the next event, or fails rather than hanging the suite.
func waitFor(t *testing.T, sub Subscription) Event {
	t.Helper()
	select {
	case event, ok := <-sub.Events:
		if !ok {
			t.Fatal("the subscription closed before the event arrived")
		}
		return event
	case <-time.After(time.Second):
		t.Fatal("no event arrived")
		return Event{}
	}
}

// drain reads whatever is already queued, so a test can assert on what comes next.
func drain(sub Subscription) {
	for {
		select {
		case <-sub.Events:
		default:
			return
		}
	}
}

func TestAnEventReachesTheMembersItNames(t *testing.T) {
	broker := NewBroker()
	anna := broker.Watch(WatchParams{MemberID: 1, Name: "Anna"})
	defer anna.Close()

	broker.Publish(Event{Kind: KindListChanged, ListUID: "list_groceries"}, []int64{1})

	if got := waitFor(t, anna).Kind; got != KindListChanged {
		t.Errorf("kind = %q", got)
	}
}

// The broker never decides who may see a List; it sends to exactly who it was told.
func TestAnEventReachesNobodyElse(t *testing.T) {
	broker := NewBroker()
	anna := broker.Watch(WatchParams{MemberID: 1, Name: "Anna"})
	defer anna.Close()
	jonas := broker.Watch(WatchParams{MemberID: 2, Name: "Jonas"})
	defer jonas.Close()

	broker.Publish(Event{Kind: KindListChanged, ListUID: "list_groceries"}, []int64{1})

	select {
	case event := <-jonas.Events:
		t.Errorf("Jonas received %v, and was not in the audience", event)
	case <-time.After(50 * time.Millisecond):
	}
}

func TestArrivingOnAListTellsTheOthers(t *testing.T) {
	broker := NewBroker()
	anna := broker.Watch(WatchParams{MemberID: 1, Name: "Anna", ListUID: "list_groceries"})
	defer anna.Close()
	drain(anna)

	jonas := broker.Watch(WatchParams{MemberID: 2, Name: "Jonas", ListUID: "list_groceries"})
	defer jonas.Close()

	event := waitFor(t, anna)
	if event.Kind != KindPresence {
		t.Fatalf("kind = %q, want presence", event.Kind)
	}
	// Anna is not told that Anna is here.
	if len(event.Watchers) != 1 || event.Watchers[0] != "Jonas" {
		t.Errorf("watchers = %v, want only Jonas", event.Watchers)
	}
}

func TestLeavingAListTellsTheOthers(t *testing.T) {
	broker := NewBroker()
	anna := broker.Watch(WatchParams{MemberID: 1, Name: "Anna", ListUID: "list_groceries"})
	defer anna.Close()
	jonas := broker.Watch(WatchParams{MemberID: 2, Name: "Jonas", ListUID: "list_groceries"})
	drain(anna)

	jonas.Close()

	event := waitFor(t, anna)
	if len(event.Watchers) != 0 {
		t.Errorf("watchers = %v, want nobody left", event.Watchers)
	}
}

// A Member with two tabs open is one person standing there.
func TestTwoTabsAreOnePerson(t *testing.T) {
	broker := NewBroker()
	first := broker.Watch(WatchParams{MemberID: 1, Name: "Anna", ListUID: "list_groceries"})
	defer first.Close()
	second := broker.Watch(WatchParams{MemberID: 1, Name: "Anna", ListUID: "list_groceries"})
	defer second.Close()

	if got := broker.WatchersOf("list_groceries"); len(got) != 1 {
		t.Errorf("watchers = %v, want one Anna", got)
	}
}

// Presence is per List: somebody reading a different one is not here.
func TestPresenceIsPerList(t *testing.T) {
	broker := NewBroker()
	anna := broker.Watch(WatchParams{MemberID: 1, Name: "Anna", ListUID: "list_groceries"})
	defer anna.Close()
	elsewhere := broker.Watch(WatchParams{MemberID: 2, Name: "Jonas", ListUID: "list_bike"})
	defer elsewhere.Close()

	if got := broker.WatchersOf("list_groceries"); len(got) != 1 {
		t.Errorf("watchers = %v, want only Anna", got)
	}
}

func TestClosingTwiceIsSafe(t *testing.T) {
	broker := NewBroker()
	sub := broker.Watch(WatchParams{MemberID: 1, Name: "Anna", ListUID: "list_groceries"})

	sub.Close()
	sub.Close()
}

// A browser that has stopped reading is a browser that has gone away.
func TestAWatcherWhoIsNotKeepingUpIsNotWaitedFor(t *testing.T) {
	broker := NewBroker()
	sub := broker.Watch(WatchParams{MemberID: 1, Name: "Anna"})
	defer sub.Close()

	for range queueDepth * 2 {
		broker.Publish(Event{Kind: KindListChanged, ListUID: "list_groceries"}, []int64{1})
	}

	if got := len(sub.Events); got != queueDepth {
		t.Errorf("queued = %d, want the queue to have stopped at %d", got, queueDepth)
	}
}

/*
Losing a List ends the stream watching it.

Presence is the one thing a stream keeps sending after access is gone. Publish works out
its audience at the moment of the event, so somebody unshared stops being told the List
changed. announcePresence sends to every open watch on that uid, and a watch is checked
when it opens and never again, so somebody unshared mid-stream kept learning who else was
reading and kept appearing to the others as present.

Only the ones named, and only on that List: a Member watching something else keeps what
they have.
*/
func TestLosingAListEndsTheStreamWatchingIt(t *testing.T) {
	broker := NewBroker()

	unshared := broker.Watch(WatchParams{MemberID: 1, Name: "Anna", ListUID: "list_shop"})
	staying := broker.Watch(WatchParams{MemberID: 2, Name: "Jonas", ListUID: "list_shop"})
	elsewhere := broker.Watch(WatchParams{MemberID: 1, Name: "Anna", ListUID: "list_jobs"})
	t.Cleanup(staying.Close)
	t.Cleanup(elsewhere.Close)

	broker.EndWatchesOn("list_shop", []int64{1})

	if _, open := <-unshared.Events; open {
		// Drained until closed: the presence events already queued come first.
		for range unshared.Events { //nolint:revive // draining is the assertion
		}
	}

	if watching := broker.WatchersOf("list_shop"); len(watching) != 1 || watching[0] != "Jonas" {
		t.Errorf("watching = %v, want only Jonas left", watching)
	}
	if watching := broker.WatchersOf("list_jobs"); len(watching) != 1 {
		t.Errorf("her stream on another List was ended too: %v", watching)
	}
}

/*
One Note has one editor, and the claim goes when the connection does.

The app saves a Note every 800ms and sends the whole thing, so two people in one Note
means the second save goes over the first with neither of them told. Preventing that was
chosen over reconciling it afterwards, which means somebody has to hold the Note and
everybody else has to be told who.

The first to open it keeps it. The claim lives on the watch rather than in a table of
its own, which is what makes a closed laptop harmless: the thing holding the Note is the
connection, and that already ends by itself.

Nobody is told they are editing their own Note, because a screen that told you somebody
was in the Note you are typing into would be telling you to stop.
*/
func TestOneNoteHasOneEditor(t *testing.T) {
	broker := NewBroker()

	first := broker.Watch(WatchParams{
		MemberID: 1, Name: "Anna", ListUID: "list_shop", EditingUID: "item_bread",
	})
	second := broker.Watch(WatchParams{
		MemberID: 2, Name: "Jonas", ListUID: "list_shop", EditingUID: "item_bread",
	})
	t.Cleanup(second.Close)

	// Jonas is told Anna has it, and Anna is told about nobody.
	if got := editorsIn(t, second.Events); len(got) != 1 || got[0].Name != "Anna" {
		t.Errorf("Jonas was told %v, want Anna holding item_bread", got)
	}
	if got := editorsIn(t, first.Events); len(got) != 0 {
		t.Errorf("Anna was told %v about a Note she has open herself", got)
	}

	// Anna closes her laptop. The Note is free, and Jonas is told so.
	first.Close()
	if got := editorsIn(t, second.Events); len(got) != 0 {
		t.Errorf("after Anna's stream ended Jonas was still told %v", got)
	}
}

// editorsIn drains what a watcher has been sent and answers with the last word on who
// is editing, since a claim arriving produces an announcement to everybody on the List.
func editorsIn(t *testing.T, events <-chan Event) []Editor {
	t.Helper()

	var last []Editor
	for {
		select {
		case event, open := <-events:
			if !open {
				return last
			}
			if event.Kind == KindPresence {
				last = event.Editing
			}
		default:
			return last
		}
	}
}
