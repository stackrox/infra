package get

import (
	"fmt"

	v1 "github.com/stackrox/infra/generated/api/v1"
)

type prettyFlavor v1.Flavor

func (p prettyFlavor) PrettyPrint() {
	fmt.Printf("ID:           %s\n", p.ID)
	fmt.Printf("Name:         %s\n", p.Name)
	fmt.Printf("Description:  %s\n", p.Description)
	fmt.Printf("Availability: %s\n", p.Availability)

	// Skip printing header/newlines if there are no parameters.
	if len(p.Parameters) == 0 {
		return
	}

	fmt.Println("Parameters:")
	for name, parameter := range p.Parameters {
		fmt.Printf("  %s:\n", name)
		fmt.Printf("    Description: %s\n", parameter.GetDescription())
		if parameter.GetOptional() {
			fmt.Printf("    Default:     %q\n", parameter.GetValue())
		} else {
			fmt.Printf("    Example:     %q\n", parameter.GetValue())
		}
	}
}
