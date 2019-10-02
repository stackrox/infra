package service

import (
	"context"
	"strings"
	"time"

	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/pkg/errors"
	"github.com/stackrox/infra/auth"
	"github.com/stackrox/infra/config"
	"github.com/stackrox/infra/generated/api/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
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

// cookieValues converts cookies stored in the gRPC-Gateway metadata into a more usable form.
func cookieValues(meta metadata.MD) map[string]string {
	cookies := meta.Get("grpcgateway-cookie")
	items := make(map[string]string)
	for _, cookie := range cookies {
		for _, part := range strings.Split(cookie, ";") {
			fields := strings.SplitN(part, "=", 2)
			switch len(fields) {
			case 2:
				items[strings.TrimSpace(fields[0])] = strings.TrimSpace(fields[1])
			case 1:
				items[strings.TrimSpace(fields[0])] = ""
			default:
				continue
			}
		}
	}
	return items
}

// UserEnricher enriches the given gRPC context with a v1.User struct, if
// possible. If there is no user, this function does not return an error, as
// anonymous API calls are a possibility. Authorization must be independently
// enforced.
func UserEnricher(cfg *config.Config) contextFunc {
	jwtUser := auth.NewUserTokenizer(time.Hour, cfg.Auth0.SessionKey)
	return func(ctx context.Context, _ *grpc.UnaryServerInfo) (context.Context, error) {

		// Extract request metadata (proxied http headers) from given context.
		meta, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, errors.New("failed to extract metadata from incoming context")
		}

		// Extract cookie key/value pairs from request metadata.
		items := cookieValues(meta)
		token, found := items["token"]
		if !found {
			return ctx, nil
		}

		// Validate the user JWT and extract the user and expiry properties.
		user, expiry, err := jwtUser.Validate(token)
		if err != nil {
			return ctx, nil
		}

		// Enrich the given context with the user and expiry.
		ts := &timestamp.Timestamp{Seconds: expiry.Unix()}
		return contextWithUser(ctx, user, ts), nil
	}
}

// EnforceAnonymousAccess eoforces authorization to API services. Specifically,
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

	// Service does not allow anonymous access. Check that an actual user is
	// accessing it.
	if _, _, found := UserFromContext(ctx); !found {
		return nil, status.Error(codes.PermissionDenied, "access denied")
	}

	return ctx, nil
}

type userContextKey struct{}

type userExpiryContextKey struct{}

// UserFromContext extracts a v1.User and the expiration date of that user from
// the given context, only if it contains one.
func UserFromContext(ctx context.Context) (*v1.User, *timestamp.Timestamp, bool) {
	userValue := ctx.Value(userContextKey{})
	if userValue == nil {
		return nil, nil, false
	}
	user := userValue.(*v1.User)

	expiryValue := ctx.Value(userExpiryContextKey{})
	if expiryValue == nil {
		return nil, nil, false
	}
	expiry := expiryValue.(*timestamp.Timestamp)

	return user, expiry, true
}

// contextWithUser returns the given context, but containing a v1.User and
// expiration date.
func contextWithUser(ctx context.Context, user *v1.User, expiry *timestamp.Timestamp) context.Context {
	ctx = context.WithValue(ctx, userExpiryContextKey{}, expiry)
	return context.WithValue(ctx, userContextKey{}, user)
}
