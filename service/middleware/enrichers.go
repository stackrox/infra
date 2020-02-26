// Package middleware provides functionality for instrumenting and enriching
// grpc connections.
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

// EnforceAccess enforces authorization to API services. Specifically,
// if a service declares that it is allowed to be accessed anonymously, access
// is allowed always. If the service does not permit anonymous access, a
// v1.User must exist in the given context for access to be allowed.
func EnforceAccess(ctx context.Context, info *grpc.UnaryServerInfo) (context.Context, error) {
	// Convert to a service.
	svc, ok := info.Server.(APIService)
	if !ok {
		return nil, status.Error(codes.Internal, "failed to cast to apiservice")
	}

	access := getAccess(ctx)

	if isAccessAllowed(info.FullMethod, svc.Access(), access) {
		return ctx, nil
	}

	// There is no authenticated principal, deny access!
	return nil, status.Error(codes.PermissionDenied, "access denied")
}

func getAccess(ctx context.Context) Access {
	// Check if an authenticated user is accessing the service.
	if _, found := UserFromContext(ctx); found {
		return Authenticated
	}

	// Check if an authenticated service account is accessing the service.
	if _, found := ServiceAccountFromContext(ctx); found {
		return Authenticated
	}

	// Check if an authenticated service account is accessing the service.
	if AdminFromContext(ctx) {
		return Admin
	}

	// Fall back to anonymous access.
	return Anonymous
}

func isAccessAllowed(method string, policy map[string]Access, access Access) bool {
	required, found := policy[method]
	if !found {
		return false
	}

	// have                  | Admin | Authenticated | Anonymous |
	// require Admin         | allow | deny          | deny      |
	// require Authenticated | deny  | allow         | deny      |
	// require Anonymous     | deny  | allow         | allow     |
	switch required {
	case Admin:
		return access == Admin
	case Authenticated:
		return access == Authenticated
	case Anonymous:
		return access == Anonymous || access == Authenticated
	default:
		panic("unknown required access level")
	}
}
