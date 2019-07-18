package server

import (
	"context"
	"log"
	"net"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/pkg/errors"
	"github.com/stackrox/infra/service"
	"google.golang.org/grpc"
)

// RunGRPCServer runs the gRPC service.
func RunGRPCServer(apiServices []service.APIService, grpcAddress string) (func(), <-chan error, error) {
	listen, err := net.Listen("tcp", grpcAddress)
	if err != nil {
		return nil, nil, err
	}

	// Create the server.
	server := grpc.NewServer()

	// Register the gRPC API service.
	for _, apiSvc := range apiServices {
		apiSvc.RegisterServiceServer(server)
	}

	// errch is a channel of size 1 which will receive the (only) error
	// returned by the gRPC server.
	errch := make(chan error, 1)

	shutdown := func() {
		log.Print("Shutting down gRPC.")
		server.GracefulStop()
	}

	go func() {
		log.Printf("starting gRPC server on %s", grpcAddress)
		if err := server.Serve(listen); err != nil {
			errch <- err
		}
	}()

	return shutdown, errch, nil
}

// RunHTTPServer runs the HTTP/REST gateway
func RunHTTPServer(apiServices []service.APIService, httpAddress string, grpcAddress string) (func(), <-chan error, error) {
	// Register the HTTP/gRPC gateway service.
	ctx := context.Background()
	conn, err := grpc.Dial(grpcAddress, grpc.WithInsecure())
	if err != nil {
		return nil, nil, errors.Wrap(err, "dialing GRPC")
	}

	mux := http.NewServeMux()

	gwMux := runtime.NewServeMux()
	for _, apiSvc := range apiServices {
		if err := apiSvc.RegisterServiceHandler(ctx, gwMux, conn); err != nil {
			return nil, nil, err
		}
	}

	mux.Handle("/v1/", gwMux)

	srv := &http.Server{
		Addr:    httpAddress,
		Handler: mux,
	}

	// errch is a channel of size 1 which will receive the (only) error
	// returned by the HTTP server.
	errch := make(chan error, 1)

	shutdown := func() {
		log.Print("shutting down HTTP server")
		srv.Shutdown(ctx)
	}

	go func() {
		log.Printf("starting HTTP server on %s", httpAddress)
		if err := srv.ListenAndServe(); err != nil {
			errch <- err
		}
	}()

	return shutdown, errch, nil
}
