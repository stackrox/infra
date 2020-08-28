// Package lifespan implements the infractl lifespan command.
package lifespan

import (
	"context"
	"errors"
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/spf13/cobra"
	"github.com/stackrox/infra/cmd/infractl/common"
	v1 "github.com/stackrox/infra/generated/api/v1"
	"google.golang.org/grpc"
)

const examples = `# Add an hour to cluster example-s3maj.
infractl lifespan example-s3maj 1h

# OR
infractl lifespan example-s3maj +1h

# Set the lifespan of cluster example-s3maj to 24h.
infractl lifespan example-s3maj =24h

# Expire cluster example-s3maj.
infractl lifespan example-s3maj =0`

// Command defines the handler for infractl lifespan.
func Command() *cobra.Command {
	// $ infractl lifespan
	return &cobra.Command{
		Use:     "lifespan CLUSTER DURATION",
		Short:   "Update cluster lifespan",
		Long:    "Lifespan updates the cluster lifespan",
		Example: examples,
		Args:    common.ArgsWithHelp(cobra.ExactArgs(2), args),
		RunE:    common.WithGRPCHandler(run),
	}
}

func args(_ *cobra.Command, args []string) error {
	if args[0] == "" {
		return errors.New("no cluster ID given")
	}
	if args[1] == "" {
		return errors.New("no duration given")
	}
	return nil
}

func run(ctx context.Context, conn *grpc.ClientConn, _ *cobra.Command, args []string) (common.PrettyPrinter, error) {
	method, lifespan, err := parseDuration(args[1])
	if err != nil {
		return nil, err
	}

	resp, err := v1.NewClusterServiceClient(conn).Lifespan(ctx, &v1.LifespanRequest{
		Id:       args[0],
		Lifespan: ptypes.DurationProto(lifespan),
		Method:   method,
	})
	if err != nil {
		return nil, err
	}

	return prettyDuration{resp}, nil
}

func parseDuration(spec string) (v1.LifespanRequest_Method, time.Duration, error) {
	if spec == "expire" {
		return v1.LifespanRequest_REPLACE, 0, nil
	}

	method := v1.LifespanRequest_ADD
	switch spec[0] {
	case '+':
		// Spec indicates that we're adding a duration, like "+5m".
		method = v1.LifespanRequest_ADD
		spec = spec[1:]
	case '-':
		// Spec indicates that we're subtracting a duration, like "-5m".
		method = v1.LifespanRequest_SUBTRACT
		spec = spec[1:]
	case '=':
		// Spec indicates that we're replacing the duration, like "=5m".
		method = v1.LifespanRequest_REPLACE
		spec = spec[1:]
	}

	// Parse the remaining spec duration.
	duration, err := time.ParseDuration(spec)
	return method, duration, err
}
