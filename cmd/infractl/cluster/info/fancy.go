package info

import (
	"fmt"
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/stackrox/infra/cmd/infractl/common"
	v1 "github.com/stackrox/infra/generated/api/v1"
)

type cluster v1.Cluster

func (r cluster) PrettyPrint() {
	var (
		createdOn, _   = ptypes.Timestamp(r.CreatedOn)
		lifespan, _    = ptypes.Duration(r.Lifespan)
		destroyedOn, _ = ptypes.Timestamp(r.DestroyedOn)
		remaining      = time.Until(createdOn.Add(lifespan))
	)

	fmt.Printf("ID:        %s\n", r.ID)
	fmt.Printf("Flavor:    %s\n", r.Flavor)
	fmt.Printf("Owner:     %s\n", r.Owner)
	fmt.Printf("Status:    %s\n", r.Status.String())
	fmt.Printf("Created:   %v\n", common.FormatTime((createdOn)))
	if destroyedOn.Unix() != 0 {
		fmt.Printf("Destroyed: %v\n", common.FormatTime(destroyedOn))
	}
	fmt.Printf("Lifespan:  %s\n", common.FormatExpiration(remaining))
}
