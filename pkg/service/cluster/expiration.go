package cluster

import (
	"context"
	"time"

	"github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	v1 "github.com/stackrox/infra/generated/api/v1"
	"github.com/stackrox/infra/pkg/logging"
	"github.com/stackrox/infra/pkg/service/cluster/helpers"
	"github.com/stackrox/infra/pkg/service/cluster/metadata"
)

func (s *clusterImpl) cleanupExpiredClusters() {
	for ; ; time.Sleep(resumeExpiredClusterInterval) {
		workflowList, err := s.argoClient.ListWorkflows()
		if err != nil {
			log.Log(logging.ERROR, "failed to list workflows", "error", err)
			continue
		}

		for _, workflow := range workflowList.Items {
			if helpers.StatusFromWorkflowStatus(workflow.Status) == v1.Status_READY && isClusterExpired(workflow) {
				s.resumeWorkflowForDeletion(&workflow)
			}
		}
	}
}

func (s *clusterImpl) resumeWorkflowForDeletion(workflow *v1alpha1.Workflow) {
	log.Log(logging.INFO, "resuming an argo workflow that has expired", "workflow-name", workflow.GetName())
	err := s.argoClient.ResumeWorkflow(workflow.GetName())
	if err != nil {
		log.Log(logging.WARN, "failed to resume argo workflow for deletion", "workflow-name", workflow.GetName(), "error", err)
	}

	err = s.bqClient.InsertClusterDeletionRecord(context.Background(), metadata.GetClusterID(workflow), workflow.GetName())
	if err != nil {
		log.Log(logging.WARN, "failed to record cluster deletion", "workflow-name", workflow.GetName(), "error", err)
	}
}
