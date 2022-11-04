package service

import (
	"context"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	v1 "github.com/stackrox/infra/generated/api/v1"
	"github.com/stackrox/infra/service/middleware"
	"google.golang.org/grpc"
)

type statusImpl struct{}

var (
	_ middleware.APIService       = (*statusImpl)(nil)
	_ v1.InfraStatusServiceServer = (*statusImpl)(nil)
)

// NewStatusService creates a new InfraStatusService.
func NewStatusService() (middleware.APIService, error) {
	return &statusImpl{}, nil
}

// GetStatus shows infra maintenance status.
func (s *statusImpl) GetStatus(ctx context.Context, _ *empty.Empty) (*v1.InfraStatus, error) {
	infraStatus := v1.InfraStatus{
		Maintainer:        "tmartens@redhat.com",
		MaintenanceActive: true,
	}
	return &infraStatus, nil
}

// Access configures access for this service.
func (s *statusImpl) Access() map[string]middleware.Access {
	return map[string]middleware.Access{
		"/v1.InfraStatusService/Status": middleware.Anonymous,
	}
}

// RegisterServiceServer registers this service with the given gRPC Server.
func (s *statusImpl) RegisterServiceServer(server *grpc.Server) {
	v1.RegisterInfraStatusServiceServer(server, s)
}

// RegisterServiceHandler registers this service with the given gRPC Gateway endpoint.
func (s *statusImpl) RegisterServiceHandler(ctx context.Context, mux *runtime.ServeMux, conn *grpc.ClientConn) error {
	return v1.RegisterInfraStatusServiceHandler(ctx, mux, conn)
}
