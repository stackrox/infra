//go:build e2e
// +build e2e

package cluster_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	utils "github.com/stackrox/infra/test/e2e"
)

func TestClusterCanRunThroughStandardLifecycle(t *testing.T) {
	utils.CheckContext()
	clusterID, err := infractlCreateCluster(
		"simulate", utils.GetUniqueClusterName("standard"),
		"--lifespan=30s",
		"--arg=create-delay-seconds=5",
		"--arg=destroy-delay-seconds=5",
	)
	assert.NoError(t, err)
	assert.NotEmpty(t, clusterID)
	assertStatusBecomes(t, clusterID, "CREATING")
	assertStatusBecomes(t, clusterID, "READY")
	assertStatusBecomes(t, clusterID, "DESTROYING")
	assertStatusBecomes(t, clusterID, "FINISHED")
}

func TestClusterCanFailInCreate(t *testing.T) {
	utils.CheckContext()
	clusterID, err := infractlCreateCluster(
		"simulate", utils.GetUniqueClusterName("create-fails"),
		"--lifespan=30s",
		"--arg=create-delay-seconds=5",
		"--arg=create-outcome=fail",
	)
	assert.NoError(t, err)
	assert.NotEmpty(t, clusterID)
	assertStatusBecomes(t, clusterID, "CREATING")
	assertStatusBecomes(t, clusterID, "FAILED")
}

func TestClusterCanFailInDestroy(t *testing.T) {
	utils.CheckContext()
	clusterID, err := infractlCreateCluster(
		"simulate", utils.GetUniqueClusterName("destroy-fails"),
		"--lifespan=30s",
		"--arg=create-delay-seconds=5",
		"--arg=destroy-delay-seconds=5",
		"--arg=destroy-outcome=fail",
	)
	assert.NoError(t, err)
	assert.NotEmpty(t, clusterID)
	assertStatusBecomes(t, clusterID, "CREATING")
	assertStatusBecomes(t, clusterID, "READY")
	assertStatusBecomes(t, clusterID, "DESTROYING")
	assertStatusBecomes(t, clusterID, "FAILED")
}

func TestClusterCanBeDeleted(t *testing.T) {
	utils.CheckContext()
	clusterID, err := infractlCreateCluster(
		"simulate", utils.GetUniqueClusterName("for-deletion"),
		"--lifespan=5m",
		"--arg=create-delay-seconds=5",
		"--arg=destroy-delay-seconds=5",
	)
	assert.NoError(t, err)
	assert.NotEmpty(t, clusterID)
	assertStatusBecomes(t, clusterID, "CREATING")
	assertStatusBecomes(t, clusterID, "READY")
	assertStatusRemainsFor(t, clusterID, "READY", 60*time.Second)
	err = infractlDeleteCluster(clusterID)
	assert.NoError(t, err)
	assertStatusBecomes(t, clusterID, "DESTROYING")
	assertStatusBecomes(t, clusterID, "FINISHED")
}

func TestClusterCanExpireByChangingLifespan(t *testing.T) {
	utils.CheckContext()
	clusterID, err := infractlCreateCluster(
		"simulate", utils.GetUniqueClusterName("for-expire"),
		"--lifespan=5m",
		"--arg=create-delay-seconds=5",
		"--arg=destroy-delay-seconds=5",
	)
	assert.NoError(t, err)
	assert.NotEmpty(t, clusterID)
	assertStatusBecomes(t, clusterID, "CREATING")
	assertStatusBecomes(t, clusterID, "READY")
	assertStatusRemainsFor(t, clusterID, "READY", 60*time.Second)
	err = infractlLifespan(clusterID, "=0")
	assert.NoError(t, err)
	assertStatusBecomes(t, clusterID, "DESTROYING")
	assertStatusBecomes(t, clusterID, "FINISHED")
}

func TestClusterCanBeCreatedWithAliasFlavor(t *testing.T) {
	utils.CheckContext()
	clusterID, err := infractlCreateCluster(
		"test-2", utils.GetUniqueClusterName("alias-positive"),
		"--lifespan=5m",
	)
	assert.NoError(t, err)
	assert.NotEmpty(t, clusterID)
	assertStatusBecomes(t, clusterID, "READY")

	cluster, err := infractlGetCluster(clusterID)
	assert.NoError(t, err)
	assert.Equal(t, "test-connect-artifact", cluster.Flavor)
}
