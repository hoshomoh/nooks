package v1

import (
	"context"
	"errors"
	"sync"
	"time"

	"connectrpc.com/connect"

	apiv1 "github.com/hoshomoh/nooks/proto/gen/nooks/api/v1"
	"github.com/hoshomoh/nooks/store"
)

// PublicService is the one page an Instance can put in front of somebody with no
// account.
//
// Everything here answers anonymously, which is why it is a service of its own: a
// method that needs no session should be impossible to add by accident.
type PublicService struct {
	store store.Store
	now   func() time.Time
	page  publicPage
}

func NewPublicService(s store.Store, now func() time.Time) *PublicService {
	if now == nil {
		now = time.Now
	}
	return &PublicService{store: s, now: now}
}

/*
publicPage is the rows the page was last built from, and what they were built for.

This is the one endpoint anybody on the internet reaches, with no session, no rate limit
and nothing bounding how much is on the List: reading it is 44ms at a thousand Items,
653ms at ten thousand and 6.6s at a hundred thousand. Paging the answer would bound one
response and do nothing about the same caller asking a thousand times, which is the
shape of the exposure and is cheap for them and linear for whoever is hosting.

The Items are kept rather than the built answer. A proto message written once and handed
to several requests at once is shared mutable state the moment anything marshals it, and
building the rows again from a slice already in memory is not what costs anything here.

Two things end an entry, because one of them cannot see everything. The version is
exact for the List and the settings, and since a List now records that it changed when
anything on it did, an add or a tick is reflected on the next request rather than
whenever a timer says so. The age is the backstop for what the version cannot see: a
Member renaming themselves changes what the page says and touches no List.
*/
type publicPage struct {
	mu    sync.Mutex
	from  publicVersion
	at    time.Time
	items []store.Item
	names map[int64]string
}

// publicVersion is everything the page is built from that can be read cheaply. Every
// field is compared, so adding one to the page means adding it here or serving it stale.
type publicVersion struct {
	listUID      string
	changedAt    time.Time
	instanceName string
	public       store.PublicList
}

// pageMaxAge bounds how long a Member's own rename can be missing from the page. Short
// enough that nobody notices, long enough that a thousand requests a second become one
// read a minute.
const pageMaxAge = time.Minute

// GetPublicList returns the published List, or nothing when there is none.
//
// It is the only List reachable this way. A Visitor cannot ask for another, and an
// Instance with nothing published answers the same way to everybody — there is no
// version of this that reveals which Lists exist.
func (s *PublicService) GetPublicList(
	ctx context.Context,
	_ *connect.Request[apiv1.GetPublicListRequest],
) (*connect.Response[apiv1.GetPublicListResponse], error) {
	settings, err := s.store.InstanceSettings(ctx)
	if err != nil {
		return nil, internalError("read instance settings", err)
	}
	if !settings.Public.IsPublished() {
		return connect.NewResponse(&apiv1.GetPublicListResponse{}), nil
	}

	list, err := s.store.ListByUID(ctx, settings.Public.ListUID)
	if err != nil {
		// The published List was deleted. The page says nothing is published rather
		// than an error: a Visitor cannot do anything about it either way.
		if errors.Is(err, store.ErrNotFound) {
			return connect.NewResponse(&apiv1.GetPublicListResponse{}), nil
		}
		return nil, internalError("read list", err)
	}

	items, names, err := s.rowsFor(ctx, list, settings)
	if err != nil {
		return nil, err
	}

	out := make([]*apiv1.PublicItem, 0, len(items))
	open := 0
	for _, item := range items {
		if !item.Done() {
			open++
		}
		out = append(out, publicItemOf(item, settings.Public, names))
	}

	return connect.NewResponse(&apiv1.GetPublicListResponse{
		Published:    true,
		InstanceName: settings.Name,
		ListName:     list.Name,
		Items:        out,
		AllowJoin:    settings.Public.AllowJoin,
		OpenCount:    int32(open),
		UpdatedAt:    lastChangedAt(list, items),
	}), nil
}

