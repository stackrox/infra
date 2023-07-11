package common

import (
	"context"
	"crypto/tls"
	"fmt"
	"os"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// GetGRPCConnection gets a grpc connection to the infra-server with the correct auth.
func GetGRPCConnection() (*grpc.ClientConn, context.Context, func(), error) {
	ctx, cancel := ContextWithTimeout()
	allDialOpts := []grpc.DialOption{
		grpc.WithPerRPCCredentials(bearerToken(token())),
	}

	// The insecure flag (--insecure) was given.
	if insecure() {
		allDialOpts = append(allDialOpts,
			grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{
				InsecureSkipVerify: true,
			})),
		)
	} else {
		allDialOpts = append(allDialOpts, grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{})))
	}

	// Dial our specified endpoint.
	conn, err := grpc.DialContext(ctx, endpoint(), allDialOpts...)
	if err != nil {
		return nil, nil, cancel, err
	}

	// done cancels the underlying context, and closes the gRPC connection.
	done := func() {
		cancel()
		if err := conn.Close(); err != nil {
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
