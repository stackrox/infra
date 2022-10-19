package common

import "github.com/spf13/cobra"

// ArgsWithHelp composes a series of positional argument check functions, and
// if any of them fail it will print the current command's help message.
func ArgsWithHelp(argFns ...cobra.PositionalArgs) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		for _, argFn := range argFns {
			if err := argFn(cmd, args); err != nil {
				cmd.Help() //nolint:errcheck
				cmd.Println()
				return err
			}
		}
		return nil
	}
}
