package platform

import "github.com/pkg/errors"

type platform struct {
	os   string
	arch string
}

var validPlatforms = map[platform]struct{}{
	{
		os:   "darwin",
		arch: "amd64",
	}: {},
	{
		os:   "darwin",
		arch: "arm64",
	}: {},
	{
		os:   "linux",
		arch: "amd64",
	}: {},
}

// Validate ensures the given OS and architecture combination is valid.
func Validate(os, arch string) error {
	if _, valid := validPlatforms[platform{
		os:   os,
		arch: arch,
	}]; valid {
		return nil
	}

	return errors.Errorf("Invalid platform: %s/%s", os, arch)
}
