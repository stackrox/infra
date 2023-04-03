package auth

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/ghodss/yaml"
	"github.com/stackrox/infra/config"
	"golang.org/x/oauth2"
)

// NewFromConfig reads and parses the given OIDC configuration file.
func NewFromConfig(oidcConfigFile string) (*OidcAuth, error) {
	cfgData, err := os.ReadFile(oidcConfigFile)
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
		jwtAccess: NewAccessTokenizer(cfg.AccessTokenClaims),
		jwtOidc:   NewOidcTokenizer(provider.Verifier(&oidc.Config{ClientID: cfg.ClientID})),
		jwtUser:   NewUserTokenizer(time.Hour, cfg.SessionSecret),
		jwtSvcAcct: serviceAccountTokenizer{
			secret:   []byte(cfg.SessionSecret),
			lifetime: cfg.TokenLifetime.Duration(),
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
