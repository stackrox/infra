package cluster

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	"github.com/argoproj/argo-workflows/v3/pkg/client/clientset/versioned"
	workflowv1 "github.com/argoproj/argo-workflows/v3/pkg/client/clientset/versioned/typed/workflow/v1alpha1"
	v1 "github.com/stackrox/infra/generated/api/v1"
	"k8s.io/client-go/kubernetes"
	k8sv1 "k8s.io/client-go/kubernetes/typed/core/v1"

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

func getK8sWorkflowsClient(workflowNamespace string) (workflowv1.WorkflowInterface, error) {
	config, err := restConfig()
	if err != nil {
		return nil, err
	}

	client, err := versioned.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return client.ArgoprojV1alpha1().Workflows(workflowNamespace), nil
}

func getK8sPodsClient(workflowNamespace string) (k8sv1.PodInterface, error) {
	config, err := restConfig()
	if err != nil {
		return nil, err
	}

	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return client.CoreV1().Pods(workflowNamespace), nil
}

func workflowStatus(workflowStatus v1alpha1.WorkflowStatus) v1.Status {
	// https://godoc.org/github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1#WorkflowStatus
	switch workflowStatus.Phase {
	case v1alpha1.WorkflowFailed, v1alpha1.WorkflowError:
		return v1.Status_FAILED

	case v1alpha1.WorkflowSucceeded:
		return v1.Status_FINISHED

	case v1alpha1.WorkflowPending:
		return v1.Status_CREATING

	case v1alpha1.WorkflowRunning:
		// https://godoc.org/github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1#Nodes
		for _, node := range workflowStatus.Nodes {
			// https://godoc.org/github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1#NodeType
			if node.Type == v1alpha1.NodeTypePod {
				if strings.Contains(node.Message, "ImagePullBackOff") {
					return v1.Status_FAILED
				}
				if strings.Contains(node.Message, "ErrImagePull") {
					return v1.Status_FAILED
				}
				if strings.Contains(node.Message, "Pod was active on the node longer than the specified deadline") {
					return v1.Status_FAILED
				}
			} else if node.Type == v1alpha1.NodeTypeSuspend {
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

	case "":
		return v1.Status_CREATING
	}

	panic("unknown situation")
}

// Returns an error with details of an aberrant condition if detected, nil otherwise.
// Intended to provide failure details to a user via slack post.
func workflowFailureDetails(workflowStatus v1alpha1.WorkflowStatus) error {
	switch workflowStatus.Phase {
	case v1alpha1.WorkflowRunning, v1alpha1.WorkflowFailed:
		for _, node := range workflowStatus.Nodes {
			if node.Type == v1alpha1.NodeTypePod {
				if strings.Contains(node.Message, "ImagePullBackOff") {
					msg := fmt.Sprintf("Workflow node `%s` has encountered an image pull back-off.", node.Name)
					return errors.New(msg)
				}
				if strings.Contains(node.Message, "ErrImagePull") {
					msg := fmt.Sprintf("Workflow node `%s` has encountered an image pull error.", node.Name)
					return errors.New(msg)
				}
				if strings.Contains(node.Message, "Pod was active on the node longer than the specified deadline") {
					msg := fmt.Sprintf("Workflow node `%s` has timed out.", node.Name)
					return errors.New(msg)
				}
			}
		}
	}
	return errors.New("")
}
