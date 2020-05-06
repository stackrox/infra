// Package flavor implements the infractl flavor ... command.
package flavor

import (
	"github.com/spf13/cobra"
	"github.com/stackrox/infra/cmd/infractl/flavor/get"
	"github.com/stackrox/infra/cmd/infractl/flavor/list"
)

// Command defines the handler for infractl flavor.
func Command() *cobra.Command {
	// $ infractl flavor
	cmd := &cobra.Command{
		Use:   "flavor",
		Short: "Interact with flavors",
		Long:  "Interact with flavors",
	}

	cmd.AddCommand(
		// $ infractl flavor get
		get.Command(),

		// $ infractl flavor list
		list.Command(),
	)

	return cmd
}
