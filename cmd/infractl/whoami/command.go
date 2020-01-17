package whoami

import (
	"fmt"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/spf13/cobra"
	"github.com/stackrox/infra/cmd/infractl/common"
	v1 "github.com/stackrox/infra/generated/api/v1"
)

// Command defines the handler for infractl whoami.
func Command() *cobra.Command {
	// $ infractl whoami
	return &cobra.Command{
		Use:   "whoami",
		Short: "Authentication information",
		Long:  "Whoami prints information about the current authentication method",
		RunE:  whoami,
	}
}

func whoami(_ *cobra.Command, _ []string) error {
	conn, ctx, done, err := common.GetGRPCConnection()
	if err != nil {
		return err
	}
	defer done()

	resp, err := v1.NewUserServiceClient(conn).Whoami(ctx, &empty.Empty{})
	if err != nil {
		return err
	}

	switch p := resp.Principal.(type) {
	case *v1.WhoamiResponse_User:
		panic("authenticating as a user is not possible in this context")
	case *v1.WhoamiResponse_ServiceAccount:
		fmt.Println("Service Account")
		fmt.Printf("  Name:        %s\n", p.ServiceAccount.GetName())
		fmt.Printf("  Description: %s\n", p.ServiceAccount.GetDescription())
	case nil:
		fmt.Println("Anonymous")
	}

	return nil
}
