package lifespan

import (
	"fmt"

	"github.com/golang/protobuf/ptypes"
	durpb "github.com/golang/protobuf/ptypes/duration"
)

type prettyDuration durpb.Duration

func (p prettyDuration) PrettyPrint() {
	delta := durpb.Duration(p)
	lifespan, _ := ptypes.Duration(&delta)

	fmt.Println(lifespan)
}
