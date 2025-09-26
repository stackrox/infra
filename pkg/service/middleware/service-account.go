package middleware

import (
	"context"
	"strings"

	"github.com/pkg/errors"
	v1 "github.com/stackrox/infra/generated/api/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type serviceAccountContextKey struct{}

// ServiceAccountEnricher enriches the given gRPC context with a
// v1.ServiceAccount struct, if possible. If there is no service account, this
// function does not return an error, as anonymous API calls are a possibility.
// Authorization must be independently enforced.
func ServiceAccountEnricher(validator func(string) (*v1.ServiceAccount, error)) contextFunc {
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

		// Validate the JWT
		svcacct, err := validator(token)
		if err != nil {
			return ctx, nil
		}

		return contextWithServiceAccount(ctx, svcacct), nil
	}
}

// ServiceAccountFromContext extracts a v1.ServiceAccount from the given
// context, if one exists.
func ServiceAccountFromContext(ctx context.Context) (*v1.ServiceAccount, bool) {
	svcacctValue := ctx.Value(serviceAccountContextKey{})
	if svcacctValue == nil {
		return nil, false
	}

	return svcacctValue.(*v1.ServiceAccount), true
}

// contextWithUser returns the given context enriched with a v1.ServiceAccount.
func contextWithServiceAccount(ctx context.Context, svcacct *v1.ServiceAccount) context.Context {
	return context.WithValue(ctx, serviceAccountContextKey{}, svcacct)
}

// bearer extracts a bearer token from the gRPC-Gateway metadata.
func bearer(meta metadata.MD) (string, bool) {
	headerValues := meta.Get("authorization")
	if len(headerValues) == 1 {
		return strings.TrimPrefix(headerValues[0], "Bearer "), true
	}
	return "", false
}
