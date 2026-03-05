//go:build e2e
// +build e2e

package cluster_test

import (
	"testing"

	"github.com/stackrox/infra/test/utils"
	"github.com/stackrox/infra/test/utils/mock"
	"github.com/stretchr/testify/assert"
)

func TestWait(t *testing.T) {
	utils.CheckContext()
	clusterID, err := mock.InfractlCreateCluster(
		"test-simulate",
		"--lifespan=60s",
	)
	assert.NoError(t, err)

	err = mock.InfractlWait(clusterID)
	assert.NoError(t, err)

	cluster, err := mock.InfractlGetCluster(clusterID)
	assert.NoError(t, err)
	assert.Equal(t, "READY", cluster.Status)
}
