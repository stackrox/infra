package cluster

import (
	"time"

	"github.com/golang/protobuf/ptypes/duration"
	"google.golang.org/protobuf/types/known/durationpb"
)

const (
	// annotationFlavorKey is the k8s annotation that contains the flavor ID.
	annotationFlavorKey = "infra.stackrox.com/flavor"

	// annotationOwnerKey is the k8s annotation that contains the owner email
	// address.
	annotationOwnerKey = "infra.stackrox.com/owner"

	// annotationLifespanKey is the k8s annotation that contains the lifespan
	// duration.
	annotationLifespanKey = "infra.stackrox.com/lifespan"

	// annotationEventKey is the k8s annotation that contains the event ID.
	annotationEventKey = "infra.stackrox.com/event"

	// annotationEventKey is the k8s annotation that contains the description.
	annotationDescriptionKey = "infra.stackrox.com/description"

	// annotationSlackKey is the k8s annotation that contains the Slack
	// notification phase.
	annotationSlackKey = "infra.stackrox.com/slack"

	// use slack direct messages instead of a channel
	annotationSlackDMKey = "infra.stackrox.com/slackdm"
)

// Annotated represents a type that has annotations.
type Annotated interface {
	GetAnnotations() map[string]string
}

// GetFlavor returns the flavor ID if it exists.
func GetFlavor(a Annotated) string {
	return a.GetAnnotations()[annotationFlavorKey]
}

// GetOwner returns the owner email address if it exists.
func GetOwner(a Annotated) string {
	return a.GetAnnotations()[annotationOwnerKey]
}

// GetLifespan returns the lifespan duration if it exists. If it does not
// exist, or is in an invalid format, a default 3 hours is returned.
func GetLifespan(a Annotated) *duration.Duration {
	lifespan, err := time.ParseDuration(a.GetAnnotations()[annotationLifespanKey])
	if err != nil {
		// Fallback to a default 3 hours.
		lifespan = 3 * time.Hour
	}

	if lifespan <= 0 {
		// Ensure that the lifespan isn't negative.
		lifespan = 0
	}
	return durationpb.New(lifespan)
}

// GetEventID returns the event ID if it exists.
func GetEventID(a Annotated) string {
	return a.GetAnnotations()[annotationEventKey]
}

// GetDescription returns the description if it exists.
func GetDescription(a Annotated) string {
	return a.GetAnnotations()[annotationDescriptionKey]
}

// GetSlack returns the Slack notification phase if it exists.
func GetSlack(a Annotated) string {
	return a.GetAnnotations()[annotationSlackKey]
}

// GetSlackDM returns the Slack DM setting for the cluster.
func GetSlackDM(a Annotated) bool {
	return a.GetAnnotations()[annotationSlackDMKey] == "yes"
}
