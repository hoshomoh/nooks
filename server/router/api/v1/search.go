package v1

import (
	"context"

	"connectrpc.com/connect"

	apiv1 "github.com/hoshomoh/nooks/proto/gen/nooks/api/v1"
	"github.com/hoshomoh/nooks/server/auth"
	"github.com/hoshomoh/nooks/store"
)

// Search finds Lists and Items by what a Member wrote in them.
//
// The store ranks and returns hits without regard to who may see them; this decides
// visibility, using the same accessTo every other method uses. Search therefore cannot
// reach further than the Member already could — which is the property that has to hold
// when the same services are reached by an Access token rather than a browser.
func (s *ListService) Search(
	ctx context.Context,
	req *connect.Request[apiv1.SearchRequest],
) (*connect.Response[apiv1.SearchResponse], error) {
	grant, err := requireGrant(ctx)
	if err != nil {
		return nil, err
	}

	hits, err := s.store.Search(ctx, req.Msg.GetQuery())
	if err != nil {
		return nil, internalError("search", err)
	}
	if len(hits) == 0 {
		return connect.NewResponse(&apiv1.SearchResponse{}), nil
	}

	// One read of the Member's Lists answers every hit's "may they see this?".
	reachable, err := s.reachableLists(ctx, grant)
	if err != nil {
		return nil, err
	}

	/*
	 * One result per thing, not per place the word was found.
	 *
	 * An Item is indexed twice — once for its label and once for its Note — so a word
	 * in both came back as two hits with the same UID, which read as two Items that
	 * were really one. The higher-ranked of the two is kept, so whichever part of the
	 * Item actually matched is the part the result shows.
	 */
	seen := make(map[string]bool, len(hits))
	out := make([]*apiv1.SearchHit, 0, len(hits))
	for _, hit := range hits {
		list, ok := reachable[hit.ListID]
		if !ok {
			continue
		}
		if seen[found(hit)] {
			continue
		}
		seen[found(hit)] = true
		out = append(out, searchHitToProto(hit, list))
	}
	return connect.NewResponse(&apiv1.SearchResponse{Hits: out}), nil
}

// reachableLists is every List the caller may at least read, by internal identity.
//
// For answering "may they see this one?" about something already found. Use
// reachableListsInOrder to show them.
func (s *ListService) reachableLists(
	ctx context.Context,
	grant auth.Grant,
) (map[int64]store.List, error) {
	lists, err := s.reachableListsInOrder(ctx, grant)
	if err != nil {
		return nil, err
	}
	reachable := make(map[int64]store.List, len(lists))
	for _, list := range lists {
		reachable[list.ID] = list
	}
	return reachable, nil
}

// reachableListsInOrder is listReach with the service's own store.
func (s *ListService) reachableListsInOrder(
	ctx context.Context,
	grant auth.Grant,
) ([]store.List, error) {
	return listReach(ctx, s.store, grant)
}

// found names the thing a hit is about, so the two ways of finding one Item agree.
func found(hit store.SearchHit) string {
	if hit.Kind == store.KindList {
		return "list:" + hit.UID
	}
	return "item:" + hit.UID
}

// searchHitToProto converts a hit, naming the List it belongs to so a result reads in
// context rather than as a bare line of text.
func searchHitToProto(hit store.SearchHit, list store.List) *apiv1.SearchHit {
	out := &apiv1.SearchHit{
		Kind:     searchKindToProto(hit.Kind),
		ListUid:  list.UID,
		Text:     hit.Text,
		ListName: list.Name,
	}
	if hit.Kind != store.KindList {
		out.ItemUid = hit.UID
	}
	return out
}

func searchKindToProto(kind store.SearchKind) apiv1.SearchHitKind {
	switch kind {
	case store.KindList:
		return apiv1.SearchHitKind_SEARCH_HIT_KIND_LIST
	case store.KindItem:
		return apiv1.SearchHitKind_SEARCH_HIT_KIND_ITEM
	case store.KindNote:
		return apiv1.SearchHitKind_SEARCH_HIT_KIND_NOTE
	default:
		return apiv1.SearchHitKind_SEARCH_HIT_KIND_UNSPECIFIED
	}
}
