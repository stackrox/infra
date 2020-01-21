package middleware

import (
	"context"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/stackrox/infra/auth"
	"github.com/stackrox/infra/config"
	v1 "github.com/stackrox/infra/generated/api/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type userContextKey struct{}

// UserEnricher enriches the given gRPC context with a v1.User struct, if
// possible. If there is no user, this function does not return an error, as
// anonymous API calls are a possibility. Authorization must be independently
// enforced.
func UserEnricher(cfg config.Config) contextFunc {
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
		user, err := jwtUser.Validate(token)
		if err != nil {
			return ctx, nil
		}

		// Enrich the given context with the user.
		return contextWithUser(ctx, user), nil
	}
}

// UserFromContext extracts a v1.User from the given context, if one exists.
func UserFromContext(ctx context.Context) (*v1.User, bool) {
	userValue := ctx.Value(userContextKey{})
	if userValue == nil {
		return nil, false
	}

	return userValue.(*v1.User), true
}

// contextWithUser returns the given context enriched with a v1.User.
func contextWithUser(ctx context.Context, user *v1.User) context.Context {
	return context.WithValue(ctx, userContextKey{}, user)
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
