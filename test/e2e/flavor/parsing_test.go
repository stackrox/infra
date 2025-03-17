//go:build e2e
// +build e2e

package flavor_test

import (
	"testing"

	"github.com/stackrox/infra/test/utils/mock"
	"github.com/stretchr/testify/assert"
)

func TestParseFlavor(t *testing.T) {
	flavor, err := mock.InfractlFlavorGet("test-gke-lite")
	assert.NoError(t, err)

	// Gets a name from metadata.name
	assert.Equal(t, "test-gke-lite", flavor.Name)

	// Availability can be set
	assert.Equal(t, "stable", flavor.Availability)

	// Parameters
	// Required parameter shows as such
	nameParam, ok := flavor.Parameters["name"]
	assert.True(t, ok, "parameter 'name' exists")
	assert.False(t, nameParam.Optional)
	assert.False(t, nameParam.Internal)

	// A parameter may have a description
	assert.Equal(t, "The name for the GKE cluster (tests required parameters)", nameParam.Description)

	// An optional parameter shows as such
	nodesParam, ok := flavor.Parameters["nodes"]
	assert.True(t, ok, "parameter 'nodes' exists")
	assert.True(t, nodesParam.Optional)
	assert.False(t, nodesParam.Internal)

	// An optional parameter may have a default value
	assert.Equal(t, "1", nodesParam.Value)

	// An optional parameter may not have a default value
	k8sVersion, ok := flavor.Parameters["k8s-version"]
	assert.True(t, ok, "parameter 'k8s-version' exists")
	assert.Equal(t, "", k8sVersion.Value)

	// Hardcoded (internal) parameters are hidden
	_, ok = flavor.Parameters["machine-type"]
	assert.False(t, ok, "internal parameter 'machine-type' is hidden")

	// Parameter order follows workflow template order
	assert.Equal(t, int32(1), nameParam.Order)
	assert.Equal(t, int32(4), k8sVersion.Order)
}

func TestDefaultAvailability(t *testing.T) {
	flavor, err := mock.InfractlFlavorGet("default-availability")
	assert.NoError(t, err)

	assert.Equal(t, "alpha", flavor.Availability)
}
