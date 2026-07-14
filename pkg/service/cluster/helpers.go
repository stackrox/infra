package cluster

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/argoproj/argo-workflows/v4/pkg/apis/workflow/v1alpha1"
	v1 "github.com/stackrox/infra/generated/api/v1"
	"github.com/stackrox/infra/pkg/logging"
	"github.com/stackrox/infra/pkg/slack"
	"google.golang.org/protobuf/types/known/timestamppb"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
)

func getClusterIDFromWorkflow(workflow *v1alpha1.Workflow) string {
	clusterID := GetClusterID(workflow)
	if clusterID == "" {
		// Prior workflows used a direct mapping from Argo workflow name to Infra cluster ID
		clusterID = workflow.GetName()
	}
	return clusterID
}

// clusterFromWorkflow converts an Argo workflow into an infra cluster.
func clusterFromWorkflow(workflow v1alpha1.Workflow) *v1.Cluster {
	cluster := &v1.Cluster{
		ID:          getClusterIDFromWorkflow(&workflow),
		Status:      workflowStatus(workflow.Status),
		Flavor:      GetFlavor(&workflow),
		Owner:       GetOwner(&workflow),
		Lifespan:    GetLifespan(&workflow),
		Description: GetDescription(&workflow),
	}

	cluster.CreatedOn = timestamppb.New(workflow.Status.StartedAt.UTC())

	if !workflow.Status.FinishedAt.Time.IsZero() {
		cluster.DestroyedOn = timestamppb.New(workflow.Status.FinishedAt.UTC())
	}

	return cluster
}

func isWorkflowExpired(workflow v1alpha1.Workflow) bool {
	lifespan := GetLifespan(&workflow).AsDuration()
	workflowExpiryTime := workflow.Status.StartedAt.Add(lifespan)
	return time.Now().After(workflowExpiryTime)
}

func isNearingExpiry(workflow v1alpha1.Workflow) bool {
	lifespan := GetLifespan(&workflow).AsDuration()
	workflowExpiryTime := workflow.Status.StartedAt.Add(lifespan)
	return time.Now().Add(nearExpiry).After(workflowExpiryTime)
}

func isClusterOneOfAllowedStatuses(workflow *v1alpha1.Workflow, allowedStatuses []v1.Status) bool {
	status := workflowStatus(workflow.Status)
	return slices.Contains(allowedStatuses, status)
}

type metaCluster struct {
	*v1.Cluster

	EventID       string
	Expired       bool
	NearingExpiry bool
	Slack         slack.Status
	SlackDM       bool
}

// metaClusterFromWorkflow() converts an Argo workflow into an infra cluster
// with additional, non-cluster, metadata.
func (s *clusterImpl) metaClusterFromWorkflow(workflow v1alpha1.Workflow) (*metaCluster, error) {
	cluster := clusterFromWorkflow(workflow)
	expired := isWorkflowExpired(workflow)
	nearingExpiry := isNearingExpiry(workflow)

	cluster, err := s.getClusterDetailsFromArtifacts(cluster, workflow)
	if err != nil {
		return nil, err
	}

	return &metaCluster{
		Cluster:       cluster,
		Slack:         slack.Status(GetSlack(&workflow)),
		SlackDM:       GetSlackDM(&workflow),
		Expired:       expired,
		NearingExpiry: nearingExpiry,
		EventID:       GetEventID(&workflow),
	}, nil
}

