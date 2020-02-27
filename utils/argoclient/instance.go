// Package argoclient helps with fetching a new instance of argo client.
package argoclient

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/argoproj/argo/pkg/client/clientset/versioned"
	workflowv1 "github.com/argoproj/argo/pkg/client/clientset/versioned/typed/workflow/v1alpha1"
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

// NewArgoClient creates a new instance of argo client.
func NewArgoClient() workflowv1.WorkflowInterface {
	config, err := restConfig()
	if err != nil {
		panic(fmt.Sprintf("%+v", err))
	}

	return versioned.NewForConfigOrDie(config).ArgoprojV1alpha1().Workflows("default")
}
