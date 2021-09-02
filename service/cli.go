package service

import (
	"context"
	"errors"
	"io"
	"log"
	"os"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	v1 "github.com/stackrox/infra/generated/api/v1"
	"github.com/stackrox/infra/service/middleware"
	"google.golang.org/grpc"
)

type cliImpl struct {}

var (
	_ middleware.APIService = (*cliImpl)(nil)
	_ v1.CliServiceServer   = (*cliImpl)(nil)
)

// NewCliService creates a new CliUpgradeService.
func NewCliService() (middleware.APIService, error) {
	return &cliImpl{}, nil
}

// Upgrade provides the a binary for the requested os.
func (s *cliImpl) Upgrade(request *v1.CliUpgradeRequest, stream v1.CliService_UpgradeServer) error {
	bufferSize := 1000 * 1024
	if request.Os != "linux" && request.Os != "darwin" {
		err := errors.New(request.Os + " is not a supported OS")
		log.Println("infractl cli upgrade:", err)
		return err
	}
	filename := "/etc/infra/static/downloads/infractl-" + request.Os + "-amd64"
	file, err := os.Open(filename)
	if err != nil {
		log.Println("Failed to open infractl binary:", err)
		return err
	}
	defer file.Close()
	buff := make([]byte, bufferSize)
	for {
		bytesRead, err := file.Read(buff)
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Println("error while reading chunk:", err)
			return err
		}
		resp := &v1.CliUpgradeResponse{FileChunk: buff[:bytesRead]}
		err = stream.Send(resp)
		if err != nil {
			log.Println("error while sending chunk:", err)
			return err
		}
	}
	return nil
}

// Access configures access for this service.
func (s *cliImpl) Access() map[string]middleware.Access {
	return map[string]middleware.Access{
		"/v1.CliUpgradeService/Download": middleware.Authenticated,
	}
}

// RegisterServiceServer registers this service with the given gRPC Server.
func (s *cliImpl) RegisterServiceServer(server *grpc.Server) {
	v1.RegisterCliServiceServer(server, s)
}

// RegisterServiceHandler registers this service with the given gRPC Gateway endpoint.
func (s *cliImpl) RegisterServiceHandler(ctx context.Context, mux *runtime.ServeMux, conn *grpc.ClientConn) error {
	return v1.RegisterCliServiceHandler(ctx, mux, conn)
}
