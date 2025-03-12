//go:build e2e
// +build e2e

package cluster_test

import (
	"testing"

	utils "github.com/stackrox/infra/test/e2e"
	"github.com/stretchr/testify/assert"
)

func TestLogs(t *testing.T) {
	utils.CheckContext()
	clusterID, err := infractlCreateCluster(
		"simulate", utils.GetUniqueClusterName("logs"),
		"--lifespan=30s",
		"--arg=create-delay-seconds=5",
		"--arg=destroy-delay-seconds=5",
	)
	assert.NoError(t, err)
	assert.NotEmpty(t, clusterID)
	assertStatusBecomes(t, clusterID, "READY")

	logs, err := infractlLogs(clusterID)
	assert.NoError(t, err)
	assert.Containsf(t, logs, "create", "logs must contain an entry for the create stage")
	assert.Containsf(t, logs, "msg=\"capturing logs\"", "logs must contain an entry confirming log collection")
	assertStatusBecomes(t, clusterID, "FINISHED")
	assert.Containsf(t, logs, "destroy", "logs must contain an entry for the destroy stage")
}
