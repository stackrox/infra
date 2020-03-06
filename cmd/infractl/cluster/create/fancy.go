package create

import (
	"fmt"

	v1 "github.com/stackrox/infra/generated/api/v1"
)

type prettyResourceByID v1.ResourceByID

func (r prettyResourceByID) PrettyPrint() {
	fmt.Printf("ID: %s\n", r.Id)
}
