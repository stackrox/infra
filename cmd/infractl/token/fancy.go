package token

import (
	"fmt"

	"github.com/stackrox/infra/cmd/infractl/common"

	v1 "github.com/stackrox/infra/generated/api/v1"
)

type prettyTokenResponse v1.TokenResponse

func (p prettyTokenResponse) PrettyPrint() {
	println("# Run the following command to configure your environment")
	fmt.Printf("export %s='%s'\n", common.TokenEnvVarName, p.Token)
}
