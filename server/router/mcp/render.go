package mcp

import (
	"fmt"
	"strings"

	apiv1 "github.com/hoshomoh/nooks/proto/gen/nooks/api/v1"
)

/*
How a thing is written down for an assistant to read.

One renderer per kind, shared by every tool that shows one. An Item read from get_list
and the same Item found by search are the same Item, and a reader that has learned to
parse one row should not meet a second shape further down the menu.
*/

// itemLine is one Item: its state, what it says, and the identifier a later call needs.
func itemLine(item *apiv1.Item) string {
	mark := " "
	if item.GetDone() {
		mark = "x"
	}

	var said strings.Builder
	fmt.Fprintf(&said, "[%s] %s", mark, item.GetLabel())
	if quantity := item.GetQuantity(); quantity != "" {
		fmt.Fprintf(&said, " (%s)", quantity)
	}
	if due := item.GetDueOn(); due != "" {
		fmt.Fprintf(&said, " due %s", due)
	}
	if note := item.GetNote(); note != "" {
		fmt.Fprintf(&said, " — note: %s", noteRow(note))
	}
	fmt.Fprintf(&said, " — %s", item.GetUid())
	return said.String()
}

/*
itemWhole is one Item read on its own: the row, then the Note underneath it in full.

The row first, so a reader that has learned to parse one from get_list meets the same
shape here. The Note is what this call is for, so it is given as written rather than cut
to fit beside anything.
*/
func itemWhole(item *apiv1.Item, list *apiv1.List) string {
	var said strings.Builder
	fmt.Fprintf(&said, "%s — on %s (%s)", itemLine(item), list.GetName(), list.GetUid())
	if by := item.GetAddedByName(); by != "" {
		fmt.Fprintf(&said, "\nAdded by %s", by)
		if token := item.GetAddedViaToken(); token != "" {
			fmt.Fprintf(&said, " via %s", token)
		}
	}
	if by := item.GetDoneByName(); by != "" {
		fmt.Fprintf(&said, "\nTicked off by %s", by)
	}
	if note := item.GetNote(); note != "" {
		fmt.Fprintf(&said, "\n\nNote:\n%s", note)
	}
	return said.String()
}

/*
listLine is one List as it appears in a menu of them, and as get_list's first row.

Everything here changes what a caller may do next or where the List can be found, which
is why each of them is worth a word: archived says why it is missing from an unfiltered
listing, read-only says why a write will be refused, and shared says the change will be
seen by somebody else.
*/
func listLine(list *apiv1.List) string {
	var said strings.Builder
	fmt.Fprintf(&said, "%s (%s) — %d open", list.GetName(), list.GetUid(), list.GetOpenCount())
	if done := list.GetDoneCount(); done > 0 {
		fmt.Fprintf(&said, ", %d done", done)
	}
	if list.GetSharing() != apiv1.Sharing_SHARING_PRIVATE {
		said.WriteString(", shared")
	}
	if list.GetIsPinned() {
		said.WriteString(", pinned")
	}
	if !list.GetCanEdit() {
		said.WriteString(", read-only")
	}
	if at := list.GetArchivedAt(); at != "" {
		fmt.Fprintf(&said, ", archived by %s", list.GetArchivedByName())
	}
	return said.String()
}

/*
pageLine says where a page of Lists sits, when there is more than one.

Spelled out rather than left to be inferred: a caller handed twenty-five rows and no
word about the rest will answer "you have twenty-five lists", and be wrong.
*/
func pageLine(res *apiv1.ListListsResponse) string {
	shown := len(res.GetLists())
	size := int(res.GetPageSize())
	page := max(int(res.GetPage()), 1)
	if shown == 0 || (page == 1 && shown < size) {
		return ""
	}

	total := fmt.Sprintf("%d", res.GetTotal())
	if res.GetAtLeast() {
		total = "at least " + total
	}
	first := (page-1)*size + 1

	said := fmt.Sprintf("Showing %d-%d of %s.", first, first+shown-1, total)
	if shown == size {
		said += fmt.Sprintf(" Ask for page %d for more.", page+1)
	}
	return said
}

/*
activityLine is one entry in the panel: when, whether it is news, and what it points at.

Unread is the only reason to read the panel twice, and the thing it points at is how a
caller acts on it, so a row of text and a timestamp was most of an entry missing.
*/
func activityLine(entry *apiv1.Activity) string {
	mark := " "
	if entry.GetUnread() {
		mark = "•"
	}

	said := fmt.Sprintf("[%s] %s — %s", mark, entry.GetCreatedAt(), entry.GetText())
	if target := entry.GetTargetUid(); target != "" {
		said += " — " + target
	}
	return said
}

// memberLine is one person: enough to name them in a sentence, and the identifier
// sharing needs.
func memberLine(member *apiv1.Member) string {
	return fmt.Sprintf("%s <%s> %s — %s",
		member.GetName(), member.GetEmail(), roleWord(member.GetRole()), member.GetUid())
}

// roleWord says a Role in the words the app uses, rather than the enum's.
func roleWord(role apiv1.Role) string {
	if role == apiv1.Role_ROLE_ADMIN {
		return "admin"
	}
	return "member"
}

// listed writes identifiers out plainly. %v on a slice prints Go's own syntax, which is
// a reader being shown the language the server happens to be written in.
func listed(uids []string) string {
	if len(uids) == 0 {
		return "none"
	}
	return strings.Join(uids, ", ")
}

// lines joins rendered rows, falling back to a sentence rather than handing an
// assistant an empty block it has to guess the meaning of.
func lines(rows []string, whenEmpty string) string {
	if len(rows) == 0 {
		return whenEmpty
	}
	return strings.Join(rows, "\n")
}

/*
noteRow keeps a Note to the part that fits on a row, and says when it did.

The part shown is the Member's own first line, because Nooks never summarises what
somebody wrote. Saying how much was held back is the half that matters: update_item
replaces a Note rather than adding to it, so a reader that took the row for the whole
Note would write back a fragment and delete the rest.
*/
func noteRow(note string) string {
	first, rest, cut := strings.Cut(note, "\n")
	if !cut {
		return note
	}
	return fmt.Sprintf("%s (+%d more lines, read with get_item)", first, strings.Count(rest, "\n")+1)
}

/*
hitLine is one search result: what matched, what it is, and what to call next.

An assistant handed the words and no identifier has found something it cannot then act
on. A Note matched inside is cut to a row like any other, for the reason noteRow gives.
*/
func hitLine(hit *apiv1.SearchHit) string {
	var said strings.Builder
	fmt.Fprintf(&said, "[%s] %s", hitWord(hit.GetKind()), noteRow(hit.GetText()))
	fmt.Fprintf(&said, " — on %s (%s)", hit.GetListName(), hit.GetListUid())
	if item := hit.GetItemUid(); item != "" {
		fmt.Fprintf(&said, " — %s", item)
	}
	return said.String()
}

// hitWord says what was found, in the words the app uses for them.
func hitWord(kind apiv1.SearchHitKind) string {
	switch kind {
	case apiv1.SearchHitKind_SEARCH_HIT_KIND_LIST:
		return "list"
	case apiv1.SearchHitKind_SEARCH_HIT_KIND_ITEM:
		return "item"
	case apiv1.SearchHitKind_SEARCH_HIT_KIND_NOTE:
		return "note"
	default:
		return "match"
	}
}
