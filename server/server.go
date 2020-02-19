// Package server facilitates running a HTTPS and gRPC server.
package server

import (
	"context"
	"log"
	"net"
	"net/http"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/pkg/errors"
	"github.com/stackrox/infra/auth"
	"github.com/stackrox/infra/config"
	"github.com/stackrox/infra/service/middleware"
	"google.golang.org/grpc"
)

const (
	addressAny       = "0.0.0.0"
	addressLocalhost = "localhost"
)

type server struct {
	services []middleware.APIService
	cfg      config.Config
	manager  TLSManager
	oauth    auth.OAuth
}

// New creates a new server that is ready to be launched.
func New(serverCfg config.Config, auth0 auth.OAuth, services ...middleware.APIService) (*server, error) {
	manager, err := NewTLSManager(serverCfg.Server)
	if err != nil {
		return nil, err
	}

	return &server{
		services: services,
		cfg:      serverCfg,
		manager:  manager,
		oauth:    auth0,
	}, nil
}

func (s *server) RunServer() (<-chan error, error) {
	grpcListenAddress := addressAny + ":" + s.cfg.Server.GRPC
	grpcConnectAddress := addressLocalhost + ":" + s.cfg.Server.GRPC
	mux := http.NewServeMux()
	errCh := make(chan error, 1)

	////////////////////////////////
	// Step 1 - Start gRPC server //
	////////////////////////////////

	// Create the server.
	server := grpc.NewServer(
		s.manager.ServerOption(),

		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			// Extract user from JWT token stored in HTTP cookie.
			middleware.ContextInterceptor(middleware.UserEnricher(s.oauth)),
			// Extract service-account from token stored in Authorization header.
			middleware.ContextInterceptor(middleware.ServiceAccountEnricher(s.cfg.ServiceAccounts)),
			// Enforce authenticated user access on resources that declare it.
			middleware.ContextInterceptor(middleware.EnforceAnonymousAccess),
		)),
	)

	// Register the gRPC API service.
	for _, apiSvc := range s.services {
		apiSvc.RegisterServiceServer(server)
	}

	listen, err := net.Listen("tcp", grpcListenAddress)
	if err != nil {
		return nil, err
	}

	log.Print("starting gRPC server")
	go func() {
		defer listen.Close() // nolint:errcheck
		defer server.Stop()

		if err := server.Serve(listen); err != nil {
			errCh <- err
		}
	}()

	/////////////////////////////////
	// Step 2 - Start HTTPS server //
	/////////////////////////////////

	log.Printf("starting HTTPS server in %s mode", s.manager.Name())
	go func() {
		defer s.manager.Listener().Close() // nolint:errcheck

		if err := http.Serve(s.manager.Listener(), mux); err != nil {
			errCh <- err
		}
	}()

	///////////////////////////////////////////
	// Step 3 - Register gRPC-Gateway routes //
	///////////////////////////////////////////

	log.Print("starting gRPC-Gateway client")
	conn, err := grpc.Dial(grpcConnectAddress, s.manager.DialOptions()...)
	if err != nil {
		return nil, errors.Wrap(err, "dialing GRPC")
	}

	// Register each service
	gwMux := runtime.NewServeMux(
		runtime.WithMarshalerOption("*", &runtime.JSONPb{Indent: "  "}),
	)

	for _, apiSvc := range s.services {
		if err := apiSvc.RegisterServiceHandler(context.Background(), gwMux, conn); err != nil {
			return nil, err
		}
	}

	// Updates http handler routes. This included "web-only" routes, like
	// login/logout/static, and also gRPC-Gateway routes.
	mux.Handle("/", http.FileServer(http.Dir(s.cfg.Server.StaticDir)))
	mux.Handle("/v1/", gwMux)
	s.oauth.Handle(mux)

	return errCh, nil
}
