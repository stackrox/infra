package flavor

import (
	"fmt"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/spf13/cobra"
	"github.com/stackrox/infra/cmd/infractl/common"
	v1 "github.com/stackrox/infra/generated/api/v1"
)

// Command defines the handler for infractl cluster flavors.
func Command() *cobra.Command {
	// $ infractl cluster flavors
	return &cobra.Command{
		Use:   "flavors",
		Short: "List cluster flavors",
		Long:  "Flavors lists the available cluster flavors",
		RunE:  flavors,
	}
}

func flavors(_ *cobra.Command, _ []string) error {
	conn, ctx, done, err := common.GetGRPCConnection()
	if err != nil {
		return err
	}
	defer done()

	resp, err := v1.NewClusterServiceClient(conn).Flavors(ctx, &empty.Empty{})
	if err != nil {
		return err
	}

	for _, flavor := range resp.Flavors {
		fmt.Printf("%s ", flavor.GetID())
		if flavor.GetID() == resp.Default {
			fmt.Printf("(default)")
		}
		fmt.Println()
		fmt.Printf("  name:         %s\n", flavor.GetName())
		fmt.Printf("  description:  %s\n", flavor.GetDescription())
		fmt.Printf("  availability: %s\n", flavor.GetAvailability())
	}

	return nil
}
