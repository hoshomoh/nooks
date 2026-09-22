package events

import (
	"fmt"
	"testing"
)

/*
What one more watcher costs when others are already on the same List.

announcePresence sends a different answer to each watcher, because each of them is the
one person their own answer leaves out. It builds that answer by scanning every open
watch, once per watcher on the List, so the work is the number watching times the
number open — and all of it happens with the broker's one mutex held, which every
Publish in the Instance waits behind.

It used to gather that answer once per watcher, walking every open watch each time.
Measured 2026-09-22 on this machine, before and after gathering it once: 208us to 112us
at ten watching, 37.9ms to 4.7ms at a hundred, 991ms to 58ms at five hundred. All of
that time is spent holding the broker's one mutex, so at five hundred it was most of a
second in which no Publish anywhere in the Instance could proceed.

It is still superlinear, and inherently so: each of n watchers is sent its own list of
the other n names. What went was walking every watch again for each of them.

None of these sizes is reached by a household. They are reachable by one signed-in
Member opening connections, because nothing bounds how many streams one account may
hold. The defense log carries that as the open question, and this is the evidence
under it.
*/
func BenchmarkJoiningACrowdedList(b *testing.B) {
	for _, crowd := range []int{10, 50, 100, 250, 500} {
		b.Run(fmt.Sprint(crowd), func(b *testing.B) {
			broker := NewBroker()
			for n := 0; n < crowd; n++ {
				broker.Watch(WatchParams{MemberID: int64(n), Name: "A", ListUID: "list_1"})
			}
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				sub := broker.Watch(WatchParams{MemberID: 99999, Name: "Z", ListUID: "list_1"})
				sub.Close()
			}
		})
	}
}

// What a publish to everybody costs at the same sizes, which is the shape presence
// would have if it did not personalise: one scan rather than one per watcher.
func BenchmarkPublishToACrowd(b *testing.B) {
	for _, crowd := range []int{10, 50, 100, 250, 500} {
		b.Run(fmt.Sprint(crowd), func(b *testing.B) {
			broker := NewBroker()
			audience := make([]int64, 0, crowd)
			for n := 0; n < crowd; n++ {
				broker.Watch(WatchParams{MemberID: int64(n), Name: "A", ListUID: "list_1"})
				audience = append(audience, int64(n))
			}
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				broker.Publish(Event{Kind: KindListChanged, ListUID: "list_1"}, audience)
			}
		})
	}
}
