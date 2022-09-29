package get

import (
	"encoding/json"

	"github.com/spf13/cobra"

	v1 "github.com/stackrox/infra/generated/api/v1"
)

type prettyFlavor v1.Flavor

func (p prettyFlavor) PrettyPrint(cmd *cobra.Command) {
	cmd.Printf("ID:           %s\n", p.ID)
	cmd.Printf("Name:         %s\n", p.Name)
	cmd.Printf("Description:  %s\n", p.Description)
	cmd.Printf("Availability: %s\n", p.Availability)

	// Skip printing header/newlines if there are no parameters.
	if len(p.Parameters) == 0 {
		return
	}

	cmd.Println("Parameters:")
	for name, parameter := range p.Parameters {
		cmd.Printf("  %s:\n", name)
		cmd.Printf("    Description: %s\n", parameter.GetDescription())
		if parameter.GetOptional() {
			cmd.Printf("    Default:     %q\n", parameter.GetValue())
		} else {
			cmd.Printf("    Example:     %q\n", parameter.GetValue())
		}
		if parameter.GetHelp() != "" {
			cmd.Printf("    %s\n", parameter.GetHelp())
		}
	}
}

func (p prettyFlavor) PrettyJSONPrint(cmd *cobra.Command) error {
	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return err
	}

	cmd.Printf("%s\n", string(data))
	return nil
}
