package cluster

import (
	"os"
	"path/filepath"

	"github.com/argoproj/argo/pkg/apis/workflow/v1alpha1"
	"github.com/argoproj/argo/pkg/client/clientset/versioned"
	workflowv1 "github.com/argoproj/argo/pkg/client/clientset/versioned/typed/workflow/v1alpha1"
	v1 "github.com/stackrox/infra/generated/api/v1"

	// Load GCP auth plugin for k8s requests
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func restConfig() (*rest.Config, error) {
	homeDir := os.Getenv("HOME")
	if homeDir != "" {
		// If there is a hone directory, and there is also a kubeconfig inside
		// that home directory, then we're running in out-of-cluster mode.
		kubeconfig := filepath.Join(homeDir, ".kube", "config")
		if _, err := os.Stat(kubeconfig); err == nil {
			return clientcmd.BuildConfigFromFlags("", kubeconfig)
		}
	}

	// Otherwise, use in-cluster mode.
	return rest.InClusterConfig()
}

func argoClient() (workflowv1.WorkflowInterface, error) {
	config, err := restConfig()
	if err != nil {
		return nil, err
	}

	return versioned.NewForConfigOrDie(config).ArgoprojV1alpha1().Workflows("default"), nil
}

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
