package cluster

import (
	"github.com/argoproj/argo/pkg/apis/workflow/v1alpha1"
	v1 "github.com/stackrox/infra/generated/api/v1"

	// Load GCP auth plugin for k8s requests
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

func workflowStatus(workflowStatus v1alpha1.WorkflowStatus) v1.Status {
	switch workflowStatus.Phase {
	case v1alpha1.NodeFailed, v1alpha1.NodeError, v1alpha1.NodeSkipped:
		return v1.Status_FAILED

	case v1alpha1.NodeSucceeded:
		return v1.Status_FINISHED

	case v1alpha1.NodePending:
		return v1.Status_CREATING

	case v1alpha1.NodeRunning:
		for _, node := range workflowStatus.Nodes {
			if node.Type == v1alpha1.NodeTypeSuspend {
				switch node.Phase {
				case v1alpha1.NodeSucceeded:
					return v1.Status_DESTROYING
				case v1alpha1.NodeError, v1alpha1.NodeFailed, v1alpha1.NodeSkipped:
					panic("a suspend should not be able to fail?")
				case v1alpha1.NodeRunning, v1alpha1.NodePending:
					return v1.Status_READY
				}
			}
		}

		// No suspend node was found, which means one hasn't been run yet, which means that this cluster is still creating.
		return v1.Status_CREATING
	}

	panic("unknown situation")
}
