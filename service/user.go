package service

import (
	"context"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/pkg/errors"
	v1 "github.com/stackrox/infra/generated/api/v1"
	"github.com/stackrox/infra/service/middleware"
	"google.golang.org/grpc"
)

type userImpl struct {
	generate func(v1.ServiceAccount) (string, error)
}

var (
	_ middleware.APIService = (*userImpl)(nil)
	_ v1.UserServiceServer  = (*userImpl)(nil)
)

// NewUserService creates a new UserService.
func NewUserService(generator func(v1.ServiceAccount) (string, error)) (middleware.APIService, error) {
	return &userImpl{
		generate: generator,
	}, nil
}

func (s *userImpl) Token(_ context.Context, req *v1.ServiceAccount) (*v1.TokenResponse, error) {
	token, err := s.generate(*req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate token")
	}

	return &v1.TokenResponse{
		Account: req,
		Token:   token,
	}, nil
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

// Access configures access for this service.
func (s *userImpl) Access() map[string]middleware.Access {
	return map[string]middleware.Access{
		"/v1.UserService/Token":  middleware.Admin,
		"/v1.UserService/Whoami": middleware.Anonymous,
	}
}

// RegisterServiceServer registers this service with the given gRPC Server.
func (s *userImpl) RegisterServiceServer(server *grpc.Server) {
	v1.RegisterUserServiceServer(server, s)
}

// RegisterServiceHandler registers this service with the given gRPC Gateway endpoint.
func (s *userImpl) RegisterServiceHandler(ctx context.Context, mux *runtime.ServeMux, conn *grpc.ClientConn) error {
	return v1.RegisterUserServiceHandler(ctx, mux, conn)
}
