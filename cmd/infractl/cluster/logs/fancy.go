package logs

import (
	"encoding/json"
	"strings"

	"github.com/spf13/cobra"

	v1 "github.com/stackrox/infra/generated/proto/api/v1"
)

type prettyLogsResponse struct {
	*v1.LogsResponse
}

func (p prettyLogsResponse) PrettyPrint(cmd *cobra.Command) {
	for _, log := range p.Logs {
		cmd.Println(log.Name)
		cmd.Println(strings.Repeat("-", len(log.Name)))
		cmd.Println(log.Message)
		cmd.Println(string(log.Body))
	}
}

func (p prettyLogsResponse) PrettyJSONPrint(cmd *cobra.Command) error {
	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return err
	}

	cmd.Printf("%s\n", string(data))
	return nil
}
