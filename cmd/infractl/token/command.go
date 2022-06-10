// Package token implements the infractl token command.
package token

import (
	"context"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/stackrox/infra/cmd/infractl/common"
	v1 "github.com/stackrox/infra/generated/api/v1"
	"google.golang.org/grpc"
)

const examples = `# Generate a service account token.
$ infractl token ci-robot 'CI service account' ci@stackrox.com`

// Command defines the handler for infractl token.
func Command() *cobra.Command {
	// $ infractl token
	return &cobra.Command{
		Use:     "token NAME DESCRIPTION EMAIL",
		Short:   "Generate tokens",
		Long:    "Generates a service account token",
		Example: examples,
		Args:    common.ArgsWithHelp(cobra.ExactArgs(3), args),
		RunE:    common.WithGRPCHandler(token),
		Hidden:  true,
	}
}

func args(_ *cobra.Command, args []string) error {
	name, description, email := args[0], args[1], args[2]
	switch {
	case name == "":
		return errors.New("no name given")
	case description == "":
		return errors.New("no description given")
	case email == "":
		return errors.New("no email given")
	case !(strings.HasSuffix(email, "@stackrox.com") || strings.HasSuffix(email, "@redhat.com")):
		return errors.Errorf("given email %q was neither a stackrox.com nor a redhat.com address", email)
	default:
		return nil
	}
}

func token(ctx context.Context, conn *grpc.ClientConn, _ *cobra.Command, args []string) (common.PrettyPrinter, error) {
	resp, err := v1.NewUserServiceClient(conn).CreateToken(ctx, &v1.ServiceAccount{
		Name:        args[0],
		Description: args[1],
		Email:       args[2],
	})
	if err != nil {
		return nil, err
	}

	return prettyTokenResponse(*resp), nil
}
