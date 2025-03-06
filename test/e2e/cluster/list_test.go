//go:build e2e
// +build e2e

package cluster_test

import (
	"fmt"
	"strconv"
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

	now := strconv.FormatInt(time.Now().Unix(), 32)
	baseClusterName := fmt.Sprintf("ls-flavor-%s", now)
	_, err := infractlCreateCluster(
		"simulate", fmt.Sprintf("%s-%s", baseClusterName, "simulate"),
		"--lifespan=30s",
		"--arg=create-delay-seconds=5",
		"--arg=destroy-delay-seconds=5",
	)
	assert.NoError(t, err)
	_, err = infractlCreateCluster(
		"gke-lite", fmt.Sprintf("%s-%s", baseClusterName, "gke-lite"),
		"--lifespan=30s",
	)
	assert.NoError(t, err)

	prefixFilter := fmt.Sprintf("--prefix=%s", baseClusterName)

	listAllClusters, err := infractlList(prefixFilter)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(listAllClusters.Clusters))

	listOnlySimulateClusters, err := infractlList(prefixFilter, "--flavor=simulate")
	assert.NoError(t, err)
	assert.Equal(t, 1, len(listOnlySimulateClusters.Clusters))
}

func TestListOfAStatus(t *testing.T) {
	utils.CheckContext()

	now := strconv.FormatInt(time.Now().Unix(), 32)
	baseClusterName := fmt.Sprintf("ls-status-%s", now)
	failedCluster, err := infractlCreateCluster(
		"simulate", fmt.Sprintf("%s-%s", baseClusterName, "simulate"),
		"--lifespan=30s",
		"--arg=create-outcome=fail",
	)
	assertStatusBecomes(t, failedCluster, "FAILED")
	assert.NoError(t, err)
	_, err = infractlCreateCluster(
		"gke-lite", fmt.Sprintf("%s-%s", baseClusterName, "gke-lite"),
		"--lifespan=30s",
	)
	assert.NoError(t, err)

	prefixFilter := fmt.Sprintf("--prefix=%s", baseClusterName)

	listAllClusters, err := infractlList(prefixFilter)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(listAllClusters.Clusters))

	listOnlyFailedClusters, err := infractlList(prefixFilter, "--status=FAILED")
	assert.NoError(t, err)
	assert.Equal(t, 1, len(listOnlyFailedClusters.Clusters))
}
