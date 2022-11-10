package cluster

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	"github.com/golang/protobuf/ptypes"
	v1 "github.com/stackrox/infra/generated/api/v1"
	"github.com/stackrox/infra/slack"
)

// clusterFromWorkflow converts an Argo workflow into a cluster.
func clusterFromWorkflow(workflow v1alpha1.Workflow) *v1.Cluster {
	var clusterID string
	clusterID, ok := workflow.GetObjectMeta().GetAnnotations()[annotationClusterID]
	if !ok {
		// Prior workflows used a direct mapping from Argo workflow name to Infra cluster ID
		clusterID = workflow.GetName()
	}

	cluster := &v1.Cluster{
		ID:          clusterID,
		Status:      workflowStatus(workflow.Status),
		Flavor:      GetFlavor(&workflow),
		Owner:       GetOwner(&workflow),
		Lifespan:    GetLifespan(&workflow),
		Description: GetDescription(&workflow),
	}

	cluster.CreatedOn, _ = ptypes.TimestampProto(workflow.Status.StartedAt.Time.UTC())

	if !workflow.Status.FinishedAt.Time.IsZero() {
		cluster.DestroyedOn, _ = ptypes.TimestampProto(workflow.Status.FinishedAt.Time.UTC())
	}

	return cluster
}

func isWorkflowExpired(workflow v1alpha1.Workflow) bool {
	lifespan, _ := ptypes.Duration(GetLifespan(&workflow))

	workflowExpiryTime := workflow.Status.StartedAt.Time.Add(lifespan)
	return time.Now().After(workflowExpiryTime)
}

func isNearingExpiry(workflow v1alpha1.Workflow) bool {
	lifespan, _ := ptypes.Duration(GetLifespan(&workflow))

	workflowExpiryTime := workflow.Status.StartedAt.Time.Add(lifespan)
	return time.Now().Add(nearExpiry).After(workflowExpiryTime)
}

type metaCluster struct {
	v1.Cluster

	EventID       string
	Expired       bool
	NearingExpiry bool
	Slack         slack.Status
	SlackDM       bool
}

type artifactData struct {
	Name        string
	Description string
	Tags        map[string]struct{}
	Data        []byte
}

// clusterFromWorkflow converts an Argo workflow into a cluster with
// additional, non-cluster, metadata.
func (s *clusterImpl) metaClusterFromWorkflow(workflow v1alpha1.Workflow) (*metaCluster, error) {
	cluster := clusterFromWorkflow(workflow)
	expired := isWorkflowExpired(workflow)
	nearingExpiry := isNearingExpiry(workflow)

	cluster, err := s.getClusterDetailsFromArtifacts(cluster, workflow)
	if err != nil {
		return nil, err
	}

	return &metaCluster{
		Cluster:       *cluster,
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
				if _, foundURL := meta.Tags[artifactTagURL]; !foundURL {
					if _, foundConnect := meta.Tags[artifactTagConnect]; !foundConnect {
						continue
					}
				}

				bucket, key := handleArtifactMigration(workflow, artifact)
				if bucket == "" || key == "" {
					continue
				}

				contents, err := s.signer.Contents(bucket, key)
				if err != nil {
					return nil, err
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

	return cluster, nil
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
		log.Printf("[WARN] Cannot figure out bucket for artifact, possibly an upgrade issue, not fatal (workflow: %q)", workflow.Name)
		log.Printf("[INFO] Artifact: %v\n", artifact)
		log.Printf("[INFO] ArtifactRepository: %v\n", workflow.Status.ArtifactRepositoryRef)
		return "", ""
	}

	return bucket, key
}

func workflowStatus(workflowStatus v1alpha1.WorkflowStatus) v1.Status {
	// https://godoc.org/github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1#WorkflowStatus
	switch workflowStatus.Phase {
	case v1alpha1.WorkflowFailed, v1alpha1.WorkflowError:
		return v1.Status_FAILED

	case v1alpha1.WorkflowSucceeded:
		return v1.Status_FINISHED

	case v1alpha1.WorkflowPending:
		return v1.Status_CREATING

	case v1alpha1.WorkflowRunning:
		// https://godoc.org/github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1#Nodes
		for _, node := range workflowStatus.Nodes {
			// https://godoc.org/github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1#NodeType
			if node.Type == v1alpha1.NodeTypePod {
				if strings.Contains(node.Message, "ImagePullBackOff") {
					return v1.Status_FAILED
				}
				if strings.Contains(node.Message, "ErrImagePull") {
					return v1.Status_FAILED
				}
				if strings.Contains(node.Message, "Pod was active on the node longer than the specified deadline") {
					return v1.Status_FAILED
				}
			} else if node.Type == v1alpha1.NodeTypeSuspend {
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

func prettyPrint(x interface{}) {
	pretty, err := json.MarshalIndent(x, "", "  ")
	if err != nil {
		log.Fatalf(err.Error())
	}
	log.Printf("[INFO] %s\n", pretty)
}
