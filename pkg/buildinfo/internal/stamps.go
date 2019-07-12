package internal

import (
	"strconv"
	"time"
)

var (
	// buildTimestampUnixSeconds is the injected built timestamp in Unix seconds.
	buildTimestampUnixSeconds string
	// BuildTimestamp is the time when this binary was built, or time.Now() if built locally.
	BuildTimestamp time.Time

	// circleciWorkflowURL is the injected CircleCI build workflow URL.
	circleciWorkflowURL string
	// CircleciWorkflowURL is the CircleCI build workflow URL, or "local" if built locally.
	CircleciWorkflowURL string

	// gitShortSha is the injected Git short SHA.
	gitShortSha         string
	// GitShortSha is the Git short SHA for the current commit, or "local" if built locally.
	GitShortSha         string

	// gitVersion is the injected Git describe version.
	gitVersion          string
	// GitVersion is the Git describe version for the current commit, or "local" if built locally.
	GitVersion          string
)

func init() {
	// Parse the given built timestamp. Fallback to now if missing. Panic if malformed.
	BuildTimestamp = mustParseUnixSecondsString(buildTimestampUnixSeconds)

	// Use given CircleCI build workflow URL. Fallback to "local" if missing.
	CircleciWorkflowURL = fallback(circleciWorkflowURL, "local")

	// Use given Git short SHA. Fallback to "local" if missing.
	GitShortSha = fallback(gitShortSha, "local")

	// Use given Git describe version. Fallback to "local" if missing.
	GitVersion = fallback(gitVersion, "local")
}

// fallback returns the given original value if it is non-empty. The given fallback is returned otherwise.
func fallback(original, fallback string) string {
	if original != "" {
		return original
	}
	return fallback
}

// mustParseUnixSecondsString parses the given Unix timestamp, and panics if it is malformed. If the timestamp is empty,
// time.Now() is returned.
func mustParseUnixSecondsString(timestamp string) time.Time {
	if timestamp == "" {
		return time.Now()
	}
	unixSecs, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		panic(err)
	}
	return time.Unix(unixSecs, 0)
}
