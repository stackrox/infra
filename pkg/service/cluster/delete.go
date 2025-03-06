package cluster

import (
	"context"

	"github.com/golang/protobuf/ptypes/duration"
	"github.com/golang/protobuf/ptypes/empty"
	v1 "github.com/stackrox/infra/generated/api/v1"
	"github.com/stackrox/infra/pkg/logging"
	"github.com/stackrox/infra/pkg/service/middleware"
)

func (s *clusterImpl) Delete(ctx context.Context, req *v1.ResourceByID) (*empty.Empty, error) {
	owner, err := middleware.GetOwnerFromContext(ctx)
	if err != nil {
		return nil, err
	}
	clusterId := req.GetId()
	log.AuditLog(logging.INFO, "cluster-delete", "received a delete request for infra cluster",
		"actor", owner,
		"cluster-id", clusterId,
	)

	workflow, err := s.argoClient.GetMostRecentArgoWorkflowFromClusterID(clusterId)
	if err != nil {
		return &empty.Empty{}, err
	}

	// Set lifespan to zero so the workflow is examined in cleanupExpiredClusters().
	lifespanReq := &v1.LifespanRequest{
		Id:       clusterId,
		Lifespan: &duration.Duration{},
		Method:   v1.LifespanRequest_REPLACE,
	}

	if _, err := s.lifespan(ctx, lifespanReq, workflow); err != nil {
		log.Log(logging.ERROR, "failed to set lifespan to 0 for argo workflow",
			"workflow-name", workflow.GetName(),
			"error", err,
		)
		return nil, err
	}

	// Resume the workflow so that it may move to the destroy phase without
	// waiting for cleanupExpiredClusters() to kick in.
	log.Log(logging.INFO, "resuming workflow for deletion", "workflow-name", workflow.GetName(), "cluster-id", req.GetId())
	s.resumeWorkflowForDeletion(workflow)
	return &empty.Empty{}, nil
}
