package cluster

import (
	"context"
	"fmt"
	"time"

	"github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/duration"
	v1 "github.com/stackrox/infra/generated/api/v1"
	"github.com/stackrox/infra/pkg/logging"
	"github.com/stackrox/infra/pkg/service/cluster/helpers"
	"github.com/stackrox/infra/pkg/service/cluster/metadata"
	"github.com/stackrox/infra/pkg/service/middleware"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// Lifespan implements ClusterService.Lifespan.
func (s *clusterImpl) Lifespan(ctx context.Context, req *v1.LifespanRequest) (*duration.Duration, error) {
	owner, err := middleware.GetOwnerFromContext(ctx)
	if err != nil {
		return nil, err
	}
	log.AuditLog(logging.INFO, "cluster-lifespan", "received a lifespan update request for infra cluster",
		"actor", owner,
		"cluster-id", req.GetId(),
		"lifespan-update-method", req.GetMethod().String(),
		"lifespan", req.GetLifespan().String(),
	)

	workflow, err := s.argoClient.GetMostRecentArgoWorkflowFromClusterID(req.GetId())
	if err != nil {
		return nil, err
	}

	return s.lifespan(ctx, req, workflow)
}

func (s *clusterImpl) lifespan(ctx context.Context, req *v1.LifespanRequest, workflow *v1alpha1.Workflow) (*duration.Duration, error) {
	log.Log(logging.INFO, "will apply a lifespan update to argo workflow",
		"workflow-name", workflow.GetName(),
		"lifespan-update-method", req.GetMethod().String(),
		"lifespan", req.GetLifespan().String(),
	)

	lifespanRequest, _ := ptypes.Duration(req.Lifespan)
	lifespanCurrent := time.Duration(0)
	lifespanUpdated := time.Duration(0)

	// If we're applying a relative lifespan (by adding or subtracting), we
	// need to know the current lifespan. Get the named workflow to obtain said
	// current lifespan.
	if req.Method != v1.LifespanRequest_REPLACE {
		lifespanCurrent, _ = ptypes.Duration(metadata.GetLifespan(workflow))
	}

	// Compute the updated lifespan using the requested method.
	switch req.Method {
	case v1.LifespanRequest_REPLACE:
		lifespanUpdated = lifespanRequest
	case v1.LifespanRequest_ADD:
		lifespanUpdated = lifespanCurrent + lifespanRequest
	case v1.LifespanRequest_SUBTRACT:
		lifespanUpdated = lifespanCurrent - lifespanRequest
	}

	// Sanity check that our updated lifespan doesn't go negative.
	if lifespanUpdated <= 0 {
		lifespanUpdated = 0
	}

	// Construct our replacement patch
	payloadBytes, err := helpers.FormatAnnotationPatch(metadata.AnnotationLifespanKey, fmt.Sprint(lifespanUpdated))
	if err != nil {
		return nil, err
	}

	// Submit the patch.
	_, err = s.k8sWorkflowsClient.Patch(ctx, workflow.GetName(), types.JSONPatchType, payloadBytes, metav1.PatchOptions{})
	if err != nil {
		log.Log(logging.ERROR, "error occurred updating the argo workflow", "workflow-name", workflow.GetName(), "error", err)
		return nil, err
	}

	// Return the remaining lifespan.
	remaining := time.Until(workflow.CreationTimestamp.Add(lifespanUpdated))
	return ptypes.DurationProto(remaining), nil
}
