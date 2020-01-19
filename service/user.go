package service

import (
	"context"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	v1 "github.com/stackrox/infra/generated/api/v1"
	"google.golang.org/grpc"
)

type userImpl struct{}

var _ APIService = (*userImpl)(nil)
var _ v1.UserServiceServer = (*userImpl)(nil)

// NewUserService creates a new UserService.
func NewUserService() APIService {
	return &userImpl{}
}

// GetVersion implements UserService.Whoami.
func (s *userImpl) Whoami(ctx context.Context, request *empty.Empty) (*v1.WhoamiResponse, error) {
	if user, expiry, found := UserFromContext(ctx); found {
		return &v1.WhoamiResponse{
			User:   user,
			Expiry: expiry,
		}, nil
	}

	return &v1.WhoamiResponse{}, nil
}

// RegisterServiceServer registers this service with the given gRPC Server.
func (s *userImpl) RegisterServiceServer(server *grpc.Server) {
	v1.RegisterUserServiceServer(server, s)
}

// RegisterServiceHandler registers this service with the given gRPC Gateway endpoint.
func (s *userImpl) RegisterServiceHandler(ctx context.Context, mux *runtime.ServeMux, conn *grpc.ClientConn) error {
	return v1.RegisterUserServiceHandler(ctx, mux, conn)
}

// RegisterServiceHandler registers this service with the given gRPC Gateway endpoint.
func (s *userImpl) AllowAnonymous() bool {
	return false
}
