package cluster

import (
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/duration"
)

const (
	// annotationFlavorKey is the k8s annotation that contains the flavor ID.
	annotationFlavorKey = "infra.stackrox.com/flavor"

	// annotationOwnerKey is the k8s annotation that contains the owner email
	// address.
	annotationOwnerKey = "infra.stackrox.com/owner"

	// AnnotationLifespanKey is the k8s annotation that contains the lifespan
	// duration.
	annotationLifespanKey = "infra.stackrox.com/lifespan"
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
	return ptypes.DurationProto(lifespan)
}
