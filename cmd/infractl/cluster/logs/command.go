// Package logs implements the infractl logs command.
package logs

import (
	"context"
	"errors"

	"github.com/spf13/cobra"
	"github.com/stackrox/infra/cmd/infractl/cluster/utils"
	"github.com/stackrox/infra/cmd/infractl/common"
	v1 "github.com/stackrox/infra/generated/api/v1"
	"google.golang.org/grpc"
)

const examples = `Lookup logs for the "example-s3maj" cluster.
$ infractl logs example-s3maj`

// Command defines the handler for infractl logs.
func Command() *cobra.Command {
	// $ infractl logs
	return &cobra.Command{
		Use:     "logs CLUSTER",
		Short:   "Get logs for a specific cluster",
		Long:    "Displays logs for a single cluster",
		Example: examples,
		Args:    common.ArgsWithHelp(cobra.ExactArgs(1), args),
		RunE:    common.WithGRPCHandler(run),
	}
}

func args(_ *cobra.Command, args []string) error {
	if args[0] == "" {
		return errors.New("no cluster ID given")
	}
	return utils.ValidateClusterName(args[0])
}

func run(ctx context.Context, conn *grpc.ClientConn, _ *cobra.Command, args []string) (common.PrettyPrinter, error) {
	req := v1.ResourceByID{Id: args[0]}

	resp, err := v1.NewClusterServiceClient(conn).Logs(ctx, &req)
	if err != nil {
		return nil, err
	}

	return prettyLogsResponse(*resp), nil
}
