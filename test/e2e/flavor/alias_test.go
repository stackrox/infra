//go:build e2e
// +build e2e

package flavor_test

import (
	"testing"

	"github.com/stackrox/infra/test/utils/mock"
	"github.com/stretchr/testify/assert"
)

func TestFindFlavorViaAlias(t *testing.T) {
	flavor, err := mock.InfractlFlavorGet("test-alias-1")
	assert.NoError(t, err)
	assert.Equal(t, "Test Connect Artifact", flavor.Name)
}
