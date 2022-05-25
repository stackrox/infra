package service

import (
	"context"
	"io"
	"log"
	"os"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/pkg/errors"
	v1 "github.com/stackrox/infra/generated/api/v1"
	"github.com/stackrox/infra/service/middleware"
	"google.golang.org/grpc"
)

const bufferSize = 1000 * 1024

type cliImpl struct{}

var (
	_ middleware.APIService = (*cliImpl)(nil)
	_ v1.CliServiceServer   = (*cliImpl)(nil)
)

// NewCliService creates a new CliUpgradeService.
func NewCliService() (middleware.APIService, error) {
	return &cliImpl{}, nil
}

// Upgrade provides the binary for the requested OS and architecture.
func (s *cliImpl) Upgrade(request *v1.CliUpgradeRequest, stream v1.CliService_UpgradeServer) error {
	if request.Os != "linux" && request.Os != "darwin" {
		err := errors.Errorf("%s is not a supported OS", request.Os)
		log.Println("infractl cli upgrade:", err)
		return err
	}
	if request.Arch != "amd64" && request.Arch != "arm64" {
		err := errors.Errorf("%s is not a supported arch", request.Arch)
		log.Println("infractl cli upgrade:", err)
		return err
	}

	filename := webRoot + "/downloads/infractl-" + request.Os + "-" + request.Arch
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
		if err := stream.Send(resp); err != nil {
			log.Println("error while sending chunk:", err)
			return err
		}
	}
	return nil
}

// Access configures access for this service.
func (s *cliImpl) Access() map[string]middleware.Access {
	return map[string]middleware.Access{
		"/v1.CliUpgradeService/Download": middleware.Anonymous,
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
