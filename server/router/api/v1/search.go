package v1

import (
	"context"

	"connectrpc.com/connect"

	apiv1 "github.com/hoshomoh/nooks/proto/gen/nooks/api/v1"
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
	member, err := requireMember(ctx)
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
	reachable, err := s.reachableLists(ctx, member)
	if err != nil {
		return nil, err
	}

	out := make([]*apiv1.SearchHit, 0, len(hits))
	for _, hit := range hits {
		list, ok := reachable[hit.ListID]
		if !ok {
			continue
		}
		out = append(out, searchHitToProto(hit, list))
	}
	return connect.NewResponse(&apiv1.SearchResponse{Hits: out}), nil
}

// reachableLists is every List the Member may at least read, by internal identity.
func (s *ListService) reachableLists(ctx context.Context, member store.Member) (map[int64]store.List, error) {
	lists, err := s.store.ListsForMember(ctx, member.ID)
	if err != nil {
		return nil, internalError("read lists", err)
	}
	shares, err := s.namedSharesFor(ctx, member)
	if err != nil {
		return nil, err
	}

	reachable := make(map[int64]store.List, len(lists))
	for _, list := range lists {
		if accessTo(list, member, shares) >= AccessRead {
			reachable[list.ID] = list
		}
	}
	return reachable, nil
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
