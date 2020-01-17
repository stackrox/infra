package common

import (
	"context"
	"os"
	"time"

	"github.com/spf13/cobra"
)

const (
	// tokenEnvVarName is the environment variable name used to pass a service
	// account token.
	tokenEnvVarName = "INFRACTL_TOKEN" // nolint:gosec

	// defaultEndpoint is the default infra-server address to connect to.
	defaultEndpoint = "localhost:8823"
)

// flags represents the collection of flag and environment variable values
// passed to infractl.
var flags struct { // nolint:gochecknoglobals
	endpoint string
	insecure bool
	token    string
	timeout  time.Duration
}

// AddConnectionFlags adds connection-related flags to infractl.
func AddConnectionFlags(c *cobra.Command) *cobra.Command {
	c.PersistentFlags().StringVarP(&flags.endpoint, "endpoint", "e", defaultEndpoint, "endpoint for service to contact")
	c.PersistentFlags().BoolVarP(&flags.insecure, "insecure", "k", false, "enable insecure connection")
	c.PersistentFlags().DurationVarP(&flags.timeout, "timeout", "t", time.Minute, "timeout for API requests")
	flags.token = os.Getenv(tokenEnvVarName)

	return c
}

// endpoint returns the given --endpoint flag value.
func endpoint() string {
	return flags.endpoint
}

// insecure returns the given --insecure flag value.
func insecure() bool {
	return flags.insecure
}

// token returns the given INFRACTL_TOKEN value.
func token() string {
	return flags.token
}

// Context returns a context and a cancel function that is bound to the given
// --timeout flag value.
func Context() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), flags.timeout)
}
