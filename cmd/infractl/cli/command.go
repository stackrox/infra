// Package cli implements the infractl cli ... command.
package cli

import (
	"github.com/spf13/cobra"
	"github.com/stackrox/infra/cmd/infractl/cli/upgrade"
)

// Command defines the handler for infractl cli.
func Command() *cobra.Command {
	// $ infractl cli
	cmd := &cobra.Command{
		Use:   "cli",
		Short: "Support for intractl",
		Long:  "Support for intractl",
	}

	cmd.AddCommand(
		// $ infractl cli upgrade
		upgrade.Command(),
	)

	return cmd
}
