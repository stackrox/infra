// Package token implements the infractl token command.
package token

import (
	"context"
	"regexp"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/stackrox/infra/cmd/infractl/common"
	v1 "github.com/stackrox/infra/generated/api/v1"
	"google.golang.org/grpc"
)

const examples = `# Generate a service account token.
$ infractl token ci-robot 'CI service account' roxbot@redhat.com`

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

func validateName(name string) error {
	if name == "" {
		return errors.New("no name given")
	}
	match, err := regexp.MatchString(`^[a-zA-Z][a-zA-Z -]*$`, name)
	if err != nil {
		return err
	}
	if !match {
		return errors.New("name must be a non-empty alphabetical string (allowed special chars: space, hyphen)")
	}

	return nil
}

func validateDescription(description string) error {
	if description == "" {
		return errors.New("no description given")
	}
	match, err := regexp.MatchString(`^[a-zA-Z][a-zA-Z0-9 .-]*$`, description)
	if err != nil {
		return err
	}
	if !match {
		return errors.New("name must be a non-empty alphanumeric string (allowed special chars: space, dot, hyphen)")
	}
	return nil
}

func validateEmail(email string) error {
	if email == "" {
		return errors.New("no email given")
	}
	if !strings.HasSuffix(email, "@redhat.com") {
		return errors.Errorf("given email %q is not a redhat.com address", email)
	}
	return nil
}

func args(_ *cobra.Command, args []string) error {
	name, description, email := args[0], args[1], args[2]
	if err := validateName(name); err != nil {
		return err
	}
	if err := validateDescription(description); err != nil {
		return err
	}
	return validateEmail(email)
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
