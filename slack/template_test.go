package slack

import (
	"fmt"
	"testing"

	v1 "github.com/stackrox/infra/generated/api/v1"
	"github.com/stretchr/testify/assert"
)

func Test(t *testing.T) {
	tests := []struct {
		title                  string
		clusterStatus          v1.Status
		slackStatus            Status
		expectedNewSlackStatus Status
		expectedNoMessages     bool
	}{
		{
			title:                  "failed nop",
			clusterStatus:          v1.Status_FAILED,
			slackStatus:            StatusFailed,
			expectedNewSlackStatus: StatusFailed,
			expectedNoMessages:     true,
		},
		{
			title:                  "failed status blank",
			clusterStatus:          v1.Status_FAILED,
			slackStatus:            "",
			expectedNewSlackStatus: StatusFailed,
		},
		{
			title:                  "failed status other",
			clusterStatus:          v1.Status_FAILED,
			slackStatus:            "qwertyuiop",
			expectedNewSlackStatus: StatusFailed,
		},

		{
			title:                  "creating nop",
			clusterStatus:          v1.Status_CREATING,
			slackStatus:            StatusCreating,
			expectedNewSlackStatus: StatusCreating,
			expectedNoMessages:     true,
		},
		{
			title:                  "creating status blank",
			clusterStatus:          v1.Status_CREATING,
			slackStatus:            "",
			expectedNewSlackStatus: StatusCreating,
		},
		{
			title:                  "creating status other",
			clusterStatus:          v1.Status_CREATING,
			slackStatus:            "qwertyuiop",
			expectedNewSlackStatus: StatusCreating,
		},

		{
			title:                  "ready nop",
			clusterStatus:          v1.Status_READY,
			slackStatus:            StatusReady,
			expectedNewSlackStatus: StatusReady,
			expectedNoMessages:     true,
		},
		{
			title:                  "ready status blank",
			clusterStatus:          v1.Status_READY,
			slackStatus:            "",
			expectedNewSlackStatus: StatusReady,
		},
		{
			title:                  "ready status other",
			clusterStatus:          v1.Status_READY,
			slackStatus:            "qwertyuiop",
			expectedNewSlackStatus: StatusReady,
		},

		{
			title:                  "destroyed (destroying) nop",
			clusterStatus:          v1.Status_DESTROYING,
			slackStatus:            StatusDestroyed,
			expectedNewSlackStatus: StatusDestroyed,
			expectedNoMessages:     true,
		},
		{
			title:                  "destroyed (destroying) status blank",
			clusterStatus:          v1.Status_DESTROYING,
			slackStatus:            "",
			expectedNewSlackStatus: StatusDestroyed,
		},
		{
			title:                  "destroyed (destroying) status other",
			clusterStatus:          v1.Status_DESTROYING,
			slackStatus:            "qwertyuiop",
			expectedNewSlackStatus: StatusDestroyed,
		},

		{
			title:                  "destroyed (finished) nop",
			clusterStatus:          v1.Status_FINISHED,
			slackStatus:            StatusDestroyed,
			expectedNewSlackStatus: StatusDestroyed,
			expectedNoMessages:     true,
		},
		{
			title:                  "destroyed (finished) status blank",
			clusterStatus:          v1.Status_FINISHED,
			slackStatus:            "",
			expectedNewSlackStatus: StatusDestroyed,
		},
		{
			title:                  "destroyed (finished) status other",
			clusterStatus:          v1.Status_FINISHED,
			slackStatus:            "qwertyuiop",
			expectedNewSlackStatus: StatusDestroyed,
		},
	}

	for index, test := range tests {
		name := fmt.Sprintf("%d %s", index+1, test.title)
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			actualNewSlackStatus, actualMessages := FormatSlackMessage(test.clusterStatus, test.slackStatus, TemplateData{})
			assert.Equal(t, actualNewSlackStatus, test.expectedNewSlackStatus)
			if test.expectedNoMessages {
				assert.Nil(t, actualMessages)
			} else {
				assert.NotNil(t, actualMessages)
			}
		})
	}
}
