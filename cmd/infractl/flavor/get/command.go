// Package get implements the infractl flavor get command.
package get

import (
	"context"
	"errors"

	"github.com/spf13/cobra"
	"github.com/stackrox/infra/cmd/infractl/common"
	v1 "github.com/stackrox/infra/generated/api/v1"
	"google.golang.org/grpc"
)

const examples = `# Display info about the "gke-default" flavor.
$ infractl flavor get gke-default`

// Command defines the handler for infractl flavor get.
func Command() *cobra.Command {
	// $ infractl flavor get
	return &cobra.Command{
		Use:     "get FLAVOR",
		Short:   "Info on a specific flavor",
		Long:    "Displays info for a single flavor",
		Example: examples,
		Args:    common.ArgsWithHelp(cobra.ExactArgs(1), args),
		RunE:    common.WithGRPCHandler(run),
	}
}

func args(_ *cobra.Command, args []string) error {
	if args[0] == "" {
		return errors.New("no flavor ID given")
	}
	return nil
}

func run(ctx context.Context, conn *grpc.ClientConn, _ *cobra.Command, args []string) (common.PrettyPrinter, error) {
	resp, err := v1.NewFlavorServiceClient(conn).Info(ctx, &v1.ResourceByID{Id: args[0]})
	if err != nil {
		return nil, err
	}

	return &prettyFlavor{*resp}, nil
}
