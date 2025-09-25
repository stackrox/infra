package service

import (
	"context"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	v1 "github.com/stackrox/infra/generated/api/v1"
	"github.com/stackrox/infra/pkg/flavor"
	"github.com/stackrox/infra/pkg/service/middleware"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type flavorImpl struct {
	registry *flavor.Registry
}

var (
	_ middleware.APIService  = (*flavorImpl)(nil)
	_ v1.FlavorServiceServer = (*flavorImpl)(nil)
)

// NewFlavorService creates a new FlavorService.
func NewFlavorService(registry *flavor.Registry) (middleware.APIService, error) {
	impl := flavorImpl{
		registry: registry,
	}

	return &impl, nil
}

// List implements FlavorService.List.
func (s *flavorImpl) List(_ context.Context, request *v1.FlavorListRequest) (*v1.FlavorListResponse, error) {
	var resp v1.FlavorListResponse
	for _, flavor := range s.registry.Flavors() {
		if flavor.GetAvailability() == v1.Flavor_janitorDelete {
			continue
		}
		if !request.GetAll() && (flavor.GetAvailability() == v1.Flavor_alpha || flavor.GetAvailability() == v1.Flavor_deprecated) {
			continue
		}
		scrubInternalParameters(&flavor)
		resp.Flavors = append(resp.GetFlavors(), &flavor)
	}

	return &resp, nil
}

// Info implements FlavorService.Info.
func (s *flavorImpl) Info(_ context.Context, flavorID *v1.ResourceByID) (*v1.Flavor, error) {
	flavor, _, found := s.registry.Get(flavorID.Id)
	if !found || flavor.GetAvailability() == v1.Flavor_janitorDelete {
		return nil, status.Errorf(codes.NotFound, "flavor %q not found", flavorID.Id)
	}
	scrubInternalParameters(&flavor)

	return &flavor, nil
}

// scrubInternalParameters drops any internal parameters from the given flavor,
// as the end user is not allowed to provide values for them.
func scrubInternalParameters(flavor *v1.Flavor) {
	newParams := make(map[string]*v1.Parameter)
	for paramName, paramValue := range flavor.Parameters {
		if paramValue.Internal {
			continue
		}
		newParams[paramName] = paramValue
	}

	flavor.Parameters = newParams
}

// Access configures access for this service.
func (s *flavorImpl) Access() map[string]middleware.Access {
	return map[string]middleware.Access{
		"/v1.FlavorService/Info": middleware.Authenticated,
		"/v1.FlavorService/List": middleware.Authenticated,
	}
}

// RegisterServiceServer registers this service with the given gRPC Server.
func (s *flavorImpl) RegisterServiceServer(server *grpc.Server) {
	v1.RegisterFlavorServiceServer(server, s)
}

// RegisterServiceHandler registers this service with the given gRPC Gateway endpoint.
func (s *flavorImpl) RegisterServiceHandler(ctx context.Context, mux *runtime.ServeMux, conn *grpc.ClientConn) error {
	return v1.RegisterFlavorServiceHandler(ctx, mux, conn)
}
