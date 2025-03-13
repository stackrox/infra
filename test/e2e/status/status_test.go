//go:build e2e
// +build e2e

package status_test

import (
	"fmt"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	statusGet "github.com/stackrox/infra/cmd/infractl/status/get"
	statusReset "github.com/stackrox/infra/cmd/infractl/status/reset"
	statusSet "github.com/stackrox/infra/cmd/infractl/status/set"
	"github.com/stackrox/infra/test/utils"
	"github.com/stackrox/infra/test/utils/mock"
)

func setup(t *testing.T) {
	err := utils.DeleteStatusConfigmap(utils.Namespace)
	assert.NoError(t, err)
}

type statusTest struct {
	title             string
	cmd               *cobra.Command
	response          mock.StatusResponse
	assertResponse    func(statusTest)
	assertLogContents func(string)
}

func TestStatusCommand(t *testing.T) {
	utils.CheckContext()
	t.Parallel()

	maintainer, err := mock.InfractlWhoami()
	assert.NoError(t, err)

	tests := []statusTest{
		{
			title:    "First infractl status get initializes inactive maintenance",
			cmd:      statusGet.Command(),
			response: mock.StatusResponse{},
			assertResponse: func(tc statusTest) {
				assert.False(t, tc.response.Status.MaintenanceActive)
				assert.Equal(t, tc.response.Status.Maintainer, "")
			},
			assertLogContents: func(podLogs string) {
				assert.Contains(t, podLogs, fmt.Sprintf("\"msg\":\"initialized infra status lazily\",\"actor\":\"%s\",\"maintenance-active\":false", maintainer))
			},
		},
		{
			title:    "infractl status set enables maintenance and makes caller maintainer",
			cmd:      statusSet.Command(),
			response: mock.StatusResponse{},
			assertResponse: func(tc statusTest) {
				maintainer, err := mock.InfractlWhoami()
				assert.NoError(t, err)
				assert.True(t, tc.response.Status.MaintenanceActive)
				assert.Equal(t, tc.response.Status.Maintainer, maintainer)
			},
			assertLogContents: func(podLogs string) {
				assert.Contains(t, podLogs, fmt.Sprintf("\"msg\":\"new status set\",\"actor\":\"%s\",\"maintainer\":\"%s\",\"maintenance-active\":true", maintainer, maintainer))
			},
		},
		{
			title:    "infractl status reset returns no active maintenance",
			cmd:      statusReset.Command(),
			response: mock.StatusResponse{},
			assertResponse: func(tc statusTest) {
				assert.False(t, tc.response.Status.MaintenanceActive)
				assert.Equal(t, tc.response.Status.Maintainer, "")
			},
			assertLogContents: func(podLogs string) {
				assert.Contains(t, podLogs, fmt.Sprintf("\"msg\":\"status was reset\",\"actor\":\"%s\",\"maintenance-active\":false", maintainer))
			},
		},
	}

	setup(t)

	for index, tc := range tests {
		name := fmt.Sprintf("%d %s", index+1, tc.title)
		t.Run(name, func(t *testing.T) {
			testStartTime := metav1.Now()

			// running command
			buf := mock.PrepareCommand(tc.cmd, true)
			err := tc.cmd.Execute()
			assert.NoError(t, err)

			// getting output from command
			err = mock.RetrieveCommandOutputJSON(buf, &tc.response)
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
