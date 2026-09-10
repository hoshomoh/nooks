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
	if len(event.Watchers) != 2 {
		t.Errorf("watchers = %v, want both of them", event.Watchers)
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
	if len(event.Watchers) != 1 || event.Watchers[0] != "Anna" {
		t.Errorf("watchers = %v, want only Anna", event.Watchers)
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
