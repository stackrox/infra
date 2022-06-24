package auth

import (
	"context"
	"fmt"
	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/ghodss/yaml"
	"github.com/stackrox/infra/config"
	"golang.org/x/oauth2"
	"os"
	"time"
)

// NewFromConfig reads and parses the given Auth0 configuration file and public key.
func NewFromConfig(auth0ConfigFile string) (*OidcAuth, error) {
	cfgData, err := os.ReadFile(auth0ConfigFile)
	if err != nil {
		return nil, err
	}

	var cfg config.AuthOidcConfig
	if err := yaml.Unmarshal(cfgData, &cfg); err != nil {
		return nil, err
	}

	provider, err := oidc.NewProvider(context.Background(), cfg.Issuer)
	if err != nil {
		return nil, err
	}

	return &OidcAuth{
		endpoint:  cfg.Endpoint,
		provider:  provider,
		jwtState:  NewStateTokenizer(time.Minute, cfg.SessionSecret),
		jwtAccess: NewAccessTokenizer(cfg.Issuer, cfg.Claims),
		jwtOidc:   NewOidcTokenizer(provider.Verifier(&oidc.Config{ClientID: cfg.ClientID})),
		jwtUser:   NewUserTokenizer(time.Hour, cfg.SessionSecret),
		jwtSvcAcct: serviceAccountTokenizer{
			secret: []byte(cfg.SessionSecret),
		},
		conf: &oauth2.Config{
			ClientID:     cfg.ClientID,
			ClientSecret: cfg.ClientSecret,
			Scopes:       []string{"email", oidc.ScopeOpenID, "profile"},
			RedirectURL:  fmt.Sprintf("https://%s/callback", cfg.Endpoint),
			Endpoint:     provider.Endpoint(),
		},
	}, nil
}
