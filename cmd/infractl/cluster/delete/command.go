// Package delete implements the infractl cluster delete command.
package delete

import (
	"context"
	"errors"

	"github.com/spf13/cobra"
	"github.com/stackrox/infra/cmd/infractl/common"
	v1 "github.com/stackrox/infra/generated/api/v1"
	"google.golang.org/grpc"
)

const examples = `# Delete cluster.
infractl cluster delete <cluster-id>

# Delete cluster example-s3maj.
infractl cluster delete example-s3maj`

// Command defines the handler for infractl cluster delete.
func Command() *cobra.Command {
	// $ infractl cluster delete
	return &cobra.Command{
		Use:     "delete CLUSTER",
		Short:   "Delete a specific cluster",
		Long:    "Deletes a specific cluster",
		Example: examples,
		Args:    common.ArgsWithHelp(cobra.ExactArgs(1), args),
		RunE:    common.WithGRPCHandler(delete),
	}
}

func args(_ *cobra.Command, args []string) error {
	clusterId := args[0]
	if clusterId == "" {
		return errors.New("no clusterid given")
	}

	return nil
}

func delete(ctx context.Context, conn *grpc.ClientConn, cmd *cobra.Command, args []string) (common.PrettyPrinter, error) {
	if len(args) != 1 {
		return nil, errors.New("invalid arguments")
	}

	req := v1.ResourceByID{
		Id: args[0],
	}

	_, err := v1.NewClusterServiceClient(conn).Delete(ctx, &req)
	if err != nil {
		return nil, err
	}

	return id(req), nil
}
