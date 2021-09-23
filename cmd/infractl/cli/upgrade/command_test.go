package upgrade

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateOSAndArch_linux_amd64(t *testing.T) {
	assert.NoError(t, validateOSAndArch("linux", "amd64"))
}

func TestValidateOSAndArch_darwin_amd64(t *testing.T) {
	assert.NoError(t, validateOSAndArch("darwin", "amd64"))
}

func TestValidateOSAndArch_darwin_arm64(t *testing.T) {
	assert.NoError(t, validateOSAndArch("darwin", "arm64"))
}

func TestValidateOSAndArch_linux_arm64(t *testing.T) {
	assert.Error(t, validateOSAndArch("linux", "arm64"))
}
