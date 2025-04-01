//go:build e2e
// +build e2e

package flavor_test

import (
	"testing"

	"github.com/stackrox/infra/test/utils/mock"
	"github.com/stretchr/testify/assert"
)

func TestListFlavors(t *testing.T) {
	flavors, err := mock.InfractlFlavorList()
	assert.NoError(t, err)
	found := false
	for _, f := range flavors.Flavors {
		if f.ID == "test-gke-lite" {
			found = true
		}
	}
	assert.True(t, found, "there is a flavor for test-gke-lite")
}

func TestJanitorFlavorsNotListed(t *testing.T) {
	flavors, err := mock.InfractlFlavorList()
	assert.NoError(t, err)
	found := false
	for _, f := range flavors.Flavors {
		if f.ID == "test-janitor-delete" {
			found = true
		}
	}
	assert.False(t, found, "janitor flavor is not returned in list")

	_, err = mock.InfractlFlavorGet("test-janitor-delete")
	assert.ErrorContains(t, err, "flavor \"test-janitor-delete\" not found")
}
