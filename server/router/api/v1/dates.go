package v1

import (
	"time"

	"connectrpc.com/connect"
	"errors"
)

// dueOnLayout is the shape of an Item's due date. An Item is due on a day, not at a
// time: nothing in Nooks is scheduled to the minute.
const dueOnLayout = "2006-01-02"

// validDueOn accepts a date or nothing at all.
func validDueOn(dueOn string) error {
	if dueOn == "" {
		return nil
	}
	if _, err := time.Parse(dueOnLayout, dueOn); err != nil {
		return connect.NewError(connect.CodeInvalidArgument,
			errors.New("a due date looks like 2026-08-25"))
	}
	return nil
}
