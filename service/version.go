package service

import (
	"context"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	v1 "github.com/stackrox/infra/generated/api/v1"
	"github.com/stackrox/infra/pkg/buildinfo"
	"github.com/stackrox/infra/service/middleware"
	"google.golang.org/grpc"
)

type versionImpl struct {
	v1.UnimplementedVersionServiceServer
}

var (
	_ middleware.APIService   = (*versionImpl)(nil)
	_ v1.VersionServiceServer = (*versionImpl)(nil)
)

// NewVersionService creates a new VersionService.
func NewVersionService() (middleware.APIService, error) {
	return &versionImpl{}, nil
}

// GetVersion implements VersionService.GetVersion.
func (s *versionImpl) GetVersion(_ context.Context, _ *empty.Empty) (*v1.Version, error) {
	return buildinfo.All(), nil
}

// Access configures access for this service.
func (s *versionImpl) Access() map[string]middleware.Access {
	return map[string]middleware.Access{
		"/v1.VersionService/GetVersion": middleware.Anonymous,
	}
}

// RegisterServiceServer registers this service with the given gRPC Server.
func (s *versionImpl) RegisterServiceServer(server *grpc.Server) {
	v1.RegisterVersionServiceServer(server, s)
}

// RegisterServiceHandler registers this service with the given gRPC Gateway endpoint.
func (s *versionImpl) RegisterServiceHandler(ctx context.Context, mux *runtime.ServeMux, conn *grpc.ClientConn) error {
	return v1.RegisterVersionServiceHandler(ctx, mux, conn)
}
