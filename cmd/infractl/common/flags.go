// Package common provides some helper functionality for building command line
// handlers.
package common

import (
	"context"
	"os"
	"time"

	"github.com/spf13/pflag"

	"github.com/spf13/cobra"
)

const (
	// TokenEnvVarName is the environment variable name used to pass a service
	// account token.
	TokenEnvVarName = "INFRA_TOKEN"

	// defaultEndpoint is the default infra-server address to connect to.
	defaultEndpoint = "infra.stackrox.com:443"
)

// flags represents the collection of flag and environment variable values
// passed to infractl.
var flags struct { // nolint:gochecknoglobals
	endpoint string
	insecure bool
	json     bool
	timeout  time.Duration
	token    string
}

// AddCommonFlags adds connection-related flags to infractl.
func AddCommonFlags(c *cobra.Command) {
	c.PersistentFlags().StringVarP(&flags.endpoint, "endpoint", "e", defaultEndpoint, "endpoint for service to contact")
	c.PersistentFlags().BoolVarP(&flags.insecure, "insecure", "k", false, "enable insecure connection")
	c.PersistentFlags().BoolVar(&flags.json, "json", false, "output as JSON")
	c.PersistentFlags().DurationVarP(&flags.timeout, "timeout", "t", time.Minute, "timeout for API requests")
	flags.token = os.Getenv(TokenEnvVarName)
}

// ContextWithTimeout returns a context and a cancel function that is bound to
// the given --timeout flag value.
func ContextWithTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), flags.timeout)
}

// endpoint returns the given --endpoint flag value.
func endpoint() string {
	return flags.endpoint
}

// insecure returns the given --insecure flag value.
func insecure() bool {
	return flags.insecure
}

// jsonOutput returns the given --json flag value.
func jsonOutput() bool {
	return flags.json
}

// token returns the given INFRACTL_TOKEN value.
func token() string {
	return flags.token
}

// MustBool looks up the named bool flag in the given flag set and panics if an
// error is returned.
func MustBool(flags *pflag.FlagSet, name string) bool {
	value, err := flags.GetBool(name)
	if err != nil {
		panic(err)
	}
	return value
}
