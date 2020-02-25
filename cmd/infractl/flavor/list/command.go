// Package list implements the infractl flavor list command.
package list

import (
	"context"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/spf13/cobra"
	"github.com/stackrox/infra/cmd/infractl/common"
	v1 "github.com/stackrox/infra/generated/api/v1"
	"google.golang.org/grpc"
)

const examples = `# List all flavors.
$ infractl flavor list`

// Command defines the handler for infractl flavor list.
func Command() *cobra.Command {
	// $ infractl flavor list
	return &cobra.Command{
		Use:     "list",
		Short:   "List flavors",
		Long:    "List the available flavors",
		Example: examples,
		Args:    common.ArgsWithHelp(cobra.ExactArgs(0)),
		RunE:    common.WithGRPCHandler(run),
	}
}

func run(ctx context.Context, conn *grpc.ClientConn, _ *cobra.Command, _ []string) (common.PrettyPrinter, error) {
	resp, err := v1.NewFlavorServiceClient(conn).List(ctx, &empty.Empty{})
	if err != nil {
		return nil, err
	}

	return prettyFlavorListResponse(*resp), nil
}
