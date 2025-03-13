package mock

import (
	infraClusterCreate "github.com/stackrox/infra/cmd/infractl/cluster/create"
	infraClusterDelete "github.com/stackrox/infra/cmd/infractl/cluster/delete"
	infraClusterGet "github.com/stackrox/infra/cmd/infractl/cluster/get"
	infraClusterLifespan "github.com/stackrox/infra/cmd/infractl/cluster/lifespan"
	infraClusterList "github.com/stackrox/infra/cmd/infractl/cluster/list"
	infraClusterLogs "github.com/stackrox/infra/cmd/infractl/cluster/logs"
	infraWhoami "github.com/stackrox/infra/cmd/infractl/whoami"
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

// InfractlLogs is a wrapper for 'infractl logs <clusterID>'.
func InfractlLogs(clusterID string) (string, error) {
	infraLogsCmd := infraClusterLogs.Command()
	buf := PrepareCommand(infraLogsCmd, false, clusterID)
	err := infraLogsCmd.Execute()
	if err != nil {
		return "", err
	}
	logs, err := RetrieveCommandOutput(buf)
	if err != nil {
		return "", err
	}
	return logs, nil
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
