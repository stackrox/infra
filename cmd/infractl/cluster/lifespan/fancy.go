package lifespan

import (
	"fmt"

	"github.com/golang/protobuf/ptypes"
	durpb "github.com/golang/protobuf/ptypes/duration"
	"github.com/stackrox/infra/cmd/infractl/common"
)

type prettyDuration struct {
	*durpb.Duration
}

func (p prettyDuration) PrettyPrint() {
	remaining, _ := ptypes.Duration(p.Duration)

	fmt.Println(common.FormatExpiration(remaining))
}
