package middleware

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// contextFunc represents a function that can be used for processing and
// transforming a context as part of a gRPC interceptor.
type contextFunc func(ctx context.Context, info *grpc.UnaryServerInfo) (context.Context, error)

// ContextInterceptor enables the interception and transformation of a gRPC context.
func ContextInterceptor(ctxFunc contextFunc) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		newCtx, err := ctxFunc(ctx, info)
		if err != nil {
			return nil, err
		}

		return handler(newCtx, req)
	}
}

// EnforceAnonymousAccess enforces authorization to API services. Specifically,
// if a service declares that it is allowed to be accessed anonymously, access
// is allowed always. If the service does not permit anonymous access, a
// v1.User must exist in the given context for access to be allowed.
func EnforceAnonymousAccess(ctx context.Context, info *grpc.UnaryServerInfo) (context.Context, error) {
	// Convert to a service.
	svc, ok := info.Server.(APIService)
	if !ok {
		return nil, status.Error(codes.Internal, "failed to cast to apiservice")
	}

	// Does this service allow anonymous access?
	if svc.AllowAnonymous() {
		return ctx, nil
	}

	// Service does not allow anonymous access. Check if an actual user is
	// accessing it.
	if _, found := UserFromContext(ctx); !found {
		// No user found, check if a service account is accessing it.
		if _, found := ServiceAccountFromContext(ctx); !found {
			return nil, status.Error(codes.PermissionDenied, "access denied")
		}
	}

	return ctx, nil
}
