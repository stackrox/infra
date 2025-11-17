// Package metrics exposes custom metrics for the infra server.
package metrics

import "github.com/prometheus/client_golang/prometheus"

var (
	// FlavorsUsedCounter is a Prometheus metric, counting the number of clusters created per flavor
	FlavorsUsedCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "infra",
			Name:      "flavors_used",
			Help:      "Number of clusters created by flavor",
		},
		[]string{"flavor"},
	)
)

func init() {
	// Metrics have to be registered to be exposed:
	prometheus.MustRegister(FlavorsUsedCounter)
}
