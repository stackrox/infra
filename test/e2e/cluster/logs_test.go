//go:build e2e
// +build e2e

package cluster_test

import (
	"testing"

	"github.com/stackrox/infra/test/utils"
	"github.com/stackrox/infra/test/utils/mock"
	"github.com/stretchr/testify/assert"
)

func TestLogs(t *testing.T) {
	utils.CheckContext()

	clusterID, err := mock.InfractlCreateCluster(
		"simulate", utils.GetUniqueClusterName("logs"),
		"--lifespan=30s",
	)
	assert.NoError(t, err)
	assert.NotEmpty(t, clusterID)
	utils.AssertStatusBecomes(t, clusterID, "READY")

	logs, err := mock.InfractlLogs(clusterID)
	assert.NoError(t, err)
	assert.Containsf(t, logs.Logs[0].Name, "create", "logs must contain an entry for the create stage")
	assert.Containsf(t, string(logs.Logs[0].Body), "msg=\"capturing logs\"", "logs must contain an entry confirming log collection")
}
