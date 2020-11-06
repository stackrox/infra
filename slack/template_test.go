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
		clusterIsNearingExpiry bool
		slackStatus            Status
		expectedNewSlackStatus Status
		expectedNoMessages     bool
	}{
		{
			title:                  "failed nop",
			clusterStatus:          v1.Status_FAILED,
			clusterIsNearingExpiry: false,
			slackStatus:            StatusFailed,
			expectedNewSlackStatus: StatusFailed,
			expectedNoMessages:     true,
		},
		{
			title:                  "failed status blank",
			clusterStatus:          v1.Status_FAILED,
			clusterIsNearingExpiry: false,
			slackStatus:            "",
			expectedNewSlackStatus: StatusFailed,
		},
		{
			title:                  "failed status other",
			clusterStatus:          v1.Status_FAILED,
			clusterIsNearingExpiry: false,
			slackStatus:            "qwertyuiop",
			expectedNewSlackStatus: StatusFailed,
		},

		{
			title:                  "creating nop",
			clusterStatus:          v1.Status_CREATING,
			clusterIsNearingExpiry: false,
			slackStatus:            StatusCreating,
			expectedNewSlackStatus: StatusCreating,
			expectedNoMessages:     true,
		},
		{
			title:                  "creating near expiration",
			clusterStatus:          v1.Status_CREATING,
			clusterIsNearingExpiry: true,
			slackStatus:            StatusCreating,
			expectedNewSlackStatus: StatusCreating,
			expectedNoMessages:     true,
		},
		{
			title:                  "creating status blank",
			clusterStatus:          v1.Status_CREATING,
			clusterIsNearingExpiry: false,
			slackStatus:            "",
			expectedNewSlackStatus: StatusCreating,
		},
		{
			title:                  "creating status other",
			clusterStatus:          v1.Status_CREATING,
			clusterIsNearingExpiry: false,
			slackStatus:            "qwertyuiop",
			expectedNewSlackStatus: StatusCreating,
		},

		{
			title:                  "ready nop",
			clusterStatus:          v1.Status_READY,
			clusterIsNearingExpiry: false,
			slackStatus:            StatusReady,
			expectedNewSlackStatus: StatusReady,
			expectedNoMessages:     true,
		},
		{
			title:                  "ready status blank",
			clusterStatus:          v1.Status_READY,
			clusterIsNearingExpiry: false,
			slackStatus:            "",
			expectedNewSlackStatus: StatusReady,
		},
		{
			title:                  "ready status other",
			clusterStatus:          v1.Status_READY,
			clusterIsNearingExpiry: false,
			slackStatus:            "qwertyuiop",
			expectedNewSlackStatus: StatusReady,
		},

		{
			title:                  "ready -> nearing expiry",
			clusterStatus:          v1.Status_READY,
			clusterIsNearingExpiry: true,
			slackStatus:            StatusReady,
			expectedNewSlackStatus: StatusNearingExpiry,
			expectedNoMessages:     false,
		},
		{
			title:                  "nearing expiry nop",
			clusterStatus:          v1.Status_READY,
			clusterIsNearingExpiry: true,
			slackStatus:            StatusNearingExpiry,
			expectedNewSlackStatus: StatusNearingExpiry,
			expectedNoMessages:     true,
		},
		{
			title:                  "nearing expiry -> ready (lifespan update)",
			clusterStatus:          v1.Status_READY,
			clusterIsNearingExpiry: false,
			slackStatus:            StatusNearingExpiry,
			expectedNewSlackStatus: StatusReady,
			expectedNoMessages:     true,
		},
		{
			title:                  "nearing expiry -> destroyed",
			clusterStatus:          v1.Status_DESTROYING,
			clusterIsNearingExpiry: true,
			slackStatus:            StatusNearingExpiry,
			expectedNewSlackStatus: StatusDestroyed,
			expectedNoMessages:     false,
		},
		{
			title:                  "nearing expiry -> destroyed (time is irrelevant)",
			clusterStatus:          v1.Status_DESTROYING,
			clusterIsNearingExpiry: false,
			slackStatus:            StatusNearingExpiry,
			expectedNewSlackStatus: StatusDestroyed,
			expectedNoMessages:     false,
		},

		{
			title:                  "destroyed (destroying) nop",
			clusterStatus:          v1.Status_DESTROYING,
			clusterIsNearingExpiry: true,
			slackStatus:            StatusDestroyed,
			expectedNewSlackStatus: StatusDestroyed,
			expectedNoMessages:     true,
		},
		{
			title:                  "destroyed (destroying) status blank",
			clusterStatus:          v1.Status_DESTROYING,
			clusterIsNearingExpiry: true,
			slackStatus:            "",
			expectedNewSlackStatus: StatusDestroyed,
		},
		{
			title:                  "destroyed (destroying) status other",
			clusterStatus:          v1.Status_DESTROYING,
			clusterIsNearingExpiry: true,
			slackStatus:            "qwertyuiop",
			expectedNewSlackStatus: StatusDestroyed,
		},
		{
			title:                  "destroyed (finished) nop",
			clusterStatus:          v1.Status_FINISHED,
			clusterIsNearingExpiry: true,
			slackStatus:            StatusDestroyed,
			expectedNewSlackStatus: StatusDestroyed,
			expectedNoMessages:     true,
		},
		{
			title:                  "destroyed (finished) status blank",
			clusterStatus:          v1.Status_FINISHED,
			clusterIsNearingExpiry: true,
			slackStatus:            "",
			expectedNewSlackStatus: StatusDestroyed,
		},
		{
			title:                  "destroyed (finished) status other",
			clusterStatus:          v1.Status_FINISHED,
			clusterIsNearingExpiry: true,
			slackStatus:            "qwertyuiop",
			expectedNewSlackStatus: StatusDestroyed,
		},
	}

	for index, test := range tests {
		name := fmt.Sprintf("%d %s", index+1, test.title)
		t.Run(name, func(t *testing.T) {
			actualNewSlackStatus, actualMessages := FormatSlackMessage(test.clusterStatus, test.clusterIsNearingExpiry, test.slackStatus, TemplateData{})
			assert.Equal(t, actualNewSlackStatus, test.expectedNewSlackStatus)
			if test.expectedNoMessages {
				assert.Nil(t, actualMessages)
			} else {
				assert.NotNil(t, actualMessages)
			}
		})
	}
}
