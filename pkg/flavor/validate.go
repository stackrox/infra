package flavor

import (
	"github.com/pkg/errors"
	"github.com/stackrox/infra/pkg/config"
)

func validateFlavor(flavorCfg config.FlavorConfig) error {
	if flavorCfg.ID == "" || flavorCfg.Name == "" || flavorCfg.Description == "" || flavorCfg.WorkflowFile == "" {
		return errors.New("flavor ID, name or description is missing")
	}
	return nil
}

func validateParameter(parameter config.Parameter) error {
	if parameter.Name == "" {
		return errors.New("parameter name is missing")
	}
	return nil
}

func validateArtifact(artifact config.Artifact) error {
	if artifact.Name == "" {
		return errors.New("artifact name is missing")
	}
	return nil
}
