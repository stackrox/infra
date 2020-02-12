package service

import (
	"context"
	"errors"
	"fmt"
	"sort"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/stackrox/infra/config"
	v1 "github.com/stackrox/infra/generated/api/v1"
	"github.com/stackrox/infra/service/middleware"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type flavorImpl struct {
	flavors       map[string]*v1.Flavor
	defaultFlavor *v1.Flavor
}

var (
	_ middleware.APIService  = (*flavorImpl)(nil)
	_ v1.FlavorServiceServer = (*flavorImpl)(nil)
)

// NewFlavorService creates a new FlavorService.
func NewFlavorService(flavorsCfg []config.FlavorConfig) (middleware.APIService, error) {
	impl := flavorImpl{
		flavors:       make(map[string]*v1.Flavor, len(flavorsCfg)),
		defaultFlavor: nil,
	}

	for _, flavorCfg := range flavorsCfg {
		if _, found := impl.flavors[flavorCfg.ID]; found {
			return nil, fmt.Errorf("duplicate flavor id %q", flavorCfg.ID)
		}

		// Sanity check and convert the configured availability.
		availability, found := v1.FlavorAvailability_value[flavorCfg.Availability]
		if !found {
			return nil, fmt.Errorf("unknown availability %q", flavorCfg.Availability)
		}

		parameters := make([]*v1.Parameter, len(flavorCfg.Parameters))
		for index, parameter := range flavorCfg.Parameters {
			parameters[index] = &v1.Parameter{
				Name:        parameter.Name,
				Description: parameter.Description,
				Example:     parameter.Example,
			}
		}

		flavor := &v1.Flavor{
			ID:           flavorCfg.ID,
			Name:         flavorCfg.Name,
			Description:  flavorCfg.Description,
			Availability: v1.FlavorAvailability(availability),
			Parameters:   parameters,
		}

		impl.flavors[flavor.ID] = flavor

		// Save off the default flavor separately.
		if flavor.Availability == v1.Flavor_default {
			// Ensure that more than one default flavor was not configured.
			if impl.defaultFlavor != nil {
				return nil, fmt.Errorf("both %q and %q configured as default flavors", impl.defaultFlavor.ID, flavor.ID)
			}
			impl.defaultFlavor = flavor
		}
	}

	// Ensure a default flavor was configured.
	if impl.defaultFlavor == nil {
		return nil, errors.New("no default flavor configured")
	}

	return &impl, nil
}

// List implements FlavorService.List.
func (s *flavorImpl) List(context.Context, *empty.Empty) (*v1.FlavorListResponse, error) {
	resp := v1.FlavorListResponse{
		Flavors: make([]*v1.Flavor, 0, len(s.flavors)),
	}

	for _, flavor := range s.flavors {
		resp.Flavors = append(resp.Flavors, flavor)
	}

	if s.defaultFlavor != nil {
		resp.Default = s.defaultFlavor.ID
	}

	sort.Slice(resp.Flavors, func(i, j int) bool {
		if resp.Flavors[i].Availability != resp.Flavors[j].Availability {
			return resp.Flavors[i].Availability > resp.Flavors[j].Availability
		}
		return resp.Flavors[i].ID < resp.Flavors[j].ID
	})

	return &resp, nil
}

// List implements FlavorService.Info.
func (s *flavorImpl) Info(_ context.Context, flavorID *v1.ResourceByID) (*v1.Flavor, error) {
	flavor, found := s.flavors[flavorID.Id]
	if !found {
		return nil, status.Errorf(codes.NotFound, "flavor %q not found", flavorID.Id)
	}

	return flavor, nil
}

// AllowAnonymous declares that this service can be called anonymously.
func (s *flavorImpl) AllowAnonymous() bool {
	return false
}

// RegisterServiceServer registers this service with the given gRPC Server.
func (s *flavorImpl) RegisterServiceServer(server *grpc.Server) {
	v1.RegisterFlavorServiceServer(server, s)
}

// RegisterServiceHandler registers this service with the given gRPC Gateway endpoint.
func (s *flavorImpl) RegisterServiceHandler(ctx context.Context, mux *runtime.ServeMux, conn *grpc.ClientConn) error {
	return v1.RegisterFlavorServiceHandler(ctx, mux, conn)
}
