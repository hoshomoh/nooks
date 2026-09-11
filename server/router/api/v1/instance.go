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

	"github.com/hoshomoh/nooks/internal/password"

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
		Name:          settings.Name,
		Version:       version.String(),
		NeedsSetup:    settings.NeedsSetup(),
		PublicSignup:  settings.PublicSignup,
		DefaultLocale: settings.DefaultLocale,
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
	settings.DefaultLocale = wanted.GetDefaultLocale()
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
		Name:          settings.Name,
		PublicSignup:  settings.PublicSignup,
		DefaultLocale: settings.DefaultLocale,
		PublicList: &apiv1.PublicListSettings{
			ListUid:   settings.Public.ListUID,
			ShowNames: settings.Public.ShowNames,
			ShowMeta:  settings.Public.ShowMeta,
			AllowJoin: settings.Public.AllowJoin,
		},
	}
}

// licence is what this build is under. Stated rather than looked up: it is a fact about
// the source, and a file read at runtime would be a file somebody could replace.
const licence = "AGPL-3.0-or-later"

/*
GetInstanceAbout reports what this copy of Nooks is and how much it holds.

Any Member, not only an Admin. It is their Instance too, and none of it — a version, a
count, the size on disk — is anybody else's business to keep from them.
*/
func (s *InstanceService) GetInstanceAbout(
	ctx context.Context,
	_ *connect.Request[apiv1.GetInstanceAboutRequest],
) (*connect.Response[apiv1.GetInstanceAboutResponse], error) {
	if _, err := requireMember(ctx); err != nil {
		return nil, err
	}

	settings, err := s.store.InstanceSettings(ctx)
	if err != nil {
		return nil, internalError("read instance settings", err)
	}
	stats, err := s.store.Stats(ctx)
	if err != nil {
		return nil, internalError("count what the instance holds", err)
	}

	return connect.NewResponse(&apiv1.GetInstanceAboutResponse{
		Version:       version.String(),
		StartedAt:     formatMoment(settings.SetupCompletedAt),
		MemberCount:   int32(stats.Members),
		ListCount:     int32(stats.Lists),
		ItemCount:     int32(stats.Items),
		StorageBytes:  stats.StorageBytes,
		Licence:       licence,
		StorageDriver: stats.Driver,
		InstanceName:  settings.Name,
	}), nil
}

// errWrongInstanceName refuses a deletion where the typed name does not match.
var errWrongInstanceName = connect.NewError(connect.CodeInvalidArgument,
	errors.New("that is not the name of this instance"))

/*
DeleteInstance empties the Instance and returns it to first run.

Two confirmations, and they do different jobs. Typing the name makes an Admin read what
they are about to lose. The password is the one that matters: an unattended browser is
how this realistically happens by accident, and a name can be copied off the screen in
front of you.

Any Admin may do it. On a household Instance the people with the keys are the people
who share the shopping, and a rule that only the founder could wipe it would strand a
household whose founder has left.
*/
func (s *InstanceService) DeleteInstance(
	ctx context.Context,
	req *connect.Request[apiv1.DeleteInstanceRequest],
) (*connect.Response[apiv1.DeleteInstanceResponse], error) {
	admin, err := requireAdmin(ctx)
	if err != nil {
		return nil, err
	}

	settings, err := s.store.InstanceSettings(ctx)
	if err != nil {
		return nil, internalError("read instance settings", err)
	}
	if !strings.EqualFold(strings.TrimSpace(req.Msg.GetInstanceName()), settings.Name) {
		return nil, errWrongInstanceName
	}

	if err := password.Verify(admin.PasswordHash, req.Msg.GetPassword()); err != nil {
		if errors.Is(err, password.ErrWrong) {
			return nil, connect.NewError(connect.CodeInvalidArgument,
				errors.New("that is not your password"))
		}
		return nil, internalError("verify password", err)
	}

	if err := s.store.ResetInstance(ctx); err != nil {
		return nil, internalError("delete instance", err)
	}
	return connect.NewResponse(&apiv1.DeleteInstanceResponse{}), nil
}
