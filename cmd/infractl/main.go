package main

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/stackrox/infra/cmd/infractl/cli"
	"github.com/stackrox/infra/cmd/infractl/cluster/artifacts"
	"github.com/stackrox/infra/cmd/infractl/cluster/create"
	"github.com/stackrox/infra/cmd/infractl/cluster/delete"
	"github.com/stackrox/infra/cmd/infractl/cluster/get"
	"github.com/stackrox/infra/cmd/infractl/cluster/lifespan"
	"github.com/stackrox/infra/cmd/infractl/cluster/list"
	"github.com/stackrox/infra/cmd/infractl/cluster/logs"
	"github.com/stackrox/infra/cmd/infractl/common"
	"github.com/stackrox/infra/cmd/infractl/flavor"
	janitorGCP "github.com/stackrox/infra/cmd/infractl/janitor/gcp"
	statusGet "github.com/stackrox/infra/cmd/infractl/status/get"
	statusReset "github.com/stackrox/infra/cmd/infractl/status/reset"
	statusSet "github.com/stackrox/infra/cmd/infractl/status/set"
	"github.com/stackrox/infra/cmd/infractl/token"
	"github.com/stackrox/infra/cmd/infractl/version"
	"github.com/stackrox/infra/cmd/infractl/whoami"
	"github.com/stackrox/infra/pkg/buildinfo"
)

func main() {
	// $ infractl
	cmd := &cobra.Command{
		SilenceUsage: true,
		Use:          os.Args[0],
		Version:      buildinfo.Version(),
	}

	statusCommand := &cobra.Command{
		Use:   "status get|set|reset",
		Short: "Modify or retrieve Server status information",
		Long:  "Get, set or reset server status",
	}
	statusCommand.AddCommand(
		statusGet.Command(),
		statusReset.Command(),
		statusSet.Command(),
	)

	janitorCommand := &cobra.Command{
		Use:   "janitor gcp",
		Short: "Runs tasks to clean up infra clusters",
		Long:  "Can be used to clean up infra clusters that have failed for various reasons and find orphaned VMs.",
	}
	janitorCommand.AddCommand(
		janitorGCP.Command(),
	)

	// For our version of Cobra, `cmd.Printf(...)` defaults to Stderr.
	// > Printf is a convenience method to Printf to the defined output, fallback to Stderr if not set.
	cmd.SetOut(os.Stdout)

	common.AddCommonFlags(cmd)

	cmd.AddCommand(
		// $ infractl artifacts
		artifacts.Command(),

		// $ infractl cli
		cli.Command(),

		// $ infractl create
		create.Command(),

		// $ infractl delete
		delete.Command(),

		// $ infractl flavor
		flavor.Command(),

		// $ infractl get
		get.Command(),

		// $ infractl janitor
		janitorCommand,

		// $ infractl lifespan
		lifespan.Command(),

		// $ infractl list
		list.Command(),

		// $ infractl logs
		logs.Command(),

		// $ infractl status
		statusCommand,

		// $ infractl token
		token.Command(),

		// $ infractl version
		version.Command(),

		// $ infractl whoami
		whoami.Command(),
	)

	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
