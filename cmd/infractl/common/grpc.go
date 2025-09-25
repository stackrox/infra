package common

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/backoff"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/keepalive"
)

// GetGRPCConnection gets a grpc connection to the infra-server with the correct auth.
func GetGRPCConnection() (*grpc.ClientConn, context.Context, func(), error) {
	// Enable gRPC debug logging
	if os.Getenv("GRPC_GO_LOG_VERBOSITY_LEVEL") == "" {
		os.Setenv("GRPC_GO_LOG_VERBOSITY_LEVEL", "99")
	}
	if os.Getenv("GRPC_GO_LOG_SEVERITY_LEVEL") == "" {
		os.Setenv("GRPC_GO_LOG_SEVERITY_LEVEL", "info")
	}

	// Set custom gRPC logger for better debugging
	grpclog.SetLoggerV2(grpclog.NewLoggerV2(os.Stdout, os.Stderr, os.Stderr))

	log.Printf("[DEBUG] Connecting to gRPC endpoint: %s, insecure: %t", endpoint(), insecure())

	ctx, cancel := ContextWithTimeout()
	allDialOpts := []grpc.DialOption{
		grpc.WithPerRPCCredentials(bearerToken(token())),
		// Add connection timeout to prevent hanging on handshake
		grpc.WithConnectParams(grpc.ConnectParams{
			Backoff: backoff.Config{
				BaseDelay:  1.0 * time.Second,
				Multiplier: 1.6,
				Jitter:     0.2,
				MaxDelay:   120 * time.Second,
			},
			MinConnectTimeout: 20 * time.Second,
		}),
		// Add keepalive settings for better connection management
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                10 * time.Second, // Send keepalive pings every 10 seconds
			Timeout:             5 * time.Second,  // Wait 5 seconds for ping ack before considering the connection dead
			PermitWithoutStream: true,             // Send pings even when no active streams
		}),
	}

	// The insecure flag (--insecure) was given.
	if insecure() {
		log.Printf("[DEBUG] Using insecure TLS configuration")
		allDialOpts = append(allDialOpts,
			grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{
				InsecureSkipVerify: true,
				// Add ALPN support for gRPC v1.67+ compatibility
				NextProtos: []string{"h2", "http/1.1"},
			})),
		)
	} else {
		log.Printf("[DEBUG] Using secure TLS configuration")
		allDialOpts = append(allDialOpts, grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{
			// Add ALPN support for gRPC v1.67+ compatibility
			NextProtos: []string{"h2", "http/1.1"},
		})))
	}

	// Dial our specified endpoint.
	log.Printf("[DEBUG] Creating gRPC client with options: %d dial options", len(allDialOpts))
	conn, err := grpc.NewClient(endpoint(), allDialOpts...)
	if err != nil {
		log.Printf("[ERROR] Failed to create gRPC client: %v", err)
		return nil, nil, cancel, err
	}

	// Log initial connection state
	state := conn.GetState()
	log.Printf("[DEBUG] Initial connection state: %v", state)

	// Start a goroutine to monitor connection state changes
	go func() {
		for {
			if !conn.WaitForStateChange(context.Background(), state) {
				break
			}
			newState := conn.GetState()
			log.Printf("[DEBUG] Connection state changed: %v -> %v", state, newState)
			state = newState

			// Log additional details for certain states
			switch newState {
			case connectivity.TransientFailure:
				log.Printf("[WARN] Connection in transient failure state")
			case connectivity.Ready:
				log.Printf("[INFO] Connection ready")
			case connectivity.Connecting:
				log.Printf("[INFO] Connection attempting to connect")
			case connectivity.Shutdown:
				log.Printf("[INFO] Connection shutdown")
				return
			}
		}
	}()

	// done cancels the underlying context, and closes the gRPC connection.
	done := func() {
		log.Printf("[DEBUG] Closing gRPC connection")
		cancel()
		if err := conn.Close(); err != nil {
			log.Printf("[ERROR] Error closing gRPC connection: %v", err)
			panic(err)
		}
	}

	return conn, ctx, done, err
}

// bearerToken implements the credentials.PerRPCCredentials interface, and sets
// a bearer token on the connection metadata.
type bearerToken string

var _ credentials.PerRPCCredentials = (*bearerToken)(nil)

func (t bearerToken) GetRequestMetadata(_ context.Context, _ ...string) (map[string]string, error) {
	// Avoid failure from the common mistake where tokens are stored with
	// trailing newlines or whitespace.
	trimmed := strings.TrimRight(string(t), "\r\n ")
	if strings.ContainsAny(trimmed, "\r\n ") {
		fmt.Fprintln(os.Stderr, "The auth token contains invalid characters")
		// To help debug issues with invalid tokens in automation, dump the
		// beginning and end. (infra tokens are typically > 300 chars, this
		// check on 100 is to ensure the entire token is not printed to logs if
		// they ever get <= 20 chars.)
		if len(trimmed) > 100 {
			fmt.Fprintf(os.Stderr, "begins: %s, end: %s\n", trimmed[0:10], trimmed[len(trimmed)-10:])
		}
		os.Exit(1)
	}
	return map[string]string{
		"authorization": "Bearer " + trimmed,
	}, nil
}

func (t bearerToken) RequireTransportSecurity() bool {
	return true
}
