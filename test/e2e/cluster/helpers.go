package cluster_test

import (
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	infraClusterCreate "github.com/stackrox/infra/cmd/infractl/cluster/create"
	infraClusterDelete "github.com/stackrox/infra/cmd/infractl/cluster/delete"
	infraClusterGet "github.com/stackrox/infra/cmd/infractl/cluster/get"
	infraClusterLifespan "github.com/stackrox/infra/cmd/infractl/cluster/lifespan"
	infraClusterList "github.com/stackrox/infra/cmd/infractl/cluster/list"
	infraWhoami "github.com/stackrox/infra/cmd/infractl/whoami"
	utils "github.com/stackrox/infra/test/e2e"
)

const defaultTimeout = 40 * time.Second

func assertStatusBecomesWithin(t *testing.T, clusterID string, desiredStatus string, timeout time.Duration) {
	tick := 1 * time.Second
	conditionMet := func() bool {
		actualStatus, err := infractlGetStatusForID(clusterID)
		if err != nil {
			log.Printf("error when requesting status for cluster: '%s'\n", err.Error())
			return false
		}
		return desiredStatus == actualStatus
	}
	assert.Eventually(t, conditionMet, timeout, tick)
}

func assertStatusBecomes(t *testing.T, clusterID string, desiredStatus string) {
	assertStatusBecomesWithin(t, clusterID, desiredStatus, defaultTimeout)
}

func assertStatusRemainsFor(t *testing.T, clusterID string, desiredStatus string, timeout time.Duration) {
	tick := 1 * time.Second
	conditionMet := func() bool {
		actualStatus, err := infractlGetStatusForID(clusterID)
		if err != nil {
			log.Printf("error when requesting status for cluster: '%s'\n", err.Error())
			return true
		}
		return desiredStatus != actualStatus
	}
	assert.Never(t, conditionMet, timeout, tick)
}

func infractlGetStatusForID(clusterID string) (string, error) {
	infraClusterGetCmd := infraClusterGet.Command()
	buf := utils.PrepareCommand(infraClusterGetCmd, true, clusterID)
	err := infraClusterGetCmd.Execute()
	if err != nil {
		return "", err
	}
	jsonData := utils.ClusterResponse{}
	err = utils.RetrieveCommandOutputJSON(buf, &jsonData)
	if err != nil {
		return "", err
	}
	return jsonData.Status, nil
}

func infractlCreateCluster(args ...string) (string, error) {
	infraClusterCreateCmd := infraClusterCreate.Command()
	buf := utils.PrepareCommand(infraClusterCreateCmd, true, args...)
	err := infraClusterCreateCmd.Execute()
	if err != nil {
		return "", err
	}
	jsonData := utils.ClusterResponse{}
	err = utils.RetrieveCommandOutputJSON(buf, &jsonData)
	if err != nil {
		return "", err
	}
	return jsonData.ID, nil
}

func infractlDeleteCluster(clusterID string) error {
	infraClusterDeleteCmd := infraClusterDelete.Command()
	utils.PrepareCommand(infraClusterDeleteCmd, false, clusterID)
	return infraClusterDeleteCmd.Execute()
}

func infractlGetCluster(clusterID string) (utils.ClusterResponse, error) {
	infraClusterGetCmd := infraClusterGet.Command()
	buf := utils.PrepareCommand(infraClusterGetCmd, true, clusterID)
	jsonData := utils.ClusterResponse{}

	err := infraClusterGetCmd.Execute()
	if err != nil {
		return jsonData, err
	}
	err = utils.RetrieveCommandOutputJSON(buf, &jsonData)
	if err != nil {
		return jsonData, err
	}

	return jsonData, nil
}

func infractlLifespan(clusterID string, lifespanUpdate string) error {
	infraClusterLifespanCmd := infraClusterLifespan.Command()
	utils.PrepareCommand(infraClusterLifespanCmd, false, clusterID, lifespanUpdate)
	return infraClusterLifespanCmd.Execute()
}

func infractlList(args ...string) (utils.ListClusterReponse, error) {
	jsonData := utils.ListClusterReponse{}
	infraClusterListCmd := infraClusterList.Command()
	buf := utils.PrepareCommand(infraClusterListCmd, true, args...)
	err := infraClusterListCmd.Execute()
	if err != nil {
		return jsonData, err
	}

	err = utils.RetrieveCommandOutputJSON(buf, &jsonData)
	if err != nil {
		return jsonData, err
	}
	return jsonData, nil
}

func infractlWhoami() (string, error) {
	whoamiCmd := infraWhoami.Command()
	buf := utils.PrepareCommand(whoamiCmd, true)
	err := whoamiCmd.Execute()
	if err != nil {
		return "", err
	}

	jsonData := utils.WhoamiResponse{}
	err = utils.RetrieveCommandOutputJSON(buf, &jsonData)
	if err != nil {
		return "", err
	}
	return jsonData.Principal.ServiceAccount.Email, nil
}
