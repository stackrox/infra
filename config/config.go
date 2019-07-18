package config

import (
	"io/ioutil"

	"github.com/BurntSushi/toml"
	"github.com/pkg/errors"
)

type Config struct {
	Server  ServerConfig  `toml:"server"`
	Storage StorageConfig `toml:"storage"`
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
	CertDir   string `toml:"certs"`
	StaticDir string `toml:"static"`
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
