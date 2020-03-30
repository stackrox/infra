package cluster

import (
	"regexp"
	"strings"
)

func simpleName(input string) string {
	reg := regexp.MustCompile("[^a-zA-Z0-9]+")
	input = reg.ReplaceAllString(input, "-")
	input = strings.ToLower(input)
	return strings.Trim(input, "-")
}
