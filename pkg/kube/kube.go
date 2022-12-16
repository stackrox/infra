// Package kube provides access to the k8s API
package kube

import (
	"os"
	"path/filepath"

	"github.com/argoproj/argo-workflows/v3/pkg/client/clientset/versioned"
	workflowv1 "github.com/argoproj/argo-workflows/v3/pkg/client/clientset/versioned/typed/workflow/v1alpha1"
	"k8s.io/client-go/kubernetes"
	k8sv1 "k8s.io/client-go/kubernetes/typed/core/v1"

	// Load GCP auth plugin for k8s requests
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// GetK8sWorkflowsClient provides access to argo workflows
func GetK8sWorkflowsClient(workflowNamespace string) (workflowv1.WorkflowInterface, error) {
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

// GetK8sPodsClient provides access to pods
func GetK8sPodsClient(workflowNamespace string) (k8sv1.PodInterface, error) {
	client, err := getGenericK8sClient()
	if err != nil {
		return nil, err
	}
	return client.CoreV1().Pods(workflowNamespace), nil
}

// GetK8sConfigMapClient provides access to ConfigMaps
func GetK8sConfigMapClient(namespace string) (k8sv1.ConfigMapInterface, error) {
	client, err := getGenericK8sClient()
	if err != nil {
		return nil, err
	}
	return client.CoreV1().ConfigMaps(namespace), nil
}

func getGenericK8sClient() (*kubernetes.Clientset, error) {
	config, err := restConfig()
	if err != nil {
		return nil, err
	}
	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func restConfig() (*rest.Config, error) {
	// Order of preference for kube config
	// 1. KUBECONFIG env var
	// 2. ~/.kube/config file
	// 3. in-cluster config
	kubeconfig := os.Getenv("KUBECONFIG")
	if kubeconfig != "" {
		return clientcmd.BuildConfigFromFlags("", kubeconfig)
	}

	homeDir := os.Getenv("HOME")
	if homeDir != "" {
		kubeconfig := filepath.Join(homeDir, ".kube", "config")
		if _, err := os.Stat(kubeconfig); err == nil {
			return clientcmd.BuildConfigFromFlags("", kubeconfig)
		}
	}

	return rest.InClusterConfig()
}
