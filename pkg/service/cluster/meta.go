package cluster

import (
	"strings"

	"github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	v1 "github.com/stackrox/infra/generated/api/v1"
	"github.com/stackrox/infra/pkg/service/cluster/helpers"
	"github.com/stackrox/infra/pkg/service/cluster/metadata"
	"github.com/stackrox/infra/pkg/slack"
)

type metaCluster struct {
	v1.Cluster

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

// metaClusterFromWorkflow converts an Argo workflow into an infra cluster
// with additional, non-cluster, metadata.
func (s *clusterImpl) metaClusterFromWorkflow(workflow v1alpha1.Workflow) (*metaCluster, error) {
	cluster := helpers.ClusterFromWorkflow(workflow)
	expired := isClusterExpired(workflow)
	nearingExpiry := isClusterNearingExpiry(workflow)

	cluster, err := s.getClusterDetailsFromArtifacts(cluster, workflow)
	if err != nil {
		return nil, err
	}

	return &metaCluster{
		Cluster:       *cluster,
		Slack:         slack.Status(metadata.GetSlack(&workflow)),
		SlackDM:       metadata.GetSlackDM(&workflow),
		Expired:       expired,
		NearingExpiry: nearingExpiry,
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

				bucket, key := helpers.HandleArtifactMigration(workflow, artifact)
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

	cluster.Parameters = helpers.ClusterParametersFromWorkflow(workflow)

	return cluster, nil
}
