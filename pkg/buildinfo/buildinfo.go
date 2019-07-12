package buildinfo

import (
	"runtime"
	"time"

	"github.com/stackrox/infra/pkg/buildinfo/internal"
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
func All() Versions {
	return Versions{
		BuildDate: internal.BuildTimestamp,
		GitCommit: internal.GitShortSha,
		GoVersion: runtime.Version(),
		Platform:  runtime.GOOS + "/" + runtime.GOARCH,
		Version:   internal.GitVersion,
		Workflow:  internal.CircleciWorkflowURL,
	}
}
