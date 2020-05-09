// Package list implements the infractl list command.
package list

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/stackrox/infra/cmd/infractl/common"
	v1 "github.com/stackrox/infra/generated/api/v1"
	"google.golang.org/grpc"
)

const examples = `# List your clusters.
$ infractl list

# List your clusters, including ones that have expired.
$ infractl list --expired

# List everyone's' clusters.
$ infractl list --all`

// Command defines the handler for infractl list.
func Command() *cobra.Command {
	// $ infractl list
	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List clusters",
		Long:    "List the available clusters",
		Example: examples,
		Args:    common.ArgsWithHelp(cobra.ExactArgs(0)),
		RunE:    common.WithGRPCHandler(run),
	}

	cmd.Flags().Bool("all", false, "include clusters not owned by you")
	cmd.Flags().Bool("expired", false, "include expired clusters")
	return cmd
}

func run(ctx context.Context, conn *grpc.ClientConn, cmd *cobra.Command, _ []string) (common.PrettyPrinter, error) {
	includeAll, _ := cmd.Flags().GetBool("all")
	includeExpired, _ := cmd.Flags().GetBool("expired")

	req := v1.ClusterListRequest{
		All:     includeAll,
		Expired: includeExpired,
	}

	resp, err := v1.NewClusterServiceClient(conn).List(ctx, &req)
	if err != nil {
		return nil, err
	}

	return prettyClusterListResponse(*resp), nil
}
