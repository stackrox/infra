package platform

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

// IsValid returns "true" if the given OS and architecture combination is valid.
func IsValid(os, arch string) bool {
	_, valid := validPlatforms[platform{
		os:   os,
		arch: arch,
	}]
	return valid
}
