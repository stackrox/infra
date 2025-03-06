package metadata

import (
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/duration"
)

const (
	// AnnotationFlavorKey is the k8s annotation that contains the flavor ID.
	AnnotationFlavorKey = "infra.stackrox.com/flavor"

	// AnnotationOwnerKey is the k8s annotation that contains the owner email
	// address.
	AnnotationOwnerKey = "infra.stackrox.com/owner"

	// AnnotationLifespanKey is the k8s annotation that contains the lifespan
	// duration.
	AnnotationLifespanKey = "infra.stackrox.com/lifespan"

	// AnnotationDescriptionKey is the k8s annotation that contains the description.
	AnnotationDescriptionKey = "infra.stackrox.com/description"

	// AnnotationSlackKey is the k8s annotation that contains the Slack
	// notification phase.
	AnnotationSlackKey = "infra.stackrox.com/slack"

	// AnnotationSlackDMKey uses slack direct messages instead of a channel
	AnnotationSlackDMKey = "infra.stackrox.com/slackdm"
)

// Annotated represents a type that has annotations.
type Annotated interface {
	GetAnnotations() map[string]string
}

// GetFlavor returns the flavor ID if it exists.
func GetFlavor(a Annotated) string {
	return a.GetAnnotations()[AnnotationFlavorKey]
}

// GetOwner returns the owner email address if it exists.
func GetOwner(a Annotated) string {
	return a.GetAnnotations()[AnnotationOwnerKey]
}

// GetLifespan returns the lifespan duration if it exists. If it does not
// exist, or is in an invalid format, a default 3 hours is returned.
func GetLifespan(a Annotated) *duration.Duration {
	lifespan, err := time.ParseDuration(a.GetAnnotations()[AnnotationLifespanKey])
	if err != nil {
		// Fallback to a default 3 hours.
		lifespan = 3 * time.Hour
	}

	if lifespan <= 0 {
		// Ensure that the lifespan isn't negative.
		lifespan = 0
	}
	return ptypes.DurationProto(lifespan)
}

// GetDescription returns the description if it exists.
func GetDescription(a Annotated) string {
	return a.GetAnnotations()[AnnotationDescriptionKey]
}

// GetSlack returns the Slack notification phase if it exists.
func GetSlack(a Annotated) string {
	return a.GetAnnotations()[AnnotationSlackKey]
}

// GetSlackDM returns the Slack DM setting for the cluster.
func GetSlackDM(a Annotated) bool {
	return a.GetAnnotations()[AnnotationSlackDMKey] == "yes"
}
