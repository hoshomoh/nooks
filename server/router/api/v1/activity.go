package v1

import (
	"context"
	"time"

	"github.com/hoshomoh/nooks/store"
)

// activityRecorder puts entries in front of the Members who need to see them.
//
// It is a small value rather than a method on one service because two of them record
// activity for the same reasons: a share comes from the List service, a join request
// from the auth service, and both land in the same panel.
type activityRecorder struct {
	store  store.Store
	now    func() time.Time
	newUID func() (string, error)
}

// record puts one entry in front of one Member.
func (r activityRecorder) record(
	ctx context.Context,
	memberID int64,
	kind store.ActivityKind,
	text, targetUID string,
) error {
	uid, err := r.newUID()
	if err != nil {
		return internalError("make an identifier", err)
	}

	_, err = r.store.CreateActivity(ctx, store.CreateActivityParams{
		UID:       uid,
		MemberID:  memberID,
		Kind:      kind,
		Text:      text,
		TargetUID: targetUID,
		At:        r.now(),
	})
	if err != nil {
		return internalError("record activity", err)
	}
	return nil
}

// tellAdmins puts one entry in front of every Admin.
//
// Requests go to all of them, so that somebody asking to join does not depend on one
// person being awake.
func (r activityRecorder) tellAdmins(
	ctx context.Context,
	kind store.ActivityKind,
	text, targetUID string,
) error {
	ids, err := r.store.AdminIDs(ctx)
	if err != nil {
		return internalError("read admins", err)
	}
	for _, id := range ids {
		if err := r.record(ctx, id, kind, text, targetUID); err != nil {
			return err
		}
	}
	return nil
}
