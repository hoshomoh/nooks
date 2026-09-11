// Package v1 implements the Connect services defined in proto/nooks/api/v1.
//
// A service here decides policy and translates between the store and the wire. It
// does not open databases, read configuration, or know how the HTTP server is built.
package v1

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"connectrpc.com/connect"

	"github.com/hoshomoh/nooks/internal/version"
	apiv1 "github.com/hoshomoh/nooks/proto/gen/nooks/api/v1"
	"github.com/hoshomoh/nooks/store"
)

// InstanceService reports what this copy of Nooks is and how it is configured.
type InstanceService struct {
	store store.Store
}

// NewInstanceService returns a service backed by the given store.
func NewInstanceService(s store.Store) *InstanceService {
	return &InstanceService{store: s}
}

// GetInstance returns the Instance's public profile. It is reachable without
// authentication because the sign-in page and the Public list both need it before
// anyone has signed in, so it must not disclose anything a Visitor may not see.
func (s *InstanceService) GetInstance(
	ctx context.Context,
	_ *connect.Request[apiv1.GetInstanceRequest],
) (*connect.Response[apiv1.GetInstanceResponse], error) {
	settings, err := s.store.InstanceSettings(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("read instance settings: %w", err))
	}

	return connect.NewResponse(&apiv1.GetInstanceResponse{
		Name:         settings.Name,
		Version:      version.String(),
		NeedsSetup:   settings.NeedsSetup(),
		PublicSignup: settings.PublicSignup,
	}), nil
}

// GetInstanceSettings returns everything an Admin may change.
//
// Admins only, and deliberately separate from GetInstance: which List is published is
// not a fact a Visitor gets to read from the outside, and the public profile is read
// before anybody has signed in.
func (s *InstanceService) GetInstanceSettings(
	ctx context.Context,
	_ *connect.Request[apiv1.GetInstanceSettingsRequest],
) (*connect.Response[apiv1.GetInstanceSettingsResponse], error) {
	if _, err := requireAdmin(ctx); err != nil {
		return nil, err
	}

	settings, err := s.store.InstanceSettings(ctx)
	if err != nil {
		return nil, internalError("read instance settings", err)
	}
	return connect.NewResponse(&apiv1.GetInstanceSettingsResponse{
		Settings: settingsToProto(settings),
	}), nil
}

// UpdateInstanceSettings replaces them.
func (s *InstanceService) UpdateInstanceSettings(
	ctx context.Context,
	req *connect.Request[apiv1.UpdateInstanceSettingsRequest],
) (*connect.Response[apiv1.UpdateInstanceSettingsResponse], error) {
	if _, err := requireAdmin(ctx); err != nil {
		return nil, err
	}

	wanted := req.Msg.GetSettings()
	name := strings.TrimSpace(wanted.GetName())
	if err := requireText(name, "a name"); err != nil {
		return nil, err
	}

	public, err := s.publicFromProto(ctx, wanted.GetPublicList())
	if err != nil {
		return nil, err
	}

	// Read first, so the parts an Admin cannot change — when first run finished — are
	// carried over rather than reset by anything this call leaves out.
	settings, err := s.store.InstanceSettings(ctx)
	if err != nil {
		return nil, internalError("read instance settings", err)
	}
	settings.Name = name
	settings.PublicSignup = wanted.GetPublicSignup()
	settings.Public = public

	if err := s.store.SaveInstanceSettings(ctx, settings); err != nil {
		return nil, internalError("save instance settings", err)
	}
	return connect.NewResponse(&apiv1.UpdateInstanceSettingsResponse{
		Settings: settingsToProto(settings),
	}), nil
}

// publicFromProto checks that the List being published is one that exists.
//
// An identifier that names nothing would publish a page that answers "nothing here" to
// everybody, which looks exactly like the Admin having turned it off.
func (s *InstanceService) publicFromProto(
	ctx context.Context,
	wanted *apiv1.PublicListSettings,
) (store.PublicList, error) {
	public := store.PublicList{
		ListUID:   wanted.GetListUid(),
		ShowNames: wanted.GetShowNames(),
		ShowMeta:  wanted.GetShowMeta(),
		AllowJoin: wanted.GetAllowJoin(),
	}
	if !public.IsPublished() {
		return public, nil
	}

	if _, err := s.store.ListByUID(ctx, public.ListUID); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return store.PublicList{}, errListNotFound
		}
		return store.PublicList{}, internalError("read list", err)
	}
	return public, nil
}

// settingsToProto converts the Instance's configuration for the wire.
func settingsToProto(settings store.InstanceSettings) *apiv1.InstanceSettings {
	return &apiv1.InstanceSettings{
		Name:         settings.Name,
		PublicSignup: settings.PublicSignup,
		PublicList: &apiv1.PublicListSettings{
			ListUid:   settings.Public.ListUID,
			ShowNames: settings.Public.ShowNames,
			ShowMeta:  settings.Public.ShowMeta,
			AllowJoin: settings.Public.AllowJoin,
		},
	}
}
