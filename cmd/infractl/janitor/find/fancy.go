package find

import (
	"encoding/json"

	"github.com/spf13/cobra"
)

type prettyJanitorFindResponse struct {
	Instances candidateMapping
}

func (p prettyJanitorFindResponse) PrettyPrint(cmd *cobra.Command) {
	cmd.Printf("INSTANCES\n------------------\n")
	for instance, clusters := range p.Instances {
		cmd.Printf("%s --> ", instance.OriginalName)
		for _, c := range clusters {
			cmd.Printf("%s ", c.ID)
		}
		cmd.Println()
	}
}

func (p prettyJanitorFindResponse) PrettyJSONPrint(cmd *cobra.Command) error {
	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return err
	}

	cmd.Printf("%s\n", string(data))
	return nil
}
