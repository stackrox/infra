// Package cluster implements the infractl cluster ... command.
package cluster

import (
	"github.com/spf13/cobra"
	"github.com/stackrox/infra/cmd/infractl/cluster/flavor"
)

// Command defines the handler for infractl cluster.
func Command() *cobra.Command {
	// $ infractl cluster
	cmd := &cobra.Command{
		Use:   "cluster",
		Short: "Cluster interactions",
		Long:  "Interact with clusters",
	}

	cmd.AddCommand(
		// $ infractl cluster flavors
		flavor.Command(),
	)

	return cmd
}
