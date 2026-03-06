// Package wait implements the infractl wait command.
package wait

import (
	"context"
	"errors"

	"github.com/spf13/cobra"
	"github.com/stackrox/infra/cmd/infractl/cluster/utils"
	"github.com/stackrox/infra/cmd/infractl/common"
	v1 "github.com/stackrox/infra/generated/api/v1"
	"google.golang.org/grpc"
)

const examples = `Wait for the "example-s3maj" cluster to become ready.
$ infractl wait example-s3maj`

// Command defines the handler for infractl wait.
func Command() *cobra.Command {
	// $ infractl wait
	cmd := &cobra.Command{
		Use:     "wait CLUSTER",
		Short:   "Wait for a specific cluster",
		Long:    "Wait for the the specific cluster to become ready.",
		Example: examples,
		Args:    common.ArgsWithHelp(cobra.ExactArgs(1), args),
		RunE:    common.WithGRPCHandler(run),
	}

	common.AddMaxWaitErrorsFlag(cmd)

	return cmd
}

func args(_ *cobra.Command, args []string) error {
	if args[0] == "" {
		return errors.New("no cluster ID given")
	}
	return utils.ValidateClusterName(args[0])
}

func run(_ context.Context, conn *grpc.ClientConn, cmd *cobra.Command, args []string) (common.PrettyPrinter, error) {
	maxWaitErrors := common.GetMaxWaitErrorsFlagValue(cmd)

	client := v1.NewClusterServiceClient(conn)
	err := common.WaitForCluster(client, &v1.ResourceByID{Id: args[0]}, maxWaitErrors)

	return prettyNoop{}, err
}
