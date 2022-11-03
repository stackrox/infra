package platform

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidate(t *testing.T) {
	testcases := []struct{
		platform platform
		valid    bool
	}{
		{
			platform: platform{os: "darwin", arch: "amd64"},
			valid:    true,
		},
		{
			platform: platform{os: "darwin", arch: "arm64"},
			valid:    true,
		},
		{
			platform: platform{os: "linux", arch: "amd64"},
			valid:    true,
		},
		{
			platform: platform{os: "linux", arch: "arm64"},
			valid:    false,
		},
	}
	for _, testcase := range testcases {
		t.Run(testcase.platform.String(), func(t *testing.T) {
			err := Validate(testcase.platform.os, testcase.platform.arch)
			if testcase.valid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}
