// Package set implements the infractl status set command.
package set

import (
	"context"
	"errors"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/spf13/cobra"
	"github.com/stackrox/infra/cmd/infractl/common"
	"github.com/stackrox/infra/cmd/infractl/status"
	v1 "github.com/stackrox/infra/generated/api/v1"

	"google.golang.org/grpc"
)

const examples = `# Print server status.
$ infractl status`

// Command defines the handler for infractl status set.
func Command() *cobra.Command {
	// $ infractl status set
	return &cobra.Command{
		Use:     "set",
		Short:   "Set Server status information",
		Long:    "Set server status",
		Example: examples,
		Args:    common.ArgsWithHelp(cobra.ExactArgs(0)),
		RunE:    common.WithGRPCHandler(run),
	}
}

func getMaintainer(ctx context.Context, conn *grpc.ClientConn) (string, error) {
	resp, err := v1.NewUserServiceClient(conn).Whoami(ctx, &empty.Empty{})
	if err != nil {
		return "", err
	}
	switch resp := resp.Principal.(type) {
	case *v1.WhoamiResponse_User:
		return "", errors.New("authenticating as a user is not possible in this context")
	case *v1.WhoamiResponse_ServiceAccount:
		return resp.ServiceAccount.GetEmail(), nil
	}
	return "", errors.New("authentication required - must provide a ServiceAccount token")
}

func run(ctx context.Context, conn *grpc.ClientConn, _ *cobra.Command, _ []string) (common.PrettyPrinter, error) {
	maintainer, err := getMaintainer(ctx, conn)
	if err != nil {
		return nil, err
	}
	infraStatus := &v1.InfraStatus{
		MaintenanceActive: true,
		Maintainer:        maintainer,
	}

	newInfraStatus, err := v1.NewInfraStatusServiceClient(conn).SetStatus(ctx, infraStatus)
	if err != nil {
		return nil, err
	}
	return status.PrettyStatusResp{
		Status: newInfraStatus,
	}, nil
}
