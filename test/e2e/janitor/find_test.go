//go:build e2e
// +build e2e

package find_test

import (
	"testing"

	"github.com/stackrox/infra/test/utils"
	"github.com/stackrox/infra/test/utils/mock"
	"github.com/stretchr/testify/assert"
)

func TestFind(t *testing.T) {
	utils.CheckContext()
	t.Parallel()

	clusterID, err := mock.InfractlCreateCluster(
		"simulate", utils.GetUniqueClusterName("find"),
		"--lifespan=30s",
	)
	assert.NoError(t, err)

	response, err := mock.InfractlJanitorFindGCP(false)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(response.Instances))
	clusters, ok := response.Instances["gke-find-test-1-exists-default-pool-83ce64af-280j"]
	assert.True(t, ok, "there must be an entry for the mocked VM")
	assert.NotEmpty(t, clusters, "the list of candidate clusters for the VM should not be empty")
	assert.Equal(t, clusterID, clusters[0].ID)

	response, err = mock.InfractlJanitorFindGCP(true)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(response.Instances))

	clusters, ok = response.Instances["gke-not-found-orphaned-default-pool-83as64af-281j"]
	assert.True(t, ok, "there must be an entry for the mocked orphaned VM")
	assert.Empty(t, clusters, "the list of candidate clusters for the VM should be empty")
}
