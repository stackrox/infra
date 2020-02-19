package artifacts

import (
	"fmt"

	v1 "github.com/stackrox/infra/generated/api/v1"
)

type clusterArtifacts v1.ClusterArtifacts

func (r clusterArtifacts) PrettyPrint() {
	for _, artifact := range r.Artifacts {
		fmt.Printf("%s\n", artifact.Name)
		fmt.Printf("  URL: %s\n", artifact.URL)
	}
}
