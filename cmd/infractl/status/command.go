// Package status implements the infractl status command.
package status

import (
	"context"
	"fmt"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/spf13/cobra"
	"github.com/stackrox/infra/cmd/infractl/common"
	v1 "github.com/stackrox/infra/generated/api/v1"
	"google.golang.org/grpc"
)

const examples = `# Print server status.
$ infractl status`

// Command defines the handler for infractl status.
func Command() *cobra.Command {
	// $ infractl status
	return &cobra.Command{
		Use:     "status",
		Short:   "Server status information",
		Long:    "Print server status",
		Example: examples,
		Args:    common.ArgsWithHelp(cobra.ExactArgs(0)),
		RunE:    common.WithGRPCHandler(run),
	}
}

func run(ctx context.Context, conn *grpc.ClientConn, _ *cobra.Command, _ []string) (common.PrettyPrinter, error) {
	var infraStatus *v1.InfraStatus

	infraStatus, err := v1.NewInfraStatusServiceClient(conn).GetStatus(ctx, &empty.Empty{})

	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(infraStatus)

	return prettyStatusResp{
		Status: infraStatus,
	}, nil
}
