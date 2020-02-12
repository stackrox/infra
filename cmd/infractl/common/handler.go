package common

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

// PrettyPrinter represents a type that knows how to render itself in a pretty,
// human-readable fashion to STDOUT.
type PrettyPrinter interface {
	// PrettyPrint renders this type in a pretty, human-readable fashion to
	// STDOUT.
	PrettyPrint()
}

// GRPCHandler represents a function that consumes a gRPC connection, and
// produces a pretty-printable type.
type GRPCHandler func(ctx context.Context, conn *grpc.ClientConn, cmd *cobra.Command, args []string) (PrettyPrinter, error)

// WithGRPCHandler performs all of the gRPC connection setup and teardown, as
// well as rendering the returned type as either JSON or in a human-readable
// fashion.
func WithGRPCHandler(handler GRPCHandler) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		// Obtain a GRPC connection if possible.
		conn, ctx, cancel, err := GetGRPCConnection()
		if err != nil {
			return err
		}
		defer cancel()

		// Invoke the given callback.
		result, err := handler(ctx, conn, cmd, args)
		if err != nil {
			return err
		}

		// The --json flag was passed, render result as json.
		if jsonOutput() {
			data, err := json.MarshalIndent(result, "", "  ")
			if err != nil {
				return err
			}

			// Print json body with a trailing newline.
			fmt.Printf("%s\n", string(data))
			return nil
		}

		// Pretty print result instead.
		result.PrettyPrint()
		return nil
	}
}
