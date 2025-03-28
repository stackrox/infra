package utils

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"math/rand"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/stackrox/infra/pkg/kube"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// Namespace is the default K8s namespace in which infra server is deployed.
	Namespace = "infra"
	// AppLabels are the default K8s labels attached to the infra server deployment.
	AppLabels = "infra-server"
)

// FindInfraPod discovers the infra server pod.
func FindInfraPod(ctx context.Context, namespace string, label string) (string, error) {
	kc, err := kube.GetK8sPodsClient(namespace)
	if err != nil {
		return "", err
	}

	labelSelector := fmt.Sprintf("app=%s", label)
	pods, err := kc.List(ctx, metav1.ListOptions{LabelSelector: labelSelector})
	if err != nil {
		return "", err
	}
	if len(pods.Items) != 1 {
		return "", fmt.Errorf("could not identify infra server pod, more than one or no pods found for labels %s", labelSelector)
	}
	return pods.Items[0].Name, nil
}

// GetPodLogs retrieves the logs for a labeled pod from a given start time.
func GetPodLogs(namespace string, label string, startTime *metav1.Time) (string, error) {
	ctx := context.Background()

	podName, err := FindInfraPod(ctx, namespace, label)
	if err != nil {
		return "", err
	}

	kc, err := kube.GetK8sPodsClient(namespace)
	if err != nil {
		return "", err
	}

	req := kc.GetLogs(podName, &corev1.PodLogOptions{SinceTime: startTime})
	podLogs, err := req.Stream(ctx)
	if err != nil {
		return "", err
	}
	defer podLogs.Close()

	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, podLogs)
	return buf.String(), err
}

// DeleteStatusConfigmap deletes the configmap named status in a given namespace.
func DeleteStatusConfigmap(namespace string) error {
	kc, err := kube.GetK8sConfigMapClient(namespace)
	if err != nil {
		return err
	}
	err = kc.Delete(context.Background(), "status", metav1.DeleteOptions{})
	if k8serrors.IsNotFound(err) {
		return nil
	}
	return err
}

// CheckContext aborts an execution if the current kubectl context is not an infra-pr cluster.
func CheckContext() {
	cmd := exec.Command("kubectl", "config", "current-context")
	out, err := cmd.Output()
	if err != nil {
		return
	}
	currentContext := string(out)
	currentContext = strings.TrimSpace(currentContext)

	pattern := `(\w+)_infra-pr-(\d+)`
	match, err := regexp.Match(pattern, []byte(currentContext))
	if !match || err != nil {
		log.Printf("Current kubectl context: %s\n", currentContext)
		log.Fatalln("Quitting test. This is not an infra PR development cluster.")
	}
}

// GetUniqueClusterName adds some uniqueness to a cluster name.
func GetUniqueClusterName(baseName string) string {
	source := rand.NewSource(time.Now().UnixNano())
	generator := rand.New(source)
	return fmt.Sprintf("%s-%d", baseName, 100+generator.Intn(899))
}
