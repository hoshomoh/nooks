package mcp

import (
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
