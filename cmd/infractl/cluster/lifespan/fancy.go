package lifespan

import (
	"encoding/json"

	"github.com/spf13/cobra"
	"google.golang.org/protobuf/types/known/durationpb"

	"github.com/stackrox/infra/cmd/infractl/common"
)

type prettyDuration struct {
	*durationpb.Duration
}

func (p prettyDuration) PrettyPrint(cmd *cobra.Command) {
	remaining := p.AsDuration()
	cmd.Println(common.FormatExpiration(remaining))
}

func (p prettyDuration) PrettyJSONPrint(cmd *cobra.Command) error {
	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return err
	}

	cmd.Printf("%s\n", string(data))
	return nil
}
