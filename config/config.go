// Package config provides configurability for the entire application.
package config

import (
	"io/ioutil"

	"github.com/BurntSushi/toml"
	"github.com/pkg/errors"
)

// Config represents the top-level configuration for infra-server.
type Config struct {
	// Auth0 is the authentication, encryption, and Auth0 configuration.
	Auth0 Auth0Config `toml:"auth0"`

	// Server is the server and TLS configuration.
	Server ServerConfig `toml:"server"`

	// StaticDir is the directory to server static assets from.
	StaticDir string `toml:"static"`

	// ServiceAccounts is a list of service accounts.
	ServiceAccounts []ServiceAccountConfig `toml:"service-account"`

	// Flavors is a list of automation flavors.
	Flavors []FlavorConfig `toml:"flavor"`
}

// Auth0Config represents the configuration used for authentication,
// encryption, and interacting with Auth0.
type Auth0Config struct {
	ClientID     string `toml:"client-id"`
	ClientSecret string `toml:"client-secret"`
	AuthURL      string `toml:"auth-url"`
	TokenURL     string `toml:"token-url"`
	CallbackURL  string `toml:"callback-url"`
	UserinfoURL  string `toml:"userinfo-url"`
	LogoutURL    string `toml:"logout-url"`
	LoginURL     string `toml:"login-url"`
	SessionKey   string `toml:"session-key"`
	PublicKey    string `toml:"public-key"`
}

// ServerConfig represents the configuration used for running the HTTP & GRPC
// servers, and providing TLS.
type ServerConfig struct {
	GRPC     string `toml:"grpc"`
	HTTPS    string `toml:"https"`
	Domain   string `toml:"domain"`
	CertFile string `toml:"cert"`
	KeyFile  string `toml:"key"`
	CertDir  string `toml:"certs"`
}

// ServiceAccountConfig represents the configuration for a single service
// account.
type ServiceAccountConfig struct {
	// Name is a human readable name for the service account.
	Name string `toml:"name"`

	// Description is a human readable description for the service account.
	Description string `toml:"description"`

	// Token is a pre-shared-key used for directly authenticating as this
	// service account.
	Token string `toml:"token"`
}

// FlavorConfig represents the configuration for a single automation flavor.
type FlavorConfig struct {
	// ID is the unique, human type-able, ID for the flavor.
	ID string `toml:"id"`

	// Name is a human readable name for the flavor.
	Name string `toml:"name"`

	// Description is a human readable description for the flavor.
	Description string `toml:"description"`

	// Availability is an availability classification level. One of "alpha",
	// "beta", "stable", or "default". Exactly 1 default flavor must be
	// configured.
	Availability string `toml:"availability"`

	// Image is a full-qualified (repo+name+tag/sha) Docker image name
	// representing the automation image for this flavor.
	Image string `toml:"image"`

	// Parameters is the list of parameters required for launching this flavor.
	Parameters []Parameter `toml:"parameter"`
}

// Parameter represents a single parameter that is needed to launch a flavor.
type Parameter struct {
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
	if _, err := toml.Decode(string(data), &cfg); err != nil {
		return nil, errors.Wrap(err, "failed to decode toml")
	}

	return &cfg, nil
}
