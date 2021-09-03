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

		// $ infractl lifespan
		lifespan.Command(),

		// $ infractl list
		list.Command(),

		// $ infractl logs
		logs.Command(),

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
