package service

import (
	"context"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/stackrox/infra/generated/api/v1"
	"github.com/stackrox/infra/pkg/buildinfo"
	"google.golang.org/grpc"
)

type versionImpl struct{}

var _ APIService = (*versionImpl)(nil)
var _ v1.VersionServiceServer = (*versionImpl)(nil)

// NewVersionService creates a new VersionService.
func NewVersionService() APIService {
	return &versionImpl{}
}

// GetVersion implements VersionService.GetVersion.
func (s *versionImpl) GetVersion(ctx context.Context, request *empty.Empty) (*v1.Version, error) {
	info := buildinfo.All()
	return &info, nil
}

// RegisterServiceServer registers this service with the given gRPC Server.
func (s *versionImpl) RegisterServiceServer(server *grpc.Server) {
	v1.RegisterVersionServiceServer(server, s)
}

// RegisterServiceHandler registers this service with the given gRPC Gateway endpoint.
func (s *versionImpl) RegisterServiceHandler(ctx context.Context, mux *runtime.ServeMux, conn *grpc.ClientConn) error {
	return v1.RegisterVersionServiceHandler(ctx, mux, conn)
}
