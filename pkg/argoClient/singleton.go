package argoClient

import (
	"fmt"
	"github.com/argoproj/argo/pkg/client/clientset/versioned"
	workflowv1 "github.com/argoproj/argo/pkg/client/clientset/versioned/typed/workflow/v1alpha1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"path/filepath"
	"sync"
)

var (
	argoCliInstance workflowv1.WorkflowInterface
	once         sync.Once
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

func initializeArgoClient() {
	config, err := restConfig()
	if err != nil {
		panic(fmt.Sprintf("%+v", err))
	}

	argoCliInstance = versioned.NewForConfigOrDie(config).ArgoprojV1alpha1().Workflows("default")
}

func Singleton() workflowv1.WorkflowInterface {
	once.Do(initializeArgoClient)
	return argoCliInstance
}