package auth

import (
	"fmt"
	"io/ioutil"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/ghodss/yaml"
	"github.com/stackrox/infra/config"
	"golang.org/x/oauth2"
)

// NewFromConfig reads and parses the given Auth0 configuration file and public key.
func NewFromConfig(auth0ConfigFile string, auth0PublicKeyPEMFile string) (*OAuth, error) {
	cfgData, err := ioutil.ReadFile(auth0ConfigFile)
	if err != nil {
		return nil, err
	}

	var cfg config.Auth0Config
	if err := yaml.Unmarshal(cfgData, &cfg); err != nil {
		return nil, err
	}

	pemData, err := ioutil.ReadFile(auth0PublicKeyPEMFile)
	if err != nil {
		return nil, err
	}

	publicKey, err := jwt.ParseRSAPublicKeyFromPEM(pemData)
	if err != nil {
		return nil, err
	}

	return &OAuth{
		endpoint: cfg.Endpoint,
		tenant:   cfg.Tenant,
		jwtState: NewStateTokenizer(time.Minute, cfg.SessionSecret),
		jwtAuth0: NewAuth0Tokenizer(0, publicKey),
		jwtUser:  NewUserTokenizer(time.Hour, cfg.SessionSecret),
		conf: &oauth2.Config{
			ClientID:     cfg.ClientID,
			ClientSecret: cfg.ClientSecret,
			Scopes:       []string{"email", "openid", "profile"},
			RedirectURL:  fmt.Sprintf("https://%s/callback", cfg.Endpoint),
			Endpoint: oauth2.Endpoint{
				AuthURL:  fmt.Sprintf("https://%s/authorize", cfg.Tenant),
				TokenURL: fmt.Sprintf("https://%s/oauth/token", cfg.Tenant),
			},
		},
	}, nil
}