// getClusterDetailsFromArtifacts - get those cluster details that are stored by workflow artifacts.
func (s *clusterImpl) getClusterDetailsFromArtifacts(cluster *v1.Cluster, workflow v1alpha1.Workflow) (*v1.Cluster, error) {

	flavorMetadata := make(map[string]*v1.FlavorArtifact)

	flavor, _, found := s.registry.Get(cluster.Flavor)
	if found && flavor.Artifacts != nil {
		flavorMetadata = flavor.Artifacts
	}

	for _, nodeStatus := range workflow.Status.Nodes {
		if nodeStatus.Outputs == nil {
			continue
		}

		for _, artifact := range nodeStatus.Outputs.Artifacts {
			if artifact.GCS == nil {
				continue
			}

			// The only artifact of concern are those explicity defined in
			// flavors.yaml artifacts section.
			if meta, found := flavorMetadata[artifact.Name]; found {

				// And only artifacts that are tagged with url or connect.
				_, foundURL := meta.Tags[artifactTagURL]
				_, foundConnect := meta.Tags[artifactTagConnect]
				if !foundURL && !foundConnect {
					continue
				}

				bucket, key := handleArtifactMigration(workflow, artifact)
				if bucket == "" || key == "" {
					continue
				}

				// Check cache first before making GCS API call
				contents, found := s.artifactCache.Get(bucket, key)
				if !found {
					// Cache miss - fetch from GCS and cache the result
					var err error
					contents, err = s.signer.Contents(bucket, key)
					if err != nil {
						return nil, err
					}
					s.artifactCache.Set(bucket, key, contents)
				}

				if _, found := meta.Tags[artifactTagURL]; found {
					cluster.URL = strings.TrimSpace(string(contents))
				}
				if _, found := meta.Tags[artifactTagConnect]; found {
					cluster.Connect = string(contents)
				}
			}
		}
	}

	cluster.Parameters = metaClusterParametersFromWorkflow(workflow)

	return cluster, nil
}

func metaClusterParametersFromWorkflow(workflow v1alpha1.Workflow) []*v1.Parameter {
	parameters := []*v1.Parameter{}
	for _, p := range workflow.Spec.Arguments.Parameters {
		description := ""
		if p.Description != nil {
			description = p.Description.String()
		}
		parameters = append(parameters, &v1.Parameter{
			Name:        p.Name,
			Description: description,
			Value:       p.GetValue(),
		})
	}

	return parameters
}

func handleArtifactMigration(workflow v1alpha1.Workflow, artifact v1alpha1.Artifact) (string, string) {
	var bucket string
	var key string

	if workflow.Status.ArtifactRepositoryRef != nil &&
		workflow.Status.ArtifactRepositoryRef.ArtifactRepository.GCS != nil &&
		workflow.Status.ArtifactRepositoryRef.ArtifactRepository.GCS.Bucket != "" {
		bucket = workflow.Status.ArtifactRepositoryRef.ArtifactRepository.GCS.Bucket
	} else if artifact.GCS != nil && artifact.GCS.Bucket != "" {
		bucket = artifact.GCS.Bucket
	}

	if artifact.GCS != nil && artifact.GCS.Key != "" {
		key = artifact.GCS.Key
	}

	if bucket == "" || key == "" {
		log.Log(logging.WARN, "cannot figure out bucket for artifact, possibly an upgrade issue, not fatal",
			"workflow-name", workflow.Name,
			"artifact", artifact,
			"artifact-repository", workflow.Status.ArtifactRepositoryRef,
		)
		return "", ""
	}

	return bucket, key
}

func workflowStatus(workflowStatus v1alpha1.WorkflowStatus) v1.Status {
	// https://godoc.org/github.com/argoproj/argo-workflows/v4/pkg/apis/workflow/v1alpha1#WorkflowStatus
	switch workflowStatus.Phase {
	case v1alpha1.WorkflowFailed, v1alpha1.WorkflowError:
		return v1.Status_FAILED

	case v1alpha1.WorkflowSucceeded:
		return v1.Status_FINISHED

	case v1alpha1.WorkflowPending:
		return v1.Status_CREATING

	case v1alpha1.WorkflowRunning:
		// https://godoc.org/github.com/argoproj/argo-workflows/v4/pkg/apis/workflow/v1alpha1#Nodes
		for _, node := range workflowStatus.Nodes {
			// https://godoc.org/github.com/argoproj/argo-workflows/v4/pkg/apis/workflow/v1alpha1#NodeType
			switch nodeType := node.Type; nodeType {
			case v1alpha1.NodeTypePod:
				if strings.Contains(node.Message, "ImagePullBackOff") {
					return v1.Status_FAILED
				}
				if strings.Contains(node.Message, "ErrImagePull") {
					return v1.Status_FAILED
				}
				if strings.Contains(node.Message, "Pod was active on the node longer than the specified deadline") {
					return v1.Status_FAILED
				}
			case v1alpha1.NodeTypeSuspend:
				switch node.Phase {
				case v1alpha1.NodeSucceeded:
					return v1.Status_DESTROYING
				case v1alpha1.NodeError, v1alpha1.NodeFailed, v1alpha1.NodeSkipped:
					panic("a suspend should not be able to fail?")
				case v1alpha1.NodeRunning, v1alpha1.NodePending:
					return v1.Status_READY
				}
			}
		}

		// No suspend node was found, which means one hasn't been run yet, which means that this cluster is still creating.
		return v1.Status_CREATING

	case "":
		return v1.Status_CREATING
	}

	panic("unknown situation")
}

