//go:build e2e
// +build e2e

package cluster_test

import (
	"fmt"
	"testing"

	"github.com/spf13/cobra"
	statusGet "github.com/stackrox/infra/cmd/infractl/status/get"
	statusReset "github.com/stackrox/infra/cmd/infractl/status/reset"
	statusSet "github.com/stackrox/infra/cmd/infractl/status/set"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	utils "github.com/stackrox/infra/test/e2e"
)

func setup(t *testing.T) {
	err := utils.DeleteStatusConfigmap(utils.Namespace)
	assert.NoError(t, err)
}

type statusTest struct {
	title             string
	cmd               *cobra.Command
	response          utils.StatusResponse
	assertResponse    func(statusTest)
	assertLogContents func(string)
}

func TestStatusCommand(t *testing.T) {
	utils.CheckContext()

	tests := []statusTest{
		{
			title:    "First infractl status get initializes inactive maintenance",
			cmd:      statusGet.Command(),
			response: utils.StatusResponse{},
			assertResponse: func(tc statusTest) {
				assert.False(t, tc.response.Status.MaintenanceActive)
				assert.Equal(t, tc.response.Status.Maintainer, "")
			},
			assertLogContents: func(podLogs string) {
				assert.Contains(t, podLogs, "[INFO] Initialized infra status lazily")
			},
		},
		{
			title:    "infractl status set enables maintenance and makes caller maintainer",
			cmd:      statusSet.Command(),
			response: utils.StatusResponse{},
			assertResponse: func(tc statusTest) {
				maintainer, err := utils.Whoami()
				assert.NoError(t, err)
				assert.True(t, tc.response.Status.MaintenanceActive)
				assert.Equal(t, tc.response.Status.Maintainer, maintainer)
			},
			assertLogContents: func(podLogs string) {
				maintainer, err := utils.Whoami()
				assert.NoError(t, err)
				assert.Contains(t, podLogs, fmt.Sprintf("[INFO] New Status was set by maintainer %s", maintainer))
			},
		},
		{
			title:    "infractl status reset returns no active maintenance",
			cmd:      statusReset.Command(),
			response: utils.StatusResponse{},
			assertResponse: func(tc statusTest) {
				assert.False(t, tc.response.Status.MaintenanceActive)
				assert.Equal(t, tc.response.Status.Maintainer, "")
			},
			assertLogContents: func(podLogs string) {
				assert.Contains(t, podLogs, "[INFO] Status was reset")
			},
		},
	}

	setup(t)

	for index, tc := range tests {
		name := fmt.Sprintf("%d %s", index+1, tc.title)
		t.Run(name, func(t *testing.T) {
			testStartTime := metav1.Now()

			// running command
			buf := utils.PrepareCommand(tc.cmd, true)
			err := tc.cmd.Execute()
			assert.NoError(t, err)

			// getting output from command
			err = utils.RetrieveCommandOutputJSON(buf, &tc.response)
			assert.NoError(t, err)

			// assert outputs
			tc.assertResponse(tc)

			// fetch infra-server logs
			podLogs, err := utils.GetPodLogs(utils.Namespace, utils.AppLabels, &testStartTime)
			assert.NoError(t, err)

			// assert log content
			tc.assertLogContents(podLogs)
		})
	}
}
