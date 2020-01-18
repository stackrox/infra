package middleware

import (
	"context"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"google.golang.org/grpc"
)

// APIService is the service interface
type APIService interface {
	AllowAnonymous() bool
	RegisterServiceServer(server *grpc.Server)
	RegisterServiceHandler(context.Context, *runtime.ServeMux, *grpc.ClientConn) error
}

type APIServiceFunc func() (APIService, error)

func Services(serviceFuncs ...APIServiceFunc) ([]APIService, error) {
	services := make([]APIService, len(serviceFuncs))
	for index, serviceFunc := range serviceFuncs {
		service, err := serviceFunc()
		if err != nil {
			return nil, err
		}
		services[index] = service
	}
	return services, nil
}
