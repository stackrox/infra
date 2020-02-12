package info

import (
	"fmt"

	v1 "github.com/stackrox/infra/generated/api/v1"
)

type flavor v1.Flavor

func (r flavor) PrettyPrint() {
	fmt.Printf("ID:           %s\n", r.ID)
	fmt.Printf("Name:         %s\n", r.Name)
	fmt.Printf("Description:  %s\n", r.Description)
	fmt.Printf("Availability: %s\n", r.Availability)

	// Skip printing header/newlines if there are no parameters.
	if len(r.Parameters) == 0 {
		return
	}

	fmt.Println("Parameters:")
	for name, parameter := range r.Parameters {
		fmt.Printf("  %s:\n", name)
		fmt.Printf("    Description: %s\n", parameter.GetDescription())
		fmt.Printf("    Example:     %q\n", parameter.GetExample())
	}
}
