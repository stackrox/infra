package upgrade

import (
	"fmt"
)

type prettyCliUpgrade struct {
	updatedFilename string
}

func (r prettyCliUpgrade) PrettyPrint() {
	fmt.Printf("Updated %s to match the infra server version\n", r.updatedFilename)
}
