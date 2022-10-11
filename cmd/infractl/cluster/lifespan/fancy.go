package lifespan

import (
	"encoding/json"

	"github.com/spf13/cobra"

	"github.com/golang/protobuf/ptypes"
	durpb "github.com/golang/protobuf/ptypes/duration"
	"github.com/stackrox/infra/cmd/infractl/common"
)

type prettyDuration struct {
	*durpb.Duration
}

func (p prettyDuration) PrettyPrint(cmd *cobra.Command) {
	remaining, _ := ptypes.Duration(p.Duration)

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
