package create

import (
	"os/exec"
	"strings"
)

const stackroxRepo = "stackrox/stackrox"

var (
	// GetRootDir runs a command to determine the top level git repository and
	// is exported so that its behavior can be overriden by a test.
	GetRootDir = func() string {
		topLevel := exec.Command("git", "rev-parse", "--show-toplevel")
		out, _ := topLevel.Output()
		return strings.TrimSpace(string(out))
	}

	// GetMakeTag runs a command to determine the make tag and
	// is exported so that its behavior can be overriden by a test.
	GetMakeTag = func(rootDir string) string {
		makeTag := exec.Command("make", "--quiet", "tag")
		makeTag.Dir = rootDir
		out, _ := makeTag.Output()
		return strings.TrimSpace(string(out))
	}
)

type currentWorkingEnvironment struct {
	gitTopLevel string
	tag         string
}

func newCurrentWorkingEnvironment() *currentWorkingEnvironment {
	rootDir := GetRootDir()
	return &currentWorkingEnvironment{
		gitTopLevel: rootDir,
		tag:         GetMakeTag(rootDir),
	}
}

func (cwe *currentWorkingEnvironment) isInStackroxRepo() bool {
	return strings.Contains(cwe.gitTopLevel, stackroxRepo)
}

func (cwe *currentWorkingEnvironment) isTagged() bool {
	return cwe.tag != ""
}

func getHyphened(tag string) string {
	return strings.ReplaceAll(tag, ".", "-")
}

func getCleaned(tag string) string {
	return strings.TrimSuffix(tag, "-dirty")
}
