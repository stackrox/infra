package whoami

import (
	"context"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/spf13/cobra"
	"github.com/stackrox/infra/cmd/infractl/common"
	v1 "github.com/stackrox/infra/generated/api/v1"
	"google.golang.org/grpc"
)

// Command defines the handler for infractl whoami.
func Command() *cobra.Command {
	// $ infractl whoami
	return &cobra.Command{
		Use:   "whoami",
		Short: "Authentication information",
		Long:  "Whoami prints information about the current authentication method",
		RunE:  common.WithGRPCHandler(whoami),
	}
}

func whoami(ctx context.Context, conn *grpc.ClientConn, _ *cobra.Command, _ []string) (common.PrettyPrinter, error) {
	resp, err := v1.NewUserServiceClient(conn).Whoami(ctx, &empty.Empty{})
	if err != nil {
		return nil, err
	}

	return whoamiResp(*resp), nil
}
