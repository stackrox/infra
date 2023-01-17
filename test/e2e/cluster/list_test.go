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
