package v1

import (
	"context"
	"errors"
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
}

// NewPublicService builds the service.
func NewPublicService(s store.Store) *PublicService {
	return &PublicService{store: s}
}

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

	items, err := s.store.ItemsOnList(ctx, list.ID)
	if err != nil {
		return nil, internalError("read items", err)
	}

	names, err := s.contributorNames(ctx, settings.Public, items)
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

	names := map[int64]string{}
	for _, item := range items {
		id := item.AddedByID
		if id == 0 || names[id] != "" {
			continue
		}
		member, err := s.store.MemberByID(ctx, id)
		if err != nil {
			if errors.Is(err, store.ErrNotFound) {
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
