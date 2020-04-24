// Package delete implements the infractl delete command.
package delete

import (
	"context"
	"errors"

	"github.com/spf13/cobra"
	"github.com/stackrox/infra/cmd/infractl/common"
	v1 "github.com/stackrox/infra/generated/api/v1"
	"google.golang.org/grpc"
)

const examples = `# Delete cluster "example-s3maj".
infractl delete example-s3maj`

// Command defines the handler for infractl delete.
func Command() *cobra.Command {
	// $ infractl delete
	return &cobra.Command{
		Use:     "delete CLUSTER",
		Short:   "Delete a specific cluster",
		Long:    "Deletes a specific cluster",
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

func run(ctx context.Context, conn *grpc.ClientConn, cmd *cobra.Command, args []string) (common.PrettyPrinter, error) {
	req := v1.ResourceByID{
		Id: args[0],
	}

	if _, err := v1.NewClusterServiceClient(conn).Delete(ctx, &req); err != nil {
		return nil, err
	}

	return id(req), nil
}
