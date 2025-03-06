package connect

import (
	"encoding/json"

	"github.com/spf13/cobra"

	v1 "github.com/stackrox/infra/generated/api/v1"
)

type prettyCluster v1.Cluster

func (p prettyCluster) PrettyPrint(cmd *cobra.Command) {
	cmd.Printf("Connected to cluster %s", p.ID)
}

func (p prettyCluster) PrettyJSONPrint(cmd *cobra.Command) error {
	data, err := json.MarshalIndent(map[string]string{"ID": p.ID}, "", "  ")
	if err != nil {
		return err
	}

	cmd.Printf("%s\n", string(data))
	return nil
}
