package cluster

import (
	"fmt"
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/slack-go/slack"
	"github.com/stackrox/infra/cmd/infractl/common"
	v1 "github.com/stackrox/infra/generated/api/v1"
)

type slackStatus string

const (
	slackStatusFailed    slackStatus = "failed"
	slackStatusDestroyed slackStatus = "destroyed"
	slackStatusReady     slackStatus = "ready"
	slackStatusCreating  slackStatus = "creating"
)

func formatSlackMessage(cluster *v1.Cluster, wfStatus v1.Status, slackStatus slackStatus, stackroxURL string) (slackStatus, []slack.MsgOption) {
	createdOn, _ := ptypes.Timestamp(cluster.CreatedOn)
	lifespan, _ := ptypes.Duration(cluster.Lifespan)
	remaining := time.Until(createdOn.Add(lifespan))

	switch {
	case wfStatus == v1.Status_FAILED && slackStatus != slackStatusFailed:
		headerText := slack.NewTextBlockObject("mrkdwn", fmt.Sprintf(
			"%s - Cluster for *%s* has failed! :fire:",
			cluster.Owner,
			cluster.Description,
		), false, false)

		infoText := slack.NewTextBlockObject("mrkdwn", fmt.Sprintf(
			"*ID*: %s\n*Flavor*: %s",
			cluster.ID,
			cluster.Flavor,
		), false, false)

		return slackStatusFailed, []slack.MsgOption{
			slack.MsgOptionBlocks(
				slack.NewSectionBlock(headerText, nil, nil),
				slack.NewSectionBlock(infoText, nil, nil),
			),
		}

	case (wfStatus == v1.Status_DESTROYING || wfStatus == v1.Status_FINISHED) && slackStatus != slackStatusDestroyed:
		headerText := slack.NewTextBlockObject("mrkdwn", fmt.Sprintf(
			"%s - Cluster for *%s* has been destroyed. :bomb:",
			cluster.Owner,
			cluster.Description,
		), false, false)

		infoText := slack.NewTextBlockObject("mrkdwn", fmt.Sprintf(
			"*ID*: %s\n*Flavor*: %s",
			cluster.ID,
			cluster.Flavor,
		), false, false)

		return slackStatusDestroyed, []slack.MsgOption{
			slack.MsgOptionBlocks(
				slack.NewSectionBlock(headerText, nil, nil),
				slack.NewSectionBlock(infoText, nil, nil),
			),
		}

	case wfStatus == v1.Status_READY && slackStatus != slackStatusReady:
		headerText := slack.NewTextBlockObject("mrkdwn", fmt.Sprintf(
			"%s - Cluster for *%s* is now ready! :parrot:",
			cluster.Owner,
			cluster.Description,
		), false, false)

		if stackroxURL != "" {
			headerText = slack.NewTextBlockObject("mrkdwn", fmt.Sprintf(
				"%s - Cluster for *%s* is now ready! :parrot:\n Browse to *%s* to login.",
				cluster.Owner,
				cluster.Description,
				stackroxURL,
			), false, false)
		}

		infoText := slack.NewTextBlockObject("mrkdwn", fmt.Sprintf(
			"*ID*: %s\n*Flavor*: %s\n*Expiration*: %s",
			cluster.ID,
			cluster.Flavor,
			common.FormatExpiration(remaining),
		), false, false)

		connect1Text := slack.NewTextBlockObject("mrkdwn", fmt.Sprintf(
			":thinking_face: To view cluster *info*, you can run:\n ```$ infractl cluster info %s```",
			cluster.ID,
		), false, false)

		connect2Text := slack.NewTextBlockObject("mrkdwn", fmt.Sprintf(
			":pencil: To read cluster *logs*, you can run:\n ```$ infractl cluster logs %s```",
			cluster.ID,
		), false, false)

		connect3Text := slack.NewTextBlockObject("mrkdwn", fmt.Sprintf(
			":package: To download cluster *artifacts*, you can run:\n ```$ infractl cluster artifacts %s```",
			cluster.ID,
		), false, false)

		return slackStatusReady, []slack.MsgOption{
			slack.MsgOptionBlocks(
				slack.NewSectionBlock(headerText, nil, nil),
				slack.NewSectionBlock(infoText, nil, nil),
				slack.NewSectionBlock(connect1Text, nil, nil),
				slack.NewSectionBlock(connect2Text, nil, nil),
				slack.NewSectionBlock(connect3Text, nil, nil),
			),
		}

	case wfStatus == v1.Status_CREATING && slackStatus != slackStatusCreating:
		headerText := slack.NewTextBlockObject("mrkdwn", fmt.Sprintf(
			"%s - Cluster for *%s* is being created. :rocket:",
			cluster.Owner,
			cluster.Description,
		), false, false)

		infoText := slack.NewTextBlockObject("mrkdwn", fmt.Sprintf(
			"*ID*: %s\n*Flavor*: %s",
			cluster.ID,
			cluster.Flavor,
		), false, false)

		connect1Text := slack.NewTextBlockObject("mrkdwn", fmt.Sprintf(
			":thinking_face: To view cluster *info*, you can run:\n ```$ infractl cluster info %s```",
			cluster.ID,
		), false, false)

		return slackStatusCreating, []slack.MsgOption{
			slack.MsgOptionBlocks(
				slack.NewSectionBlock(headerText, nil, nil),
				slack.NewSectionBlock(infoText, nil, nil),
				slack.NewSectionBlock(connect1Text, nil, nil),
			),
		}

	default:
		return slackStatus, nil
	}
}
