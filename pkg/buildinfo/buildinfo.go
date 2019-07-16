package buildinfo

import (
	"runtime"
	"time"

	"github.com/stackrox/infra/generated/api/v1"
	"github.com/stackrox/infra/pkg/buildinfo/internal"
	"gopkg.in/golang/protobuf.v1/ptypes"
)

// Versions represents a collection of various pieces of version information.
type Versions struct {
	BuildDate time.Time `json:"BuildDate"`
	GitCommit string    `json:"GitCommit"`
	GoVersion string    `json:"GoVersion"`
	Platform  string    `json:"Platform"`
	Version   string    `json:"Version"`
	Workflow  string    `json:"Workflow"`
}

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
