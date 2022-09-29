package token

import (
	"encoding/json"

	"github.com/spf13/cobra"

	"github.com/stackrox/infra/cmd/infractl/common"
	v1 "github.com/stackrox/infra/generated/api/v1"
)

type prettyTokenResponse v1.TokenResponse

func (p prettyTokenResponse) PrettyPrint(cmd *cobra.Command) {
	cmd.Println("# Run the following command to configure your environment")
	cmd.Printf("export %s='%s'\n", common.TokenEnvVarName, p.Token)
}

func (p prettyTokenResponse) PrettyJSONPrint(cmd *cobra.Command) error {
	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return err
	}

	cmd.Printf("%s\n", string(data))
	return nil
}
