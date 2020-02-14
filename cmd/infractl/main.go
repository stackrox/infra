package main

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/stackrox/infra/cmd/infractl/cluster"
	"github.com/stackrox/infra/cmd/infractl/common"
	"github.com/stackrox/infra/cmd/infractl/flavor"
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
		// $ infractl cluster
		cluster.Command(),
		// $ infractl flavor
		flavor.Command(),
		// $ infractl whoami
		whoami.Command(),
		// $ infractl version
		version.Command(),
	)

	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
