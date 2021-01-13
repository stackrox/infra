package slack

import (
	"bytes"
	"strings"
	"text/template"

	"github.com/slack-go/slack"
	v1 "github.com/stackrox/infra/generated/api/v1"
)

// TemplateData represents the available context that is passed when executing
// Slack message templates.
type TemplateData struct {
	Description    string
	Flavor         string
	ID             string
	Remaining      string
	Scheduled      bool
	URL            string
	OwnerEmail     string
	OwnerID        string
	FailureDetails string
}

// Status represents which lifecycle stage a cluster has most recently sent a
// slack message for.
type Status string

const (
	// StatusSkip is for when cluster should not result in Slack messages.
	StatusSkip Status = "skip"

	// StatusFailed is for when a cluster has failed.
	StatusFailed Status = "failed"

	// StatusDestroyed is for when a cluster is being deleted.
	StatusDestroyed Status = "destroyed"

	// StatusReady is for when a cluster is ready.
	StatusReady Status = "ready"

	// StatusNearingExpiry is for when a cluster is close to expiry.
	StatusNearingExpiry Status = "nearing_expiry"

	// StatusCreating is for when a cluster is being created.
	StatusCreating Status = "creating"
)

var (
	templatesFailed = []string{ // nolint:gochecknoglobals
		"<@{{.OwnerID}}> - Your {{if .Scheduled}}scheduled {{end}}{{if .Description}}*{{.Description}}* {{else}}*{{.ID}}* {{end}}cluster has failed!{{if .FailureDetails}} {{.FailureDetails}}{{end}} :fire:",
	}

	templatesDestroyed = []string{ // nolint:gochecknoglobals
		":skull_and_crossbones: The {{if .Scheduled}}scheduled {{end}}{{if .Description}}*{{.Description}}* {{else}}*{{.ID}}* {{end}}cluster has been destroyed.",
	}

	templatesReady = []string{ // nolint:gochecknoglobals
		"<@{{.OwnerID}}> - Your {{if .Scheduled}}scheduled {{end}}{{if .Description}}*{{.Description}}* {{else}}*{{.ID}}* {{end}}cluster is now ready! :parrot:",
		"{{if .URL}}:earth_americas: Browse to *{{.URL}}* to login.{{end}}",
		":clock2: This cluster has about *{{.Remaining}}* before it is destroyed.",
		":thinking_face: To view cluster *info*, you can run:\n```$ infractl get {{.ID}}```",
		":pencil: To read cluster *logs*, you can run:\n```$ infractl logs {{.ID}}```",
		":package: To download cluster *artifacts*, you can run:\n```$ infractl artifacts {{.ID}}```",
	}

	templatesCreating = []string{ // nolint:gochecknoglobals
		"<@{{.OwnerID}}> - Your {{if .Scheduled}}scheduled {{end}}{{if .Description}}*{{.Description}}* {{else}}*{{.ID}}* {{end}}cluster is being created. :rocket:",
		":clock2: This cluster has about *{{.Remaining}}* before it is destroyed.",
		":thinking_face: To view cluster *info*, you can run:\n ```$ infractl get {{.ID}}```",
		":link: Or go to: https://infra.rox.systems/cluster/{{.ID}}",
	}

	templatesNearingExpiry = []string{ // nolint:gochecknoglobals
		"<@{{.OwnerID}}> - Your {{if .Scheduled}}scheduled {{end}}{{if .Description}}*{{.Description}}* {{else}}*{{.ID}}* {{end}}cluster has about *{{.Remaining}}*. :skull_and_crossbones:",
		":clock2: To buy more time, you can run:\n```$ infractl lifespan {{.ID}} '+1h'```",
		":link: Or go to: https://infra.rox.systems/cluster/{{.ID}}",
	}
)

func templateBlocks(context TemplateData, templates []string) []slack.MsgOption {
	blocks := make([]slack.Block, 0, len(templates))
	for _, raw := range templates {
		tpl := template.Must(template.New("template").Parse(raw))
		var buf bytes.Buffer
		if err := tpl.Execute(&buf, context); err != nil {
			panic(err)
		}

		if strings.TrimSpace(buf.String()) == "" {
			continue
		}

		blocks = append(blocks,
			slack.NewSectionBlock(
				slack.NewTextBlockObject(
					slack.MarkdownType,
					buf.String(),
					false,
					false,
				),
				nil,
				nil,
			),
		)
	}

	return []slack.MsgOption{
		slack.MsgOptionBlocks(
			blocks...,
		),
	}
}

// FormatSlackMessage formats the correct Slack message given the current cluster state.
func FormatSlackMessage(wfStatus v1.Status, clusterIsNearingExpiry bool, slackStatus Status, contextData TemplateData) (Status, []slack.MsgOption) {
	switch {
	case slackStatus == StatusSkip:
		return StatusSkip, nil

	case wfStatus == v1.Status_FAILED && slackStatus != StatusFailed:
		return StatusFailed, templateBlocks(contextData, templatesFailed)

	case (wfStatus == v1.Status_DESTROYING || wfStatus == v1.Status_FINISHED) && slackStatus != StatusDestroyed:
		return StatusDestroyed, templateBlocks(contextData, templatesDestroyed)

	case wfStatus == v1.Status_READY && slackStatus != StatusReady && slackStatus != StatusNearingExpiry:
		return StatusReady, templateBlocks(contextData, templatesReady)

	case wfStatus == v1.Status_READY && slackStatus == StatusReady && clusterIsNearingExpiry:
		return StatusNearingExpiry, templateBlocks(contextData, templatesNearingExpiry)

	case wfStatus == v1.Status_READY && slackStatus == StatusNearingExpiry && !clusterIsNearingExpiry:
		return StatusReady, nil

	case wfStatus == v1.Status_CREATING && slackStatus != StatusCreating:
		return StatusCreating, templateBlocks(contextData, templatesCreating)

	default:
		return slackStatus, nil
	}
}
