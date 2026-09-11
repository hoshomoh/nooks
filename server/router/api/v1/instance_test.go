package v1

import (
	"context"
	"errors"
	"testing"
	"time"

	"connectrpc.com/connect"

	apiv1 "github.com/hoshomoh/nooks/proto/gen/nooks/api/v1"
	"github.com/hoshomoh/nooks/store"
)

// fakeStore stands in for a database so that service policy can be tested without one.
//
// It embeds store.Store so that only the methods a test actually exercises need to be
// written; anything else panics with a nil-pointer dereference, which is the right
// outcome for a method the test did not expect to be called.
type fakeStore struct {
	store.Store

	settings store.InstanceSettings
	err      error
}

func (f *fakeStore) InstanceSettings(context.Context) (store.InstanceSettings, error) {
	return f.settings, f.err
}

func TestGetInstanceBeforeFirstRun(t *testing.T) {
	svc := NewInstanceService(&fakeStore{})

	got, err := svc.GetInstance(t.Context(), connect.NewRequest(&apiv1.GetInstanceRequest{}))
	if err != nil {
		t.Fatalf("GetInstance: %v", err)
	}
	if !got.Msg.GetNeedsSetup() {
		t.Error("NeedsSetup = false on a fresh Instance, want true")
	}
	if got.Msg.GetName() != "" {
		t.Errorf("Name = %q on a fresh Instance, want empty", got.Msg.GetName())
	}
	if got.Msg.GetVersion() == "" {
		t.Error("Version is empty, want the running build")
	}
}

func TestGetInstanceAfterFirstRun(t *testing.T) {
	svc := NewInstanceService(&fakeStore{settings: store.InstanceSettings{
		Name:             "Brunnen Street",
		PublicSignup:     true,
		SetupCompletedAt: time.Date(2026, time.March, 3, 0, 0, 0, 0, time.UTC),
	}})

	got, err := svc.GetInstance(t.Context(), connect.NewRequest(&apiv1.GetInstanceRequest{}))
	if err != nil {
		t.Fatalf("GetInstance: %v", err)
	}
	if got.Msg.GetNeedsSetup() {
		t.Error("NeedsSetup = true after first run, want false")
	}
	if got.Msg.GetName() != "Brunnen Street" {
		t.Errorf("Name = %q, want %q", got.Msg.GetName(), "Brunnen Street")
	}
	if !got.Msg.GetPublicSignup() {
		t.Error("PublicSignup = false, want true")
	}
}

func TestGetInstanceReportsStoreFailureAsInternal(t *testing.T) {
	svc := NewInstanceService(&fakeStore{err: errors.New("database is gone")})

	_, err := svc.GetInstance(t.Context(), connect.NewRequest(&apiv1.GetInstanceRequest{}))
	if err == nil {
		t.Fatal("GetInstance succeeded, want an error")
	}
	if got := connect.CodeOf(err); got != connect.CodeInternal {
		t.Errorf("code = %v, want %v", got, connect.CodeInternal)
	}
}

// instanceFixture is an Instance with an Admin and a Member.
func newInstanceSettingsFixture(t *testing.T) listFixture {
	t.Helper()
	f := newListFixture(t)
	settings, err := f.store.InstanceSettings(t.Context())
	if err != nil {
		t.Fatalf("InstanceSettings: %v", err)
	}
	settings.Name = "Brunnen Street"
	if err := f.store.SaveInstanceSettings(t.Context(), settings); err != nil {
		t.Fatalf("SaveInstanceSettings: %v", err)
	}
	return f
}

// Which List is published is not a fact a Visitor gets to read from the outside.
func TestOnlyAnAdminReadsInstanceSettings(t *testing.T) {
	f := newInstanceSettingsFixture(t)
	svc := NewInstanceService(f.store)

	_, err := svc.GetInstanceSettings(f.as(t, f.jonas), connect.NewRequest(
		&apiv1.GetInstanceSettingsRequest{},
	))
	if got := connect.CodeOf(err); got != connect.CodePermissionDenied {
		t.Errorf("code = %v, want permission_denied", got)
	}
}

