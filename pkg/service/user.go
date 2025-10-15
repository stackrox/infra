package service

import (
	"context"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/pkg/errors"
	v1 "github.com/stackrox/infra/generated/api/v1"
	"github.com/stackrox/infra/pkg/service/middleware"
	"google.golang.org/grpc"
)

type userImpl struct {
	v1.UnimplementedUserServiceServer
	generate func(*v1.ServiceAccount) (string, error)
}

var (
	_ middleware.APIService = (*userImpl)(nil)
	_ v1.UserServiceServer  = (*userImpl)(nil)
)

// NewUserService creates a new UserService.
func NewUserService(generator func(*v1.ServiceAccount) (string, error)) (middleware.APIService, error) {
	return &userImpl{
		generate: generator,
	}, nil
}

// CreateToken implements UserService.CreateToken.
func (s *userImpl) CreateToken(_ context.Context, req *v1.ServiceAccount) (*v1.TokenResponse, error) {
	// Generate the service account token.
	token, err := s.generate(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate token")
	}

	return &v1.TokenResponse{
		Account: req,
		Token:   token,
	}, nil
}

// Token implements UserService.Token.
func (s *userImpl) Token(ctx context.Context, _ *empty.Empty) (*v1.TokenResponse, error) {
	// Extract the calling user from the context.
	user, found := middleware.UserFromContext(ctx)
	if !found {
		return nil, errors.New("not called by a user")
	}

	// Synthesize a service account from the current user.
	svcacct := v1.ServiceAccount{
		Name:        user.Name,
		Description: "Personal service account for " + user.Email,
		Email:       user.Email,
	}

	return s.CreateToken(ctx, &svcacct)
}

// Whoami implements UserService.Whoami.
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
		"/v1.UserService/Token":       middleware.Authenticated,
		"/v1.UserService/CreateToken": middleware.Admin,
		"/v1.UserService/Whoami":      middleware.Anonymous,
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
