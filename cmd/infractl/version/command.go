package version

import (
	"fmt"

	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/spf13/cobra"
	"github.com/stackrox/infra/cmd/infractl/common"
	v1 "github.com/stackrox/infra/generated/api/v1"
	"github.com/stackrox/infra/pkg/buildinfo"
)

// Command defines the handler for infractl version.
func Command() *cobra.Command {
	// $ infractl version
	return &cobra.Command{
		Use:   "version",
		Short: "Version information",
		Long:  "Version prints the client and server version",
		Run:   version,
	}
}

func version(_ *cobra.Command, _ []string) {
	var (
		clientVersion = buildinfo.All()
		serverVersion *v1.Version
	)

	// Attempt to get the server version if possible. If not, then continue
	// normal operation, and ignore any errors.
	if conn, ctx, close, err := common.GetGRPCConnection(); err == nil {
		defer close()
		serverVersion, _ = v1.NewVersionServiceClient(conn).GetVersion(ctx, &empty.Empty{})
	}

	printVersion("Client", clientVersion)
	printVersion("Server", serverVersion)
}

// printVersion pretty prints the given version indented under the given title.
func printVersion(title string, version *v1.Version) {
	if version == nil {
		fmt.Printf("%s: unknown\n", title)
		return
	}

	timestamp, _ := ptypes.Timestamp(version.GetBuildDate())
	fmt.Printf("%s\n", title)
	fmt.Printf("  Version:    %s\n", version.GetVersion())
	fmt.Printf("  Commit:     %s\n", version.GetGitCommit())
	fmt.Printf("  Workflow:   %s\n", version.GetWorkflow())
	fmt.Printf("  Build Date: %v\n", timestamp)
	fmt.Printf("  Go Version: %s\n", version.GetGoVersion())
	fmt.Printf("  Platform:   %s\n", version.GetPlatform())
}
