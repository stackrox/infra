// Package cluster implements the infractl cluster ... command.
package cluster

import (
	"github.com/spf13/cobra"
	"github.com/stackrox/infra/cmd/infractl/cluster/create"
	"github.com/stackrox/infra/cmd/infractl/cluster/info"
	"github.com/stackrox/infra/cmd/infractl/cluster/lifespan"
	"github.com/stackrox/infra/cmd/infractl/cluster/list"
)

// Command defines the handler for infractl cluster.
func Command() *cobra.Command {
	// $ infractl cluster
	cmd := &cobra.Command{
		Use:   "cluster",
		Short: "cluster interactions",
		Long:  "Interact with clusters",
	}

	cmd.AddCommand(
		// $ infractl cluster create
		create.Command(),

		// $ infractl cluster info
		info.Command(),

		// $ infractl cluster list
		list.Command(),

		// $ infractl cluster lifespan
		lifespan.Command(),
	)

	return cmd
}
