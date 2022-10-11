package artifacts

import (
	"encoding/json"

	"github.com/spf13/cobra"

	v1 "github.com/stackrox/infra/generated/api/v1"
)

type prettyClusterArtifacts v1.ClusterArtifacts

func (p prettyClusterArtifacts) PrettyPrint(cmd *cobra.Command) {
	for _, artifact := range p.Artifacts {
		cmd.Printf("%s\n", artifact.Name)
		if artifact.Description != "" {
			cmd.Printf("  Description: %s\n", artifact.Description)
		}
		cmd.Printf("  URL: %s\n", artifact.URL)
	}
}

func (p prettyClusterArtifacts) PrettyJSONPrint(cmd *cobra.Command) error {
	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return err
	}

	cmd.Printf("%s\n", string(data))
	return nil
}
