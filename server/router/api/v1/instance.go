// Package v1 implements the Connect services defined in proto/nooks/api/v1.
//
// A service here decides policy and translates between the store and the wire. It
// does not open databases, read configuration, or know how the HTTP server is built.
package v1

import (
	"context"
	"fmt"

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
