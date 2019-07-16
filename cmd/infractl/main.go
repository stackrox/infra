package main

import (
	"fmt"

	"github.com/stackrox/infra/pkg/buildinfo"
)

func main() {
	fmt.Printf("%+v\n", buildinfo.All())
}
