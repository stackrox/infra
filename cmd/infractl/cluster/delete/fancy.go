package delete

import (
	"encoding/json"

	"github.com/spf13/cobra"

	v1 "github.com/stackrox/infra/generated/api/v1"
)

type id v1.ResourceByID

func (p id) PrettyPrint(cmd *cobra.Command) {
	cmd.Printf("ID: %s\n", p.Id)
}

func (p id) PrettyJSONPrint(cmd *cobra.Command) error {
	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return err
	}

	cmd.Printf("%s\n", string(data))
	return nil
}
