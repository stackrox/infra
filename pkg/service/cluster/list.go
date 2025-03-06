package cluster

import (
	"context"

	v1 "github.com/stackrox/infra/generated/api/v1"
	"github.com/stackrox/infra/pkg/logging"
	"github.com/stackrox/infra/pkg/service/middleware"
)

// List implements ClusterService.List.
func (s *clusterImpl) List(ctx context.Context, request *v1.ClusterListRequest) (*v1.ClusterListResponse, error) {
	workflowList, err := s.argoClient.ListWorkflows()
	if err != nil {
		return nil, err
	}

	// Obtain the email of the current principal.
	var email string
	if user, found := middleware.UserFromContext(ctx); found {
		email = user.Email
	} else if svcacct, found := middleware.ServiceAccountFromContext(ctx); found {
		email = svcacct.Email
	}

	clusters := make([]*v1.Cluster, 0, len(workflowList.Items))

	// Loop over all of the workflows, and keep only the ones that match our
	// request criteria.
	for _, workflow := range workflowList.Items {
		// This cluster is expired, and we did not request to include expired
		// clusters.
		if !request.Expired && isClusterExpired(workflow) {
			continue
		}

		// TODO(perf): move this to a listOption for the WorkflowListRequest to do the selection on K8s side
		// This cluster is not ours, and we did not request to include all clusters.
		if !request.All && !isClusterOwnedByCurrentUser(&workflow, email) {
			continue
		}

		if request.Prefix != "" && !hasClusterNamePrefix(&workflow, request.Prefix) {
			continue
		}

		metacluster, err := s.metaClusterFromWorkflow(workflow)
		if err != nil {
			log.Log(logging.ERROR, "failed to convert argo workflow to infra meta-cluster", "workflow-name", workflow.GetName(), "error", err)
			continue
		}

		// This cluster wasn't rejected, so we'll keep it for the response.
		clusters = append(clusters, &metacluster.Cluster)
	}

	resp := &v1.ClusterListResponse{
		Clusters: clusters,
	}

	return resp, nil
}
