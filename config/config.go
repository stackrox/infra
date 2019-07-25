package config

import (
	"io/ioutil"

	"github.com/BurntSushi/toml"
	"github.com/pkg/errors"
)

type Config struct {
	Auth0   Auth0Config   `toml:"auth0"`
	Server  ServerConfig  `toml:"server"`
	Storage StorageConfig `toml:"storage"`
}

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

type ServerConfig struct {
	GRPC     string `toml:"grpc"`
	HTTP     string `toml:"http"`
	HTTPS    string `toml:"https"`
	Domain   string `toml:"domain"`
	CertFile string `toml:"cert"`
	KeyFile  string `toml:"key"`
}

type StorageConfig struct {
	CertDir     string `toml:"certs"`
	StaticDir   string `toml:"static"`
	SessionsDir string `toml:"sessions"`
}

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
