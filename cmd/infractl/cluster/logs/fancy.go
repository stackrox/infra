package logs

import (
	"fmt"
	"strings"

	v1 "github.com/stackrox/infra/generated/api/v1"
)

type prettyLogsResponse v1.LogsResponse

func (r prettyLogsResponse) PrettyPrint() {
	for _, log := range r.Logs {
		fmt.Println(log.Name)
		fmt.Println(strings.Repeat("-", len(log.Name)))
		fmt.Println(log.Message)
		fmt.Println(string(log.Body))
	}
}
