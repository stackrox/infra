package list

import (
	"fmt"

	v1 "github.com/stackrox/infra/generated/api/v1"
)

type prettyFlavorListResponse v1.FlavorListResponse

func (p prettyFlavorListResponse) PrettyPrint() {
	for _, flavor := range p.Flavors {
		fmt.Printf("%s ", flavor.GetID())
		if flavor.GetID() == p.Default {
			fmt.Printf("(default)")
		}
		fmt.Println()
		fmt.Printf("  Name:         %s\n", flavor.GetName())
		fmt.Printf("  Description:  %s\n", flavor.GetDescription())
		fmt.Printf("  Availability: %s\n", flavor.GetAvailability())
	}
}
