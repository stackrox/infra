package status

import (
	"encoding/json"

	"github.com/spf13/cobra"

	v1 "github.com/stackrox/infra/generated/api/v1"
)

// PrettyStatusResp is a struct wrapping an InfraStatus
type PrettyStatusResp struct {
	Status *v1.InfraStatus
}

// PrettyPrint prints the infra status pretty
func (p PrettyStatusResp) PrettyPrint(cmd *cobra.Command) {
	cmd.Printf("  Maintenance active: %v\n", p.Status.GetMaintenanceActive())
	cmd.Printf("  Maintainer:         %s\n", p.Status.GetMaintainer())
}

// PrettyJSONPrint prints the infra status as JSON
func (p PrettyStatusResp) PrettyJSONPrint(cmd *cobra.Command) error {
	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return err
	}

	cmd.Printf("%s\n", string(data))
	return nil
}
