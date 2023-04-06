// Package list implements the infractl flavor list command.
package list

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/stackrox/infra/cmd/infractl/common"
	v1 "github.com/stackrox/infra/generated/proto/api/v1"
	"google.golang.org/grpc"
)

const examples = `# List all flavors except alpha ones.
$ infractl flavor list

# List all flavors.
$ infractl flavor list --all`

// Command defines the handler for infractl flavor list.
func Command() *cobra.Command {
	// $ infractl flavor list
	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List flavors",
		Long:    "List the available flavors",
		Example: examples,
		Args:    common.ArgsWithHelp(cobra.ExactArgs(0)),
		RunE:    common.WithGRPCHandler(run),
	}

	cmd.Flags().Bool("all", false, "include alpha flavors")
	return cmd
}

func run(ctx context.Context, conn *grpc.ClientConn, cmd *cobra.Command, _ []string) (common.PrettyPrinter, error) {
	includeAll := common.MustBool(cmd.Flags(), "all")

	req := v1.FlavorListRequest{
		All: includeAll,
	}

	resp, err := v1.NewFlavorServiceClient(conn).List(ctx, &req)
	if err != nil {
		return nil, err
	}

	return prettyFlavorListResponse{resp}, nil
}
