// Package token implements the infractl token command.
package token

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/stackrox/infra/cmd/infractl/common"
	v1 "github.com/stackrox/infra/generated/api/v1"
	"google.golang.org/grpc"
)

// Command defines the handler for infractl token.
func Command() *cobra.Command {
	// $ infractl token
	return &cobra.Command{
		Use: "token <name> <description> <email>",
		//Short: "Authentication information",
		//Long:  "token prints information about the current authentication method",
		RunE:   common.WithGRPCHandler(token),
		Hidden: true,
	}
}

func token(ctx context.Context, conn *grpc.ClientConn, _ *cobra.Command, _ []string) (common.PrettyPrinter, error) {
	resp, err := v1.NewUserServiceClient(conn).Token(ctx, &v1.ServiceAccount{
		Name:        "test name",
		Description: "test description",
		Email:       "test@stackrox.com",
	})
	if err != nil {
		return nil, err
	}

	return prettyTokenResponse(*resp), nil
}
