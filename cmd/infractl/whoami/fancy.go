package whoami

import (
	"encoding/json"

	"github.com/spf13/cobra"

	v1 "github.com/stackrox/infra/generated/api/v1"
)

type prettyWhoamiResp v1.WhoamiResponse

func (p prettyWhoamiResp) PrettyPrint(cmd *cobra.Command) {
	switch p := p.Principal.(type) {
	case *v1.WhoamiResponse_User:
		panic("authenticating as a user is not possible in this context")
	case *v1.WhoamiResponse_ServiceAccount:
		cmd.Println("Service Account")
		cmd.Printf("  Name:        %s\n", p.ServiceAccount.GetName())
		cmd.Printf("  Description: %s\n", p.ServiceAccount.GetDescription())
		cmd.Printf("  Email:       %s\n", p.ServiceAccount.GetEmail())
	case nil:
		cmd.Println("Anonymous")
	}
}

func (p prettyWhoamiResp) PrettyJSONPrint(cmd *cobra.Command) error {
	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return err
	}

	cmd.Printf("%s\n", string(data))
	return nil
}
