package wait

import (
	"github.com/spf13/cobra"
)

// prettyNoop does not output anything because 'wait' command does not need output.
type prettyNoop struct {
}

func (p prettyNoop) PrettyPrint(cmd *cobra.Command) {
	cmd.Printf("\n")
}

func (p prettyNoop) PrettyJSONPrint(cmd *cobra.Command) error {
	cmd.Printf("{}\n")
	return nil
}