/*
rowsFor is what the page is made of, read again only when it could have changed.

Held under one lock rather than a read-write pair. The work being protected is a map
lookup and a comparison; the read it avoids is the whole List. Two Visitors arriving
together on a cold page both read, which is the ordinary cost of not holding a lock
across a database call, and is the right way round on the endpoint that must never let
one slow read hold up another request.
*/
func (s *PublicService) rowsFor(
	ctx context.Context,
	list store.List,
	settings store.InstanceSettings,
) ([]store.Item, map[int64]string, error) {
	want := publicVersion{
		listUID:      list.UID,
		changedAt:    list.UpdatedAt,
		instanceName: settings.Name,
		public:       settings.Public,
	}

	now := s.now()
	s.page.mu.Lock()
	if s.page.from == want && s.page.items != nil && now.Sub(s.page.at) < pageMaxAge {
		items, names := s.page.items, s.page.names
		s.page.mu.Unlock()
		return items, names, nil
	}
	s.page.mu.Unlock()

	items, err := s.store.ItemsOnList(ctx, list.ID)
	if err != nil {
		return nil, nil, internalError("read items", err)
	}
	names, err := s.contributorNames(ctx, settings.Public, items)
	if err != nil {
		return nil, nil, err
	}

	s.page.mu.Lock()
	s.page.from, s.page.at, s.page.items, s.page.names = want, now, items, names
	s.page.mu.Unlock()

	return items, names, nil
}

/*
lastChangedAt is when the page last had something to say.

The List's own timestamp moves when it is renamed or reshared, which a Visitor cannot
see and does not care about. What they came for is whether anything on the list changed,
so the newest Item wins where there is one.
*/
func lastChangedAt(list store.List, items []store.Item) string {
	newest := list.UpdatedAt
	for _, item := range items {
		if item.UpdatedAt.After(newest) {
			newest = item.UpdatedAt
		}
	}
	if newest.IsZero() {
		return ""
	}
	return newest.Format(time.RFC3339)
}

// contributorNames looks up who added what, but only when the page shows it.
//
// Reading them when they will not be rendered would put them in the response for
// anybody watching the wire, which is the leak the setting exists to prevent.
func (s *PublicService) contributorNames(
	ctx context.Context,
	public store.PublicList,
	items []store.Item,
) (map[int64]string, error) {
	if !public.ShowNames {
		return nil, nil
	}

	// Asked about once, by whether the answer is here rather than by whether it says
	// anything. Removing a Member empties their name and keeps the row, so a tombstone
	// answers with "" and a check for a non-empty name never remembers them: every Item
	// a removed person added would read them again, on the one endpoint a stranger can
	// reach, over a List with no bound on how many Items it holds.
	names := map[int64]string{}
	for _, item := range items {
		id := item.AddedByID
		if _, known := names[id]; id == 0 || known {
			continue
		}
		member, err := s.store.MemberByID(ctx, id)
		if err != nil {
			if errors.Is(err, store.ErrNotFound) {
				// Remembered as nameless rather than skipped, so a row nothing can name
				// is read once and not once per Item.
				names[id] = ""
				continue
			}
			return nil, internalError("read member", err)
		}
		names[id] = member.Name
	}
	return names, nil
}

// publicItemOf shows one Item the way the Instance has chosen to show it.
func publicItemOf(
	item store.Item,
	public store.PublicList,
	names map[int64]string,
) *apiv1.PublicItem {
	out := &apiv1.PublicItem{Uid: item.UID, Label: item.Label, Done: item.Done()}
	if public.ShowMeta {
		out.Quantity = item.Quantity
		out.DueOn = item.DueOn
	}
	if public.ShowNames {
		out.AddedByName = names[item.AddedByID]
	}
	return out
}
