package middleware

import (
	"context"

	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type adminContextKey struct{}

// AdminEnricher enriches the given gRPC context with an admin, if possible. If
// the caller is not an admin, this function does not return an error, as
// anonymous API calls are a possibility. Authorization must be independently
// enforced.
func AdminEnricher(password string) contextFunc {
	return func(ctx context.Context, _ *grpc.UnaryServerInfo) (context.Context, error) {
		// Extract request metadata (proxied http headers) from given context.
		meta, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, errors.New("failed to extract metadata from incoming context")
		}

		// Extract the bearer token value from the Authorization header.
		token, found := bearer(meta)
		if !found {
			return ctx, nil
		}

		// Check if the given token matches the admin password.
		if token != password {
			return ctx, nil
		}

		return contextWithAdmin(ctx), nil
	}
}

// AdminInContext determines if an admin value is in the given context.
func AdminInContext(ctx context.Context) bool {
	_, ok := ctx.Value(adminContextKey{}).(struct{})
	return ok
}

// contextWithAdmin returns the given context enriched as an admin.
func contextWithAdmin(ctx context.Context) context.Context {
	return context.WithValue(ctx, adminContextKey{}, struct{}{})
}
