package common

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

type Fancifier interface {
	Fancy()
}

type GRPCHandler func(ctx context.Context, conn *grpc.ClientConn, cmd *cobra.Command, args []string) (Fancifier, error)

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
		result.Fancy()
		return nil
	}
}
