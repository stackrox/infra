package upgrade

import (
	"fmt"
)

type prettyCliUpgrade struct {
	bytes []byte
}

func (r prettyCliUpgrade) PrettyPrint() {
	fmt.Printf("woot! %v\n", len(r.bytes))
}

