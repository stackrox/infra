// Package create implements the infractl cluster create command.
package create

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/spf13/cobra"
	"github.com/stackrox/infra/cmd/infractl/common"
	v1 "github.com/stackrox/infra/generated/api/v1"
	"google.golang.org/grpc"
)

// Command defines the handler for infractl cluster create.
func Command() *cobra.Command {
	// $ infractl cluster create
	cmd := &cobra.Command{
		Use:   "create",
		Short: "create on a specific cluster",
		Long:  "create displays create on a specific cluster",
		RunE:  common.WithGRPCHandler(create),
	}

	cmd.Flags().StringArray("arg", []string{}, "repeated key=value parameter pairs")
	cmd.Flags().Duration("lifespan", 3*time.Hour, "initial lifespan of the cluster")
	return cmd
}

func create(ctx context.Context, conn *grpc.ClientConn, cmd *cobra.Command, args []string) (common.PrettyPrinter, error) {
	params, _ := cmd.Flags().GetStringArray("arg")
	lifespan, _ := cmd.Flags().GetDuration("lifespan")

	if len(args) != 1 {
		return nil, errors.New("invalid arguments")
	}

	req := v1.CreateClusterRequest{
		ID:         args[0],
		Parameters: make(map[string]string),
		Lifespan:   ptypes.DurationProto(lifespan),
	}

	for _, arg := range params {
		parts := strings.SplitN(arg, "=", 2)
		if len(parts) != 2 || parts[1] == "" {
			return nil, fmt.Errorf("bad parameter argument %q", arg)
		}
		req.Parameters[parts[0]] = parts[1]
	}

	resp, err := v1.NewClusterServiceClient(conn).Create(ctx, &req)
	if err != nil {
		return nil, err
	}

	return id(*resp), nil
}
