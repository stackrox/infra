// Package version implements the infractl version command.
package version

import (
	"context"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/spf13/cobra"
	"github.com/stackrox/infra/cmd/infractl/common"
	v1 "github.com/stackrox/infra/generated/api/v1"
	"github.com/stackrox/infra/pkg/buildinfo"
	"google.golang.org/grpc"
)

const examples = `# Print client and server version.
$ infractl version`

// Command defines the handler for infractl version.
func Command() *cobra.Command {
	// $ infractl version
	return &cobra.Command{
		Use:     "version",
		Short:   "Version information",
		Long:    "Print client and server version",
		Example: examples,
		Args:    common.ArgsWithHelp(cobra.ExactArgs(0)),
		RunE:    common.WithGRPCHandler(run),
	}
}

func run(ctx context.Context, conn *grpc.ClientConn, _ *cobra.Command, _ []string) (common.PrettyPrinter, error) {
	clientVersion := buildinfo.All()
	var serverVersion *v1.Version

	// Attempt to get the server version if possible. If not, then continue
	// normal operation, and ignore any errors.
	serverVersion, _ = v1.NewVersionServiceClient(conn).GetVersion(ctx, &empty.Empty{})

	return prettyVersionResp{
		Client: clientVersion,
		Server: serverVersion,
	}, nil
}
