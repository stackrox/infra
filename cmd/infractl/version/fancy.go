package version

import (
	"fmt"

	"github.com/golang/protobuf/ptypes"
	v1 "github.com/stackrox/infra/generated/api/v1"
)

type versionResp struct {
	Client *v1.Version `json:"Client"`
	Server *v1.Version `json:"Server"`
}

func (r versionResp) PrettyPrint() {
	printVersion("Client", r.Client)
	printVersion("Server", r.Server)
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
