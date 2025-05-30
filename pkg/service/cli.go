// Package service implements the backend services for the infra server.
package service

import (
	"context"
	"io"
	"os"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	v1 "github.com/stackrox/infra/generated/api/v1"
	"github.com/stackrox/infra/pkg/logging"
	"github.com/stackrox/infra/pkg/platform"
	"github.com/stackrox/infra/pkg/service/middleware"
	"google.golang.org/grpc"
)

const bufferSize = 1000 * 1024

type cliImpl struct {
	staticDir string
}

var (
	log = logging.CreateProductionLogger()

	_ middleware.APIService = (*cliImpl)(nil)
	_ v1.CliServiceServer   = (*cliImpl)(nil)
)

// NewCliService creates a new CliUpgradeService.
func NewCliService(staticDir string) (middleware.APIService, error) {
	return &cliImpl{staticDir: staticDir}, nil
}

// Upgrade provides the binary for the requested OS and architecture.
func (s *cliImpl) Upgrade(request *v1.CliUpgradeRequest, stream v1.CliService_UpgradeServer) error {
	if err := platform.Validate(request.GetOs(), request.GetArch()); err != nil {
		log.Log(logging.INFO, "failed to validate platform for infractl upgrade", "error", err)
		return err
	}

	filename := s.staticDir + "/downloads/infractl-" + request.GetOs() + "-" + request.GetArch()
	file, err := os.Open(filename)
	if err != nil {
		log.Log(logging.ERROR, "failed to open infractl binary", "error", err)
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
			log.Log(logging.ERROR, "error while reading infractl chunk", "error", err)
			return err
		}
		resp := &v1.CliUpgradeResponse{FileChunk: buff[:bytesRead]}
		if err := stream.Send(resp); err != nil {
			log.Log(logging.ERROR, "error while sending infractl chunk", "error", err)
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
