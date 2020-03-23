package list

import (
	"fmt"
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/stackrox/infra/cmd/infractl/common"
	v1 "github.com/stackrox/infra/generated/api/v1"
)

type prettyClusterListResponse v1.ClusterListResponse

func (p prettyClusterListResponse) PrettyPrint() {
	for _, cluster := range p.Clusters {
		var (
			createdOn, _   = ptypes.Timestamp(cluster.GetCreatedOn())
			lifespan, _    = ptypes.Duration(cluster.GetLifespan())
			destroyedOn, _ = ptypes.Timestamp(cluster.GetDestroyedOn())
			remaining      = time.Until(createdOn.Add(lifespan))
		)

		fmt.Printf("%s \n", cluster.GetID())
		fmt.Printf("  Flavor:      %s\n", cluster.GetFlavor())
		fmt.Printf("  Owner:       %s\n", cluster.GetOwner())
		fmt.Printf("  Description: %s\n", cluster.GetDescription())
		fmt.Printf("  Status:      %s\n", cluster.GetStatus())
		fmt.Printf("  Created:     %v\n", common.FormatTime(createdOn))
		if destroyedOn.Unix() != 0 {
			fmt.Printf("  Destroyed:   %v\n", common.FormatTime(destroyedOn))
		}
		fmt.Printf("  Lifespan:    %s\n", common.FormatExpiration(remaining))
	}
}
