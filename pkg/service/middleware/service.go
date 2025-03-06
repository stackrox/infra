package middleware

import (
	"context"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"google.golang.org/grpc"
)

// APIService is the service interface
type APIService interface {
	Access() map[string]Access
	RegisterServiceServer(server *grpc.Server)
	RegisterServiceHandler(context.Context, *runtime.ServeMux, *grpc.ClientConn) error
}

// APIServiceFunc represents a function that is capable of making a APIService.
type APIServiceFunc func() (APIService, error)

// Services process the given APIServiceFunc list, and returns the resulting
// APIService list. If any errors occur, that error is returned immediately.
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
