package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
	"github.com/stackrox/infra/cmd/infractl/common"
	"github.com/stackrox/infra/pkg/kube"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	whoami "github.com/stackrox/infra/cmd/infractl/whoami"
)

const (
	Namespace = "infra"
	AppLabels = "infra-server"
)

func PrepareCommand(cmd *cobra.Command, asJson bool) *bytes.Buffer {
	common.AddCommonFlags(cmd)

	defaultArgs := []string{"--endpoint=localhost:8443", "--insecure"}
	if asJson {
		defaultArgs = append(defaultArgs, "--json")
	}

	cmd.SetArgs(defaultArgs)
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	return buf
}

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
		return "", errors.New("could not identify infra server pod, more than one or no pods found")
	}
	return pods.Items[0].Name, nil
}

func GetPodLogs(namespace string, label string, testStartTime *metav1.Time) (string, error) {
	ctx := context.Background()

	podName, err := FindInfraPod(ctx, namespace, label)
	if err != nil {
		return "", err
	}

	kc, err := kube.GetK8sPodsClient(namespace)
	if err != nil {
		return "", err
	}

	req := kc.GetLogs(podName, &corev1.PodLogOptions{
		SinceTime: testStartTime,
	})
	podLogs, err := req.Stream(ctx)
	if err != nil {
		return "", err
	}
	defer podLogs.Close()

	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, podLogs)
	return buf.String(), err
}

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

func RetrieveCommandOutput(buf *bytes.Buffer) (string, error) {
	data, err := io.ReadAll(buf)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func RetrieveCommandOutputJson(buf *bytes.Buffer, outJson interface{}) error {
	data, err := io.ReadAll(buf)
	if err != nil {
		return err
	}
	err = json.Unmarshal(data, &outJson)
	if err != nil {
		return err
	}
	return nil
}

func Whoami() (string, error) {
	whoamiCmd := whoami.Command()
	buf := PrepareCommand(whoamiCmd, true)
	err := whoamiCmd.Execute()
	if err != nil {
		return "", err
	}

	jsonData := WhoamiResponse{}
	err = RetrieveCommandOutputJson(buf, &jsonData)
	if err != nil {
		return "", err
	}
	return jsonData.Principal.ServiceAccount.Email, nil
}

// TODO: implement equivalent in Go
// e2e_setup() {
// 	# safety check, must be an infra-pr cluster
// 	context="$(kubectl config current-context)"
// 	if ! [[ "$context" =~ infra-pr-[[:digit:]]+ ]]; then
// 	  echo "kubectl config current-context: $context"
// 	  echo "Quitting test. This is not an infra PR development cluster."
// 	  exit 1
// 	fi
//   }

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
		fmt.Printf("Current kubectl context: %s\n", currentContext)
		fmt.Println("Quitting test. This is not an infra PR development cluster.")
		os.Exit(1)
	}
}
