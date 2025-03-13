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
