package config

import (
	"fmt"
	"testing"
	"time"

	"github.com/ghodss/yaml"
	"github.com/stretchr/testify/require"
)

func TestMarshal(t *testing.T) {
	tests := []struct {
		yaml     string
		duration time.Duration
	}{
		{
			yaml: "duration: 0s\n",
		},
		{
			yaml:     "duration: 0s\n",
			duration: 0,
		},
		{
			yaml:     "duration: 1m0s\n",
			duration: time.Minute,
		},
	}

	for index, test := range tests {
		name := fmt.Sprintf("%d", index+1)
		t.Run(name, func(t *testing.T) {
			var cfg = config{
				Duration: JSONDuration(test.duration),
			}
			yaml, err := yaml.Marshal(cfg)
			require.NoError(t, err)

			require.Equal(t, test.yaml, string(yaml))
		})
	}
}

func TestUnmarshal(t *testing.T) {
	tests := []struct {
		yaml     string
		duration time.Duration
	}{
		{
			yaml:     "",
			duration: 0,
		},
		{
			yaml:     "duration: 0",
			duration: 0,
		},
		{
			yaml:     "duration: 60",
			duration: time.Nanosecond * 60,
		},
		{
			yaml:     "duration: 60s",
			duration: time.Minute,
		},
	}

	for index, test := range tests {
		name := fmt.Sprintf("%d", index+1)
		t.Run(name, func(t *testing.T) {
			var cfg config
			err := yaml.Unmarshal([]byte(test.yaml), &cfg)
			require.NoError(t, err)

			require.Equal(t, test.duration, cfg.Duration.Duration())
		})
	}
}

type config struct {
	Duration JSONDuration `json:"duration"`
}
