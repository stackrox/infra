// Package server facilitates running a HTTPS and gRPC server.
package server

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"os"
	"strings"

	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/pkg/errors"
	"github.com/stackrox/infra/auth"
	"github.com/stackrox/infra/config"
	"github.com/stackrox/infra/pkg/logging"
	"github.com/stackrox/infra/service/middleware"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

var log = logging.CreateProductionLogger()

type server struct {
	services []middleware.APIService
	cfg      config.Config
	oidc     auth.OidcAuth
}

// New creates a new server that is ready to be launched.
func New(serverCfg config.Config, oidc auth.OidcAuth, services ...middleware.APIService) *server {
	return &server{
		services: services,
		cfg:      serverCfg,
		oidc:     oidc,
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
			middleware.ContextInterceptor(middleware.UserEnricher(s.oidc)),
			// Extract service-account from token stored in Authorization header.
			middleware.ContextInterceptor(middleware.ServiceAccountEnricher(s.oidc.ValidateServiceAccountToken)),

			middleware.ContextInterceptor(middleware.AdminEnricher(s.cfg.Password)),
			// Enforce authenticated user access on resources that declare it.
			middleware.ContextInterceptor(middleware.EnforceAccess),

			// Collect and expose Prometheus metrics
			grpc_prometheus.UnaryServerInterceptor,
		)),
		grpc.StreamInterceptor(
			// Collect and expose Prometheus metrics
			grpc_prometheus.StreamServerInterceptor,
		),
	)

	// Register the gRPC API service.
	for _, apiSvc := range s.services {
		apiSvc.RegisterServiceServer(server)
	}

	grpc_prometheus.Register(server)

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

	log.Log(logging.INFO, "starting gRPC server", "listen-address", listenAddress)
	go func() {
		if err := http.ListenAndServeTLS(listenAddress, s.cfg.Server.CertFile, s.cfg.Server.KeyFile, h2c.NewHandler(muxHandler, &http2.Server{})); err != nil {
			errCh <- err
		}
	}()

	dialOption, err := grpcLocalCredentials(s.cfg.Server.CertFile)
	if err != nil {
		return nil, err
	}

	log.Log(logging.INFO, "starting gRPC-Gateway client", "connect-address", connectAddress)
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

	routeMux := http.NewServeMux()

	// Updates http handler routes. This included "web-only" routes, like
	// login/logout/static, and also gRPC-Gateway routes.
	routeMux.Handle("/", serveApplicationResources(s.cfg.Server.StaticDir, s.oidc))
	routeMux.Handle("/v1/", gwMux)
	routeMux.Handle("/metrics", promhttp.Handler())
	s.oidc.Handle(routeMux)

	mux.Handle("/",
		wrapHealthCheck(
			wrapCanonicalRedirect(
				s.oidc.Endpoint(),
				routeMux,
			),
		),
	)

	return errCh, nil
}

// serveApplicationResources handles requests for SPA endpoints as well as
// regular resources.
func serveApplicationResources(dir string, oidc auth.OidcAuth) http.Handler {
	type rule struct {
		path      string
		spa       bool
		anonymous bool
		prefix    bool
	}

	// List of path rules, roughly ordered from most-likely matched to
	// least-likely matched.
	rules := []rule{
		{
			path:   "/static/",
			prefix: true,
		},
		{
			path:      "/manifest.json",
			anonymous: true,
		},
		{
			path:      "/favicon.ico",
			anonymous: true,
		},
		{
			path:      "/logout-page.html",
			anonymous: true,
		},
		{
			path:   "/downloads/",
			prefix: true,
		},
		{
			path:   "/",
			spa:    true,
			prefix: true,
		},
	}

	fs := http.FileServer(http.Dir(dir))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestPath := r.URL.Path

		for _, rule := range rules {
			if rule.prefix {
				// If this rule is supposed to match a path prefix, and that
				// path prefix isn't matched, move onto the next rule.
				if !strings.HasPrefix(requestPath, rule.path) {
					continue
				}
			} else {
				// If this rule is supposed to match a path exactly, and that
				// path isn't exactly matched, move onto the next rule.
				if requestPath != rule.path {
					continue
				}
			}

			// If the path is a path in the SPA, set the path to be the root,
			// so that the index.html is served.
			if rule.spa {
				r.URL.Path = "/"
			}

			if rule.anonymous {
				// Serve this path anonymously (without any authentication).
				fs.ServeHTTP(w, r)
			} else {
				// Serve this path with authentication.
				oidc.Authorized(fs).ServeHTTP(w, r)
			}

			return
		}
		// No rules matched, so serve this path with authentication by default.
		oidc.Authorized(fs).ServeHTTP(w, r)
	})
}

// wrapCanonicalRedirect redirects proxied requests for non-canonical endpoints
// to the canonical endpoint.
//
// Examples:
//
//	http://example.com      --> https://example.com (non https)
//	https://old.example.com --> https://example.com (CNAME)
func wrapCanonicalRedirect(endpoint string, wrapped http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		headerXForwardedFor := r.Header.Get("X-Forwarded-For")
		headerXForwardedProto := r.Header.Get("X-Forwarded-Proto")
		headerVia := r.Header.Get("Via")

		// It doesn't appear that this request came through an Ingress, process
		// the request normally.
		if headerVia == "" || headerXForwardedFor == "" || headerXForwardedProto == "" {
			wrapped.ServeHTTP(w, r)
			return
		}

		// Compare the endpoint the browser thinks it's talking with to the
		// endpoint it should be talking with. If the match, process the
		// request normally.
		requestEndpoint := headerXForwardedProto + "://" + r.Host
		canonicalEndpoint := "https://" + endpoint
		if requestEndpoint == canonicalEndpoint {
			wrapped.ServeHTTP(w, r)
			return
		}

		// There was a mismatch, so redirect to the canonical URL.
		http.Redirect(w, r, canonicalEndpoint, http.StatusMovedPermanently)
	})
}

// wrapHealthCheck determines if the given request is a health check, and
// responds appropriately with a 200 OK status code.
func wrapHealthCheck(wrapped http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasPrefix(r.UserAgent(), "kube-probe"):
			// Kubernetes internal service health check.
			w.WriteHeader(http.StatusOK)

		case strings.HasPrefix(r.UserAgent(), "GoogleHC"):
			// GCP backend health check.
			w.WriteHeader(http.StatusOK)

		default:
			wrapped.ServeHTTP(w, r)
		}
	})
}

func grpcLocalCredentials(certFile string) (grpc.DialOption, error) {
	// Read the x509 PEM public certificate file
	pem, err := os.ReadFile(certFile)
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