func TestPublishingAList(t *testing.T) {
	f := newInstanceSettingsFixture(t)
	svc := NewInstanceService(f.store)
	uid := f.createList(t, f.anna, "Groceries")

	res, err := svc.UpdateInstanceSettings(f.as(t, f.anna), connect.NewRequest(
		&apiv1.UpdateInstanceSettingsRequest{
			Settings: &apiv1.InstanceSettings{
				Name: "Brunnen Street",
				PublicList: &apiv1.PublicListSettings{
					ListUid: uid, ShowMeta: true, AllowJoin: true,
				},
			},
		},
	))
	if err != nil {
		t.Fatalf("UpdateInstanceSettings: %v", err)
	}
	if got := res.Msg.GetSettings().GetPublicList().GetListUid(); got != uid {
		t.Errorf("published %q, want %q", got, uid)
	}

	page, err := NewPublicService(f.store).GetPublicList(
		t.Context(), connect.NewRequest(&apiv1.GetPublicListRequest{}),
	)
	if err != nil {
		t.Fatalf("GetPublicList: %v", err)
	}
	if !page.Msg.GetPublished() {
		t.Error("the page is not published after publishing it")
	}
}

// Publishing a List that is not there would answer "nothing here" to everybody, which
// looks exactly like having turned the page off.
func TestPublishingAListThatIsNotThere(t *testing.T) {
	f := newInstanceSettingsFixture(t)
	svc := NewInstanceService(f.store)

	_, err := svc.UpdateInstanceSettings(f.as(t, f.anna), connect.NewRequest(
		&apiv1.UpdateInstanceSettingsRequest{
			Settings: &apiv1.InstanceSettings{
				Name:       "Brunnen Street",
				PublicList: &apiv1.PublicListSettings{ListUid: "list_nope"},
			},
		},
	))
	if got := connect.CodeOf(err); got != connect.CodeNotFound {
		t.Errorf("code = %v, want not_found", got)
	}
}

// An Instance always has a name: the sidebar and every printed sheet carry it.
func TestAnInstanceCannotBeLeftUnnamed(t *testing.T) {
	f := newInstanceSettingsFixture(t)
	svc := NewInstanceService(f.store)

	_, err := svc.UpdateInstanceSettings(f.as(t, f.anna), connect.NewRequest(
		&apiv1.UpdateInstanceSettingsRequest{Settings: &apiv1.InstanceSettings{Name: "  "}},
	))
	if got := connect.CodeOf(err); got != connect.CodeInvalidArgument {
		t.Errorf("code = %v, want invalid_argument", got)
	}
}

// Taking the page down is setting it to nothing, not a second kind of state.
func TestUnpublishing(t *testing.T) {
	f := newInstanceSettingsFixture(t)
	svc := NewInstanceService(f.store)
	uid := f.createList(t, f.anna, "Groceries")

	publish := func(listUID string) {
		t.Helper()
		if _, err := svc.UpdateInstanceSettings(f.as(t, f.anna), connect.NewRequest(
			&apiv1.UpdateInstanceSettingsRequest{
				Settings: &apiv1.InstanceSettings{
					Name:       "Brunnen Street",
					PublicList: &apiv1.PublicListSettings{ListUid: listUID},
				},
			},
		)); err != nil {
			t.Fatalf("UpdateInstanceSettings: %v", err)
		}
	}

	publish(uid)
	publish("")

	page, err := NewPublicService(f.store).GetPublicList(
		t.Context(), connect.NewRequest(&apiv1.GetPublicListRequest{}),
	)
	if err != nil {
		t.Fatalf("GetPublicList: %v", err)
	}
	if page.Msg.GetPublished() {
		t.Error("still published after being taken down")
	}
}
