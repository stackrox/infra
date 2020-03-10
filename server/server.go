// Package server facilitates running a HTTPS and gRPC server.
package server

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
	"strings"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/pkg/errors"
	"github.com/stackrox/infra/auth"
	"github.com/stackrox/infra/config"
	"github.com/stackrox/infra/service/middleware"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type server struct {
	services []middleware.APIService
	cfg      config.Config
	oauth    auth.OAuth
}

// New creates a new server that is ready to be launched.
func New(serverCfg config.Config, auth0 auth.OAuth, services ...middleware.APIService) *server {
	return &server{
		services: services,
		cfg:      serverCfg,
		oauth:    auth0,
	}
}

func (s *server) RunServer() (<-chan error, error) {
	// listenAddress is the address that the server will listen on. Must bind
	// to INADDR_ANY in order for the server to be reachable outside the
	// container.
	listenAddress := fmt.Sprintf("0.0.0.0:%d", s.cfg.Server.Port)

	// connectAddress is the address that the (gRPC-Gateway) client will
	// connect to. Can be localhost as the connection doesn't leave the
	// container.
	connectAddress := fmt.Sprintf("localhost:%d", s.cfg.Server.Port)

	mux := http.NewServeMux()
	errCh := make(chan error, 1)

	// Create the server.
	server := grpc.NewServer(
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			// Extract user from JWT token stored in HTTP cookie.
			middleware.ContextInterceptor(middleware.UserEnricher(s.oauth)),
			// Extract service-account from token stored in Authorization header.
			middleware.ContextInterceptor(middleware.ServiceAccountEnricher(s.oauth.ValidateServiceAccountToken)),

			middleware.ContextInterceptor(middleware.AdminEnricher(s.cfg.Password)),
			// Enforce authenticated user access on resources that declare it.
			middleware.ContextInterceptor(middleware.EnforceAccess),
		)),
	)

	// Register the gRPC API service.
	for _, apiSvc := range s.services {
		apiSvc.RegisterServiceServer(server)
	}

	// muxHandler is a HTTP handler that can route both HTTP/2 gRPC and HTTP1.1
	// requests.
	muxHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Is the current request gRPC?
		if r.ProtoMajor == 2 && strings.HasPrefix(r.Header.Get("Content-Type"), "application/grpc") {
			server.ServeHTTP(w, r)
			return
		}

		// Fallback to HTTP
		mux.ServeHTTP(w, r)
	})

	log.Printf("Starting gRPC server on %s", listenAddress)
	go func() {
		if err := http.ListenAndServeTLS(listenAddress, s.cfg.Server.CertFile, s.cfg.Server.KeyFile, h2c.NewHandler(muxHandler, &http2.Server{})); err != nil {
			errCh <- err
		}
	}()

	dialOption, err := grpcLocalCredentials(s.cfg.Server.CertFile)
	if err != nil {
		return nil, err
	}

	log.Printf("Starting gRPC-Gateway client on %s", connectAddress)
	conn, err := grpc.Dial(connectAddress, dialOption)
	if err != nil {
		return nil, errors.Wrap(err, "dialing gRPC")
	}

	gwMux := runtime.NewServeMux(
		runtime.WithMarshalerOption("*", &runtime.JSONPb{Indent: "  "}),
	)

	// Register each service
	for _, apiSvc := range s.services {
		if err := apiSvc.RegisterServiceHandler(context.Background(), gwMux, conn); err != nil {
			return nil, err
		}
	}

	// Updates http handler routes. This included "web-only" routes, like
	// login/logout/static, and also gRPC-Gateway routes.
	mux.Handle("/", http.FileServer(http.Dir(s.cfg.Server.StaticDir)))
	mux.Handle("/v1/", gwMux)
	mux.Handle("/downloads/", s.oauth.AuthorizedFunc(downloadsHandler(s.cfg.Server.AssetsDir)))
	s.oauth.Handle(mux)

	return errCh, nil
}

func grpcLocalCredentials(certFile string) (grpc.DialOption, error) {
	// Read the x509 PEM public certificate file
	pem, err := ioutil.ReadFile(certFile)
	if err != nil {
		return nil, err
	}

	// Create an empty certificate pool, and add our single "CA" certificate to
	// it. This allows us to trust the local server specifically, as its
	// serving the same exact certificate.
	rootCAs := x509.NewCertPool()
	if !rootCAs.AppendCertsFromPEM(pem) {
		return nil, fmt.Errorf("no root CA certs parsed from file %q", certFile)
	}

	return grpc.WithTransportCredentials(
		credentials.NewTLS(&tls.Config{
			RootCAs:    rootCAs,
			ServerName: "localhost",
		}),
	), nil
}

func downloadsHandler(downloadPath string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Need to remove the rest of the URL path so that it just contains the file wanted
		r.URL.Path = filepath.Base(r.URL.Path)
		http.FileServer(http.Dir(downloadPath)).ServeHTTP(w, r)
	}
}
