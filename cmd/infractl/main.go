package main

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/stackrox/infra/cmd/infractl/common"
	"github.com/stackrox/infra/cmd/infractl/version"
	"github.com/stackrox/infra/cmd/infractl/whoami"
	"github.com/stackrox/infra/pkg/buildinfo"
)

func main() {
	// $ infractl
	c := &cobra.Command{
		SilenceUsage: true,
		Use:          os.Args[0],
		Version:      buildinfo.Version(),
	}

	common.AddCommonFlags(c)

	c.AddCommand(
		// $ infractl whoami
		whoami.Command(),
		// $ infractl version
		version.Command(),
	)

	if err := c.Execute(); err != nil {
		os.Exit(1)
	}
}
