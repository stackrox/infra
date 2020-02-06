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
	//manager  autocert.Manager
	cfg config.Config
	//handler   http.Handler
	//tlsConfig *tls.Config
	manager TLSManager

	oauth *auth.OAuth
}

func New(cfg config.Config, services ...middleware.APIService) (*server, error) {
	manager, err := NewTLSManager(cfg.Server)
	if err != nil {
		return nil, err
	}

	oauth, err := auth.NewOAuth(s.cfg.Auth0)
	if err != nil {
		panic(err)
	}

	return &server{
		services: services,
		cfg:      cfg,
		manager:  manager,
		oauth: oauth,
	}, nil
}

func shutdown(fns []func()) func() {
	return func() {
		for _, fn := range fns {
			fn()
		}
	}
}

func (s *server) RunServer() (func(), <-chan error, error) {
	var (
		grpcListenAddress  = addressAny + ":" + s.cfg.Server.GRPC
		grpcConnectAddress = addressLocalhost + ":" + s.cfg.Server.GRPC

		// errch is a channel of size 1 which will receive the (only) error
		// returned by the server.
		errch = make(chan error, 1)

		mux = http.NewServeMux()

		shutdowns []func()
	)

	mux.Handle("/", http.FileServer(http.Dir(s.cfg.Storage.StaticDir)))
	mux.Handle("/callback", http.HandlerFunc(s.oauth.CallbackHandler))
	mux.Handle("/login", http.HandlerFunc(s.oauth.LoginHandler))
	mux.Handle("/logout", http.HandlerFunc(s.oauth.LogoutHandler))

	/////////////////////////////////
	// Step 1 - Start HTTPS server //
	/////////////////////////////////

	shutdowns = append(shutdowns, func() {
		s.manager.Listener().Close()
	})
	log.Printf("starting HTTPS server in %s mode", s.manager.Name())
	go func() {
		if err := http.Serve(s.manager.Listener(), mux); err != nil {
			errch <- err
		}
	}()

	////////////////////////////////
	// Step 2 - Start gRPC server //
	////////////////////////////////

	listen, err := net.Listen("tcp", grpcListenAddress)
	if err != nil {
		return shutdown(shutdowns), nil, err
	}
	shutdowns = append(shutdowns, func() {
		listen.Close()
	})

	// Create the server.
	server := grpc.NewServer(
		s.manager.ServerOption(),

		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			// Extract user from JWT token stored in HTTP cookie.
			middleware.ContextInterceptor(middleware.UserEnricher(s.cfg)),
			// Extract service-account from token stored in Authorization header.
			middleware.ContextInterceptor(middleware.ServiceAccountEnricher(s.cfg)),
			// Enforce authenticated user access on resources that declare it.
			middleware.ContextInterceptor(middleware.EnforceAnonymousAccess),
		)),
	)

	// Register the gRPC API service.
	for _, apiSvc := range s.services {
		apiSvc.RegisterServiceServer(server)
	}

	shutdowns = append(shutdowns, func() {
		server.Stop()
	})
	log.Print("starting gRPC server")
	go func() {
		if err := server.Serve(listen); err != nil {
			errch <- err
		}
	}()

	///////////////////////////////////////////
	// Step 3 - Register gRPC-Gateway routes //
	///////////////////////////////////////////

	log.Print("starting gRPC-Gateway client")
	conn, err := grpc.Dial(grpcConnectAddress, s.manager.DialOptions()...)
	if err != nil {
		return shutdown(shutdowns), nil, errors.Wrap(err, "dialing GRPC")
	}
	shutdowns = append(shutdowns, func() {
		conn.Close()
	})

	// Register each service
	gwMux := runtime.NewServeMux(
		runtime.WithMarshalerOption("*", &runtime.JSONPb{Indent: "  "}),
	)

	ctx, ctxcancel := context.WithCancel(context.Background())
	shutdowns = append(shutdowns, ctxcancel)

	for _, apiSvc := range s.services {
		if err := apiSvc.RegisterServiceHandler(ctx, gwMux, conn); err != nil {
			return shutdown(shutdowns), nil, err
		}
	}

	// Updates http handler routes. This included "web-only" routes, like
	// login/logout/static, and now includes gRPC-Gateway routes.
	mux.Handle("/v1/", gwMux)

	return shutdown(shutdowns), errch, nil
}
