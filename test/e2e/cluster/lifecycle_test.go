//go:build e2e
// +build e2e

package cluster_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/stackrox/infra/test/utils"
	"github.com/stackrox/infra/test/utils/mock"
)

func TestClusterCanRunThroughStandardLifecycle(t *testing.T) {
	utils.CheckContext()
	clusterID, err := mock.InfractlCreateCluster(
		"test-simulate", utils.GetUniqueClusterName("standard"),
		"--lifespan=10s",
	)
	assert.NoError(t, err)
	assert.NotEmpty(t, clusterID)
	utils.AssertStatusBecomes(t, clusterID, "CREATING")
	utils.AssertStatusBecomes(t, clusterID, "READY")
	utils.AssertStatusBecomes(t, clusterID, "DESTROYING")
	utils.AssertStatusBecomes(t, clusterID, "FINISHED")
}

func TestClusterCanFailInCreate(t *testing.T) {
	utils.CheckContext()
	clusterID, err := mock.InfractlCreateCluster(
		"test-simulate", utils.GetUniqueClusterName("create-fails"),
		"--lifespan=30s",
		"--arg=create-outcome=fail",
	)
	assert.NoError(t, err)
	assert.NotEmpty(t, clusterID)
	utils.AssertStatusBecomes(t, clusterID, "CREATING")
	utils.AssertStatusBecomes(t, clusterID, "FAILED")
}

func TestClusterCanFailInDestroy(t *testing.T) {
	utils.CheckContext()
	clusterID, err := mock.InfractlCreateCluster(
		"test-simulate", utils.GetUniqueClusterName("destroy-fails"),
		"--lifespan=20s",
		"--arg=destroy-outcome=fail",
	)
	assert.NoError(t, err)
	assert.NotEmpty(t, clusterID)
	utils.AssertStatusBecomes(t, clusterID, "CREATING")
	utils.AssertStatusBecomes(t, clusterID, "READY")
	utils.AssertStatusBecomes(t, clusterID, "DESTROYING")
	utils.AssertStatusBecomes(t, clusterID, "FAILED")
}

func TestClusterCanBeDeleted(t *testing.T) {
	utils.CheckContext()
	clusterID, err := mock.InfractlCreateCluster(
		"test-simulate", utils.GetUniqueClusterName("for-deletion"),
		"--lifespan=5m",
	)
	assert.NoError(t, err)
	assert.NotEmpty(t, clusterID)
	utils.AssertStatusBecomes(t, clusterID, "CREATING")
	utils.AssertStatusBecomes(t, clusterID, "READY")
	// Checking that the cluster doesn't go into DESTROYING mode on its own
	utils.AssertStatusRemainsFor(t, clusterID, "READY", 20*time.Second)
	err = mock.InfractlDeleteCluster(clusterID)
	assert.NoError(t, err)
	utils.AssertStatusBecomes(t, clusterID, "DESTROYING")
	utils.AssertStatusBecomes(t, clusterID, "FINISHED")
}

func TestClusterCanExpireByChangingLifespan(t *testing.T) {
	utils.CheckContext()
	clusterID, err := mock.InfractlCreateCluster(
		"test-simulate", utils.GetUniqueClusterName("for-expire"),
		"--lifespan=5m",
	)
	assert.NoError(t, err)
	assert.NotEmpty(t, clusterID)
	utils.AssertStatusBecomes(t, clusterID, "CREATING")
	utils.AssertStatusBecomes(t, clusterID, "READY")
	// Checking that the cluster doesn't go into DESTROYING mode on its own
	utils.AssertStatusRemainsFor(t, clusterID, "READY", 20*time.Second)
	err = mock.InfractlLifespan(clusterID, "=0")
	assert.NoError(t, err)
	utils.AssertStatusBecomes(t, clusterID, "DESTROYING")
	utils.AssertStatusBecomes(t, clusterID, "FINISHED")
}

func TestClusterCanBeCreatedWithAliasFlavor(t *testing.T) {
	utils.CheckContext()
	clusterID, err := mock.InfractlCreateCluster(
		"test-alias-1", utils.GetUniqueClusterName("alias-positive"),
		"--lifespan=30s",
	)
	assert.NoError(t, err)
	assert.NotEmpty(t, clusterID)

	cluster, err := mock.InfractlGetCluster(clusterID)
	assert.NoError(t, err)
	assert.Equal(t, "test-connect-artifact", cluster.Flavor)
}

func TestClusterWontBeCreatedIfAliasNotFound(t *testing.T) {
	utils.CheckContext()
	_, err := mock.InfractlCreateCluster(
		"test-alias-not-set", utils.GetUniqueClusterName("alias-negative"),
		"--lifespan=30s",
	)
	assert.ErrorContains(t, err, "flavor \"test-alias-not-set\" not found")
}
