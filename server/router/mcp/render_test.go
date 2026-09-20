package mcp

import (
	"strings"
	"testing"

	apiv1 "github.com/hoshomoh/nooks/proto/gen/nooks/api/v1"
)

// answered is one page of Lists as the service sends it back.
func answered(shown, total, page int, atLeast bool) *apiv1.ListListsResponse {
	return &apiv1.ListListsResponse{
		Lists:    make([]*apiv1.List, shown),
		Total:    int32(total),
		Page:     int32(page),
		PageSize: 25,
		AtLeast:  atLeast,
	}
}

/*
An assistant handed twenty-five rows and no word about the rest will answer "you have
twenty-five lists", and be wrong. This is the line that stops it.
*/
func TestAnAssistantIsToldWhenThereIsMore(t *testing.T) {
	for _, one := range []struct {
		what string
		res  *apiv1.ListListsResponse
		want string
	}{
		{
			what: "a full page with more behind it",
			res:  answered(25, 60, 1, false),
			want: "Showing 1-25 of 60. Ask for page 2 for more.",
		},
		{
			what: "the last page",
			res:  answered(10, 60, 3, false),
			want: "Showing 51-60 of 60.",
		},
		{
			what: "more than anyone counted",
			res:  answered(25, 1000, 1, true),
			want: "Showing 1-25 of at least 1000. Ask for page 2 for more.",
		},
	} {
		if got := pageLine(one.res); got != one.want {
			t.Errorf("%s: %q, want %q", one.what, got, one.want)
		}
	}
}

// A household with eight Lists is not told it is looking at 1-8 of 8.
func TestOnePageOfListsSaysNothingAboutPages(t *testing.T) {
	if got := pageLine(answered(8, 8, 1, false)); got != "" {
		t.Errorf("said %q, want nothing", got)
	}
	if got := pageLine(answered(0, 0, 1, false)); got != "" {
		t.Errorf("said %q for no Lists at all, want nothing", got)
	}
}

/*
A row that shortens a Note has to say so.

update_item replaces a Note rather than adding to it, so an assistant that read one line
of five and wrote back what it thought the Note was would delete the other four. The
count is what tells it to go and read the rest first.
*/
func TestAShortenedNoteSaysHowMuchIsMissing(t *testing.T) {
	item := &apiv1.Item{
		Uid:   "item_coffee",
		Label: "Coffee",
		Note:  "### Where\nSaturday market.\nThey pack up around two.",
	}

	said := itemLine(item)
	if !strings.Contains(said, "### Where") {
		t.Errorf("row = %q, want the Member's own first line", said)
	}
	if !strings.Contains(said, "+2 more lines") {
		t.Errorf("row = %q, want it to say how much it held back", said)
	}
	if !strings.Contains(said, "get_item") {
		t.Errorf("row = %q, want it to say where the rest is", said)
	}
}

// A Note that fits says nothing about lines, because there are none missing.
func TestANoteThatFitsIsLeftAlone(t *testing.T) {
	item := &apiv1.Item{Uid: "item_coffee", Label: "Coffee", Note: "Saturday market."}

	said := itemLine(item)
	if !strings.Contains(said, "note: Saturday market.") {
		t.Errorf("row = %q, want the Note as written", said)
	}
	if strings.Contains(said, "more lines") {
		t.Errorf("row = %q, want nothing about missing lines", said)
	}
}

/*
A search result an assistant cannot act on is a search result it cannot use.

Every other row ends with the identifier a later call needs. These did not carry one at
all, so finding something was as far as an assistant could get.
*/
func TestASearchResultCarriesWhatToCallNext(t *testing.T) {
	hit := &apiv1.SearchHit{
		Kind:     apiv1.SearchHitKind_SEARCH_HIT_KIND_ITEM,
		ListUid:  "list_groceries",
		ItemUid:  "item_coffee",
		Text:     "Coffee",
		ListName: "Groceries",
	}

	said := hitLine(hit)
	for _, want := range []string{"item", "Coffee", "Groceries", "item_coffee", "list_groceries"} {
		if !strings.Contains(said, want) {
			t.Errorf("hit = %q, want it to carry %q", said, want)
		}
	}
}

// A List matched by its own name has no Item to name, and still says where it is.
func TestAListHitNamesTheList(t *testing.T) {
	hit := &apiv1.SearchHit{
		Kind:     apiv1.SearchHitKind_SEARCH_HIT_KIND_LIST,
		ListUid:  "list_groceries",
		Text:     "Groceries",
		ListName: "Groceries",
	}

	said := hitLine(hit)
	if !strings.Contains(said, "[list]") {
		t.Errorf("hit = %q, want it to say what was found", said)
	}
	if !strings.Contains(said, "list_groceries") {
		t.Errorf("hit = %q, want the List's identifier", said)
	}
}

