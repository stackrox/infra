// Package config provides configurability for the entire application.
package config

import (
	"io/ioutil"

	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
)

// Config represents the top-level configuration for infra-server.
type Config struct {
	// Server is the server and TLS configuration.
	Server ServerConfig `json:"server"`

	Password string `json:"password"`
}

// Auth0Config represents the configuration for integrating with Auth0.
type Auth0Config struct {
	// Tenant is the full Auth0 tenant name. An example value would be
	// "example.auth0.com".
	Tenant string `json:"tenant"`

	// ClientID is the client ID for the Auth0 application integration.
	ClientID string `json:"clientID"`

	// ClientSecret is the client secret for the Auth0 application integration.
	ClientSecret string `json:"clientSecret"`

	// Endpoint is the server hostname with optional port used for redirecting
	// requests back from Auth0. An example value would be "localhost:8443" or
	// "example.com".
	Endpoint string `json:"endpoint"`

	// SessionSecret is an arbitrary string used in the signing of session
	// tokens. Changing this value would invalidate current sessions.
	SessionSecret string `json:"sessionSecret"`
}

// ServerConfig represents the configuration used for running the HTTP & GRPC
// servers, and providing TLS.
type ServerConfig struct {
	Port      int    `json:"port"`
	CertFile  string `json:"cert"`
	KeyFile   string `json:"key"`
	StaticDir string `json:"static"`
}

// FlavorConfig represents the configuration for a single automation flavor.
type FlavorConfig struct {
	// ID is the unique, human type-able, ID for the flavor.
	ID string `json:"id"`

	// Name is a human readable name for the flavor.
	Name string `json:"name"`

	// Description is a human readable description for the flavor.
	Description string `json:"description"`

	// Availability is an availability classification level. One of "alpha",
	// "beta", "stable", or "default". Exactly 1 default flavor must be
	// configured.
	Availability string `json:"availability"`

	// Parameters is the list of parameters required for launching this flavor.
	Parameters []parameter `json:"parameters"`

	WorkflowFile string `json:"workflow"`
}

// parameter represents a single parameter that is needed to launch a flavor.
type parameter struct {
	// Name is the unique name of the parameter.
	Name string `toml:"name"`

	// Description is a human readable description for the parameter.
	Description string `toml:"description"`

	// Example is an arbitrary hint for a valid parameter value.
	Example string `toml:"example"`
}

// Load reads and parses the given configuration file.
func Load(filename string) (*Config, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read config file %q", filename)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
