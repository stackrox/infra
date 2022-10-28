package status

import (
	"encoding/json"

	"github.com/spf13/cobra"

	v1 "github.com/stackrox/infra/generated/api/v1"
)

type prettyStatusResp struct {
	Status *v1.InfraStatus `json:"Status"`
}

func (p prettyStatusResp) PrettyPrint(cmd *cobra.Command) {
	cmd.Printf("  Maintenance active: %v\n", p.Status.GetMaintenanceActive())
	cmd.Printf("  Maintainer:         %s\n", p.Status.GetMaintainer())
}

func (p prettyStatusResp) PrettyJSONPrint(cmd *cobra.Command) error {
	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return err
	}

	cmd.Printf("%s\n", string(data))
	return nil
}
