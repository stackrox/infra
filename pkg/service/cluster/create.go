package cluster

import (
	"context"
	"fmt"
	"time"

	"github.com/golang/protobuf/ptypes"
	v1 "github.com/stackrox/infra/generated/api/v1"
	"github.com/stackrox/infra/pkg/logging"
	"github.com/stackrox/infra/pkg/service/cluster/helpers"
	"github.com/stackrox/infra/pkg/service/cluster/metadata"
	"github.com/stackrox/infra/pkg/service/metrics"
	"github.com/stackrox/infra/pkg/service/middleware"
	"github.com/stackrox/infra/pkg/slack"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Create implements ClusterService.Create.
func (s *clusterImpl) Create(ctx context.Context, req *v1.CreateClusterRequest) (*v1.ResourceByID, error) {
	owner, err := middleware.GetOwnerFromContext(ctx)
	if err != nil {
		return nil, err
	}

	log.AuditLog(logging.INFO, "cluster-create", "received a create request for flavor",
		"actor", owner,
		"flavor-id", req.GetID(),
	)
	return s.create(req, owner)
}

func (s *clusterImpl) create(req *v1.CreateClusterRequest, owner string) (*v1.ResourceByID, error) {
	flav, workflow, found := s.registry.Get(req.ID)
	if !found {
		return nil, status.Errorf(codes.NotFound, "flavor %q not found", req.ID)
	}

	// Combine any hardcoded or default workflow parameters with the user
	// provided parameters. Or return an error if the user provided
	// insufficient or superfluous parameters.
	workflowParams, err := checkAndEnrichParameters(flav.Parameters, req.Parameters)
	if err != nil {
		return nil, err
	}
	workflow.Spec.Arguments.Parameters = workflowParams

	// Use the user supplied name as the root of Argo workflow name and the Infra cluster Id.
	clusterID, ok := req.Parameters["name"]
	if ok {
		workflow.ObjectMeta.GenerateName = clusterID + "-"
	} else {
		return nil, fmt.Errorf("parameter 'name' was not provided")
	}

	// Make sure there is no running argo workflow for infra cluster with the same ID
	existingWorkflow, _ := s.argoClient.GetMostRecentArgoWorkflowFromClusterID(clusterID)
	if existingWorkflow != nil {
		switch helpers.StatusFromWorkflowStatus(existingWorkflow.Status) {
		case v1.Status_FAILED, v1.Status_FINISHED:
			// It is ok to reuse a cluster ID from a failed or finished workflow.
			log.Log(logging.INFO, "a completed argo workflow exists",
				"workflow-name", existingWorkflow.GetName(),
				"cluster-id", clusterID,
				"workflow-phase", existingWorkflow.Status.Phase,
			)

		default:
			log.Log(logging.WARN, "infra cluster create failed due to an existing busy argo workflow",
				"workflow-name", existingWorkflow.GetName(),
				"cluster-id", clusterID,
				"workflow-phase", existingWorkflow.Status.Phase,
			)
			return nil, status.Errorf(
				codes.AlreadyExists,
				"An infra cluster ID %q already exists in state %s.",
				clusterID, helpers.StatusFromWorkflowStatus(existingWorkflow.Status).String(),
			)
		}
	}

	// Determine the lifespan for this cluster. Apply some sanity/bounds
	// checking on provided lifespans.
	lifespan, _ := ptypes.Duration(req.Lifespan)
	if lifespan <= 0 {
		lifespan = 3 * time.Hour
	}

	var slackStatus slack.Status
	if req.GetNoSlack() {
		slackStatus = slack.StatusSkip
	}

	slackDM := "no"
	if req.GetSlackDM() {
		slackDM = "yes"
	}

	// Set workflow metadata annotations.
	workflow.SetAnnotations(map[string]string{
		metadata.AnnotationDescriptionKey: req.Description,
		metadata.AnnotationFlavorKey:      flav.ID,
		metadata.AnnotationLifespanKey:    fmt.Sprint(lifespan),
		metadata.AnnotationOwnerKey:       owner,
		metadata.AnnotationSlackKey:       string(slackStatus),
		metadata.AnnotationSlackDMKey:     slackDM,
	})

	workflow.SetLabels(map[string]string{
		metadata.LabelClusterID: clusterID,
	})

	log.Log(logging.INFO, "will create an infra cluster",
		"flavor-id", flav.GetID(),
		"cluster-id", clusterID,
		"cluster-owner", owner,
	)

	created, err := s.argoClient.CreateWorkflow(&workflow)
	if err != nil {
		log.Log(logging.ERROR, "creating a new cluster failed", "cluster-id", clusterID)
	}
	log.Log(logging.INFO, "created an argo workflow for a new infra cluster",
		"workflow-name", created.GetName(),
		"cluster-id", clusterID,
	)

	metrics.FlavorsUsedCounter.WithLabelValues(flav.ID).Inc()

	err = s.bqClient.InsertClusterCreationRecord(context.Background(), clusterID, created.GetName(), flav.GetID(), owner)
	if err != nil {
		log.Log(logging.WARN, "failed to record cluster creation", "cluster-id", clusterID, "error", err)
	}

	return &v1.ResourceByID{Id: clusterID}, nil
}
