package service

import (
	"context"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/stackrox/infra/flavor"
	v1 "github.com/stackrox/infra/generated/api/v1"
	"github.com/stackrox/infra/service/middleware"
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
func (s *flavorImpl) List(context.Context, *empty.Empty) (*v1.FlavorListResponse, error) {
	var resp v1.FlavorListResponse
	for _, flavor := range s.registry.Flavors() {
		flavor := flavor
		scrubInternalParameters(&flavor)
		resp.Flavors = append(resp.Flavors, &flavor)
	}

	return &resp, nil
}

// List implements FlavorService.Info.
func (s *flavorImpl) Info(_ context.Context, flavorID *v1.ResourceByID) (*v1.Flavor, error) {
	flavor, _, found := s.registry.Get(flavorID.Id)
	if !found {
		return nil, status.Errorf(codes.NotFound, "flavor %q not found", flavorID.Id)
	}
	scrubInternalParameters(&flavor)

	return &flavor, nil
}

// scrubInternalParameters drops any internal parameters from the given flavor,
// as the end user is not allowed to provide values for them.
func scrubInternalParameters(flavor *v1.Flavor) {
	for paramName, paramValue := range flavor.Parameters {
		if paramValue.Internal {
			delete(flavor.Parameters, paramName)
		}
	}
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
