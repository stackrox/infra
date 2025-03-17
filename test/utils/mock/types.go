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
	ID         string
	Status     string
	Flavor     string
	Parameters []v1.Parameter
}

// ListClusterReponse maps to the JSON response for infractl list operations.
type ListClusterReponse struct {
	Clusters []struct {
		ID string
	}
}

// JanitorFindResponse maps to the JSON response for infractl janitor find-gcp operations.
type JanitorFindResponse struct {
	Instances map[string][]*v1.Cluster
}

// FlavorResponse maps to the JSON response for infractl flavor get operations.
type FlavorResponse struct {
	ID           string
	Name         string
	Description  string
	Availability string
	Parameters   map[string]v1.Parameter
	Artifacts    map[string]v1.FlavorArtifact
}

type FlavorListResponse struct {
	Flavors []FlavorResponse
}
