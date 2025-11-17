// Package buildinfo provides information about the built binary and the
// environment in which it was built.
package buildinfo

import (
	"runtime"

	v1 "github.com/stackrox/infra/generated/api/v1"
	"github.com/stackrox/infra/pkg/buildinfo/internal"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// All returns all of the various pieces of version information.
func All() *v1.Version {
	ts := timestamppb.New(internal.BuildTimestamp)
	return &v1.Version{
		BuildDate: ts,
		GitCommit: internal.GitShortSha,
		GoVersion: runtime.Version(),
		Platform:  runtime.GOOS + "/" + runtime.GOARCH,
		Version:   internal.GitVersion,
		Workflow:  internal.CircleciWorkflowURL,
	}
}

// Version returns only the Git version.
func Version() string {
	return internal.GitVersion
}
