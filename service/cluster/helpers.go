package cluster

import (
	"encoding/json"
	"log"
	"strings"
	"time"

	"github.com/argoproj/argo/pkg/apis/workflow/v1alpha1"
	"github.com/golang/protobuf/ptypes"
	v1 "github.com/stackrox/infra/generated/api/v1"
	"github.com/stackrox/infra/slack"
)

// clusterFromWorkflow converts an Argo workflow into a cluster.
func clusterFromWorkflow(workflow v1alpha1.Workflow) *v1.Cluster {
	cluster := &v1.Cluster{
		ID:          workflow.GetName(),
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

				contents, err := s.signer.Contents(artifact.GCS.Bucket, artifact.GCS.Key)
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

func prettyPrint(x interface{}) {
	pretty, err := json.MarshalIndent(x, "", "  ")
	if err != nil {
		log.Fatalf(err.Error())
	}
	log.Printf("[INFO] %s\n", pretty)
}
