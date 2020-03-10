// Package cluster implements the infractl cluster ... command.
package cluster

import (
	"github.com/spf13/cobra"
	"github.com/stackrox/infra/cmd/infractl/cluster/artifacts"
	"github.com/stackrox/infra/cmd/infractl/cluster/create"
	"github.com/stackrox/infra/cmd/infractl/cluster/delete"
	"github.com/stackrox/infra/cmd/infractl/cluster/info"
	"github.com/stackrox/infra/cmd/infractl/cluster/lifespan"
	"github.com/stackrox/infra/cmd/infractl/cluster/list"
	"github.com/stackrox/infra/cmd/infractl/cluster/logs"
)

// Command defines the handler for infractl cluster.
func Command() *cobra.Command {
	// $ infractl cluster
	cmd := &cobra.Command{
		Use:   "cluster",
		Short: "Interact with clusters",
		Long:  "Interact with clusters",
	}

	cmd.AddCommand(
		// $ infractl cluster artifacts
		artifacts.Command(),

		// $ infractl cluster create
		create.Command(),

		// $ infractl cluster delete
		delete.Command(),

		// $ infractl cluster info
		info.Command(),

		// $ infractl cluster lifespan
		lifespan.Command(),

		// $ infractl cluster list
		list.Command(),

		// $ infractl cluster logs
		logs.Command(),
	)

	return cmd
}
