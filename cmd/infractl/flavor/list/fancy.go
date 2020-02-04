package list

import (
	"fmt"

	v1 "github.com/stackrox/infra/generated/api/v1"
)

type flavorListResponse v1.FlavorListResponse

func (r flavorListResponse) PrettyPrint() {
	for _, flavor := range r.Flavors {
		fmt.Printf("%s ", flavor.GetID())
		if flavor.GetID() == r.Default {
			fmt.Printf("(default)")
		}
		fmt.Println()
		fmt.Printf("  Name:         %s\n", flavor.GetName())
		fmt.Printf("  Description:  %s\n", flavor.GetDescription())
		fmt.Printf("  Availability: %s\n", flavor.GetAvailability())
	}
}
