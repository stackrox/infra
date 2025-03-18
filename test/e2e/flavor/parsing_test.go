//go:build e2e
// +build e2e

package flavor_test

import (
	"testing"

	"github.com/stackrox/infra/pkg/flavor"
	"github.com/stackrox/infra/test/utils/mock"
	"github.com/stretchr/testify/assert"
)

func TestParseFlavor(t *testing.T) {
	flavor, err := mock.InfractlFlavorGet("test-gke-lite")
	assert.NoError(t, err)

	// Has a name
	assert.Equal(t, "Test GKE Lite", flavor.Name)

	// Availability can be set
	assert.Equal(t, "test", flavor.Availability)

	// Parameters
	// Required parameter shows as such
	nameParam, ok := flavor.Parameters["name"]
	assert.True(t, ok, "parameter 'name' exists")
	assert.False(t, nameParam.Optional)
	assert.False(t, nameParam.Internal)

	// A parameter may have a description
	assert.Equal(t, "cluster name", nameParam.Description)

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

func TestFlavorMustHaveName(t *testing.T) {
	registry, err := flavor.NewFromConfig("../../fixtures/flavors/must-have-name.yaml")
	assert.Nil(t, registry)
	assert.ErrorContains(t, err, "flavor ID, name or description is missing")
}

func TestFlavorMustHaveValidAvailability(t *testing.T) {
	registry, err := flavor.NewFromConfig("../../fixtures/flavors/must-have-valid-availability.yaml")
	assert.Nil(t, registry)
	assert.ErrorContains(t, err, "unknown availability for flavor")
}

func TestFlavorParametersMustHaveDescription(t *testing.T) {
	registry, err := flavor.NewFromConfig("../../fixtures/flavors/parameters-must-have-name.yaml")
	assert.Nil(t, registry)
	assert.ErrorContains(t, err, "failed to validate parameters for flavor")
}
