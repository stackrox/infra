package server

import (
	"context"
	"log"
	"net"
	"net/http"
	"strings"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/pkg/errors"
	"github.com/stackrox/infra/config"
	"github.com/stackrox/infra/service/middleware"
	"golang.org/x/crypto/acme/autocert"
	"google.golang.org/grpc"
)

// RunGRPCServer runs the gRPC service.
func RunGRPCServer(apiServices []middleware.APIService, cfg *config.Config) (func(), <-chan error, error) {
	listen, err := net.Listen("tcp", cfg.Server.GRPC)
	if err != nil {
		return nil, nil, err
	}

	// Create the server.
	server := grpc.NewServer(
		// Wire up context interceptors.
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			// Extract user from JWT token stored in HTTP cookie.
			middleware.ContextInterceptor(middleware.UserEnricher(cfg)),
			// Extract service-account from token stored in Authorization header.
			middleware.ContextInterceptor(middleware.ServiceAccountEnricher(cfg)),
			// Enforce authenticated user access on resources that declare it.
			middleware.ContextInterceptor(middleware.EnforceAnonymousAccess),
		)),
	)

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
		log.Printf("starting gRPC server on %s", cfg.Server.GRPC)
		if err := server.Serve(listen); err != nil {
			errch <- err
		}
	}()

	return shutdown, errch, nil
}

// RunHTTPServer runs the HTTP/REST gateway
func RunHTTPServer(apiServices []middleware.APIService, cfg *config.Config) (func(), <-chan error, error) {
	// Register the HTTP/gRPC gateway service.
	ctx := context.Background()
	conn, err := grpc.Dial(cfg.Server.GRPC, grpc.WithInsecure())
	if err != nil {
		return nil, nil, errors.Wrap(err, "dialing GRPC")
	}

	// Register each service
	gwMux := runtime.NewServeMux()
	for _, apiSvc := range apiServices {
		if err := apiSvc.RegisterServiceHandler(ctx, gwMux, conn); err != nil {
			return nil, nil, err
		}
	}

	mux := buildRoutes(gwMux, cfg)
	shutdown, errch := startHTTP(ctx, mux, cfg)
	return shutdown, errch, nil
}

func startHTTP(ctx context.Context, handler http.Handler, cfg *config.Config) (func(), <-chan error) {
	// errch is a channel of size 1 which will receive the (only) error
	// returned by the HTTP server.
	errch := make(chan error, 1)

	switch {
	case cfg.Server.CertFile != "" && cfg.Server.KeyFile != "":
		// If a cert and key files are configured, start both the http and https servers.
		log.Print("starting HTTP+HTTPS server in local certificate mode")

		httpServer := http.Server{
			Addr:    cfg.Server.HTTP,
			Handler: handlerRedirectToHTTPS(cfg.Server.HTTPS),
		}

		httpsServer := http.Server{
			Addr:    cfg.Server.HTTPS,
			Handler: handler,
		}

		shutdown := func() {
			log.Print("shutting down HTTP server")
			httpServer.Shutdown(ctx) // nolint:errcheck
			log.Print("shutting down HTTPS server")
			httpsServer.Shutdown(ctx) // nolint:errcheck
		}

		go func() {
			log.Printf("starting HTTP server on %s", httpServer.Addr)
			if err := httpServer.ListenAndServe(); err != nil {
				errch <- err
			}
		}()

		go func() {
			log.Printf("starting HTTPS server on %s", httpsServer.Addr)
			if err := httpsServer.ListenAndServeTLS(cfg.Server.CertFile, cfg.Server.KeyFile); err != nil {
				errch <- err
			}
		}()

		return shutdown, errch

	case cfg.Server.HTTPS != "" && cfg.Server.Domain != "":
		// If https and a domain is configured, start both the http and https servers with Let’s Encrypt.
		log.Print("starting HTTP+HTTPS server in Let’s Encrypt mode")

		certManager := autocert.Manager{
			Cache:      autocert.DirCache(cfg.Storage.CertDir),
			HostPolicy: autocert.HostWhitelist(cfg.Server.Domain),
			Prompt:     autocert.AcceptTOS,
		}

		httpServer := http.Server{
			Addr:    cfg.Server.HTTP,
			Handler: certManager.HTTPHandler(nil),
		}

		httpsServer := http.Server{
			Addr:      cfg.Server.HTTPS,
			Handler:   handler,
			TLSConfig: certManager.TLSConfig(),
		}

		shutdown := func() {
			log.Print("shutting down HTTP server")
			httpServer.Shutdown(ctx) // nolint:errcheck
			log.Print("shutting down HTTPS server")
			httpsServer.Shutdown(ctx) // nolint:errcheck
		}

		go func() {
			log.Printf("starting HTTP server on %s", httpServer.Addr)
			if err := httpServer.ListenAndServe(); err != nil {
				errch <- err
			}
		}()

		go func() {
			log.Printf("starting HTTPS server on %s (%s)", httpsServer.Addr, cfg.Server.Domain)
			if err := httpsServer.ListenAndServeTLS("", ""); err != nil {
				errch <- err
			}
		}()

		return shutdown, errch

	default:
		// Otherwise, only start the http server.
		log.Print("Starting HTTP server only")

		httpServer := http.Server{
			Addr:    cfg.Server.HTTP,
			Handler: handler,
		}

		shutdown := func() {
			log.Print("shutting down HTTP server")
			httpServer.Shutdown(ctx) // nolint:errcheck
		}

		go func() {
			log.Printf("starting HTTP server on %s", httpServer.Addr)
			if err := httpServer.ListenAndServe(); err != nil {
				errch <- err
			}
		}()

		return shutdown, errch
	}
}

// handlerRedirectToHTTPS returns a http.Handler that redirects requests to use HTTPS.
func handlerRedirectToHTTPS(httpsEndpoint string) http.Handler {
	httpsPort := "443"
	chunks := strings.SplitN(httpsEndpoint, ":", 2)
	if len(chunks) == 2 {
		httpsPort = chunks[1]
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" && r.Method != "HEAD" {
			http.Error(w, "Use HTTPS", http.StatusBadRequest)
			return
		}

		host := strings.SplitN(r.Host, ":", 2)[0]
		target := "https://" + host + ":" + httpsPort + r.URL.RequestURI()
		http.Redirect(w, r, target, http.StatusFound)
	})
}
