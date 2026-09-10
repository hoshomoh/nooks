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
type fakeStore struct {
	settings store.InstanceSettings
	err      error
}

func (f *fakeStore) InstanceSettings(context.Context) (store.InstanceSettings, error) {
	return f.settings, f.err
}

func (f *fakeStore) SaveInstanceSettings(context.Context, store.InstanceSettings) error {
	return f.err
}

func (f *fakeStore) Close() error { return nil }

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
