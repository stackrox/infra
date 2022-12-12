//go:build e2e
// +build e2e

// #!/usr/bin/env bats

// # shellcheck disable=SC1091
// source "$BATS_TEST_DIRNAME/../test/bats-lib.sh"
// load_bats_support

// delete_status_configmap() {
//   kubectl delete configmap/status -n infra || true
// }

// infractl() {
//   bin/infractl -e localhost:8443 -k "$@"
// }

// setup_file() {
//   e2e_setup
//   delete_status_configmap
// }

// @test "status reset returns no active maintenance" {
//   status="$(infractl status reset --json | jq -r '.Status')"
//   assert_success
//   assert_equal "$status" "{}"

//   run kubectl -n infra logs -l app=infra-server
//   assert_output --partial "[INFO] Status was reset"
// }

// @test "status set returns active maintenance with maintainer" {
//   whoami="$(infractl whoami --json | jq -r '.Principal.ServiceAccount.Email')"
//   status="$(infractl status set --json | jq -r '.Status')"
//   maintenanceActive="$(echo "$status" | jq -r '.MaintenanceActive')"
//   maintainer="$(echo "$status" | jq -r '.Maintainer')"
//   assert_success
//   assert_equal "$maintenanceActive" "true"
//   assert_equal "$maintainer" "$whoami"

//   run kubectl -n infra logs -l app=infra-server
//   assert_output --partial "[INFO] New Status was set by maintainer $maintainer"
// }

// @test "status get returns no active maintenance after lazy initialization" {
//     delete_status_configmap
//     status="$(infractl status get --json | jq -r '.Status')"
//     assert_success
//     assert_equal "$status" "{}"

//   run kubectl -n infra logs -l app=infra-server
//   assert_output --partial "[INFO] Initialized infra status lazily"
// }

package cluster_test

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stackrox/infra/cmd/infractl/common"
	statusGet "github.com/stackrox/infra/cmd/infractl/status/get"
	"github.com/stackrox/infra/pkg/kube"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func PrepareCommand(cmd *cobra.Command) *bytes.Buffer {
	common.AddCommonFlags(cmd)
	cmd.SetArgs([]string{"--endpoint=localhost:8443"})
	b := bytes.NewBufferString("")
	cmd.SetOut(b)
	return b
}

func GetPodLogs(podName string, podNamespace string) (string, error) {
	kc, err := kube.GetK8sPodsClient(podNamespace)
	if err != nil {
		return "", err
	}
	req := kc.GetLogs(podName, &corev1.PodLogOptions{})
	podLogs, err := req.Stream(context.Background())
	if err != nil {
		return "", err
	}
	defer podLogs.Close()
	buf := new(bytes.Buffer)
	if _, err := io.Copy(buf, podLogs); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func DeleteStatusConfigmap(namespace string) error {
	kc, err := kube.GetK8sConfigMapClient(namespace)
	if err != nil {
		return err
	}
	err = kc.Delete(context.Background(), "status", metav1.DeleteOptions{})
	return err
}

func TestResetReturnsNoActiveMaintenance(t *testing.T) {
	statusGetCmd := statusGet.Command()
	b := PrepareCommand(statusGetCmd)
	err := statusGetCmd.Execute()
	assert.NoError(t, err)

	out, err := ioutil.ReadAll(b)
	assert.NoError(t, err)

	assert.Contains(t, string(out), "Maintenance active: false")
	assert.Contains(t, string(out), "Maintainer: ")

	// podLogs, err := GetPodLogs("infra-server-deployment-86dd7fb475-9ljmn", "infra")
	// assert.NoError(t, err)
	// assert.Contains(t, podLogs, "[INFO] Initialized infra status lazily")
}
