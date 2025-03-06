//go:build e2e
// +build e2e

package cluster_test

import (
	"fmt"
	"testing"
	"time"

	utils "github.com/stackrox/infra/test/e2e"
	"github.com/stretchr/testify/assert"
)

func TestListCreated(t *testing.T) {
	utils.CheckContext()
	clusterID, err := infractlCreateCluster(
		"simulate", utils.GetUniqueClusterName("list-created"),
		"--lifespan=30s",
		"--arg=create-delay-seconds=5",
		"--arg=destroy-delay-seconds=5",
	)
	assert.NoError(t, err)
	assert.NotEmpty(t, clusterID)
	listedClusters, err := infractlList("--prefix=list-created")
	assert.NoError(t, err)
	assert.Equal(t, 1, len(listedClusters.Clusters))
}

func TestListExpired(t *testing.T) {
	utils.CheckContext()
	clusterID, err := infractlCreateCluster(
		"simulate", utils.GetUniqueClusterName("list-created"),
		"--lifespan=20s",
		"--arg=create-delay-seconds=5",
		"--arg=destroy-delay-seconds=5",
	)
	assert.NoError(t, err)
	assert.NotEmpty(t, clusterID)
	assertStatusBecomesWithin(t, clusterID, "FINISHED", 60*time.Second)
	listedClusters, err := infractlList(fmt.Sprintf("--prefix=%s", clusterID))
	assert.NoError(t, err)
	assert.Equal(t, 0, len(listedClusters.Clusters))
	listedClusters, err = infractlList(fmt.Sprintf("--prefix=%s", clusterID), "--expired")
	assert.NoError(t, err)
	assert.Equal(t, 1, len(listedClusters.Clusters))
}

func TestListOfAFlavor(t *testing.T) {
	utils.CheckContext()

	prefix := "ls-flavor"
	_, err := infractlCreateCluster(
		"simulate", utils.GetUniqueClusterName(prefix),
		"--lifespan=30s",
		"--arg=create-delay-seconds=5",
		"--arg=destroy-delay-seconds=5",
	)
	assert.NoError(t, err)
	_, err = infractlCreateCluster(
		"simulate-2", utils.GetUniqueClusterName(prefix),
		"--lifespan=30s",
	)
	assert.NoError(t, err)

	listAllClusters, err := infractlList(fmt.Sprintf("--prefix=%s", prefix))
	assert.NoError(t, err)
	assert.Equal(t, 2, len(listAllClusters.Clusters))

	listOnlySimulateClusters, err := infractlList(fmt.Sprintf("--prefix=%s", prefix), "--flavor=simulate")
	assert.NoError(t, err)
	assert.Equal(t, 1, len(listOnlySimulateClusters.Clusters))
}

func TestListOfAStatus(t *testing.T) {
	utils.CheckContext()
	prefix := "ls-simulate"

	failedCluster, err := infractlCreateCluster(
		"simulate", utils.GetUniqueClusterName(prefix),
		"--lifespan=30s",
		"--arg=create-outcome=fail",
	)
	assertStatusBecomes(t, failedCluster, "FAILED")
	assert.NoError(t, err)
	_, err = infractlCreateCluster(
		"simulate", utils.GetUniqueClusterName(prefix),
		"--lifespan=30s",
	)
	assert.NoError(t, err)

	listAllClusters, err := infractlList(fmt.Sprintf("--prefix=%s", prefix))
	assert.NoError(t, err)
	assert.Equal(t, 2, len(listAllClusters.Clusters))

	listOnlyFailedClusters, err := infractlList(fmt.Sprintf("--prefix=%s", prefix), "--status=FAILED")
	assert.NoError(t, err)
	assert.Equal(t, 1, len(listOnlyFailedClusters.Clusters))
}
