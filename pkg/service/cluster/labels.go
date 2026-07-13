package cluster

const (
	// labelClusterId is the label key used to map an infra cluster to
	// an argo workflow.
	labelClusterID = "infra.stackrox.com/cluster-id"

	// labelOwner is the label key for the cluster owner email.
	labelOwner = "infra.stackrox.com/owner"

	// labelFlavor is the label key for the cluster flavor ID.
	labelFlavor = "infra.stackrox.com/flavor"
)

// Labeled represents a type that has labels.
type Labeled interface {
	GetLabels() map[string]string
}

// GetClusterID returns the Cluster ID if it exists.
func GetClusterID(a Labeled) string {
	return a.GetLabels()[labelClusterID]
}
