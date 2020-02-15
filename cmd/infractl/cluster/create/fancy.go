package create

import (
	"fmt"

	v1 "github.com/stackrox/infra/generated/api/v1"
)

type id v1.ResourceByID

func (r id) PrettyPrint() {
	fmt.Printf("ID: %s\n", r.Id)
}
