package list

import (
	"encoding/json"
	"time"

	"github.com/spf13/cobra"

	"github.com/stackrox/infra/cmd/infractl/common"
	v1 "github.com/stackrox/infra/generated/api/v1"
)

type prettyClusterListResponse struct {
	*v1.ClusterListResponse
	QuietMode bool
}

func (p prettyClusterListResponse) PrettyPrint(cmd *cobra.Command) {
	for _, cluster := range p.Clusters {
		var (
			createdOn   = cluster.GetCreatedOn().AsTime()
			lifespan    = cluster.GetLifespan().AsDuration()
			destroyedOn = cluster.GetDestroyedOn().AsTime()
			remaining   = time.Until(createdOn.Add(lifespan))
		)

		cmd.Printf("%s \n", cluster.GetID())
		if !p.QuietMode {
			cmd.Printf("  Flavor:      %s\n", cluster.GetFlavor())
			cmd.Printf("  Owner:       %s\n", cluster.GetOwner())
			cmd.Printf("  Description: %s\n", cluster.GetDescription())
			cmd.Printf("  Status:      %s\n", cluster.GetStatus())
			cmd.Printf("  Created:     %v\n", common.FormatTime(createdOn))
			if destroyedOn.Unix() != 0 {
				cmd.Printf("  Destroyed:   %v\n", common.FormatTime(destroyedOn))
			}
			cmd.Printf("  Lifespan:    %s\n", common.FormatExpiration(remaining))
		}
	}
}

func (p prettyClusterListResponse) PrettyJSONPrint(cmd *cobra.Command) error {
	data, err := json.MarshalIndent(p.ClusterListResponse, "", "  ")
	if err != nil {
		return err
	}

	cmd.Printf("%s\n", string(data))
	return nil
}
