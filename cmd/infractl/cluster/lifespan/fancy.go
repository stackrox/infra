package lifespan

import (
	"fmt"

	"github.com/golang/protobuf/ptypes"
	durpb "github.com/golang/protobuf/ptypes/duration"
)

type dur durpb.Duration

func (r dur) PrettyPrint() {
	delta := durpb.Duration(r)
	lifespan, _ := ptypes.Duration(&delta)

	fmt.Println(lifespan)
}
