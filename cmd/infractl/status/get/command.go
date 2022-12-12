// Package get implements the infractl status get command.
package get

import (
	"context"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/spf13/cobra"
	"github.com/stackrox/infra/cmd/infractl/common"
	"github.com/stackrox/infra/cmd/infractl/status"
	v1 "github.com/stackrox/infra/generated/api/v1"
	"google.golang.org/grpc"
)

const examples = `# Print server status.
$ infractl status get`

// Command defines the handler for infractl status get.
func Command() *cobra.Command {
	// $ infractl status get
	return &cobra.Command{
		Use:     "get",
		Short:   "Server status information",
		Long:    "Print server status",
		Example: examples,
		Args:    common.ArgsWithHelp(cobra.ExactArgs(0)),
		RunE:    common.WithGRPCHandler(run),
	}
}

func run(ctx context.Context, conn *grpc.ClientConn, _ *cobra.Command, _ []string) (common.PrettyPrinter, error) {
	infraStatus, err := v1.NewInfraStatusServiceClient(conn).GetStatus(ctx, &empty.Empty{})
	if err != nil {
		return nil, err
	}

	return status.PrettyStatusResp{
		Status: infraStatus,
	}, nil
}
