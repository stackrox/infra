package list

import (
	"encoding/json"

	"github.com/spf13/cobra"

	v1 "github.com/stackrox/infra/generated/proto/api/v1"
)

type prettyFlavorListResponse v1.FlavorListResponse

func (p prettyFlavorListResponse) PrettyPrint(cmd *cobra.Command) {
	for _, flavor := range p.Flavors {
		cmd.Printf("%s ", flavor.GetID())
		if flavor.GetID() == p.Default {
			cmd.Printf("(default)")
		}
		cmd.Println()
		cmd.Printf("  Name:         %s\n", flavor.GetName())
		cmd.Printf("  Description:  %s\n", flavor.GetDescription())
		cmd.Printf("  Availability: %s\n", flavor.GetAvailability())
	}
}

func (p prettyFlavorListResponse) PrettyJSONPrint(cmd *cobra.Command) error {
	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return err
	}

	cmd.Printf("%s\n", string(data))
	return nil
}
