package cluster

import (
	"slices"
	"strings"
	"time"

	"github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	"github.com/golang/protobuf/ptypes"
	"github.com/stackrox/infra/pkg/service/cluster/metadata"
)

// when is a cluster considered near expiration
var nearExpiry = 30 * time.Minute

func hasClusterNamePrefix(workflow *v1alpha1.Workflow, prefix string) bool {
	return strings.HasPrefix(metadata.GetClusterID(workflow), prefix)
}

func isClusterExpired(workflow v1alpha1.Workflow) bool {
	lifespan, _ := ptypes.Duration(metadata.GetLifespan(&workflow))

	workflowExpiryTime := workflow.Status.StartedAt.Time.Add(lifespan)
	return time.Now().After(workflowExpiryTime)
}

func isClusterNearingExpiry(workflow v1alpha1.Workflow) bool {
	lifespan, _ := ptypes.Duration(metadata.GetLifespan(&workflow))

	workflowExpiryTime := workflow.Status.StartedAt.Time.Add(lifespan)
	return time.Now().Add(nearExpiry).After(workflowExpiryTime)
}

func isClusterOneOfAllowedFlavors(workflow *v1alpha1.Workflow, allowedFlavors []string) bool {
	flavor := metadata.GetFlavor(workflow)
	return slices.Contains(allowedFlavors, flavor)
}

func isClusterOwnedByCurrentUser(workflow *v1alpha1.Workflow, email string) bool {
	return metadata.GetOwner(workflow) == email
}
