// Package reset implements the infractl status reset command.
package reset

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/stackrox/infra/cmd/infractl/common"
	"github.com/stackrox/infra/cmd/infractl/status"
	v1 "github.com/stackrox/infra/generated/api/v1"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

const examples = `# Resets server status.
$ infractl status reset`

// Command defines the handler for infractl status reset.
func Command() *cobra.Command {
	// $ infractl status reset
	return &cobra.Command{
		Use:     "reset",
		Short:   "Reset Server status information",
		Long:    "Reset server status",
		Example: examples,
		Args:    common.ArgsWithHelp(cobra.ExactArgs(0)),
		RunE:    common.WithGRPCHandler(run),
	}
}

func run(ctx context.Context, conn *grpc.ClientConn, _ *cobra.Command, _ []string) (common.PrettyPrinter, error) {
	updatedInfraStatus, err := v1.NewInfraStatusServiceClient(conn).ResetStatus(ctx, &emptypb.Empty{})
	if err != nil {
		return nil, err
	}
	return status.PrettyStatusResp{
		Status: updatedInfraStatus,
	}, nil
}
