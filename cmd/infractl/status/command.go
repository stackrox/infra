// Package version implements the infractl version command.
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

// Command defines the handler for infractl version.
func Command() *cobra.Command {
	// $ infractl version
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

	// Attempt to get the server version if possible. If not, then continue
	// normal operation, and ignore any errors.
	infraStatus, _ = v1.NewInfraStatusServiceClient(conn).GetStatus(ctx, &empty.Empty{})

	fmt.Println(infraStatus)

	return prettyStatusResp{
		Status: infraStatus,
	}, nil
}
