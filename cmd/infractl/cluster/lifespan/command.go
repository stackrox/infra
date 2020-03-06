// Package lifespan implements the infractl cluster lifespan command.
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

const examples = `# Set the lifespan of cluster example-s3maj to 30 minutes.
infractl cluster lifespan example-s3maj 30m

# Expire cluster example-s3maj.
infractl cluster lifespan example-s3maj 0`

// Command defines the handler for infractl cluster lifespan.
func Command() *cobra.Command {
	// $ infractl cluster lifespan
	return &cobra.Command{
		Use:     "lifespan CLUSTER DURATION",
		Short:   "update cluster lifespan",
		Long:    "lifespan updates the cluster lifespan",
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
	lifespan, err := time.ParseDuration(args[1])
	if err != nil {
		return nil, err
	}

	resp, err := v1.NewClusterServiceClient(conn).Lifespan(ctx, &v1.LifespanRequest{
		Id:       args[0],
		Lifespan: ptypes.DurationProto(lifespan),
	})
	if err != nil {
		return nil, err
	}

	return prettyDuration(*resp), nil
}
