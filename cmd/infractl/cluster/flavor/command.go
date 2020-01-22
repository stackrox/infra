package flavor

import (
	"context"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/spf13/cobra"
	"github.com/stackrox/infra/cmd/infractl/common"
	v1 "github.com/stackrox/infra/generated/api/v1"
	"google.golang.org/grpc"
)

// Command defines the handler for infractl cluster flavors.
func Command() *cobra.Command {
	// $ infractl cluster flavors
	return &cobra.Command{
		Use:   "flavors",
		Short: "List cluster flavors",
		Long:  "Flavors lists the available cluster flavors",
		RunE:  common.WithGRPCHandler(flavors),
	}
}

func flavors(ctx context.Context, conn *grpc.ClientConn, _ *cobra.Command, _ []string) (common.PrettyPrinter, error) {
	resp, err := v1.NewClusterServiceClient(conn).Flavors(ctx, &empty.Empty{})
	if err != nil {
		return nil, err
	}

	return clusterFlavorsResp(*resp), nil
}
