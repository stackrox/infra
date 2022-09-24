package common

import (
	"bytes"
	"context"
	"github.com/golang/protobuf/jsonpb"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/runtime/protoiface"
)

// PrettyPrinter represents a type that knows how to render itself in a pretty,
// human-readable fashion to STDOUT.
type PrettyPrinter interface {
	// PrettyPrint renders this type in a pretty, human-readable fashion to
	// STDOUT.
	PrettyPrint()
	protoiface.MessageV1
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

		checkForVersionDiff(ctx, conn, cmd)

		// Invoke the given callback.
		result, err := handler(ctx, conn, cmd, args)
		if err != nil {
			return err
		}

		// The --json flag was passed, render result as json.
		if jsonOutput() {
			var b bytes.Buffer
			// EmitDefaults needs to be true for 0 enum values to show up, e.g., cluster status
			m := jsonpb.Marshaler{EnumsAsInts: true, EmitDefaults: true, Indent: "  ", OrigName: true}
			if err := m.Marshal(&b, result); err != nil {
				return err
			}

			// Print json body with a trailing newline.
			cmd.Printf("%s\n", b.String())
			return nil
		}

		// Pretty print result instead.
		result.PrettyPrint()
		return nil
	}
}
