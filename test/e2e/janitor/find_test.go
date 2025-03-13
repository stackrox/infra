//go:build e2e
// +build e2e

package find_test

import (
	"testing"

	"github.com/stackrox/infra/test/utils"
	"github.com/stackrox/infra/test/utils/mock"
	"github.com/stretchr/testify/assert"
)

func TestReminder(t *testing.T) {
	utils.CheckContext()
	clusterID, err := mock.InfractlCreateCluster(
		"simulate", utils.GetUniqueClusterName("logs"),
		"--lifespan=30s",
		"--arg=create-delay-seconds=5",
		"--arg=destroy-delay-seconds=5",
	)
	assert.NoError(t, err)
	assert.NotEmpty(t, clusterID)
	utils.AssertStatusBecomes(t, clusterID, "READY")

	logs, err := mock.InfractlLogs(clusterID)
	assert.NoError(t, err)
	assert.Containsf(t, logs, "create", "logs must contain an entry for the create stage")
	assert.Containsf(t, logs, "msg=\"capturing logs\"", "logs must contain an entry confirming log collection")
}

func TestFind(t *testing.T) {
	utils.CheckContext()

	_, err := mock.InfractlCreateCluster(
		"simulate", utils.GetUniqueClusterName("find"),
		"--lifespan=30s",
	)
	assert.NoError(t, err)

	response, err := mock.InfractlJanitorFindGCP(false)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(response.Instances))

	response, err = mock.InfractlJanitorFindGCP(true)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(response.Instances))

	clusters, ok := response.Instances["gke-not-found-orphaned-default-pool-83as64af-281j"]
	assert.True(t, ok, "there must be an entry for the mocked orphaned VM")
	assert.Empty(t, clusters, "the list of candidate clusters for the VM should be empty")
}
