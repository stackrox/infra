// Package info implements the infractl cluster info command.
package info

import (
	"context"
	"errors"

	"github.com/spf13/cobra"
	"github.com/stackrox/infra/cmd/infractl/common"
	v1 "github.com/stackrox/infra/generated/api/v1"
	"google.golang.org/grpc"
)

// Command defines the handler for infractl cluster info.
func Command() *cobra.Command {
	// $ infractl cluster info
	return &cobra.Command{
		Use:     "info <cluster id>",
		Short:   "Info on a specific cluster",
		Long:    "Info displays info on a specific cluster",
		Example: "  $ infractl cluster info example-s3maj",
		RunE:    common.WithGRPCHandler(info),
	}
}

func info(ctx context.Context, conn *grpc.ClientConn, _ *cobra.Command, args []string) (common.PrettyPrinter, error) {
	if len(args) != 1 {
		return nil, errors.New("invalid arguments")
	}

	resp, err := v1.NewClusterServiceClient(conn).Info(ctx, &v1.ResourceByID{Id: args[0]})
	if err != nil {
		return nil, err
	}

	return cluster(*resp), nil
}
