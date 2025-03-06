package cluster

import (
	"context"
	"time"

	"github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	"github.com/golang/protobuf/ptypes"
	"github.com/stackrox/infra/cmd/infractl/common"
	v1 "github.com/stackrox/infra/generated/api/v1"
	"github.com/stackrox/infra/pkg/logging"
	"github.com/stackrox/infra/pkg/service/cluster/helpers"
	"github.com/stackrox/infra/pkg/service/cluster/metadata"
	"github.com/stackrox/infra/pkg/slack"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func (s *clusterImpl) startSlackCheck() {
	for ; ; time.Sleep(slackCheckInterval) {
		workflowList, err := s.argoClient.ListWorkflows()
		if err != nil {
			log.Log(logging.ERROR, "failed to list workflows", "error", err)
			continue
		}

		for _, workflow := range workflowList.Items {
			s.slackCheckWorkflow(workflow)
		}
	}
}

func (s *clusterImpl) slackCheckWorkflow(workflow v1alpha1.Workflow) {
	if slack.IsSlackComplete(slack.Status(metadata.GetSlack(&workflow))) {
		return
	}

	metacluster, err := s.metaClusterFromWorkflow(workflow)
	if err != nil {
		log.Log(logging.ERROR, "failed to convert workflow to meta-cluster", "workflow-name", workflow.Name, "error", err)
		return
	}

	// Generate a Slack message for our current cluster state.
	failureDetails := helpers.WorkflowFailureDetails(workflow.Status).Error()
	data := slackTemplateContext(s.slackClient, metacluster, failureDetails)
	newSlackStatus, message := slack.FormatSlackMessage(metacluster.Status, metacluster.NearingExpiry, metacluster.Slack, data)

	// Only bother to send a message if there is one to send.
	if message != nil {
		sent := false
		user, found := s.slackClient.LookupUser(metacluster.Owner)
		if found && metacluster.SlackDM {
			if err := s.slackClient.PostMessageToUser(user, message...); err != nil {
				log.Log(logging.ERROR, "failed to send Slack message directly to user", "user-email", user.Profile.Email, "error", err)
			} else {
				sent = true
			}
		}
		if !sent {
			if err := s.slackClient.PostMessage(message...); err != nil {
				log.Log(logging.ERROR, "failed to send Slack message", "error", err)
				return
			}
		}

		if metacluster.Status == v1.Status_FAILED {
			clusterID := metadata.GetClusterID(&workflow)
			err = s.bqClient.InsertClusterDeletionRecord(context.Background(), clusterID, workflow.GetName())
			if err != nil {
				log.Log(logging.WARN, "failed to record cluster deletion", "cluster-id", clusterID, "error", err)
			}
		}
	}

	// Only bother to update workflow annotation if our phase has
	// transitioned.
	if newSlackStatus != metacluster.Slack {
		// Construct our replacement patch
		payloadBytes, err := helpers.FormatAnnotationPatch(metadata.AnnotationSlackKey, string(newSlackStatus))
		if err != nil {
			log.Log(logging.ERROR, "failed to format Slack annotation patch", "error", err)
			return
		}

		// Submit the patch.
		_, err = s.k8sWorkflowsClient.Patch(context.Background(), workflow.GetName(), types.JSONPatchType, payloadBytes, metav1.PatchOptions{})
		if err != nil {
			log.Log(logging.ERROR, "failed to patch Slack annotation",
				"cluster-id", metacluster.Cluster.ID,
				"workflow-name", workflow.GetName(),
				"error", err,
			)
			return
		}
	}
}

func slackTemplateContext(client slack.Slacker, cluster *metaCluster, failureDetails string) slack.TemplateData {
	createdOn, _ := ptypes.Timestamp(cluster.CreatedOn)
	lifespan, _ := ptypes.Duration(cluster.Lifespan)
	remaining := time.Until(createdOn.Add(lifespan))

	data := slack.TemplateData{
		Description:    cluster.Description,
		Flavor:         cluster.Flavor,
		ID:             cluster.ID,
		OwnerEmail:     cluster.Owner,
		Remaining:      common.FormatExpiration(remaining),
		URL:            cluster.URL,
		FailureDetails: failureDetails,
	}

	if user, found := client.LookupUser(cluster.Owner); found {
		data.OwnerID = user.ID
	}

	return data
}