// Returns an error with details of an aberrant condition if detected, nil otherwise.
// Intended to provide failure details to a user via slack post.
func workflowFailureDetails(workflowStatus v1alpha1.WorkflowStatus) error {
	switch workflowStatus.Phase {
	case v1alpha1.WorkflowRunning, v1alpha1.WorkflowFailed:
		for _, node := range workflowStatus.Nodes {
			if node.Type == v1alpha1.NodeTypePod {
				if strings.Contains(node.Message, "ImagePullBackOff") {
					msg := fmt.Sprintf("Workflow node `%s` has encountered an image pull back-off.", node.Name)
					return errors.New(msg)
				}
				if strings.Contains(node.Message, "ErrImagePull") {
					msg := fmt.Sprintf("Workflow node `%s` has encountered an image pull error.", node.Name)
					return errors.New(msg)
				}
				if strings.Contains(node.Message, "Pod was active on the node longer than the specified deadline") {
					msg := fmt.Sprintf("Workflow node `%s` has timed out.", node.Name)
					return errors.New(msg)
				}
			}
		}
	}
	return errors.New("")
}

// emailToLabelValue converts an email address to a Kubernetes label-safe value.
// Kubernetes label values must match ([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9] and be at most 63 characters.
func emailToLabelValue(email string) string {
	// Replace characters that aren't valid in Kubernetes labels
	result := strings.ReplaceAll(email, "@", ".at.")
	result = strings.ReplaceAll(result, "+", ".plus.")

	// Ensure max length of 63 characters
	if len(result) > 63 {
		result = result[:63]
	}

	// Ensure it starts and ends with alphanumeric (trim invalid leading/trailing chars)
	result = strings.TrimRight(result, "._-")
	result = strings.TrimLeft(result, "._-")

	return result
}

// buildLabelSelector constructs a Kubernetes label selector from a ClusterListRequest.
// This enables server-side filtering to reduce the amount of data transferred and processed.
func buildLabelSelector(req *v1.ClusterListRequest, email string) (labels.Selector, error) {
	selector := labels.NewSelector()

	// Filter out deleted workflows unless --expired is specified
	// (deleted workflows are a subset of expired workflows)
	if !req.Expired {
		requirement, err := labels.NewRequirement(labelDeleted, selection.NotEquals, []string{"true"})
		if err != nil {
			return nil, err
		}
		selector = selector.Add(*requirement)
	}

	// Filter by owner if not requesting all clusters
	if !req.All && email != "" {
		labelSafeEmail := emailToLabelValue(email)
		requirement, err := labels.NewRequirement(labelOwner, selection.Equals, []string{labelSafeEmail})
		if err != nil {
			return nil, err
		}
		selector = selector.Add(*requirement)
	}

	// Filter by allowed flavors if specified
	if len(req.AllowedFlavors) > 0 {
		requirement, err := labels.NewRequirement(labelFlavor, selection.In, req.AllowedFlavors)
		if err != nil {
			return nil, err
		}
		selector = selector.Add(*requirement)
	}

	return selector, nil
}
