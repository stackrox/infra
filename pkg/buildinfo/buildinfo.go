package buildinfo

import (
	"runtime"

	"github.com/golang/protobuf/ptypes"
	"github.com/stackrox/infra/generated/api/v1"
	"github.com/stackrox/infra/pkg/buildinfo/internal"
)

// All returns all of the various pieces of version information.
func All() v1.Version {
	ts, _ := ptypes.TimestampProto(internal.BuildTimestamp)
	return v1.Version{
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
