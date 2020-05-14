package cluster

import (
	"fmt"
	"testing"

	"github.com/slack-go/slack"

	v1 "github.com/stackrox/infra/generated/api/v1"
	"github.com/stretchr/testify/assert"
)

func Test(t *testing.T) {
	tests := []struct {
		title                  string
		clusterStatus          v1.Status
		slackStatus            slackStatus
		expectedNewSlackStatus slackStatus
		expectedNoMessages     bool
	}{
		{
			title:                  "failed nop",
			clusterStatus:          v1.Status_FAILED,
			slackStatus:            slackStatusFailed,
			expectedNewSlackStatus: slackStatusFailed,
			expectedNoMessages:     true,
		},
		{
			title:                  "failed status blank",
			clusterStatus:          v1.Status_FAILED,
			slackStatus:            "",
			expectedNewSlackStatus: slackStatusFailed,
		},
		{
			title:                  "failed status other",
			clusterStatus:          v1.Status_FAILED,
			slackStatus:            "qwertyuiop",
			expectedNewSlackStatus: slackStatusFailed,
		},

		{
			title:                  "creating nop",
			clusterStatus:          v1.Status_CREATING,
			slackStatus:            slackStatusCreating,
			expectedNewSlackStatus: slackStatusCreating,
			expectedNoMessages:     true,
		},
		{
			title:                  "creating status blank",
			clusterStatus:          v1.Status_CREATING,
			slackStatus:            "",
			expectedNewSlackStatus: slackStatusCreating,
		},
		{
			title:                  "creating status other",
			clusterStatus:          v1.Status_CREATING,
			slackStatus:            "qwertyuiop",
			expectedNewSlackStatus: slackStatusCreating,
		},

		{
			title:                  "ready nop",
			clusterStatus:          v1.Status_READY,
			slackStatus:            slackStatusReady,
			expectedNewSlackStatus: slackStatusReady,
			expectedNoMessages:     true,
		},
		{
			title:                  "ready status blank",
			clusterStatus:          v1.Status_READY,
			slackStatus:            "",
			expectedNewSlackStatus: slackStatusReady,
		},
		{
			title:                  "ready status other",
			clusterStatus:          v1.Status_READY,
			slackStatus:            "qwertyuiop",
			expectedNewSlackStatus: slackStatusReady,
		},

		{
			title:                  "destroyed (destroying) nop",
			clusterStatus:          v1.Status_DESTROYING,
			slackStatus:            slackStatusDestroyed,
			expectedNewSlackStatus: slackStatusDestroyed,
			expectedNoMessages:     true,
		},
		{
			title:                  "destroyed (destroying) status blank",
			clusterStatus:          v1.Status_DESTROYING,
			slackStatus:            "",
			expectedNewSlackStatus: slackStatusDestroyed,
		},
		{
			title:                  "destroyed (destroying) status other",
			clusterStatus:          v1.Status_DESTROYING,
			slackStatus:            "qwertyuiop",
			expectedNewSlackStatus: slackStatusDestroyed,
		},

		{
			title:                  "destroyed (finished) nop",
			clusterStatus:          v1.Status_FINISHED,
			slackStatus:            slackStatusDestroyed,
			expectedNewSlackStatus: slackStatusDestroyed,
			expectedNoMessages:     true,
		},
		{
			title:                  "destroyed (finished) status blank",
			clusterStatus:          v1.Status_FINISHED,
			slackStatus:            "",
			expectedNewSlackStatus: slackStatusDestroyed,
		},
		{
			title:                  "destroyed (finished) status other",
			clusterStatus:          v1.Status_FINISHED,
			slackStatus:            "qwertyuiop",
			expectedNewSlackStatus: slackStatusDestroyed,
		},
	}

	var dummy metaCluster

	data := slackTemplateContext(mockClient("example@example.com"), &dummy)

	for index, test := range tests {
		name := fmt.Sprintf("%d %s", index+1, test.title)
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			actualNewSlackStatus, actualMessages := formatSlackMessage(test.clusterStatus, test.slackStatus, data)
			assert.Equal(t, actualNewSlackStatus, test.expectedNewSlackStatus)
			if test.expectedNoMessages {
				assert.Nil(t, actualMessages)
			} else {
				assert.NotNil(t, actualMessages)
			}
		})
	}
}

type mockClient string

func (m mockClient) PostMessage(channelID string, options ...slack.MsgOption) (string, string, error) {
	panic("unimplemented")
}

func (m mockClient) LookupUser(email string) (slack.User, bool) {
	return slack.User{ID: string(m)}, true
}

var _ Slacker = (*mockClient)(nil)
