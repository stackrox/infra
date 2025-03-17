//go:build e2e
// +build e2e

package cluster_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stackrox/infra/test/utils"
	"github.com/stackrox/infra/test/utils/mock"
	"github.com/stretchr/testify/assert"
)

func TestListCreated(t *testing.T) {
	utils.CheckContext()
	clusterID, err := mock.InfractlCreateCluster(
		"simulate", utils.GetUniqueClusterName("list-created"),
		"--lifespan=10s",
	)
	assert.NoError(t, err)
	assert.NotEmpty(t, clusterID)
	listedClusters, err := mock.InfractlList(fmt.Sprintf("--prefix=%s", clusterID))
	assert.NoError(t, err)
	assert.Equal(t, 1, len(listedClusters.Clusters))
}

func TestListExpired(t *testing.T) {
	utils.CheckContext()
	clusterID, err := mock.InfractlCreateCluster(
		"simulate", utils.GetUniqueClusterName("list-expired"),
		"--lifespan=5s",
	)
	assert.NoError(t, err)
	assert.NotEmpty(t, clusterID)
	utils.AssertStatusBecomesWithin(t, clusterID, "FINISHED", 60*time.Second)
	listedClusters, err := mock.InfractlList(fmt.Sprintf("--prefix=%s", clusterID))
	assert.NoError(t, err)
	assert.Equal(t, 0, len(listedClusters.Clusters))
	listedClusters, err = mock.InfractlList(fmt.Sprintf("--prefix=%s", clusterID), "--expired")
	assert.NoError(t, err)
	assert.Equal(t, 1, len(listedClusters.Clusters))
}

func TestListOfAFlavor(t *testing.T) {
	utils.CheckContext()

	commonPrefix := utils.GetUniqueClusterName("ls-flavor")

	_, err := mock.InfractlCreateCluster(
		"simulate", fmt.Sprintf("%s%d", commonPrefix, 1),
		"--lifespan=10s",
	)
	assert.NoError(t, err)
	_, err = mock.InfractlCreateCluster(
		"simulate-2", fmt.Sprintf("%s%d", commonPrefix, 2),
		"--lifespan=10s",
	)
	assert.NoError(t, err)

	listAllClusters, err := mock.InfractlList(fmt.Sprintf("--prefix=%s", commonPrefix), "--expired")
	assert.NoError(t, err)
	assert.Equal(t, 2, len(listAllClusters.Clusters))

	listOnlySimulateClusters, err := mock.InfractlList(fmt.Sprintf("--prefix=%s", commonPrefix), "--flavor=simulate", "--expired")
	assert.NoError(t, err)
	assert.Equal(t, 1, len(listOnlySimulateClusters.Clusters))
}

func TestListOfAStatus(t *testing.T) {
	utils.CheckContext()

	commonPrefix := utils.GetUniqueClusterName("ls-status")

	_, err := mock.InfractlCreateCluster(
		"simulate", fmt.Sprintf("%s%d", commonPrefix, 1),
		"--lifespan=10s",
	)
	assert.NoError(t, err)

	failedCluster, err := mock.InfractlCreateCluster(
		"simulate", fmt.Sprintf("%s%d", commonPrefix, 2),
		"--lifespan=10s",
		"--arg=create-outcome=fail",
	)
	utils.AssertStatusBecomes(t, failedCluster, "FAILED")
	assert.NoError(t, err)

	listAllClusters, err := mock.InfractlList(fmt.Sprintf("--prefix=%s", commonPrefix), "--expired")
	assert.NoError(t, err)
	assert.Equal(t, 2, len(listAllClusters.Clusters))

	listOnlyFailedClusters, err := mock.InfractlList(fmt.Sprintf("--prefix=%s", commonPrefix), "--status=FAILED", "--expired")
	assert.NoError(t, err)
	assert.Equal(t, 1, len(listOnlyFailedClusters.Clusters))
}
