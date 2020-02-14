// Package list implements the infractl cluster list command.
package list

import (
	"context"
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/spf13/cobra"
	"github.com/stackrox/infra/cmd/infractl/common"
	v1 "github.com/stackrox/infra/generated/api/v1"
	"google.golang.org/grpc"
)

// Command defines the handler for infractl cluster list.
func Command() *cobra.Command {
	// $ infractl cluster list
	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List clusters",
		Long:    "List lists the available clusters",
		Example: "  $ infractl cluster list",
		RunE:    common.WithGRPCHandler(list),
	}

	cmd.Flags().Duration("expired-cutoff", time.Hour, "do not show cluster that expired before a cutoff")
	return cmd
}

func list(ctx context.Context, conn *grpc.ClientConn, cmd *cobra.Command, _ []string) (common.PrettyPrinter, error) {
	cutoff, _ := cmd.Flags().GetDuration("expired-cutoff")

	resp, err := v1.NewClusterServiceClient(conn).List(ctx, &empty.Empty{})
	if err != nil {
		return nil, err
	}

	var results v1.ClusterListResponse
	for _, cluster := range resp.Clusters {
		createdOn, _ := ptypes.Timestamp(cluster.GetCreatedOn())
		lifespan, _ := ptypes.Duration(cluster.GetLifespan())
		expiredBy := time.Since(createdOn.Add(lifespan))

		if expiredBy < cutoff || cluster.GetStatus() == v1.Status_READY || cluster.GetStatus() == v1.Status_CREATING {
			results.Clusters = append(results.Clusters, cluster)
		}
	}

	return clusterListResponse(results), nil
}
