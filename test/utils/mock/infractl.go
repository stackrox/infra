package mock

import (
	"os"

	infraClusterCreate "github.com/stackrox/infra/cmd/infractl/cluster/create"
	infraClusterDelete "github.com/stackrox/infra/cmd/infractl/cluster/delete"
	infraClusterGet "github.com/stackrox/infra/cmd/infractl/cluster/get"
	infraClusterLifespan "github.com/stackrox/infra/cmd/infractl/cluster/lifespan"
	infraClusterList "github.com/stackrox/infra/cmd/infractl/cluster/list"
	infraClusterLogs "github.com/stackrox/infra/cmd/infractl/cluster/logs"
	infraFlavorGet "github.com/stackrox/infra/cmd/infractl/flavor/get"
	infraJanitorFind "github.com/stackrox/infra/cmd/infractl/janitor/find"
	infraWhoami "github.com/stackrox/infra/cmd/infractl/whoami"
	v1 "github.com/stackrox/infra/generated/api/v1"
)

// InfractlGetStatusForID is a wrapper for 'infractl get <clusterID> --json'.
func InfractlGetStatusForID(clusterID string) (string, error) {
	jsonData, err := InfractlGetCluster(clusterID)
	if err != nil {
		return "", err
	}
	return jsonData.Status, nil
}

// InfractlCreateCluster is a wrapper for 'infractl create ...'.
func InfractlCreateCluster(args ...string) (string, error) {
	infraClusterCreateCmd := infraClusterCreate.Command()
	buf := PrepareCommand(infraClusterCreateCmd, true, args...)
	err := infraClusterCreateCmd.Execute()
	if err != nil {
		return "", err
	}
	jsonData := ClusterResponse{}
	err = RetrieveCommandOutputJSON(buf, &jsonData)
	if err != nil {
		return "", err
	}
	return jsonData.ID, nil
}

// InfractlDeleteCluster is a wrapper for 'infractl delete <clusterID>'.
func InfractlDeleteCluster(clusterID string) error {
	infraClusterDeleteCmd := infraClusterDelete.Command()
	PrepareCommand(infraClusterDeleteCmd, false, clusterID)
	return infraClusterDeleteCmd.Execute()
}

// InfractlGetCluster is a wrapper for 'infractl get <clusterID' --json'.
func InfractlGetCluster(clusterID string) (ClusterResponse, error) {
	infraClusterGetCmd := infraClusterGet.Command()
	buf := PrepareCommand(infraClusterGetCmd, true, clusterID)
	jsonData := ClusterResponse{}

	err := infraClusterGetCmd.Execute()
	if err != nil {
		return jsonData, err
	}
	err = RetrieveCommandOutputJSON(buf, &jsonData)
	if err != nil {
		return jsonData, err
	}

	return jsonData, nil
}

// InfractlLifespan is a wrapper for 'infractl lifespan <clusterID> <lifespanUpdate>'.
func InfractlLifespan(clusterID string, lifespanUpdate string) error {
	infraClusterLifespanCmd := infraClusterLifespan.Command()
	PrepareCommand(infraClusterLifespanCmd, false, clusterID, lifespanUpdate)
	return infraClusterLifespanCmd.Execute()
}

// InfractlList is a wrapper for 'infractl list <args>'.
func InfractlList(args ...string) (ListClusterReponse, error) {
	jsonData := ListClusterReponse{}
	infraClusterListCmd := infraClusterList.Command()
	buf := PrepareCommand(infraClusterListCmd, true, args...)
	err := infraClusterListCmd.Execute()
	if err != nil {
		return jsonData, err
	}

	err = RetrieveCommandOutputJSON(buf, &jsonData)
	if err != nil {
		return jsonData, err
	}
	return jsonData, nil
}

// InfractlLogs is a wrapper for 'infractl logs <clusterID> --json'.
func InfractlLogs(clusterID string) (v1.LogsResponse, error) {
	jsonData := v1.LogsResponse{}
	infraLogsCmd := infraClusterLogs.Command()
	buf := PrepareCommand(infraLogsCmd, true, clusterID)
	err := infraLogsCmd.Execute()
	if err != nil {
		return jsonData, err
	}
	err = RetrieveCommandOutputJSON(buf, &jsonData)
	if err != nil {
		return jsonData, err
	}
	return jsonData, nil
}

// InfractlWhoami is a wrapper for 'infractl whoami'.
func InfractlWhoami() (string, error) {
	whoamiCmd := infraWhoami.Command()
	buf := PrepareCommand(whoamiCmd, true)
	err := whoamiCmd.Execute()
	if err != nil {
		return "", err
	}

	jsonData := WhoamiResponse{}
	err = RetrieveCommandOutputJSON(buf, &jsonData)
	if err != nil {
		return "", err
	}
	return jsonData.Principal.ServiceAccount.Email, nil
}

// InfractlJanitorFindGCP is a wrapper for infractl janitor find-gcp'.
func InfractlJanitorFindGCP(quiet bool) (JanitorFindResponse, error) {
	findGCPCommand := infraJanitorFind.Command()

	jsonData := JanitorFindResponse{}
	args := []string{}
	if quiet {
		args = append(args, "--quiet")
	}
	buf := PrepareCommand(findGCPCommand, true, args...)

	instancesFixtureFile := "../../fixtures/gcp-instances.json"
	file, err := os.Open(instancesFixtureFile)
	if err != nil {
		return jsonData, err
	}

	defer file.Close()
	os.Stdin = file

	if err := findGCPCommand.Execute(); err != nil {
		return jsonData, err
	}

	if err := RetrieveCommandOutputJSON(buf, &jsonData); err != nil {
		return jsonData, err
	}
	return jsonData, nil
}

// InfractlFlavorGet is a wrapper for 'infractl flavor get <flavorID>'.
func InfractlFlavorGet(flavorID string) (FlavorResponse, error) {
	flavorGetCommand := infraFlavorGet.Command()
	jsonData := FlavorResponse{}
	buf := PrepareCommand(flavorGetCommand, true, flavorID)
	err := flavorGetCommand.Execute()
	if err != nil {
		return jsonData, err
	}

	err = RetrieveCommandOutputJSON(buf, &jsonData)
	if err != nil {
		return jsonData, err
	}
	return jsonData, nil
}
