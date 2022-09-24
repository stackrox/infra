package get

import (
	"fmt"
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/stackrox/infra/cmd/infractl/common"
	v1 "github.com/stackrox/infra/generated/api/v1"
)

type prettyCluster struct {
	*v1.Cluster
}

func (p *prettyCluster) PrettyPrint() {
	//p := *pc
	var (
		createdOn, _   = ptypes.Timestamp(p.CreatedOn)
		lifespan, _    = ptypes.Duration(p.Lifespan)
		destroyedOn, _ = ptypes.Timestamp(p.DestroyedOn)
		remaining      = time.Until(createdOn.Add(lifespan))
	)

	fmt.Printf("ID:          %s\n", p.ID)
	fmt.Printf("Flavor:      %s\n", p.Flavor)
	fmt.Printf("Owner:       %s\n", p.Owner)
	if p.Description != "" {
		fmt.Printf("Description: %s\n", p.Description)
	}
	fmt.Printf("Status:      %s\n", p.Status.String())
	fmt.Printf("Created:     %v\n", common.FormatTime(createdOn))
	if p.URL != "" {
		fmt.Printf("URL:         %s\n", p.URL)
	}
	if destroyedOn.Unix() != 0 {
		fmt.Printf("Destroyed:   %v\n", common.FormatTime(destroyedOn))
	}
	fmt.Printf("Lifespan:    %s\n", common.FormatExpiration(remaining))
}
