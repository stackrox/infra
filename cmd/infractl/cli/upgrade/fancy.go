package upgrade

import (
	"encoding/json"

	"github.com/spf13/cobra"
)

type prettyCliUpgrade struct {
	updatedFilename string
}

func (p prettyCliUpgrade) PrettyPrint(cmd *cobra.Command) {
	cmd.Printf("Updated %s to match the infra server version\n", p.updatedFilename)
}

func (p prettyCliUpgrade) PrettyJSONPrint(cmd *cobra.Command) error {
	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return err
	}

	cmd.Printf("%s\n", string(data))
	return nil
}
