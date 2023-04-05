package version

import (
	"encoding/json"

	"github.com/spf13/cobra"

	"github.com/golang/protobuf/ptypes"
	v1 "github.com/stackrox/infra/generated/proto/api/v1"
)

type prettyVersionResp struct {
	Client *v1.Version `json:"Client"`
	Server *v1.Version `json:"Server"`
}

func (p prettyVersionResp) PrettyPrint(cmd *cobra.Command) {
	printVersion(cmd, "Client", p.Client)
	printVersion(cmd, "Server", p.Server)
}

func (p prettyVersionResp) PrettyJSONPrint(cmd *cobra.Command) error {
	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return err
	}

	cmd.Printf("%s\n", string(data))
	return nil
}

// printVersion pretty prints the given version indented under the given title.
func printVersion(cmd *cobra.Command, title string, version *v1.Version) {
	if version == nil {
		cmd.Printf("%s: unknown\n", title)
		return
	}

	timestamp, _ := ptypes.Timestamp(version.GetBuildDate())
	cmd.Printf("%s\n", title)
	cmd.Printf("  Version:    %s\n", version.GetVersion())
	cmd.Printf("  Commit:     %s\n", version.GetGitCommit())
	cmd.Printf("  Workflow:   %s\n", version.GetWorkflow())
	cmd.Printf("  Build Date: %v\n", timestamp)
	cmd.Printf("  Go Version: %s\n", version.GetGoVersion())
	cmd.Printf("  Platform:   %s\n", version.GetPlatform())
}
