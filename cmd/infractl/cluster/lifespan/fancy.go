package lifespan

import (
	"fmt"

	"github.com/golang/protobuf/ptypes"
	durpb "github.com/golang/protobuf/ptypes/duration"
	"github.com/stackrox/infra/cmd/infractl/common"
)

type prettyDuration durpb.Duration

func (p prettyDuration) PrettyPrint() {
	delta := durpb.Duration(p)
	remaining, _ := ptypes.Duration(&delta)

	fmt.Println(common.FormatExpiration(remaining))
}
