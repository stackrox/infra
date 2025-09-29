// Package whoami implements the infractl whoami command.
package whoami

import (
	"context"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/spf13/cobra"
	"github.com/stackrox/infra/cmd/infractl/common"
	v1 "github.com/stackrox/infra/generated/api/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const examples = `# Print current user.
$ infractl whoami`

// Command defines the handler for infractl whoami.
func Command() *cobra.Command {
	// $ infractl whoami
	return &cobra.Command{
		Use:     "whoami",
		Short:   "Authentication information",
		Long:    "Display information about the current user",
		Example: examples,
		Args:    common.ArgsWithHelp(cobra.ExactArgs(0)),
		RunE:    common.WithGRPCHandler(run),
	}
}

func run(ctx context.Context, conn *grpc.ClientConn, _ *cobra.Command, _ []string) (common.PrettyPrinter, error) {
	resp, err := v1.NewUserServiceClient(conn).Whoami(ctx, &empty.Empty{})
	if err != nil {
		if serr, ok := status.FromError(err); ok {
			if serr.Code() == codes.PermissionDenied {
				return &prettyWhoamiResp{&v1.WhoamiResponse{}}, nil
			}
		}
		return nil, err
	}

	return prettyWhoamiResp{resp}, nil
}
