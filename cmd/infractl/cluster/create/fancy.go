package create

import (
	"encoding/json"

	"github.com/spf13/cobra"

	v1 "github.com/stackrox/infra/generated/proto/api/v1"
)

type prettyResourceByID v1.ResourceByID

func (p prettyResourceByID) PrettyPrint(cmd *cobra.Command) {
	cmd.Printf("ID: %s\n", p.Id)
}

func (p prettyResourceByID) PrettyJSONPrint(cmd *cobra.Command) error {
	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return err
	}

	cmd.Printf("%s\n", string(data))
	return nil
}
