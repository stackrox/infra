package cluster

import (
	"context"

	v1 "github.com/stackrox/infra/generated/api/v1"
	"github.com/stackrox/infra/pkg/service/cluster/helpers"
)

// Info implements ClusterService.Info.
func (s *clusterImpl) Info(_ context.Context, clusterID *v1.ResourceByID) (*v1.Cluster, error) {
	workflow, err := s.argoClient.GetMostRecentArgoWorkflowFromClusterID(clusterID.GetId())
	if err != nil {
		return nil, err
	}

	return helpers.ClusterFromWorkflow(*workflow), nil
}
