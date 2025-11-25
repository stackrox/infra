// Package config provides configurability for the entire application.
package config

import (
	"log"
	"os"

	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
	"github.com/stackrox/infra/pkg/auth/claimrule"
)

// Config represents the top-level configuration for infra-server.
type Config struct {
	// Server is the server and TLS configuration.
	Server ServerConfig `json:"server"`

	// Server administrator password.
	Password string `json:"password"`

	// Google BigQuery integration configuration
	BigQuery *BigQueryConfig `json:"bigQuery"`

	// Slack notification configuration.
	Slack *SlackConfig `json:"slack"`

	// LocalDeploy disables authentication and uses HTTP instead of HTTPS when set to true.
	// This should only be set for local development deployments.
	LocalDeploy bool `json:"localDeploy"`
}

// BigQueryConfig represents the configuration for integrating with Google BigQuery
// to record cluster lifetime.
type BigQueryConfig struct {
	CredentialsFile string `json:"credentialsFile"`
	Environment     string `json:"environment"`
	Project         string `json:"project"`
	Dataset         string `json:"dataset"`
	CreationTable   string `json:"creationTable"`
	DeletionTable   string `json:"deletionTable"`
}

// AuthOidcConfig represents the configuration for integrating with OIDC provider.
type AuthOidcConfig struct {
	// Issuer is the full URL provided by OIDC provider. An example:
	// "https://auth.stage.redhat.com/auth/realms/EmployeeIDP".
	Issuer string `json:"issuer"`

	// ClientID is the client ID for the OIDC application integration.
	ClientID string `json:"clientID"`

	// ClientSecret is the client secret for the OIDC application integration.
	ClientSecret string `json:"clientSecret"`

	// Endpoint is the server hostname with optional port used for redirecting
	// requests back from OIDC provider. An example value would be
	// "localhost:8443" or "example.com".
	Endpoint string `json:"endpoint"`

	// SessionSecret is an arbitrary string used in the signing of session
	// tokens. Changing this value would invalidate current sessions.
	SessionSecret string `json:"sessionSecret"`

	// AccessTokenClaims are the list of defined claim rules used to validate
	// access token claims provided by the OIDC Provider.
	// All claims have to be fulfilled.
	AccessTokenClaims *claimrule.ClaimRules `json:"accessTokenClaims"`

	// TokenLifeTime is the duration for which generated service account tokens
	// shall be valid.
	TokenLifetime JSONDuration `json:"tokenLifetime"`
}

// ServerConfig represents the configuration used for running the HTTP & GRPC
// servers, and providing TLS.
type ServerConfig struct {
	Port                    int    `json:"port"`
	CertFile                string `json:"cert"`
	KeyFile                 string `json:"key"`
	StaticDir               string `json:"static"`
	MetricsPort             int    `json:"metricsPort"`
	MetricsIncludeHistogram bool   `json:"metricsIncludeHistogram"`
}

// SlackConfig represents the configuration used for sending cluster lifecycle
// notifications via Slack.
type SlackConfig struct {
	// Token is the Slack App token provisioned when an App is registered.
	Token string `json:"token"`

	// Channel is the channel ID of where to send messages.
	Channel string `json:"channel"`
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
	Parameters []Parameter `json:"parameters"`

	// Artifacts is the list of artifacts produced by this flavor.
	Artifacts []Artifact `json:"artifacts"`

	// WorkflowFile is the filename of an Argo workflow definition.
	WorkflowFile string `json:"workflow"`

	// Aliases are alternative IDs of the flavor.
	Aliases []string `json:"aliases"`
}

// Parameter represents a single Parameter that is needed to launch a flavor.
type Parameter struct {
	// Name is the unique name of the parameter.
	Name string `json:"name"`

	// Description is a human readable description for the parameter.
	Description string `json:"description"`

	// Value represents an example, default, or hardcoded value, depending on
	// the kind configured.
	Value string `json:"value"`

	// Kind represents the type of parameter (and corresponding value):
	// required - The user must specify a value. The configured value is used
	// as an example.
	// optional - The user may specify a value. The configured value is used
	// as a default.
	// hardcoded - The user may not specify a value. The configured value is
	// used as the only value.
	Kind parameterKind `json:"kind"`

	// For parameters that can use more explicit help
	Help string `json:"help"`

	// Indicates that the value for this parameter can be provided from the
	// contents of a file.
	FromFile bool `json:"fromFile"`
}

// Artifact represents a single Artifact that is produced by this flavor.
type Artifact struct {
	// Name is the unique name of the artifact.
	Name string `json:"name"`

	// Description is a human readable description for the artifact.
	Description string `json:"description"`

	// Tags is a list of artifact tags.
	Tags []string `json:"tags"`
}

// Load reads and parses the given configuration file.
func Load(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read config file %q", filename)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	// Override with LOCAL_DEPLOY environment variable if set
	if os.Getenv("LOCAL_DEPLOY") == "true" {
		cfg.LocalDeploy = true
		log.Printf("LOCAL_DEPLOY enabled - authentication will be bypassed and HTTP will be used")

		// Set default values for local deploy if not configured
		if cfg.Server.Port == 0 {
			cfg.Server.Port = 8443
		}
		if cfg.Server.MetricsPort == 0 {
			cfg.Server.MetricsPort = 9101
		}
		if cfg.Server.StaticDir == "" {
			cfg.Server.StaticDir = "/etc/infra/static"
		}
	}

	return &cfg, nil
}
