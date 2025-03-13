package mock

import (
	v1 "github.com/stackrox/infra/generated/api/v1"
)

// StatusResponse helper maps to the JSON response for infractl status operations.
type StatusResponse struct {
	Status v1.InfraStatus
}

// WhoamiResponse helper maps to the JSON response for infractl whoami operations.
type WhoamiResponse struct {
	Principal v1.WhoamiResponse_ServiceAccount
}

// ClusterResponse helper maps to the JSON response for infractl operations related to clusters.
// We use this instead of v1.Cluster because Go cannot parse the Status string back to the enum.
type ClusterResponse struct {
	ID     string
	Status string
	Flavor string
}

// ListClusterReponse maps to the JSON response for infractl list operations.
type ListClusterReponse struct {
	Clusters []struct {
		ID string
	}
}

type JanitorFindResponse struct {
	Instances map[string][]*v1.Cluster
}
