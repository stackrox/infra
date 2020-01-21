package common

import (
	"context"

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
		allDialOpts = append(allDialOpts, grpc.WithInsecure())
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

func (t bearerToken) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	return map[string]string{
		"authorization": "Bearer " + string(t),
	}, nil
}

func (t bearerToken) RequireTransportSecurity() bool {
	return false
}