// A Note is indexed as the whole markdown a Member wrote, so a hit inside one arrives
// here as the lot. A row is a row.
func TestANoteHitIsKeptToARow(t *testing.T) {
	hit := &apiv1.SearchHit{
		Kind:     apiv1.SearchHitKind_SEARCH_HIT_KIND_NOTE,
		ListUid:  "list_groceries",
		ItemUid:  "item_coffee",
		Text:     "### Where\nSaturday market.\nThey pack up around two.",
		ListName: "Groceries",
	}

	said := hitLine(hit)
	if strings.Count(said, "\n") != 0 {
		t.Errorf("hit = %q, want one row", said)
	}
	if !strings.Contains(said, "+2 more lines") {
		t.Errorf("hit = %q, want it to say what it held back", said)
	}
}

/*
get_item is where a whole Note lives, so it gives one.

The row first, so a reader that learned the shape from get_list meets the same one, and
then what only this call has.
*/
func TestGetItemGivesTheNoteWhole(t *testing.T) {
	note := "### Where\nSaturday market.\nThey pack up around two."
	said := itemWhole(
		&apiv1.Item{Uid: "item_coffee", Label: "Coffee", Note: note, AddedByName: "Anna"},
		&apiv1.List{Uid: "list_groceries", Name: "Groceries"},
	)

	if !strings.Contains(said, note) {
		t.Errorf("get_item said %q, want the Note byte for byte", said)
	}
	if !strings.Contains(said, "Groceries") || !strings.Contains(said, "list_groceries") {
		t.Errorf("get_item said %q, want the List it is on", said)
	}
	if !strings.Contains(said, "Anna") {
		t.Errorf("get_item said %q, want who put it there", said)
	}
}

/*
A List row says what changes what a caller may do next.

Archived says why it was missing from an unfiltered listing, and read-only says why a
write is about to be refused. Both were left out, so an assistant met the refusal with
no way to have expected it.
*/
func TestAListRowSaysWhatWouldSurpriseACaller(t *testing.T) {
	said := listLine(&apiv1.List{
		Uid: "list_move", Name: "Move", OpenCount: 2, DoneCount: 5,
		Sharing: apiv1.Sharing_SHARING_INSTANCE, CanEdit: false,
		ArchivedAt: "2026-08-04T10:00:00Z", ArchivedByName: "Anna",
	})

	for _, want := range []string{"2 open", "5 done", "shared", "read-only", "archived by Anna"} {
		if !strings.Contains(said, want) {
			t.Errorf("row = %q, want it to say %q", said, want)
		}
	}
}

// A plain private List says none of it, because none of it is true.
func TestAPlainListSaysNothingExtra(t *testing.T) {
	said := listLine(&apiv1.List{
		Uid: "list_groceries", Name: "Groceries", OpenCount: 7,
		Sharing: apiv1.Sharing_SHARING_PRIVATE, CanEdit: true,
	})

	if said != "Groceries (list_groceries) — 7 open" {
		t.Errorf("row = %q, want only what is true of it", said)
	}
}

// Unread is the only reason to read the panel twice.
func TestAnActivityRowSaysWhetherItIsNews(t *testing.T) {
	unread := activityLine(&apiv1.Activity{
		CreatedAt: "2026-09-20T09:00:00Z", Text: "Anna shared Groceries",
		TargetUid: "list_groceries", Unread: true,
	})
	if !strings.Contains(unread, "•") || !strings.Contains(unread, "list_groceries") {
		t.Errorf("entry = %q, want it marked unread and pointing somewhere", unread)
	}

	read := activityLine(&apiv1.Activity{CreatedAt: "2026-09-20T09:00:00Z", Text: "Old news"})
	if strings.Contains(read, "•") {
		t.Errorf("entry = %q, want nothing marking a read one", read)
	}
}

// %v on a slice prints Go's own syntax, which is a reader being shown the language the
// server happens to be written in.
func TestIdentifiersAreWrittenOutPlainly(t *testing.T) {
	if got := listed([]string{"mem_anna", "mem_jonas"}); got != "mem_anna, mem_jonas" {
		t.Errorf("listed = %q", got)
	}
	if got := listed(nil); got != "none" {
		t.Errorf("listed nothing = %q, want a word rather than a blank", got)
	}
}
