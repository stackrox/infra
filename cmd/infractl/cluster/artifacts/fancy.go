package artifacts

import (
	"fmt"

	v1 "github.com/stackrox/infra/generated/api/v1"
)

type prettyClusterArtifacts v1.ClusterArtifacts

func (r prettyClusterArtifacts) PrettyPrint() {
	for _, artifact := range r.Artifacts {
		fmt.Printf("%s\n", artifact.Name)
		if artifact.Description != "" {
			fmt.Printf("  Description: %s\n", artifact.Description)
		}
		fmt.Printf("  URL: %s\n", artifact.URL)
	}
}
