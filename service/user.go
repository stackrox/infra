package service

import (
	"context"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	v1 "github.com/stackrox/infra/generated/api/v1"
	"github.com/stackrox/infra/service/middleware"
	"google.golang.org/grpc"
)

type userImpl struct{}

var (
	_ middleware.APIService = (*userImpl)(nil)
	_ v1.UserServiceServer  = (*userImpl)(nil)
)

// NewUserService creates a new UserService.
func NewUserService() middleware.APIService {
	return &userImpl{}
}

// GetVersion implements UserService.Whoami.
func (s *userImpl) Whoami(ctx context.Context, _ *empty.Empty) (*v1.WhoamiResponse, error) {
	if user, found := middleware.UserFromContext(ctx); found {
		return &v1.WhoamiResponse{
			Principal: &v1.WhoamiResponse_User{
				User: user,
			},
		}, nil
	}

	if svcacct, found := middleware.ServiceAccountFromContext(ctx); found {
		return &v1.WhoamiResponse{
			Principal: &v1.WhoamiResponse_ServiceAccount{
				ServiceAccount: svcacct,
			},
		}, nil
	}

	return &v1.WhoamiResponse{}, nil
}

// AllowAnonymous declares that the user service can be called anonymously.
func (s *userImpl) AllowAnonymous() bool {
	return true
}

// RegisterServiceServer registers this service with the given gRPC Server.
func (s *userImpl) RegisterServiceServer(server *grpc.Server) {
	v1.RegisterUserServiceServer(server, s)
}

// RegisterServiceHandler registers this service with the given gRPC Gateway endpoint.
func (s *userImpl) RegisterServiceHandler(ctx context.Context, mux *runtime.ServeMux, conn *grpc.ClientConn) error {
	return v1.RegisterUserServiceHandler(ctx, mux, conn)
}
