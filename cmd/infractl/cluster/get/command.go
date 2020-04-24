// Package get implements the infractl info command.
package get

import (
	"context"
	"errors"

	"github.com/spf13/cobra"
	"github.com/stackrox/infra/cmd/infractl/common"
	v1 "github.com/stackrox/infra/generated/api/v1"
	"google.golang.org/grpc"
)

const examples = `Lookup info for the "example-s3maj" cluster.
$ infractl get example-s3maj`

// Command defines the handler for infractl get.
func Command() *cobra.Command {
	// $ infractl get
	return &cobra.Command{
		Use:     "get CLUSTER",
		Short:   "Get info for a specific cluster",
		Long:    "Displays info for a single cluster",
		Example: examples,
		Args:    common.ArgsWithHelp(cobra.ExactArgs(1), args),
		RunE:    common.WithGRPCHandler(run),
	}
}

func args(_ *cobra.Command, args []string) error {
	if args[0] == "" {
		return errors.New("no cluster ID given")
	}
	return nil
}

func run(ctx context.Context, conn *grpc.ClientConn, _ *cobra.Command, args []string) (common.PrettyPrinter, error) {
	resp, err := v1.NewClusterServiceClient(conn).Info(ctx, &v1.ResourceByID{Id: args[0]})
	if err != nil {
		return nil, err
	}

	return prettyCluster(*resp), nil
}
