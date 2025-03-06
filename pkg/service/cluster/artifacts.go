package cluster

import (
	"context"

	v1 "github.com/stackrox/infra/generated/api/v1"
	"github.com/stackrox/infra/pkg/service/cluster/helpers"
	"github.com/stackrox/infra/pkg/service/cluster/metadata"
)

// Artifacts implements ClusterService.Artifacts.
func (s *clusterImpl) Artifacts(_ context.Context, clusterID *v1.ResourceByID) (*v1.ClusterArtifacts, error) {
	workflow, err := s.argoClient.GetMostRecentArgoWorkflowFromClusterID(clusterID.GetId())
	if err != nil {
		return nil, err
	}

	flavorMetadata := make(map[string]*v1.FlavorArtifact)
	flavorName := metadata.GetFlavor(workflow)
	flavor, _, found := s.registry.Get(flavorName)
	if found && flavor.Artifacts != nil {
		flavorMetadata = flavor.Artifacts
	}

	resp := v1.ClusterArtifacts{}

	for _, nodeStatus := range workflow.Status.Nodes {
		if nodeStatus.Outputs != nil {
			for _, artifact := range nodeStatus.Outputs.Artifacts {
				if artifact.GCS == nil {
					continue
				}

				var description string

				meta, found := flavorMetadata[artifact.Name]
				if found {
					if _, isInternal := meta.Tags[artifactTagInternal]; isInternal {
						continue
					}

					description = meta.Description
				}

				bucket, key := helpers.HandleArtifactMigration(*workflow, artifact)
				if bucket == "" || key == "" {
					continue
				}

				url, err := s.signer.Generate(bucket, key)
				if err != nil {
					return nil, err
				}

				var mode int32 = artifactDefaultMode
				if artifact.Mode != nil {
					mode = *artifact.Mode
				}

				resp.Artifacts = append(resp.Artifacts, &v1.Artifact{
					Name:        artifact.Name,
					Description: description,
					URL:         url,
					Mode:        mode,
				})
			}
		}
	}

	return &resp, nil
}
