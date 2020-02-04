package cluster

import (
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/duration"
)

const (
	// AnnotationFlavor is the k8s annotation that contains the flavor ID.
	AnnotationFlavor = "infra.stackrox.com/flavor"

	// AnnotationOwner is the k8s annotation that contains the owner email
	// address.
	AnnotationOwner = "infra.stackrox.com/owner"

	// AnnotationLifespan is the k8s annotation that contains the lifespan
	// duration.
	AnnotationLifespan = "infra.stackrox.com/lifespan"
)

// Annotated represents a type that has annotations.
type Annotated interface {
	GetAnnotations() map[string]string
}

// GetFlavor returns the flavor ID if it exists.
func GetFlavor(a Annotated) string {
	return a.GetAnnotations()[AnnotationFlavor]
}

// GetOwner returns the owner email address if it exists.
func GetOwner(a Annotated) string {
	return a.GetAnnotations()[AnnotationOwner]
}

// GetLifespan returns the lifespan duration if it exists. If it does not
// exist, or is in an invalid format, a default 3 hours is returned.
func GetLifespan(a Annotated) *duration.Duration {
	lifespan, err := time.ParseDuration(a.GetAnnotations()[AnnotationLifespan])
	if err != nil {
		// Fallback to a default 3 hours.
		lifespan = 3 * time.Hour
	}

	return ptypes.DurationProto(lifespan)
}
