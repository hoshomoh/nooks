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
		fmt.Fprintf(&said, " — note: %s", firstLine(note))
	}
	fmt.Fprintf(&said, " — %s", item.GetUid())
	return said.String()
}

// listLine is one List as it appears in a menu of them.
func listLine(list *apiv1.List) string {
	said := fmt.Sprintf("%s (%s) — %d open", list.GetName(), list.GetUid(), list.GetOpenCount())
	if list.GetIsPinned() {
		said += ", pinned"
	}
	if !list.GetCanEdit() {
		said += ", read-only"
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

// lines joins rendered rows, falling back to a sentence rather than handing an
// assistant an empty block it has to guess the meaning of.
func lines(rows []string, whenEmpty string) string {
	if len(rows) == 0 {
		return whenEmpty
	}
	return strings.Join(rows, "\n")
}

// firstLine keeps a Note to the part that fits on a row. Nooks never summarises what a
// Member wrote, so this is their own first line rather than anything generated.
func firstLine(said string) string {
	if cut := strings.IndexByte(said, '\n'); cut >= 0 {
		return said[:cut]
	}
	return said
}
