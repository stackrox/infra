package cluster

import (
	"bytes"
	"strings"
	"text/template"
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/slack-go/slack"
	"github.com/stackrox/infra/cmd/infractl/common"
	v1 "github.com/stackrox/infra/generated/api/v1"
)

type slackTemplateData struct {
	Description string
	Flavor      string
	ID          string
	Remaining   string
	Scheduled   bool
	URL         string

	OwnerEmail string
	OwnerID    string
}

type slackStatus string

const (
	slackStatusSkip      slackStatus = "skip"
	slackStatusFailed    slackStatus = "failed"
	slackStatusDestroyed slackStatus = "destroyed"
	slackStatusReady     slackStatus = "ready"
	slackStatusCreating  slackStatus = "creating"
)

var (
	templatesFailed = []string{ // nolint:gochecknoglobals
		"<@{{.OwnerID}}> - Your {{if .Scheduled}}scheduled {{end}}{{if .Description}}*{{.Description}}* {{else}}*{{.ID}}* {{end}}cluster has failed! :fire:",
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
	}
)

func templateBlocks(context slackTemplateData, templates []string) []slack.MsgOption {
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

func formatSlackMessage(wfStatus v1.Status, slackStatus slackStatus, contextData slackTemplateData) (slackStatus, []slack.MsgOption) {
	switch {
	case slackStatus == slackStatusSkip:
		return slackStatusSkip, nil

	case wfStatus == v1.Status_FAILED && slackStatus != slackStatusFailed:
		return slackStatusFailed, templateBlocks(contextData, templatesFailed)

	case (wfStatus == v1.Status_DESTROYING || wfStatus == v1.Status_FINISHED) && slackStatus != slackStatusDestroyed:
		return slackStatusDestroyed, templateBlocks(contextData, templatesDestroyed)

	case wfStatus == v1.Status_READY && slackStatus != slackStatusReady:
		return slackStatusReady, templateBlocks(contextData, templatesReady)

	case wfStatus == v1.Status_CREATING && slackStatus != slackStatusCreating:
		return slackStatusCreating, templateBlocks(contextData, templatesCreating)

	default:
		return slackStatus, nil
	}
}

func slackTemplateContext(client Slacker, cluster *metaCluster) slackTemplateData {
	createdOn, _ := ptypes.Timestamp(cluster.CreatedOn)
	lifespan, _ := ptypes.Duration(cluster.Lifespan)
	remaining := time.Until(createdOn.Add(lifespan))

	data := slackTemplateData{
		Description: cluster.Description,
		Flavor:      cluster.Flavor,
		ID:          cluster.ID,
		OwnerEmail:  cluster.Owner,
		Remaining:   common.FormatExpiration(remaining),
		Scheduled:   cluster.EventID != "",
		URL:         cluster.URL,
	}

	if user, found := client.LookupUser(cluster.Owner); found {
		data.OwnerID = user.ID
	}

	return data
}
