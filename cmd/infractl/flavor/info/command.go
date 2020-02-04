// Package info implements the infractl flavor info command.
package info

import (
	"context"
	"errors"

	"github.com/spf13/cobra"
	"github.com/stackrox/infra/cmd/infractl/common"
	v1 "github.com/stackrox/infra/generated/api/v1"
	"google.golang.org/grpc"
)

// Command defines the handler for infractl flavor info.
func Command() *cobra.Command {
	// $ infractl flavor info
	return &cobra.Command{
		Use:   "info",
		Short: "Info on a specific flavor",
		Long:  "Info displays info on a specific flavor",
		RunE:  common.WithGRPCHandler(info),
	}
}

func info(ctx context.Context, conn *grpc.ClientConn, _ *cobra.Command, args []string) (common.PrettyPrinter, error) {
	if len(args) != 1 {
		return nil, errors.New("invalid arguments")
	}

	resp, err := v1.NewFlavorServiceClient(conn).Info(ctx, &v1.ResourceByID{Id: args[0]})
	if err != nil {
		return nil, err
	}

	return flavor(*resp), nil
}
